// Copyright (c) 2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package blockdag

import (
	"sync"
)

// virtualBlock is a virtual block whose parents are the tip of the DAG.
type virtualBlock struct {
	mtx      sync.Mutex
	phantomK uint32
	blockNode
}

// newVirtualBlock creates and returns a new virtualBlock.
func newVirtualBlock(tips blockSet, phantomK uint32) *virtualBlock {
	// The mutex is intentionally not held since this is a constructor.
	var virtual virtualBlock
	virtual.phantomK = phantomK
	virtual.setTips(tips)

	return &virtual
}

// setTips replaces the tips of the virtual block with the blocks in the
// given blockSet. This only differs from the exported version in that it
// is up to the caller to ensure the lock is held.
//
// This function MUST be called with the view mutex locked (for writes).
func (v *virtualBlock) setTips(tips blockSet) {
	v.blockNode = *newBlockNode(nil, tips, v.phantomK)
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
	tips := newSet()
	tips.add(newTip)

	for tipHash, tip := range v.Tips() {
		isParent := false
		for parentHash := range newTip.parents {
			if tipHash == parentHash {
				isParent = true
				break
			}
		}
		if !isParent {
			tips.add(tip)
		}
	}

	v.setTips(tips)
}

// addTip adds the given tip to the set of tips in the virtual block.
// All former tips that happen to be the given tip's parents are removed
// from the set.
//
// This function is safe for concurrent access.
func (v *virtualBlock) AddTip(newTip *blockNode) {
	v.mtx.Lock()
	v.addTip(newTip)
	v.mtx.Unlock()
}

// Tips returns the current tip block nodes for the DAG.  It will return
// an empty blockSet if there is no tip.
//
// This function is safe for concurrent access.
func (v *virtualBlock) Tips() blockSet {
	return v.parents
}

// SelectedTip returns the current selected tip for the DAG.
// It will return nil if there is no tip.
//
// This function is safe for concurrent access.
func (v *virtualBlock) SelectedTip() *blockNode {
	return v.selectedParent
}
