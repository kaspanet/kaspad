// Copyright (c) 2015-2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package blocknode

import (
	"math"

	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/kaspanet/kaspad/util/mstime"

	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/util/daghash"
)

// Status is representing the validation state of the block.
type Status byte

const (
	// StatusDataStored indicates that the block's payload is stored on disk.
	StatusDataStored Status = iota

	// StatusValid indicates that the block has been fully validated.
	StatusValid

	// StatusValidateFailed indicates that the block has failed validation.
	StatusValidateFailed

	// StatusInvalidAncestor indicates that one of the block's ancestors has
	// has failed validation, thus the block is also invalid.
	StatusInvalidAncestor

	// StatusUTXOPendingVerification indicates that the block is pending verification against its past UTXO-Set, either
	// because it was not yet verified since the block was never in the selected parent chain, or if the
	// block violates finality.
	StatusUTXOPendingVerification

	// StatusDisqualifiedFromChain indicates that the block is not eligible to be a selected parent.
	StatusDisqualifiedFromChain
)

var BlockStatusToString = map[Status]string{
	StatusDataStored:              "StatusDataStored",
	StatusValid:                   "StatusValid",
	StatusValidateFailed:          "StatusValidateFailed",
	StatusInvalidAncestor:         "StatusInvalidAncestor",
	StatusUTXOPendingVerification: "StatusUTXOPendingVerification",
	StatusDisqualifiedFromChain:   "StatusDisqualifiedFromChain",
}

func (status Status) String() string {
	return BlockStatusToString[status]
}

// KnownValid returns whether the block is known to be valid. This will return
// false for a valid block that has not been fully validated yet.
func (status Status) KnownValid() bool {
	return status == StatusValid
}

// KnownInvalid returns whether the block is known to be invalid. This may be
// because the block itself failed validation or any of its ancestors is
// invalid. This will return false for invalid blocks that have not been proven
// invalid yet.
func (status Status) KnownInvalid() bool {
	return status == StatusValidateFailed || status == StatusInvalidAncestor
}

// Node represents a block within the block DAG. The DAG is stored into
// the block database.
type Node struct {
	// NOTE: Additions, deletions, or modifications to the order of the
	// definitions in this struct should not be changed without considering
	// how it affects alignment on 64-bit platforms. The current order is
	// specifically crafted to result in minimal padding. There will be
	// hundreds of thousands of these in memory, so a few extra bytes of
	// padding adds up.

	// Parents is the parent blocks for this node.
	Parents Set

	// SelectedParent is the selected parent for this node.
	// The selected parent is the parent that if chosen will maximize the blue score of this block
	SelectedParent *Node

	// Children are all the blocks that refer to this block as a parent
	Children Set

	// Blues are all blue blocks in this block's worldview that are in its merge set
	Blues []*Node

	// Reds are all red blocks in this block's worldview that are in its merge set
	Reds []*Node

	// BlueScore is the count of all the blue blocks in this block's past
	BlueScore uint64

	// BluesAnticoneSizes is a map holding the set of Blues affected by this block and their
	// modified blue anticone size.
	BluesAnticoneSizes map[*Node]dagconfig.KType

	// Hash is the double sha 256 of the block.
	Hash *daghash.Hash

	// Some fields from block headers to aid in reconstructing headers
	// from memory. These must be treated as immutable and are intentionally
	// ordered to avoid padding on 64-bit platforms.
	Version              int32
	Bits                 uint32
	Nonce                uint64
	Timestamp            int64
	HashMerkleRoot       *daghash.Hash
	AcceptedIDMerkleRoot *daghash.Hash
	UTXOCommitment       *daghash.Hash

	// Status is a bitfield representing the validation state of the block. The
	// Status field, unlike the other fields, may be written to and so should
	// only be accessed using the concurrent-safe BlockNodeStatus method on
	// Index once the node has been added to the global Index.
	Status Status
}

// NewNode returns a new block node for the given block header and Parents, and the
// anticone of its selected parent (parent with highest blue score).
// selectedParentAnticone is used to update reachability data we store for future reachability queries.
// This function is NOT safe for concurrent access.
func NewNode(blockHeader *appmessage.BlockHeader, parents Set, timestamp int64) (node *Node) {
	node = &Node{
		Parents:            parents,
		Children:           make(Set),
		BlueScore:          math.MaxUint64, // Initialized to the max value to avoid collisions with the genesis block
		Timestamp:          timestamp,
		BluesAnticoneSizes: make(map[*Node]dagconfig.KType),
	}

	// blockHeader is nil only for the virtual block
	if blockHeader != nil {
		node.Hash = blockHeader.BlockHash()
		node.Version = blockHeader.Version
		node.Bits = blockHeader.Bits
		node.Nonce = blockHeader.Nonce
		node.Timestamp = blockHeader.Timestamp.UnixMilliseconds()
		node.HashMerkleRoot = blockHeader.HashMerkleRoot
		node.AcceptedIDMerkleRoot = blockHeader.AcceptedIDMerkleRoot
		node.UTXOCommitment = blockHeader.UTXOCommitment
	} else {
		node.Hash = &daghash.ZeroHash
	}

	if len(parents) == 0 {
		// The genesis block is defined to have a BlueScore of 0
		node.BlueScore = 0
		return node
	}

	return node
}

// UpdateParentsChildren updates the node's Parents to point to new node
func (node *Node) UpdateParentsChildren() {
	for parent := range node.Parents {
		parent.Children.Add(node)
	}
}

func (node *Node) Less(other *Node) bool {
	if node.BlueScore == other.BlueScore {
		return daghash.Less(node.Hash, other.Hash)
	}

	return node.BlueScore < other.BlueScore
}

// Header constructs a block header from the node and returns it.
//
// This function is safe for concurrent access.
func (node *Node) Header() *appmessage.BlockHeader {
	// No lock is needed because all accessed fields are immutable.
	return &appmessage.BlockHeader{
		Version:              node.Version,
		ParentHashes:         node.ParentHashes(),
		HashMerkleRoot:       node.HashMerkleRoot,
		AcceptedIDMerkleRoot: node.AcceptedIDMerkleRoot,
		UTXOCommitment:       node.UTXOCommitment,
		Timestamp:            node.Time(),
		Bits:                 node.Bits,
		Nonce:                node.Nonce,
	}
}

// SelectedAncestor returns the ancestor block node at the provided blue score by following
// the selected-Parents chain backwards from this node. The returned block will be nil when a
// blue score is requested that is higher than the blue score of the passed node.
//
// This function is safe for concurrent access.
func (node *Node) SelectedAncestor(BlueScore uint64) *Node {
	if BlueScore > node.BlueScore {
		return nil
	}

	n := node
	for n != nil && n.BlueScore > BlueScore {
		n = n.SelectedParent
	}

	return n
}

// RelativeAncestor returns the ancestor block node a relative 'distance' of
// blue blocks before this node. This is equivalent to calling Ancestor with
// the node's blue score minus provided distance.
//
// This function is safe for concurrent access.
func (node *Node) RelativeAncestor(distance uint64) *Node {
	return node.SelectedAncestor(node.BlueScore - distance)
}

func (node *Node) ParentHashes() []*daghash.Hash {
	return node.Parents.Hashes()
}

// IsGenesis returns if the current block is the genesis block
func (node *Node) IsGenesis() bool {
	return len(node.Parents) == 0
}

// String returns a string that contains the block Hash.
func (node Node) String() string {
	return node.Hash.String()
}

func (node *Node) Time() mstime.Time {
	return mstime.UnixMilliseconds(node.Timestamp)
}

func (node *Node) BlockAtDepth(depth uint64) *Node {
	if node.BlueScore <= depth { // to prevent overflow of requiredBlueScore
		depth = node.BlueScore
	}

	current := node
	requiredBlueScore := node.BlueScore - depth

	for current.BlueScore >= requiredBlueScore {
		if current.IsGenesis() {
			return current
		}
		current = current.SelectedParent
	}

	return current
}
