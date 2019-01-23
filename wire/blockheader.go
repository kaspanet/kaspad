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
// not including the list of parent block headers.
// Version 4 bytes + Timestamp 8 bytes + Bits 4 bytes + Nonce 8 bytes +
// + NumParentBlocks 1 byte + HashMerkleRoot hash + IDMerkleRoot hash.
// To get total size of block header len(ParentHashes) * daghash.HashSize should be
// added to this value
const BaseBlockHeaderPayload = 25 + 2*(daghash.HashSize)

// MaxNumParentBlocks is the maximum number of parent blocks a block can reference.
// Currently set to 255 as the maximum number NumParentBlocks can be due to it being a byte
const MaxNumParentBlocks = 255

// MaxBlockHeaderPayload is the maximum number of bytes a block header can be.
// BaseBlockHeaderPayload + up to MaxNumParentBlocks hashes of parent blocks
const MaxBlockHeaderPayload = BaseBlockHeaderPayload + (MaxNumParentBlocks * daghash.HashSize)

// BlockHeader defines information about a block and is used in the bitcoin
// block (MsgBlock) and headers (MsgHeader) messages.
type BlockHeader struct {
	// Version of the block.  This is not the same as the protocol version.
	Version int32

	// Hashes of the parent block headers in the blockDAG.
	ParentHashes []daghash.Hash

	// HashMerkleRoot is the merkle tree reference to hash of all transactions for the block.
	HashMerkleRoot daghash.Hash

	// IDMerkleRoot is the merkle tree reference to hash of all transactions' IDs for the block.
	IDMerkleRoot daghash.Hash

	// Time the block was created.
	Timestamp time.Time

	// Difficulty target for the block.
	Bits uint32

	// Nonce used to generate the block.
	Nonce uint64
}

// NumParentBlocks return the number of entries in ParentHashes
func (h *BlockHeader) NumParentBlocks() byte {
	return byte(len(h.ParentHashes))
}

// BlockHash computes the block identifier hash for the given block header.
func (h *BlockHeader) BlockHash() daghash.Hash {
	// Encode the header and double sha256 everything prior to the number of
	// transactions.  Ignore the error returns since there is no way the
	// encode could fail except being out of memory which would cause a
	// run-time panic.
	buf := bytes.NewBuffer(make([]byte, 0, BaseBlockHeaderPayload+h.NumParentBlocks()))
	_ = writeBlockHeader(buf, 0, h)

	return daghash.DoubleHashH(buf.Bytes())
}

// SelectedParentHash returns the hash of the selected block header.
func (h *BlockHeader) SelectedParentHash() *daghash.Hash {
	if h.NumParentBlocks() == 0 {
		return nil
	}

	return &h.ParentHashes[0]
}

// IsGenesis returns true iff this block is a genesis block
func (h *BlockHeader) IsGenesis() bool {
	return h.NumParentBlocks() == 0
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
	return BaseBlockHeaderPayload + int(h.NumParentBlocks())*daghash.HashSize
}

// NewBlockHeader returns a new BlockHeader using the provided version, previous
// block hash, hash merkle root, ID merkle root difficulty bits, and nonce used to generate the
// block with defaults or calclulated values for the remaining fields.
func NewBlockHeader(version int32, parentHashes []daghash.Hash, hashMerkleRoot *daghash.Hash,
	idMerkleRoot *daghash.Hash, bits uint32, nonce uint64) *BlockHeader {

	// Limit the timestamp to one second precision since the protocol
	// doesn't support better.
	return &BlockHeader{
		Version:        version,
		ParentHashes:   parentHashes,
		HashMerkleRoot: *hashMerkleRoot,
		IDMerkleRoot:   *idMerkleRoot,
		Timestamp:      time.Unix(time.Now().Unix(), 0),
		Bits:           bits,
		Nonce:          nonce,
	}
}

// readBlockHeader reads a bitcoin block header from r.  See Deserialize for
// decoding block headers stored to disk, such as in a database, as opposed to
// decoding from the wire.
func readBlockHeader(r io.Reader, pver uint32, bh *BlockHeader) error {
	var numParentBlocks byte
	err := readElements(r, &bh.Version, &numParentBlocks)
	if err != nil {
		return err
	}

	bh.ParentHashes = make([]daghash.Hash, numParentBlocks)
	for i := byte(0); i < numParentBlocks; i++ {
		err := readElement(r, &bh.ParentHashes[i])
		if err != nil {
			return err
		}
	}
	return readElements(r, &bh.HashMerkleRoot, &bh.IDMerkleRoot, (*int64Time)(&bh.Timestamp), &bh.Bits, &bh.Nonce)
}

// writeBlockHeader writes a bitcoin block header to w.  See Serialize for
// encoding block headers to be stored to disk, such as in a database, as
// opposed to encoding for the wire.
func writeBlockHeader(w io.Writer, pver uint32, bh *BlockHeader) error {
	sec := int64(bh.Timestamp.Unix())
	return writeElements(w, bh.Version, bh.NumParentBlocks(), &bh.ParentHashes, &bh.HashMerkleRoot, &bh.IDMerkleRoot,
		sec, bh.Bits, bh.Nonce)
}
