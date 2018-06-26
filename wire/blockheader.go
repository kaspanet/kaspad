// Copyright (c) 2013-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package wire

import (
	"bytes"
	"io"
	"time"

	"github.com/daglabs/btcd/dagconfig/daghash"
)

// BaseBlockHeaderPayload is the base number of bytes a block header can be,
// not including the list of previous block headers.
// Version 4 bytes + Timestamp 4 bytes + Bits 4 bytes + Nonce 4 bytes +
// + NumPrevBlocks 1 byte + MerkleRoot hash.
// To get total size of block header len(PrevBlocks) * daghash.HashSize should be
// added to this value
const BaseBlockHeaderPayload = 17 + (daghash.HashSize)

// MaxNumPrevBlocks is the maximum number of previous blocks a block can reference.
// Currently set to 255 as the maximum number NumPrevBlocks can be due to it being a byte
const MaxNumPrevBlocks = 255

// MaxBlockHeaderPayload is the maximum number of bytes a block header can be.
// BaseBlockHeaderPayload + up to MaxNumPrevBlocks hashes of previous blocks
const MaxBlockHeaderPayload = BaseBlockHeaderPayload + (MaxNumPrevBlocks * daghash.HashSize)

// BlockHeader defines information about a block and is used in the bitcoin
// block (MsgBlock) and headers (MsgHeader) messages.
type BlockHeader struct {
	// Version of the block.  This is not the same as the protocol version.
	Version int32

	// Number of entries in PrevBlocks
	NumPrevBlocks byte

	// Hashes of the previous block headers in the blockDAG.
	PrevBlocks []daghash.Hash

	// Merkle tree reference to hash of all transactions for the block.
	MerkleRoot daghash.Hash

	// Time the block was created.  This is, unfortunately, encoded as a
	// uint32 on the wire and therefore is limited to 2106.
	Timestamp time.Time

	// Difficulty target for the block.
	Bits uint32

	// Nonce used to generate the block.
	Nonce uint32
}

// BlockHash computes the block identifier hash for the given block header.
func (h *BlockHeader) BlockHash() daghash.Hash {
	// Encode the header and double sha256 everything prior to the number of
	// transactions.  Ignore the error returns since there is no way the
	// encode could fail except being out of memory which would cause a
	// run-time panic.
	buf := bytes.NewBuffer(make([]byte, 0, BaseBlockHeaderPayload+len(h.PrevBlocks)))
	_ = writeBlockHeader(buf, 0, h)

	return daghash.DoubleHashH(buf.Bytes())
}

// BestPrevBlock returns the hash of the selected block header.
func (header *BlockHeader) SelectedPrevBlock() *daghash.Hash {
	if header.NumPrevBlocks == 0 {
		return nil
	}

	return &header.PrevBlocks[0]
}

// BtcDecode decodes r using the bitcoin protocol encoding into the receiver.
// This is part of the Message interface implementation.
// See Deserialize for decoding block headers stored to disk, such as in a
// database, as opposed to decoding block headers from the wire.
func (h *BlockHeader) BtcDecode(r io.Reader, pver uint32) error {
	return readBlockHeader(r, pver, h)
}

// BtcEncode encodes the receiver to w using the bitcoin protocol encoding.
// This is part of the Message interface implementation.
// See Serialize for encoding block headers to be stored to disk, such as in a
// database, as opposed to encoding block headers for the wire.
func (h *BlockHeader) BtcEncode(w io.Writer, pver uint32) error {
	return writeBlockHeader(w, pver, h)
}

// Deserialize decodes a block header from r into the receiver using a format
// that is suitable for long-term storage such as a database while respecting
// the Version field.
func (h *BlockHeader) Deserialize(r io.Reader) error {
	// At the current time, there is no difference between the wire encoding
	// at protocol version 0 and the stable long-term storage format.  As
	// a result, make use of readBlockHeader.
	return readBlockHeader(r, 0, h)
}

// Serialize encodes a block header from r into the receiver using a format
// that is suitable for long-term storage such as a database while respecting
// the Version field.
func (h *BlockHeader) Serialize(w io.Writer) error {
	// At the current time, there is no difference between the wire encoding
	// at protocol version 0 and the stable long-term storage format.  As
	// a result, make use of writeBlockHeader.
	return writeBlockHeader(w, 0, h)
}

// SerializeSize returns the number of bytes it would take to serialize the
// block header.
func (h *BlockHeader) SerializeSize() int {
	return BaseBlockHeaderPayload + int(h.NumPrevBlocks)*daghash.HashSize
}

// NewBlockHeader returns a new BlockHeader using the provided version, previous
// block hash, merkle root hash, difficulty bits, and nonce used to generate the
// block with defaults or calclulated values for the remaining fields.
func NewBlockHeader(version int32, prevHashes []daghash.Hash, merkleRootHash *daghash.Hash,
	bits uint32, nonce uint32) *BlockHeader {

	// Limit the timestamp to one second precision since the protocol
	// doesn't support better.
	return &BlockHeader{
		Version:       version,
		NumPrevBlocks: byte(len(prevHashes)),
		PrevBlocks:    prevHashes,
		MerkleRoot:    *merkleRootHash,
		Timestamp:     time.Unix(time.Now().Unix(), 0),
		Bits:          bits,
		Nonce:         nonce,
	}
}

// readBlockHeader reads a bitcoin block header from r.  See Deserialize for
// decoding block headers stored to disk, such as in a database, as opposed to
// decoding from the wire.
func readBlockHeader(r io.Reader, pver uint32, bh *BlockHeader) error {
	err := readElements(r, &bh.Version, &bh.NumPrevBlocks)
	if err != nil {
		return err
	}

	bh.PrevBlocks = make([]daghash.Hash, bh.NumPrevBlocks)
	for i := byte(0); i < bh.NumPrevBlocks; i++ {
		err := readElement(r, &bh.PrevBlocks[i])
		if err != nil {
			return err
		}
	}
	return readElements(r, &bh.MerkleRoot, (*uint32Time)(&bh.Timestamp), &bh.Bits, &bh.Nonce)
}

// writeBlockHeader writes a bitcoin block header to w.  See Serialize for
// encoding block headers to be stored to disk, such as in a database, as
// opposed to encoding for the wire.
func writeBlockHeader(w io.Writer, pver uint32, bh *BlockHeader) error {
	sec := uint32(bh.Timestamp.Unix())
	return writeElements(w, bh.Version, bh.NumPrevBlocks, &bh.PrevBlocks, &bh.MerkleRoot,
		sec, bh.Bits, bh.Nonce)
}
