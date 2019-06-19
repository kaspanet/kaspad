package blockdag

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/daglabs/btcd/util/subnetworkid"
	"io"
	"math"

	"github.com/daglabs/btcd/database"
	"github.com/daglabs/btcd/util"
	"github.com/daglabs/btcd/util/daghash"
	"github.com/daglabs/btcd/util/txsort"
	"github.com/daglabs/btcd/wire"
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
var feeBucket = []byte("fees")

// getBluesFeeData returns the compactFeeData for all nodes's blues,
// used to calculate the fees this blockNode needs to pay
func (node *blockNode) getBluesFeeData(dag *BlockDAG) (map[daghash.Hash]compactFeeData, error) {
	bluesFeeData := make(map[daghash.Hash]compactFeeData)

	err := dag.db.View(func(dbTx database.Tx) error {
		for _, blueBlock := range node.blues {
			feeData, err := dbFetchFeeData(dbTx, blueBlock.hash)
			if err != nil {
				return fmt.Errorf("Error getting fee data for block %s: %s", blueBlock.hash, err)
			}

			bluesFeeData[*blueBlock.hash] = feeData
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return bluesFeeData, nil
}

func dbStoreFeeData(dbTx database.Tx, blockHash *daghash.Hash, feeData compactFeeData) error {
	feeBucket, err := dbTx.Metadata().CreateBucketIfNotExists(feeBucket)
	if err != nil {
		return fmt.Errorf("Error creating or retrieving fee bucket: %s", err)
	}

	return feeBucket.Put(blockHash.CloneBytes(), feeData)
}

func dbFetchFeeData(dbTx database.Tx, blockHash *daghash.Hash) (compactFeeData, error) {
	feeBucket := dbTx.Metadata().Bucket(feeBucket)
	if feeBucket == nil {
		return nil, errors.New("Fee bucket does not exist")
	}

	feeData := feeBucket.Get(blockHash.CloneBytes())
	if feeData == nil {
		return nil, fmt.Errorf("No fee data found for block %s", blockHash)
	}

	return feeData, nil
}

// The following functions deal with building and validating the coinbase transaction

func (node *blockNode) validateCoinbaseTransaction(dag *BlockDAG, block *util.Block, txsAcceptanceData MultiBlockTxsAcceptanceData) error {
	if node.isGenesis() {
		return nil
	}
	blockCoinbaseTx := block.CoinbaseTransaction().MsgTx()
	pkScript, extraData, err := DeserializeCoinbasePayload(blockCoinbaseTx)
	if err != nil {
		return err
	}
	expectedCoinbaseTransaction, err := node.expectedCoinbaseTransaction(dag, txsAcceptanceData, pkScript, extraData)
	if err != nil {
		return err
	}

	if !expectedCoinbaseTransaction.Hash().IsEqual(block.CoinbaseTransaction().Hash()) {
		return ruleError(ErrBadCoinbaseTransaction, "Coinbase transaction is not built as expected")
	}

	return nil
}

// expectedCoinbaseTransaction returns the coinbase transaction for the current block
func (node *blockNode) expectedCoinbaseTransaction(dag *BlockDAG, txsAcceptanceData MultiBlockTxsAcceptanceData, pkScript []byte, extraData []byte) (*util.Tx, error) {
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
	payload, err := SerializeCoinbasePayload(pkScript, extraData)
	if err != nil {
		return nil, err
	}
	coinbaseTx := wire.NewSubnetworkMsgTx(wire.TxVersion, txIns, txOuts, subnetworkid.SubnetworkIDCoinbase, 0, payload)
	sortedCoinbaseTx := txsort.Sort(coinbaseTx)
	return util.NewTx(sortedCoinbaseTx), nil
}

// SerializeCoinbasePayload builds the coinbase payload based on the provided pkScript and extra data.
func SerializeCoinbasePayload(pkScript []byte, extraData []byte) ([]byte, error) {
	w := &bytes.Buffer{}
	err := wire.WriteVarInt(w, uint64(len(pkScript)))
	if err != nil {
		return nil, err
	}
	_, err = w.Write(pkScript)
	if err != nil {
		return nil, err
	}
	_, err = w.Write(extraData)
	if err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

// DeserializeCoinbasePayload deserialize the coinbase payload to its component (pkScript and extra data).
func DeserializeCoinbasePayload(tx *wire.MsgTx) (pkScript []byte, extraData []byte, err error) {
	r := bytes.NewReader(tx.Payload)
	pkScriptLen, err := wire.ReadVarInt(r)
	if err != nil {
		return nil, nil, err
	}
	pkScript = make([]byte, pkScriptLen)
	_, err = r.Read(pkScript)
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
	return pkScript, extraData, nil
}

// feeInputAndOutputForBlueBlock calculates the input and output that should go into the coinbase transaction of blueBlock
// If blueBlock gets no fee - returns only txIn and nil for txOut
func coinbaseInputAndOutputForBlueBlock(dag *BlockDAG, blueBlock *blockNode,
	txsAcceptanceData MultiBlockTxsAcceptanceData, feeData map[daghash.Hash]compactFeeData) (
	*wire.TxIn, *wire.TxOut, error) {

	blockTxsAcceptanceData, ok := txsAcceptanceData[*blueBlock.hash]
	if !ok {
		return nil, nil, fmt.Errorf("No txsAcceptanceData for block %s", blueBlock.hash)
	}
	blockFeeData, ok := feeData[*blueBlock.hash]
	if !ok {
		return nil, nil, fmt.Errorf("No feeData for block %s", blueBlock.hash)
	}

	if len(blockTxsAcceptanceData) != blockFeeData.Len() {
		return nil, nil, fmt.Errorf(
			"length of accepted transaction data(%d) and fee data(%d) is not equal for block %s",
			len(blockTxsAcceptanceData), blockFeeData.Len(), blueBlock.hash)
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

	for _, txAcceptanceData := range blockTxsAcceptanceData {
		fee, err := feeIterator.next()
		if err != nil {
			return nil, nil, fmt.Errorf("Error retrieving fee from compactFeeData iterator: %s", err)
		}
		if txAcceptanceData.IsAccepted {
			totalFees += fee
		}
	}

	totalReward := CalcBlockSubsidy(blueBlock.height, dag.dagParams) + totalFees

	if totalReward == 0 {
		return txIn, nil, nil
	}

	// the PkScript for the coinbase is parsed from the coinbase payload
	pkScript, _, err := DeserializeCoinbasePayload(blockTxsAcceptanceData[0].Tx.MsgTx())
	if err != nil {
		return nil, nil, err
	}

	txOut := &wire.TxOut{
		Value:    totalReward,
		PkScript: pkScript,
	}

	return txIn, txOut, nil
}
