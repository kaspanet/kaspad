// Copyright (c) 2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package blockdag

import (
	"sync"
)

// approxNodesPerWeek is an approximation of the number of new blocks there are
// in a week on average.
const approxNodesPerWeek = 6 * 24 * 7

// log2FloorMasks defines the masks to use when quickly calculating
// floor(log2(x)) in a constant log2(32) = 5 steps, where x is a uint32, using
// shifts.  They are derived from (2^(2^x) - 1) * (2^(2^x)), for x in 4..0.
var log2FloorMasks = []uint32{0xffff0000, 0xff00, 0xf0, 0xc, 0x2}

// fastLog2Floor calculates and returns floor(log2(x)) in a constant 5 steps.
func fastLog2Floor(n uint32) uint8 {
	rv := uint8(0)
	exponent := uint8(16)
	for i := 0; i < 5; i++ {
		if n&log2FloorMasks[i] != 0 {
			rv += exponent
			n >>= exponent
		}
		exponent >>= 1
	}
	return rv
}

// virtualBlock is a virtual block whose parents are the tip of the DAG.
type virtualBlock struct {
	mtx      sync.Mutex
	nodes    []*blockNode
	phantomK uint32
	blockNode
}

// newVirtualBlock creates and returns a new virtualBlock.
func newVirtualBlock(tips blockSet, phantomK uint32) *virtualBlock {
	// The mutex is intentionally not held since this is a constructor.
	var virtual virtualBlock
	virtual.phantomK = phantomK
	virtual.setTip(tips.first())
	if tips != nil {
		virtual.setTips(tips)
	}

	return &virtual
}

// Tips returns the current tip block nodes for the chain view.  It will return
// an empty slice if there is no tip.
//
// This function is safe for concurrent access.
func (v *virtualBlock) Tips() blockSet {
	v.mtx.Lock()
	defer func() {
		v.mtx.Unlock()
	}()

	return v.parents
}

// SelecedTip returns the current selected tip block node for the chain view.
// It will return nil if there is no tip.
//
// This function is safe for concurrent access.
func (v *virtualBlock) SelectedTip() *blockNode {
	return v.Tips().first()
}

// setTip sets the chain view to use the provided block node as the current tip
// and ensures the view is consistent by populating it with the nodes obtained
// by walking backwards all the way to genesis block as necessary.  Further
// calls will only perform the minimum work needed, so switching between chain
// tips is efficient.  This only differs from the exported version in that it is
// up to the caller to ensure the lock is held.
//
// This function MUST be called with the view mutex locked (for writes).
func (v *virtualBlock) setTip(node *blockNode) {
	if node == nil {
		return
	}

	v.setTips(setFromSlice(node))

	// Create or resize the slice that will hold the block nodes to the
	// provided tip height.  When creating the slice, it is created with
	// some additional capacity for the underlying array as append would do
	// in order to reduce overhead when extending the chain later.  As long
	// as the underlying array already has enough capacity, simply expand or
	// contract the slice accordingly.  The additional capacity is chosen
	// such that the array should only have to be extended about once a
	// week.
	needed := node.height + 1
	if int32(cap(v.nodes)) < needed {
		nodes := make([]*blockNode, needed, needed+approxNodesPerWeek)
		copy(nodes, v.nodes)
		v.nodes = nodes
	} else {
		prevLen := int32(len(v.nodes))
		v.nodes = v.nodes[0:needed]
		for i := prevLen; i < needed; i++ {
			v.nodes[i] = nil
		}
	}

	for node != nil && v.nodes[node.height] != node {
		v.nodes[node.height] = node
		node = node.selectedParent
	}
}

// SetTip sets the chain view to use the provided block node as the current tip
// and ensures the view is consistent by populating it with the nodes obtained
// by walking backwards all the way to genesis block as necessary.  Further
// calls will only perform the minimum work needed, so switching between chain
// tips is efficient.
//
// This function is safe for concurrent access.
func (v *virtualBlock) SetTip(node *blockNode) {
	v.mtx.Lock()
	v.setTip(node)
	v.mtx.Unlock()
}

func (v *virtualBlock) setTips(tips blockSet) {
	v.blockNode = *newBlockNode(nil, tips, v.phantomK)
}

func (v *virtualBlock) SetTips(tips blockSet) {
	v.mtx.Lock()
	v.setTips(tips)
	v.mtx.Unlock()
}

// blockLocator returns a block locator for the passed block node.  The passed
// node can be nil in which case the block locator for the current tip
// associated with the view will be returned.  This only differs from the
// exported version in that it is up to the caller to ensure the lock is held.
//
// See the exported BlockLocator function comments for more details.
//
// This function MUST be called with the view mutex locked (for reads).
func (v *virtualBlock) blockLocator(node *blockNode) BlockLocator {
	// Use the current tip if requested.
	if node == nil {
		node = v.selectedParent
	}
	if node == nil {
		return nil
	}

	// Calculate the max number of entries that will ultimately be in the
	// block locator.  See the description of the algorithm for how these
	// numbers are derived.
	var maxEntries uint8
	if node.height <= 12 {
		maxEntries = uint8(node.height) + 1
	} else {
		// Requested hash itself + previous 10 entries + genesis block.
		// Then floor(log2(height-10)) entries for the skip portion.
		adjustedHeight := uint32(node.height) - 10
		maxEntries = 12 + fastLog2Floor(adjustedHeight)
	}
	locator := make(BlockLocator, 0, maxEntries)

	step := int32(1)
	for node != nil {
		locator = append(locator, &node.hash)

		// Nothing more to add once the genesis block has been added.
		if node.height == 0 {
			break
		}

		// Calculate height of previous node to include ensuring the
		// final node is the genesis block.
		height := node.height - step
		if height < 0 {
			height = 0
		}

		// walk backwards through the nodes to the correct ancestor.
		node = node.Ancestor(height)

		// Once 11 entries have been included, start doubling the
		// distance between included hashes.
		if len(locator) > 10 {
			step *= 2
		}
	}

	return locator
}

// BlockLocator returns a block locator for the passed block node.  The passed
// node can be nil in which case the block locator for the current tip
// associated with the view will be returned.
//
// See the BlockLocator type for details on the algorithm used to create a block
// locator.
//
// This function is safe for concurrent access.
func (v *virtualBlock) BlockLocator(node *blockNode) BlockLocator {
	v.mtx.Lock()
	locator := v.blockLocator(node)
	v.mtx.Unlock()
	return locator
}
