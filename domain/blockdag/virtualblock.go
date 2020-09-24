// Copyright (c) 2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package blockdag

import (
	"sync"

	"github.com/kaspanet/kaspad/util/daghash"
)

// virtualBlock is a virtual block whose parents are the tips of the DAG.
type virtualBlock struct {
	mtx     sync.Mutex
	dag     *BlockDAG
	utxoSet *FullUTXOSet
	*blockNode

	// selectedParentChainSet is a block set that includes all the blocks
	// that belong to the chain of selected parents from the virtual block.
	selectedParentChainSet blockSet

	// selectedParentChainSlice is an ordered slice that includes all the
	// blocks that belong the the chain of selected parents from the
	// virtual block.
	selectedParentChainSlice []*blockNode
}

// newVirtualBlock creates and returns a new VirtualBlock.
func newVirtualBlock(dag *BlockDAG, parents blockSet) *virtualBlock {
	// The mutex is intentionally not held since this is a constructor.
	var virtual virtualBlock
	virtual.dag = dag
	virtual.utxoSet = NewFullUTXOSetFromContext(dag.databaseContext, dag.maxUTXOCacheSize)
	virtual.selectedParentChainSet = newBlockSet()
	virtual.selectedParentChainSlice = nil
	virtual.blockNode, _ = dag.newBlockNode(nil, parents)

	return &virtual
}

// updateSelectedParentSet updates the selectedParentSet to match the
// new selected parent of the virtual block.
// Every time the new selected parent is not a child of
// the old one, it updates the selected path by removing from
// the path blocks that are selected ancestors of the old selected
// parent and are not selected ancestors of the new one, and adding
// blocks that are selected ancestors of the new selected parent
// and aren't selected ancestors of the old one.
func (v *virtualBlock) updateSelectedParentSet(oldSelectedParent *blockNode) *selectedParentChainUpdates {
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
	// Also, save the hashes of the removed blocks to removedChainBlockHashes
	removeCount := 0
	var removedChainBlockHashes []*daghash.Hash
	if intersectionNode != nil {
		for node := oldSelectedParent; !node.hash.IsEqual(intersectionNode.hash); node = node.selectedParent {
			v.selectedParentChainSet.remove(node)
			removedChainBlockHashes = append(removedChainBlockHashes, node.hash)
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
	// Also, save the hashes of the added blocks to addedChainBlockHashes
	var addedChainBlockHashes []*daghash.Hash
	for _, node := range nodesToAdd {
		v.selectedParentChainSet.add(node)
		addedChainBlockHashes = append(addedChainBlockHashes, node.hash)
	}
	v.selectedParentChainSlice = append(v.selectedParentChainSlice, nodesToAdd...)

	return &selectedParentChainUpdates{
		removedChainBlockHashes: removedChainBlockHashes,
		addedChainBlockHashes:   addedChainBlockHashes,
	}
}
