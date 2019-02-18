package blockdag

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/daglabs/btcd/dagconfig/daghash"
	"github.com/daglabs/btcd/database"
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

func dbStoreFeeData(dbTx database.Tx, blockHash *daghash.Hash, feeData compactFeeData) error {
	feeBucket, err := dbTx.Metadata().CreateBucketIfNotExists([]byte("fees"))
	if err != nil {
		return fmt.Errorf("Error creating or retrieving fee bucket: %s", err)
	}

	return feeBucket.Put(blockHash.CloneBytes(), feeData)
}
