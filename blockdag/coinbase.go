package blockdag

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"github.com/kaspanet/kaspad/dbaccess"
	"github.com/kaspanet/kaspad/util/subnetworkid"
	"github.com/pkg/errors"
	"io"
	"math"

	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/kaspanet/kaspad/util/txsort"
	"github.com/kaspanet/kaspad/wire"
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
func (node *blockNode) getBluesFeeData(dag *BlockDAG) (map[daghash.Hash]compactFeeData, error) {
	bluesFeeData := make(map[daghash.Hash]compactFeeData)

	for _, blueBlock := range node.blues {
		feeData, err := dbaccess.FetchFeeData(dbaccess.NoTx(), blueBlock.hash)
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
	scriptPubKey, extraData, err := DeserializeCoinbasePayload(blockCoinbaseTx)
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
	bluesFeeData, err := node.getBluesFeeData(dag)
	if err != nil {
		return nil, err
	}

	txIns := []*wire.TxIn{}
	txOuts := []*wire.TxOut{}

	for _, blue := range node.blues {
		txIn, txOut, err := coinbaseInputAndOutputForBlueBlock(dag, blue, txsAcceptanceData, bluesFeeData)
		if err != nil {
			return nil, err
		}
		txIns = append(txIns, txIn)
		if txOut != nil {
			txOuts = append(txOuts, txOut)
		}
	}
	payload, err := SerializeCoinbasePayload(scriptPubKey, extraData)
	if err != nil {
		return nil, err
	}
	coinbaseTx := wire.NewSubnetworkMsgTx(wire.TxVersion, txIns, txOuts, subnetworkid.SubnetworkIDCoinbase, 0, payload)
	sortedCoinbaseTx := txsort.Sort(coinbaseTx)
	return util.NewTx(sortedCoinbaseTx), nil
}

// SerializeCoinbasePayload builds the coinbase payload based on the provided scriptPubKey and extra data.
func SerializeCoinbasePayload(scriptPubKey []byte, extraData []byte) ([]byte, error) {
	w := &bytes.Buffer{}
	err := wire.WriteVarInt(w, uint64(len(scriptPubKey)))
	if err != nil {
		return nil, err
	}
	_, err = w.Write(scriptPubKey)
	if err != nil {
		return nil, err
	}
	_, err = w.Write(extraData)
	if err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

// DeserializeCoinbasePayload deserialize the coinbase payload to its component (scriptPubKey and extra data).
func DeserializeCoinbasePayload(tx *wire.MsgTx) (scriptPubKey []byte, extraData []byte, err error) {
	r := bytes.NewReader(tx.Payload)
	scriptPubKeyLen, err := wire.ReadVarInt(r)
	if err != nil {
		return nil, nil, err
	}
	scriptPubKey = make([]byte, scriptPubKeyLen)
	_, err = r.Read(scriptPubKey)
	if err != nil {
		return nil, nil, err
	}
	extraData = make([]byte, r.Len())
	if r.Len() != 0 {
		_, err = r.Read(extraData)
		if err != nil {
			return nil, nil, err
		}
	}
	return scriptPubKey, extraData, nil
}

// feeInputAndOutputForBlueBlock calculates the input and output that should go into the coinbase transaction of blueBlock
// If blueBlock gets no fee - returns only txIn and nil for txOut
func coinbaseInputAndOutputForBlueBlock(dag *BlockDAG, blueBlock *blockNode,
	txsAcceptanceData MultiBlockTxsAcceptanceData, feeData map[daghash.Hash]compactFeeData) (
	*wire.TxIn, *wire.TxOut, error) {

	blockTxsAcceptanceData, ok := txsAcceptanceData.FindAcceptanceData(blueBlock.hash)
	if !ok {
		return nil, nil, errors.Errorf("No txsAcceptanceData for block %s", blueBlock.hash)
	}
	blockFeeData, ok := feeData[*blueBlock.hash]
	if !ok {
		return nil, nil, errors.Errorf("No feeData for block %s", blueBlock.hash)
	}

	if len(blockTxsAcceptanceData.TxAcceptanceData) != blockFeeData.Len() {
		return nil, nil, errors.Errorf(
			"length of accepted transaction data(%d) and fee data(%d) is not equal for block %s",
			len(blockTxsAcceptanceData.TxAcceptanceData), blockFeeData.Len(), blueBlock.hash)
	}

	txIn := &wire.TxIn{
		SignatureScript: []byte{},
		PreviousOutpoint: wire.Outpoint{
			TxID:  daghash.TxID(*blueBlock.hash),
			Index: math.MaxUint32,
		},
		Sequence: wire.MaxTxInSequenceNum,
	}

	totalFees := uint64(0)
	feeIterator := blockFeeData.iterator()

	for _, txAcceptanceData := range blockTxsAcceptanceData.TxAcceptanceData {
		fee, err := feeIterator.next()
		if err != nil {
			return nil, nil, errors.Errorf("Error retrieving fee from compactFeeData iterator: %s", err)
		}
		if txAcceptanceData.IsAccepted {
			totalFees += fee
		}
	}

	totalReward := CalcBlockSubsidy(blueBlock.blueScore, dag.dagParams) + totalFees

	if totalReward == 0 {
		return txIn, nil, nil
	}

	// the ScriptPubKey for the coinbase is parsed from the coinbase payload
	scriptPubKey, _, err := DeserializeCoinbasePayload(blockTxsAcceptanceData.TxAcceptanceData[0].Tx.MsgTx())
	if err != nil {
		return nil, nil, err
	}

	txOut := &wire.TxOut{
		Value:        totalReward,
		ScriptPubKey: scriptPubKey,
	}

	return txIn, txOut, nil
}
