package blockdag

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"

	"github.com/daglabs/btcd/dagconfig/daghash"
	"github.com/daglabs/btcd/database"
	"github.com/daglabs/btcd/util/txsort"
	"github.com/daglabs/btcd/wire"
)

// compactFeeData is a specialized data type to store a compact list of fees
// inside a block.
// Every transaction gets a single uint64 value, stored as a plain binary list.
// The transactions are ordered the same way they are ordered inside the block, making it easy
// to traverse all transactions in a block and extract it's fee from the fee accumulator.
//
// compactFeeFactory is used to create such a list.
// compactFeeIterator is used to iterate over such a list.

type compactFeeData []byte

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

// following functions relate to storing and retreiving fee data from the dabase
var feeBucket = []byte("fees")

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

// following function deal with building the fee transaction

// buildFeeTransaction returns the expected fee transaction for the current block
func (node *blockNode) buildFeeTransaction(dag *BlockDAG, acceptedTxsData AcceptedTxsData) (*wire.MsgTx, error) {
	bluesFeeData, err := node.getBluesFeeData(dag)
	if err != nil {
		return nil, err
	}

	feeTx := wire.NewMsgTx(wire.TxVersion)

	for _, blue := range node.blues {
		txIn, txOut, err := feeInputAndOutputForBlueBlock(blue, acceptedTxsData, bluesFeeData)
		if err != nil {
			return nil, err
		}
		if txIn == nil && txOut == nil {
			continue
		}

		feeTx.AddTxIn(txIn)
		feeTx.AddTxOut(txOut)
	}
	return txsort.Sort(feeTx), nil
}

// feeInputAndOutputForBlueBlock calculatres the input and output that should go into the fee transaction
// for given blueNode
// If block gets no fee - returns nil in all return values
func feeInputAndOutputForBlueBlock(blueBlock *blockNode, acceptedTxsData AcceptedTxsData, feeData map[daghash.Hash]compactFeeData) (
	*wire.TxIn, *wire.TxOut, error) {

	blockAcceptedTxsData := acceptedTxsData[blueBlock.hash]
	blockFeeData := feeData[blueBlock.hash]

	if len(blockAcceptedTxsData) != len(blockFeeData) {
		return nil, nil, fmt.Errorf(
			"length of accepted transaction data and fee data is not equal for block %s", blueBlock.hash)
	}

	txIn := &wire.TxIn{
		SignatureScript: []byte{},
		PreviousOutPoint: wire.OutPoint{
			TxID:  daghash.TxID(blueBlock.hash),
			Index: math.MaxUint32,
		},
		Sequence: wire.MaxTxInSequenceNum,
	}

	totalFees := uint64(0)
	feeIterator := blockFeeData.iterator()

	for _, txAcceptedData := range blockAcceptedTxsData {
		fee, err := feeIterator.next()
		if err != nil {
			return nil, nil, fmt.Errorf("Error retrieving fee from compactFeeData iterator: %s", err)
		}
		if txAcceptedData.IsAccepted {
			totalFees += fee
		}
	}

	if totalFees == 0 {
		return nil, nil, nil
	}

	// the scriptPubKey for the fee is the same as the coinbase's first scriptPubKey
	pkScript := blockAcceptedTxsData[0].Tx.MsgTx().TxOut[0].PkScript

	txOut := &wire.TxOut{
		Value:    totalFees,
		PkScript: pkScript,
	}

	return txIn, txOut, nil
}
