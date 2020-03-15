// Copyright (c) 2015-2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package blockdag

import (
	"fmt"
	"github.com/kaspanet/kaspad/dagconfig"
	"github.com/pkg/errors"
	"math"
	"time"

	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/kaspanet/kaspad/wire"
)

// blockStatus is a bit field representing the validation state of the block.
type blockStatus byte

const (
	// statusDataStored indicates that the block's payload is stored on disk.
	statusDataStored blockStatus = 1 << iota

	// statusValid indicates that the block has been fully validated.
	statusValid

	// statusValidateFailed indicates that the block has failed validation.
	statusValidateFailed

	// statusInvalidAncestor indicates that one of the block's ancestors has
	// has failed validation, thus the block is also invalid.
	statusInvalidAncestor
)

// KnownValid returns whether the block is known to be valid. This will return
// false for a valid block that has not been fully validated yet.
func (status blockStatus) KnownValid() bool {
	return status&statusValid != 0
}

// KnownInvalid returns whether the block is known to be invalid. This may be
// because the block itself failed validation or any of its ancestors is
// invalid. This will return false for invalid blocks that have not been proven
// invalid yet.
func (status blockStatus) KnownInvalid() bool {
	return status&(statusValidateFailed|statusInvalidAncestor) != 0
}

// blockNode represents a block within the block DAG. The DAG is stored into
// the block database.
type blockNode struct {
	// NOTE: Additions, deletions, or modifications to the order of the
	// definitions in this struct should not be changed without considering
	// how it affects alignment on 64-bit platforms. The current order is
	// specifically crafted to result in minimal padding. There will be
	// hundreds of thousands of these in memory, so a few extra bytes of
	// padding adds up.

	// parents is the parent blocks for this node.
	parents blockSet

	// selectedParent is the selected parent for this node.
	// The selected parent is the parent that if chosen will maximize the blue score of this block
	selectedParent *blockNode

	// children are all the blocks that refer to this block as a parent
	children blockSet

	// blues are all blue blocks in this block's worldview that are in its selected parent anticone
	blues []*blockNode

	// blueScore is the count of all the blue blocks in this block's past
	blueScore uint64

	// bluesAnticoneSizes is a map holding the set of blues affected by this block and their
	// modified blue anticone size.
	bluesAnticoneSizes map[*blockNode]dagconfig.KType

	// hash is the double sha 256 of the block.
	hash *daghash.Hash

	// Some fields from block headers to aid in  reconstructing headers
	// from memory. These must be treated as immutable and are intentionally
	// ordered to avoid padding on 64-bit platforms.
	version              int32
	bits                 uint32
	nonce                uint64
	timestamp            int64
	hashMerkleRoot       *daghash.Hash
	acceptedIDMerkleRoot *daghash.Hash
	utxoCommitment       *daghash.Hash

	// status is a bitfield representing the validation state of the block. The
	// status field, unlike the other fields, may be written to and so should
	// only be accessed using the concurrent-safe NodeStatus method on
	// blockIndex once the node has been added to the global index.
	status blockStatus

	// isFinalized determines whether the node is below the finality point.
	isFinalized bool
}

// newBlockNode returns a new block node for the given block header and parents, and the
// anticone of its selected parent (parent with highest blue score).
// selectedParentAnticone is used to update reachability data we store for future reachability queries.
// This function is NOT safe for concurrent access.
func (dag *BlockDAG) newBlockNode(blockHeader *wire.BlockHeader, parents blockSet) (node *blockNode, selectedParentAnticone []*blockNode) {
	node = &blockNode{
		parents:            parents,
		children:           make(blockSet),
		blueScore:          math.MaxUint64, // Initialized to the max value to avoid collisions with the genesis block
		timestamp:          dag.Now().Unix(),
		bluesAnticoneSizes: make(map[*blockNode]dagconfig.KType),
	}

	// blockHeader is nil only for the virtual block
	if blockHeader != nil {
		node.hash = blockHeader.BlockHash()
		node.version = blockHeader.Version
		node.bits = blockHeader.Bits
		node.nonce = blockHeader.Nonce
		node.timestamp = blockHeader.Timestamp.Unix()
		node.hashMerkleRoot = blockHeader.HashMerkleRoot
		node.acceptedIDMerkleRoot = blockHeader.AcceptedIDMerkleRoot
		node.utxoCommitment = blockHeader.UTXOCommitment
	} else {
		node.hash = &daghash.ZeroHash
	}

	if len(parents) == 0 {
		// The genesis block is defined to have a blueScore of 0
		node.blueScore = 0
		return node, nil
	}

	selectedParentAnticone, err := dag.ghostdag(node)
	if err != nil {
		panic(errors.Wrap(err, "unexpected error in GHOSTDAG"))
	}
	return node, selectedParentAnticone
}

// updateParentsChildren updates the node's parents to point to new node
func (node *blockNode) updateParentsChildren() {
	for parent := range node.parents {
		parent.children.add(node)
	}
}

func (node *blockNode) less(other *blockNode) bool {
	if node.blueScore == other.blueScore {
		return daghash.Less(node.hash, other.hash)
	}

	return node.blueScore < other.blueScore
}

// Header constructs a block header from the node and returns it.
//
// This function is safe for concurrent access.
func (node *blockNode) Header() *wire.BlockHeader {
	// No lock is needed because all accessed fields are immutable.
	return &wire.BlockHeader{
		Version:              node.version,
		ParentHashes:         node.ParentHashes(),
		HashMerkleRoot:       node.hashMerkleRoot,
		AcceptedIDMerkleRoot: node.acceptedIDMerkleRoot,
		UTXOCommitment:       node.utxoCommitment,
		Timestamp:            time.Unix(node.timestamp, 0),
		Bits:                 node.bits,
		Nonce:                node.nonce,
	}
}

// SelectedAncestor returns the ancestor block node at the provided blue score by following
// the selected-parents chain backwards from this node. The returned block will be nil when a
// blue score is requested that is higher than the blue score of the passed node.
//
// This function is safe for concurrent access.
func (node *blockNode) SelectedAncestor(blueScore uint64) *blockNode {
	if blueScore > node.blueScore {
		return nil
	}

	n := node
	for n != nil && n.blueScore > blueScore {
		n = n.selectedParent
	}

	return n
}

// RelativeAncestor returns the ancestor block node a relative 'distance' of
// blue blocks before this node. This is equivalent to calling Ancestor with
// the node's blue score minus provided distance.
//
// This function is safe for concurrent access.
func (node *blockNode) RelativeAncestor(distance uint64) *blockNode {
	return node.SelectedAncestor(node.blueScore - distance)
}

// CalcPastMedianTime returns the median time of the previous few blocks
// prior to, and including, the block node.
//
// This function is safe for concurrent access.
func (node *blockNode) PastMedianTime(dag *BlockDAG) time.Time {
	window := blueBlockWindow(node, 2*dag.TimestampDeviationTolerance-1)
	medianTimestamp, err := window.medianTimestamp()
	if err != nil {
		panic(fmt.Sprintf("blueBlockWindow: %s", err))
	}
	return time.Unix(medianTimestamp, 0)
}

func (node *blockNode) ParentHashes() []*daghash.Hash {
	return node.parents.hashes()
}

// isGenesis returns if the current block is the genesis block
func (node *blockNode) isGenesis() bool {
	return len(node.parents) == 0
}

func (node *blockNode) finalityScore(dag *BlockDAG) uint64 {
	return node.blueScore / uint64(dag.dagParams.FinalityInterval)
}

// String returns a string that contains the block hash.
func (node blockNode) String() string {
	return node.hash.String()
}
