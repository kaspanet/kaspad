package blockdag

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"io"
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
