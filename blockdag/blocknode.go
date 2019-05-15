// Copyright (c) 2015-2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package blockdag

import (
	"fmt"
	"sort"
	"time"

	"github.com/daglabs/btcd/util/daghash"
	"github.com/daglabs/btcd/wire"
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

	// statusNone indicates that the block has no validation state flags set.
	//
	// NOTE: This must be defined last in order to avoid influencing iota.
	statusNone blockStatus = 0
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
	// how it affects alignment on 64-bit platforms.  The current order is
	// specifically crafted to result in minimal padding.  There will be
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

	// diff is the UTXO representation of the block
	// A block's UTXO is reconstituted by applying diffWith on every block in the chain of diffChildren
	// from the virtual block down to the block. See diffChild
	diff *UTXODiff

	// diffChild is the child that diff will be built from. See diff
	diffChild *blockNode

	// hash is the double sha 256 of the block.
	hash *daghash.Hash

	// height is the position in the block DAG.
	height uint64

	// chainHeight is the number of hops you need to go down the selected parent chain in order to get to the genesis block.
	chainHeight uint64

	// Some fields from block headers to aid in best chain selection and
	// reconstructing headers from memory.  These must be treated as
	// immutable and are intentionally ordered to avoid padding on 64-bit
	// platforms.
	version              int32
	bits                 uint32
	nonce                uint64
	timestamp            int64
	hashMerkleRoot       *daghash.Hash
	idMerkleRoot         *daghash.Hash
	acceptedIDMerkleRoot *daghash.Hash
	utxoCommitment       *daghash.Hash

	// status is a bitfield representing the validation state of the block. The
	// status field, unlike the other fields, may be written to and so should
	// only be accessed using the concurrent-safe NodeStatus method on
	// blockIndex once the node has been added to the global index.
	status blockStatus
}

// initBlockNode initializes a block node from the given header and parent nodes.
// This function is NOT safe for concurrent access.  It must only be called when
// initially creating a node.
func initBlockNode(node *blockNode, blockHeader *wire.BlockHeader, parents blockSet, phantomK uint32) {
	*node = blockNode{
		parents:   parents,
		children:  make(blockSet),
		timestamp: time.Now().Unix(),
	}

	// blockHeader is nil only for the virtual block
	if blockHeader != nil {
		node.hash = blockHeader.BlockHash()
		node.version = blockHeader.Version
		node.bits = blockHeader.Bits
		node.nonce = blockHeader.Nonce
		node.timestamp = blockHeader.Timestamp.Unix()
		node.hashMerkleRoot = blockHeader.HashMerkleRoot
		node.idMerkleRoot = blockHeader.IDMerkleRoot
		node.acceptedIDMerkleRoot = blockHeader.AcceptedIDMerkleRoot
		node.utxoCommitment = blockHeader.UTXOCommitment
	} else {
		node.hash = &daghash.ZeroHash
	}

	if len(parents) > 0 {
		node.blues, node.selectedParent, node.blueScore = phantom(node, phantomK)
		node.height = calculateNodeHeight(node)
		node.chainHeight = calculateChainHeight(node)
	}
}

func calculateNodeHeight(node *blockNode) uint64 {
	if node.isGenesis() {
		return 0
	}
	return node.parents.maxHeight() + 1
}

func calculateChainHeight(node *blockNode) uint64 {
	if node.isGenesis() {
		return 0
	}
	return node.selectedParent.chainHeight + 1
}

// newBlockNode returns a new block node for the given block header and parent
//nodes. This function is NOT safe for concurrent access.
func newBlockNode(blockHeader *wire.BlockHeader, parents blockSet, phantomK uint32) *blockNode {
	var node blockNode
	initBlockNode(&node, blockHeader, parents, phantomK)
	return &node
}

// updateParentsChildren updates the node's parents to point to new node
func (node *blockNode) updateParentsChildren() {
	for _, parent := range node.parents {
		parent.children.add(node)
	}
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
		IDMerkleRoot:         node.idMerkleRoot,
		AcceptedIDMerkleRoot: node.acceptedIDMerkleRoot,
		UTXOCommitment:       node.utxoCommitment,
		Timestamp:            time.Unix(node.timestamp, 0),
		Bits:                 node.bits,
		Nonce:                node.nonce,
	}
}

// SelectedAncestor returns the ancestor block node at the provided chain-height by following
// the selected-parents chain backwards from this node. The returned block will be nil when a
// height is requested that is after the height of the passed node.
//
// This function is safe for concurrent access.
func (node *blockNode) SelectedAncestor(chainHeight uint64) *blockNode {
	if chainHeight < 0 || chainHeight > node.chainHeight {
		return nil
	}

	n := node
	for ; n != nil && n.chainHeight != chainHeight; n = n.selectedParent {
		// Intentionally left blank
	}

	return n
}

// RelativeAncestor returns the ancestor block node a relative 'distance' of
// chain-blocks before this node. This is equivalent to calling Ancestor with
// the node's chain-height minus provided distance.
//
// This function is safe for concurrent access.
func (node *blockNode) RelativeAncestor(distance uint64) *blockNode {
	return node.SelectedAncestor(node.chainHeight - distance)
}

// PastMedianTime returns the median time of the previous few blocks
// prior to, and including, the block node.
//
// This function is safe for concurrent access.
func (node *blockNode) PastMedianTime() time.Time {
	// Create a slice of the previous few block timestamps used to calculate
	// the median per the number defined by the constant medianTimeBlocks.
	// If there aren't enough blocks yet - pad remaining with genesis block's timestamp.
	timestamps := make([]int64, medianTimeBlocks)
	iterNode := node
	for i := 0; i < medianTimeBlocks; i++ {
		timestamps[i] = iterNode.timestamp

		if !iterNode.isGenesis() {
			iterNode = iterNode.selectedParent
		}
	}

	sort.Sort(timeSorter(timestamps))

	// Note: This works when medianTimeBlockCount is an odd number.
	// If it is to be changed to an even number - must take avarage of two middle values
	// Since medianTimeBlockCount is a constant, we can skip the odd/even check
	medianTimestamp := timestamps[medianTimeBlocks/2]
	return time.Unix(medianTimestamp, 0)
}

func (node *blockNode) ParentHashes() []*daghash.Hash {
	return node.parents.hashes()
}

// isGenesis returns if the current block is the genesis block
func (node *blockNode) isGenesis() bool {
	return len(node.parents) == 0
}

func (node *blockNode) finalityScore() uint64 {
	return node.blueScore / FinalityInterval
}

// String returns a string that contains the block hash and height.
func (node blockNode) String() string {
	return fmt.Sprintf("%s (%d)", node.hash, node.height)
}
