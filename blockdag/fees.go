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

// The fee accumulator is a specialized data type to store a compact list of fees
// inside a block.
// Every transaction gets a single uint64 value, stored as a plain binary list.
// The transactions are ordered the same way they are ordered inside the block, making it easy
// to traverse all transactions in a block and extract it's fee from the fee accumulator.
//
// feeAccumulatorWriter is used to create such a list.
// feeAccumulatorReader is used to read such a list.

type feeAccumulatorWriter struct {
	data   *bytes.Buffer
	writer *bufio.Writer
}

func newFeeAccumulatorWriter() *feeAccumulatorWriter {
	buffer := bytes.NewBuffer([]byte{})
	return &feeAccumulatorWriter{
		data:   buffer,
		writer: bufio.NewWriter(buffer),
	}
}

func (faw *feeAccumulatorWriter) addTxFee(txFee uint64) error {
	return binary.Write(faw.writer, binary.LittleEndian, txFee)
}

func (faw *feeAccumulatorWriter) bytes() ([]byte, error) {
	err := faw.writer.Flush()

	return faw.data.Bytes(), err
}

type feeAccumulatorReader struct {
	data   *bytes.Buffer
	reader io.Reader
}

func newFeeAccumulatorReader(data []byte) *feeAccumulatorReader {
	buffer := bytes.NewBuffer(data)
	return &feeAccumulatorReader{
		data:   buffer,
		reader: bufio.NewReader(buffer),
	}
}

func (far *feeAccumulatorReader) nextTxFee() (uint64, error) {
	var txFee uint64

	err := binary.Read(far.reader, binary.LittleEndian, &txFee)

	return txFee, err
}

// following functions relate to storing and retreiving fee data from the dabase

func dbStoreFeeData(dbTx database.Tx, blockHash *daghash.Hash, feeData []byte) error {
	feeBucket, err := dbTx.Metadata().CreateBucketIfNotExists([]byte("fees"))
	if err != nil {
		return fmt.Errorf("Error creating or retrieving fee bucket: %s", err)
	}

	return feeBucket.Put(blockHash.CloneBytes(), feeData)
}
