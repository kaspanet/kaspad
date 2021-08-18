// Copyright (c) 2013-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package appmessage

import (
	"math"
	"math/big"

	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"

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

// MsgBlockHeader defines information about a block and is used in the kaspa
// block (MsgBlock) and headers (MsgHeader) messages.
type MsgBlockHeader struct {
	baseMessage

	// Version of the block. This is not the same as the protocol version.
	Version uint16

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

	// DAASCore is the DAA score of the block.
	DAAScore uint64

	// BlueWork is the blue work of the block.
	BlueWork *big.Int

	PruningPoint *externalapi.DomainHash
}

// NumParentBlocks return the number of entries in ParentHashes
func (h *MsgBlockHeader) NumParentBlocks() byte {
	numParents := len(h.ParentHashes)
	if numParents > math.MaxUint8 {
		panic(errors.Errorf("number of parents is %d, which is more than one byte can fit", numParents))
	}
	return byte(numParents)
}

// BlockHash computes the block identifier hash for the given block header.
func (h *MsgBlockHeader) BlockHash() *externalapi.DomainHash {
	return consensushashing.HeaderHash(BlockHeaderToDomainBlockHeader(h))
}

// IsGenesis returns true iff this block is a genesis block
func (h *MsgBlockHeader) IsGenesis() bool {
	return h.NumParentBlocks() == 0
}

// NewBlockHeader returns a new MsgBlockHeader using the provided version, previous
// block hash, hash merkle root, accepted ID merkle root, difficulty bits, and nonce used to generate the
// block with defaults or calclulated values for the remaining fields.
func NewBlockHeader(version uint16, parentHashes []*externalapi.DomainHash, hashMerkleRoot *externalapi.DomainHash,
	acceptedIDMerkleRoot *externalapi.DomainHash, utxoCommitment *externalapi.DomainHash, bits uint32, nonce uint64,
	daaScore uint64, blueWork *big.Int, pruningPoint *externalapi.DomainHash) *MsgBlockHeader {

	// Limit the timestamp to one millisecond precision since the protocol
	// doesn't support better.
	return &MsgBlockHeader{
		Version:              version,
		ParentHashes:         parentHashes,
		HashMerkleRoot:       hashMerkleRoot,
		AcceptedIDMerkleRoot: acceptedIDMerkleRoot,
		UTXOCommitment:       utxoCommitment,
		Timestamp:            mstime.Now(),
		Bits:                 bits,
		Nonce:                nonce,
		DAAScore:             daaScore,
		BlueWork:             blueWork,
		PruningPoint:         pruningPoint,
	}
}
