// Copyright (c) 2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package blockdag

import (
	"sync"
)

// virtualBlock is a virtual block whose parents are the tips of the DAG.
type virtualBlock struct {
	mtx      sync.Mutex
	phantomK uint32
	utxoSet  *FullUTXOSet
	blockNode

	// selectedParentChainSet is a block set that includes all the blocks
	// that belong to the chain of selected parents from the virtual block.
	selectedParentChainSet blockSet

	// selectedParentChainSlice is an ordered slice that includes all the
	// blocks that belong the the chain of selected parents from the
	// virtual block.
	selectedParentChainSlice []*blockNode
}

// newVirtualBlock creates and returns a new VirtualBlock.
func newVirtualBlock(tips blockSet, phantomK uint32) *virtualBlock {
	// The mutex is intentionally not held since this is a constructor.
	var virtual virtualBlock
	virtual.phantomK = phantomK
	virtual.utxoSet = NewFullUTXOSet()
	virtual.selectedParentChainSet = newSet()
	virtual.selectedParentChainSlice = nil
	virtual.setTips(tips)

	return &virtual
}

// clone creates and returns a clone of the virtual block.
func (v *virtualBlock) clone() *virtualBlock {
	return &virtualBlock{
		phantomK:               v.phantomK,
		utxoSet:                v.utxoSet,
		blockNode:              v.blockNode,
		selectedParentChainSet: v.selectedParentChainSet,
	}
}

// setTips replaces the tips of the virtual block with the blocks in the
// given blockSet. This only differs from the exported version in that it
// is up to the caller to ensure the lock is held.
//
// This function MUST be called with the view mutex locked (for writes).
func (v *virtualBlock) setTips(tips blockSet) {
	oldSelectedParent := v.selectedParent
	v.blockNode = *newBlockNode(nil, tips, v.phantomK)
	v.updateSelectedParentSet(oldSelectedParent)
}

// updateSelectedParentSet updates the selectedParentSet to match the
// new selected parent of the virtual block.
// Every time the new selected parent is not a child of
// the old one, it updates the selected path by removing from
// the path blocks that are selected ancestors of the old selected
// parent and are not selected ancestors of the new one, and adding
// blocks that are selected ancestors of the new selected parent
// and aren't selected ancestors of the old one.
func (v *virtualBlock) updateSelectedParentSet(oldSelectedParent *blockNode) {
	var intersectionNode *blockNode
	nodesToAdd := make([]*blockNode, 0)
	for node := v.blockNode.selectedParent; intersectionNode == nil && node != nil; node = node.selectedParent {
		if v.selectedParentChainSet.contains(node) {
			intersectionNode = node
		} else {
			nodesToAdd = append(nodesToAdd, node)
		}
	}

	if intersectionNode == nil && oldSelectedParent != nil {
		panic("updateSelectedParentSet: Cannot find intersection node. The block index may be corrupted.")
	}

	// Remove the nodes in the set from the oldSelectedParent down to the intersectionNode
	removeCount := 0
	if intersectionNode != nil {
		for node := oldSelectedParent; !node.hash.IsEqual(intersectionNode.hash); node = node.selectedParent {
			v.selectedParentChainSet.remove(node)
			removeCount++
		}
	}
	// Remove the last removeCount nodes from the slice
	v.selectedParentChainSlice = v.selectedParentChainSlice[:len(v.selectedParentChainSlice)-removeCount]

	// Reverse nodesToAdd, since we collected them in reverse order
	for left, right := 0, len(nodesToAdd)-1; left < right; left, right = left+1, right-1 {
		nodesToAdd[left], nodesToAdd[right] = nodesToAdd[right], nodesToAdd[left]
	}
	// Add the nodes to the set and to the slice
	for _, node := range nodesToAdd {
		v.selectedParentChainSet.add(node)
	}
	v.selectedParentChainSlice = append(v.selectedParentChainSlice, nodesToAdd...)
}

// SetTips replaces the tips of the virtual block with the blocks in the
// given blockSet.
//
// This function is safe for concurrent access.
func (v *virtualBlock) SetTips(tips blockSet) {
	v.mtx.Lock()
	v.setTips(tips)
	v.mtx.Unlock()
}

// addTip adds the given tip to the set of tips in the virtual block.
// All former tips that happen to be the given tips parents are removed
// from the set. This only differs from the exported version in that it
// is up to the caller to ensure the lock is held.
//
// This function MUST be called with the view mutex locked (for writes).
func (v *virtualBlock) addTip(newTip *blockNode) {
	updatedTips := v.tips().clone()
	for _, parent := range newTip.parents {
		updatedTips.remove(parent)
	}

	updatedTips.add(newTip)
	v.setTips(updatedTips)
}

// AddTip adds the given tip to the set of tips in the virtual block.
// All former tips that happen to be the given tip's parents are removed
// from the set.
//
// This function is safe for concurrent access.
func (v *virtualBlock) AddTip(newTip *blockNode) {
	v.mtx.Lock()
	v.addTip(newTip)
	v.mtx.Unlock()
}

// tips returns the current tip block nodes for the DAG.  It will return
// an empty blockSet if there is no tip.
//
// This function is safe for concurrent access.
func (v *virtualBlock) tips() blockSet {
	return v.parents
}
