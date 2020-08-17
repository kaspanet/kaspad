package blockdag

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"io"

	"github.com/kaspanet/kaspad/infrastructure/db/dbaccess"
	"github.com/kaspanet/kaspad/network/domainmessage"
	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/util/coinbasepayload"
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/kaspanet/kaspad/util/subnetworkid"
	"github.com/kaspanet/kaspad/util/txsort"
	"github.com/pkg/errors"
)

// compactFeeData is a specialized data type to store a compact list of fees
// inside a block.
// Every transaction gets a single uint64 value, stored as a plain binary list.
// The transactions are ordered the same way they are ordered inside the block, making it easy
// to traverse every transaction in a block and extract its fee.
//
// compactFeeFactory is used to create such a list.
// compactFeeIterator is used to iterate over such a list.

type compactFeeData []byte

func (cfd compactFeeData) Len() int {
	return len(cfd) / 8
}

type compactFeeFactory struct {
	buffer *bytes.Buffer
	writer *bufio.Writer
}

func newCompactFeeFactory() *compactFeeFactory {
	buffer := bytes.NewBuffer([]byte{})
	return &compactFeeFactory{
		buffer: buffer,
		writer: bufio.NewWriter(buffer),
	}
}

func (cfw *compactFeeFactory) add(txFee uint64) error {
	return binary.Write(cfw.writer, binary.LittleEndian, txFee)
}

func (cfw *compactFeeFactory) data() (compactFeeData, error) {
	err := cfw.writer.Flush()

	return compactFeeData(cfw.buffer.Bytes()), err
}

type compactFeeIterator struct {
	reader io.Reader
}

func (cfd compactFeeData) iterator() *compactFeeIterator {
	return &compactFeeIterator{
		reader: bufio.NewReader(bytes.NewBuffer(cfd)),
	}
}

func (cfr *compactFeeIterator) next() (uint64, error) {
	var txFee uint64

	err := binary.Read(cfr.reader, binary.LittleEndian, &txFee)

	return txFee, err
}

// The following functions relate to storing and retrieving fee data from the database

// getBluesFeeData returns the compactFeeData for all nodes's blues,
// used to calculate the fees this blockNode needs to pay
func (dag *BlockDAG) getBluesFeeData(node *blockNode) (map[daghash.Hash]compactFeeData, error) {
	bluesFeeData := make(map[daghash.Hash]compactFeeData)

	for _, blueBlock := range node.blues {
		feeData, err := dbaccess.FetchFeeData(dag.databaseContext, blueBlock.hash)
		if err != nil {
			return nil, err
		}

		bluesFeeData[*blueBlock.hash] = feeData
	}

	return bluesFeeData, nil
}

// The following functions deal with building and validating the coinbase transaction

func (node *blockNode) validateCoinbaseTransaction(dag *BlockDAG, block *util.Block, txsAcceptanceData MultiBlockTxsAcceptanceData) error {
	if node.isGenesis() {
		return nil
	}
	blockCoinbaseTx := block.CoinbaseTransaction().MsgTx()
	_, scriptPubKey, extraData, err := coinbasepayload.DeserializeCoinbasePayload(blockCoinbaseTx)
	if errors.Is(err, coinbasepayload.ErrIncorrectScriptPubKeyLen) {
		return ruleError(ErrBadCoinbaseTransaction, err.Error())
	}
	if err != nil {
		return err
	}
	expectedCoinbaseTransaction, err := node.expectedCoinbaseTransaction(dag, txsAcceptanceData, scriptPubKey, extraData)
	if err != nil {
		return err
	}

	if !expectedCoinbaseTransaction.Hash().IsEqual(block.CoinbaseTransaction().Hash()) {
		return ruleError(ErrBadCoinbaseTransaction, "Coinbase transaction is not built as expected")
	}

	return nil
}

// expectedCoinbaseTransaction returns the coinbase transaction for the current block
func (node *blockNode) expectedCoinbaseTransaction(dag *BlockDAG, txsAcceptanceData MultiBlockTxsAcceptanceData, scriptPubKey []byte, extraData []byte) (*util.Tx, error) {
	bluesFeeData, err := dag.getBluesFeeData(node)
	if err != nil {
		return nil, err
	}

	txIns := []*domainmessage.TxIn{}
	txOuts := []*domainmessage.TxOut{}

	for _, blue := range node.blues {
		txOut, err := coinbaseOutputForBlueBlock(dag, blue, txsAcceptanceData, bluesFeeData)
		if err != nil {
			return nil, err
		}
		if txOut != nil {
			txOuts = append(txOuts, txOut)
		}
	}
	payload, err := coinbasepayload.SerializeCoinbasePayload(node.blueScore, scriptPubKey, extraData)
	if err != nil {
		return nil, err
	}
	coinbaseTx := domainmessage.NewSubnetworkMsgTx(domainmessage.TxVersion, txIns, txOuts, subnetworkid.SubnetworkIDCoinbase, 0, payload)
	sortedCoinbaseTx := txsort.Sort(coinbaseTx)
	return util.NewTx(sortedCoinbaseTx), nil
}

// coinbaseOutputForBlueBlock calculates the output that should go into the coinbase transaction of blueBlock
// If blueBlock gets no fee - returns nil for txOut
func coinbaseOutputForBlueBlock(dag *BlockDAG, blueBlock *blockNode,
	txsAcceptanceData MultiBlockTxsAcceptanceData, feeData map[daghash.Hash]compactFeeData) (*domainmessage.TxOut, error) {

	blockTxsAcceptanceData, ok := txsAcceptanceData.FindAcceptanceData(blueBlock.hash)
	if !ok {
		return nil, errors.Errorf("No txsAcceptanceData for block %s", blueBlock.hash)
	}
	blockFeeData, ok := feeData[*blueBlock.hash]
	if !ok {
		return nil, errors.Errorf("No feeData for block %s", blueBlock.hash)
	}

	if len(blockTxsAcceptanceData.TxAcceptanceData) != blockFeeData.Len() {
		return nil, errors.Errorf(
			"length of accepted transaction data(%d) and fee data(%d) is not equal for block %s",
			len(blockTxsAcceptanceData.TxAcceptanceData), blockFeeData.Len(), blueBlock.hash)
	}

	totalFees := uint64(0)
	feeIterator := blockFeeData.iterator()

	for _, txAcceptanceData := range blockTxsAcceptanceData.TxAcceptanceData {
		fee, err := feeIterator.next()
		if err != nil {
			return nil, errors.Errorf("Error retrieving fee from compactFeeData iterator: %s", err)
		}
		if txAcceptanceData.IsAccepted {
			totalFees += fee
		}
	}

	totalReward := CalcBlockSubsidy(blueBlock.blueScore, dag.Params) + totalFees

	if totalReward == 0 {
		return nil, nil
	}

	// the ScriptPubKey for the coinbase is parsed from the coinbase payload
	_, scriptPubKey, _, err := coinbasepayload.DeserializeCoinbasePayload(blockTxsAcceptanceData.TxAcceptanceData[0].Tx.MsgTx())
	if err != nil {
		return nil, err
	}

	txOut := &domainmessage.TxOut{
		Value:        totalReward,
		ScriptPubKey: scriptPubKey,
	}

	return txOut, nil
}

// NextBlockCoinbaseTransaction prepares the coinbase transaction for the next mined block
//
// This function CAN'T be called with the DAG lock held.
func (dag *BlockDAG) NextBlockCoinbaseTransaction(scriptPubKey []byte, extraData []byte) (*util.Tx, error) {
	dag.dagLock.RLock()
	defer dag.dagLock.RUnlock()

	return dag.NextBlockCoinbaseTransactionNoLock(scriptPubKey, extraData)
}

// NextBlockCoinbaseTransactionNoLock prepares the coinbase transaction for the next mined block
//
// This function MUST be called with the DAG read-lock held
func (dag *BlockDAG) NextBlockCoinbaseTransactionNoLock(scriptPubKey []byte, extraData []byte) (*util.Tx, error) {
	txsAcceptanceData, err := dag.TxsAcceptedByVirtual()
	if err != nil {
		return nil, err
	}
	return dag.virtual.blockNode.expectedCoinbaseTransaction(dag, txsAcceptanceData, scriptPubKey, extraData)
}
