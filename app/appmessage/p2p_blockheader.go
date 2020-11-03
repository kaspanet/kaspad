// Copyright (c) 2013-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package appmessage

import (
	"io"
	"math"

	"github.com/kaspanet/kaspad/domain/consensus/utils/hashserialization"

	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/util/mstime"
	"github.com/pkg/errors"
)

// BaseBlockHeaderPayload is the base number of bytes a block header can be,
// not including the list of parent block headers.
// Version 4 bytes + Timestamp 8 bytes + Bits 4 bytes + Nonce 8 bytes +
// + NumParentBlocks 1 byte + HashMerkleRoot hash +
// + AcceptedIDMerkleRoot hash + UTXOCommitment hash.
// To get total size of block header len(ParentHashes) * externalapi.DomainHashSize should be
// added to this value
const BaseBlockHeaderPayload = 25 + 3*(externalapi.DomainHashSize)

// MaxNumParentBlocks is the maximum number of parent blocks a block can reference.
// Currently set to 255 as the maximum number NumParentBlocks can be due to it being a byte
const MaxNumParentBlocks = 255

// MaxBlockHeaderPayload is the maximum number of bytes a block header can be.
// BaseBlockHeaderPayload + up to MaxNumParentBlocks hashes of parent blocks
const MaxBlockHeaderPayload = BaseBlockHeaderPayload + (MaxNumParentBlocks * externalapi.DomainHashSize)

// BlockHeader defines information about a block and is used in the kaspa
// block (MsgBlock) and headers (MsgHeader) messages.
type BlockHeader struct {
	// Version of the block. This is not the same as the protocol version.
	Version int32

	// Hashes of the parent block headers in the blockDAG.
	ParentHashes []*externalapi.DomainHash

	// HashMerkleRoot is the merkle tree reference to hash of all transactions for the block.
	HashMerkleRoot *externalapi.DomainHash

	// AcceptedIDMerkleRoot is merkle tree reference to hash all transactions
	// accepted form the block.Blues
	AcceptedIDMerkleRoot *externalapi.DomainHash

	// UTXOCommitment is an ECMH UTXO commitment to the block UTXO.
	UTXOCommitment *externalapi.DomainHash

	// Time the block was created.
	Timestamp mstime.Time

	// Difficulty target for the block.
	Bits uint32

	// Nonce used to generate the block.
	Nonce uint64
}

// NumParentBlocks return the number of entries in ParentHashes
func (h *BlockHeader) NumParentBlocks() byte {
	numParents := len(h.ParentHashes)
	if numParents > math.MaxUint8 {
		panic(errors.Errorf("number of parents is %d, which is more than one byte can fit", numParents))
	}
	return byte(numParents)
}

// BlockHash computes the block identifier hash for the given block header.
func (h *BlockHeader) BlockHash() *externalapi.DomainHash {
	return hashserialization.HeaderHash(BlockHeaderToDomainBlockHeader(h))
}

// IsGenesis returns true iff this block is a genesis block
func (h *BlockHeader) IsGenesis() bool {
	return h.NumParentBlocks() == 0
}

// KaspaDecode decodes r using the kaspa protocol encoding into the receiver.
// This is part of the Message interface implementation.
// See Deserialize for decoding block headers stored to disk, such as in a
// database, as opposed to decoding block headers from the appmessage.
func (h *BlockHeader) KaspaDecode(r io.Reader, pver uint32) error {
	return readBlockHeader(r, pver, h)
}

// KaspaEncode encodes the receiver to w using the kaspa protocol encoding.
// This is part of the Message interface implementation.
// See Serialize for encoding block headers to be stored to disk, such as in a
// database, as opposed to encoding block headers for the appmessage.
func (h *BlockHeader) KaspaEncode(w io.Writer, pver uint32) error {
	return writeBlockHeader(w, pver, h)
}

// Deserialize decodes a block header from r into the receiver using a format
// that is suitable for long-term storage such as a database while respecting
// the Version field.
func (h *BlockHeader) Deserialize(r io.Reader) error {
	// At the current time, there is no difference between the appmessage encoding
	// at protocol version 0 and the stable long-term storage format. As
	// a result, make use of readBlockHeader.
	return readBlockHeader(r, 0, h)
}

// Serialize encodes a block header from r into the receiver using a format
// that is suitable for long-term storage such as a database while respecting
// the Version field.
func (h *BlockHeader) Serialize(w io.Writer) error {
	// At the current time, there is no difference between the appmessage encoding
	// at protocol version 0 and the stable long-term storage format. As
	// a result, make use of writeBlockHeader.
	return writeBlockHeader(w, 0, h)
}

// SerializeSize returns the number of bytes it would take to serialize the
// block header.
func (h *BlockHeader) SerializeSize() int {
	return BaseBlockHeaderPayload + int(h.NumParentBlocks())*externalapi.DomainHashSize
}

// NewBlockHeader returns a new BlockHeader using the provided version, previous
// block hash, hash merkle root, accepted ID merkle root, difficulty bits, and nonce used to generate the
// block with defaults or calclulated values for the remaining fields.
func NewBlockHeader(version int32, parentHashes []*externalapi.DomainHash, hashMerkleRoot *externalapi.DomainHash,
	acceptedIDMerkleRoot *externalapi.DomainHash, utxoCommitment *externalapi.DomainHash, bits uint32, nonce uint64) *BlockHeader {

	// Limit the timestamp to one millisecond precision since the protocol
	// doesn't support better.
	return &BlockHeader{
		Version:              version,
		ParentHashes:         parentHashes,
		HashMerkleRoot:       hashMerkleRoot,
		AcceptedIDMerkleRoot: acceptedIDMerkleRoot,
		UTXOCommitment:       utxoCommitment,
		Timestamp:            mstime.Now(),
		Bits:                 bits,
		Nonce:                nonce,
	}
}

// readBlockHeader reads a kaspa block header from r. See Deserialize for
// decoding block headers stored to disk, such as in a database, as opposed to
// decoding from the appmessage.
func readBlockHeader(r io.Reader, pver uint32, bh *BlockHeader) error {
	var numParentBlocks byte
	err := readElements(r, &bh.Version, &numParentBlocks)
	if err != nil {
		return err
	}

	bh.ParentHashes = make([]*externalapi.DomainHash, numParentBlocks)
	for i := byte(0); i < numParentBlocks; i++ {
		hash := &externalapi.DomainHash{}
		err := ReadElement(r, hash)
		if err != nil {
			return err
		}
		bh.ParentHashes[i] = hash
	}
	bh.HashMerkleRoot = &externalapi.DomainHash{}
	bh.AcceptedIDMerkleRoot = &externalapi.DomainHash{}
	bh.UTXOCommitment = &externalapi.DomainHash{}
	return readElements(r, bh.HashMerkleRoot, bh.AcceptedIDMerkleRoot, bh.UTXOCommitment,
		(*int64Time)(&bh.Timestamp), &bh.Bits, &bh.Nonce)
}

// writeBlockHeader writes a kaspa block header to w. See Serialize for
// encoding block headers to be stored to disk, such as in a database, as
// opposed to encoding for the appmessage.
func writeBlockHeader(w io.Writer, pver uint32, bh *BlockHeader) error {
	timestamp := bh.Timestamp.UnixMilliseconds()
	if err := writeElements(w, bh.Version, bh.NumParentBlocks()); err != nil {
		return err
	}
	for _, hash := range bh.ParentHashes {
		if err := WriteElement(w, hash); err != nil {
			return err
		}
	}
	return writeElements(w, bh.HashMerkleRoot, bh.AcceptedIDMerkleRoot, bh.UTXOCommitment, timestamp, bh.Bits, bh.Nonce)
}
