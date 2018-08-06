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
	mtx   sync.Mutex
	nodes []*blockNode
	blockNode
}

// newVirtualBlock creates and returns a new virtualBlock.
func newVirtualBlock(tip *blockNode) *virtualBlock {
	// The mutex is intentionally not held since this is a constructor.
	var c virtualBlock
	c.setTip(tip)
	return &c
}

// tip returns the current tip block node for the chain view.  It will return
// nil if there is no tip.  This only differs from the exported version in that
// it is up to the caller to ensure the lock is held.
//
// This function MUST be called with the view mutex locked (for reads).
func (c *virtualBlock) tip() *blockNode {
	if len(c.nodes) == 0 {
		return nil
	}

	return c.nodes[len(c.nodes)-1]
}

// Tips returns the current tip block nodes for the chain view.  It will return
// an empty slice if there is no tip.
//
// This function is safe for concurrent access.
func (c *virtualBlock) Tips() blockSet {
	c.mtx.Lock()
	tip := c.tip()
	c.mtx.Unlock()

	if tip == nil { // TODO: (Stas) This is wrong. Modified only to satisfy compilation.
		return newSet()
	}

	return setFromSlice(tip) // TODO: (Stas) This is wrong. Modified only to satisfy compilation.
}

// SelecedTip returns the current selected tip block node for the chain view.
// It will return nil if there is no tip.
//
// This function is safe for concurrent access.
func (c *virtualBlock) SelectedTip() *blockNode {
	return c.Tips().first()
}

// setTip sets the chain view to use the provided block node as the current tip
// and ensures the view is consistent by populating it with the nodes obtained
// by walking backwards all the way to genesis block as necessary.  Further
// calls will only perform the minimum work needed, so switching between chain
// tips is efficient.  This only differs from the exported version in that it is
// up to the caller to ensure the lock is held.
//
// This function MUST be called with the view mutex locked (for writes).
func (c *virtualBlock) setTip(node *blockNode) {
	if node == nil {
		// Keep the backing array around for potential future use.
		c.nodes = c.nodes[:0]
		return
	}

	// Create or resize the slice that will hold the block nodes to the
	// provided tip height.  When creating the slice, it is created with
	// some additional capacity for the underlying array as append would do
	// in order to reduce overhead when extending the chain later.  As long
	// as the underlying array already has enough capacity, simply expand or
	// contract the slice accordingly.  The additional capacity is chosen
	// such that the array should only have to be extended about once a
	// week.
	needed := node.height + 1
	if int32(cap(c.nodes)) < needed {
		nodes := make([]*blockNode, needed, needed+approxNodesPerWeek)
		copy(nodes, c.nodes)
		c.nodes = nodes
	} else {
		prevLen := int32(len(c.nodes))
		c.nodes = c.nodes[0:needed]
		for i := prevLen; i < needed; i++ {
			c.nodes[i] = nil
		}
	}

	for node != nil && c.nodes[node.height] != node {
		c.nodes[node.height] = node
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
func (c *virtualBlock) SetTip(node *blockNode) {
	c.mtx.Lock()
	c.setTip(node)
	c.mtx.Unlock()
}

// nodeByHeight returns the block node at the specified height.  Nil will be
// returned if the height does not exist.  This only differs from the exported
// version in that it is up to the caller to ensure the lock is held.
//
// This function MUST be called with the view mutex locked (for reads).
func (c *virtualBlock) nodeByHeight(height int32) *blockNode {
	if height < 0 || height >= int32(len(c.nodes)) {
		return nil
	}

	return c.nodes[height]
}

// NodeByHeight returns the block node at the specified height.  Nil will be
// returned if the height does not exist.
//
// This function is safe for concurrent access.
func (c *virtualBlock) NodeByHeight(height int32) *blockNode {
	c.mtx.Lock()
	node := c.nodeByHeight(height)
	c.mtx.Unlock()
	return node
}

// Equals returns whether or not two chain views are the same.  Uninitialized
// views (tip set to nil) are considered equal.
//
// This function is safe for concurrent access.
func (c *virtualBlock) Equals(other *virtualBlock) bool {
	c.mtx.Lock()
	other.mtx.Lock()
	equals := len(c.nodes) == len(other.nodes) && c.tip() == other.tip()
	other.mtx.Unlock()
	c.mtx.Unlock()
	return equals
}

// contains returns whether or not the chain view contains the passed block
// node.  This only differs from the exported version in that it is up to the
// caller to ensure the lock is held.
//
// This function MUST be called with the view mutex locked (for reads).
func (c *virtualBlock) contains(node *blockNode) bool {
	return c.nodeByHeight(node.height) == node
}

// Contains returns whether or not the chain view contains the passed block
// node.
//
// This function is safe for concurrent access.
func (c *virtualBlock) Contains(node *blockNode) bool {
	c.mtx.Lock()
	contains := c.contains(node)
	c.mtx.Unlock()
	return contains
}

// next returns the successor to the provided node for the chain view.  It will
// return nil if there is no successor or the provided node is not part of the
// view.  This only differs from the exported version in that it is up to the
// caller to ensure the lock is held.
//
// See the comment on the exported function for more details.
//
// This function MUST be called with the view mutex locked (for reads).
func (c *virtualBlock) next(node *blockNode) *blockNode {
	if node == nil || !c.contains(node) {
		return nil
	}

	return c.nodeByHeight(node.height + 1)
}

// Next returns the successor to the provided node for the chain view.  It will
// return nil if there is no successfor or the provided node is not part of the
// view.
//
// For example, assume a block chain with a side chain as depicted below:
//   genesis -> 1 -> 2 -> 3 -> 4  -> 5 ->  6  -> 7  -> 8
//                         \-> 4a -> 5a -> 6a
//
// Further, assume the view is for the longer chain depicted above.  That is to
// say it consists of:
//   genesis -> 1 -> 2 -> 3 -> 4 -> 5 -> 6 -> 7 -> 8
//
// Invoking this function with block node 5 would return block node 6 while
// invoking it with block node 5a would return nil since that node is not part
// of the view.
//
// This function is safe for concurrent access.
func (c *virtualBlock) Next(node *blockNode) *blockNode {
	c.mtx.Lock()
	next := c.next(node)
	c.mtx.Unlock()
	return next
}

// blockLocator returns a block locator for the passed block node.  The passed
// node can be nil in which case the block locator for the current tip
// associated with the view will be returned.  This only differs from the
// exported version in that it is up to the caller to ensure the lock is held.
//
// See the exported BlockLocator function comments for more details.
//
// This function MUST be called with the view mutex locked (for reads).
func (c *virtualBlock) blockLocator(node *blockNode) BlockLocator {
	// Use the current tip if requested.
	if node == nil {
		node = c.tip()
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

		// When the node is in the current chain view, all of its
		// ancestors must be too, so use a much faster O(1) lookup in
		// that case.  Otherwise, fall back to walking backwards through
		// the nodes of the other chain to the correct ancestor.
		if c.contains(node) {
			node = c.nodes[height]
		} else {
			node = node.Ancestor(height)
		}

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
func (c *virtualBlock) BlockLocator(node *blockNode) BlockLocator {
	c.mtx.Lock()
	locator := c.blockLocator(node)
	c.mtx.Unlock()
	return locator
}
