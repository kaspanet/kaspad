// Copyright (c) 2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package blockdag

import (
	"sync"

	"github.com/kaspanet/kaspad/domain/blocknode"
	"github.com/kaspanet/kaspad/domain/utxo"
	"github.com/kaspanet/kaspad/util/daghash"
)

// virtualBlock is a virtual block whose parents are the tips of the DAG.
type virtualBlock struct {
	mtx     sync.Mutex
	utxoSet *utxo.FullUTXOSet
	*blocknode.Node

	// SelectedParentChainSet is a block set that includes all the blocks
	// that belong to the chain of selected parents from the virtual block.
	SelectedParentChainSet blocknode.Set

	// SelectedParentChainSlice is an ordered slice that includes all the
	// blocks that belong the the chain of selected parents from the
	// virtual block.
	SelectedParentChainSlice []*blocknode.Node
}

// newVirtualBlock creates and returns a new VirtualBlock.
func newVirtualBlock(utxoSet *utxo.FullUTXOSet, parents blocknode.Set, timestamp int64) *virtualBlock {
	// The mutex is intentionally not held since this is a constructor.
	var virtual virtualBlock
	virtual.utxoSet = utxoSet
	virtual.SelectedParentChainSet = blocknode.NewSet()
	virtual.SelectedParentChainSlice = nil
	virtual.Node = blocknode.NewNode(nil, parents, timestamp)

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
func (v *virtualBlock) updateSelectedParentSet(oldSelectedParent *blocknode.Node) *selectedParentChainUpdates {
	var intersectionNode *blocknode.Node
	nodesToAdd := make([]*blocknode.Node, 0)
	for node := v.Node.SelectedParent; intersectionNode == nil && node != nil; node = node.SelectedParent {
		if v.SelectedParentChainSet.Contains(node) {
			intersectionNode = node
		} else {
			nodesToAdd = append(nodesToAdd, node)
		}
	}

	if intersectionNode == nil && oldSelectedParent != nil {
		panic("updateSelectedParentSet: Cannot find intersection node. The block Index may be corrupted.")
	}

	// Remove the nodes in the set from the oldSelectedParent down to the intersectionNode
	// Also, save the hashes of the removed blocks to removedChainBlockHashes
	removeCount := 0
	var removedChainBlockHashes []*daghash.Hash
	if intersectionNode != nil {
		for node := oldSelectedParent; !node.Hash.IsEqual(intersectionNode.Hash); node = node.SelectedParent {
			v.SelectedParentChainSet.Remove(node)
			removedChainBlockHashes = append(removedChainBlockHashes, node.Hash)
			removeCount++
		}
	}
	// Remove the last removeCount nodes from the slice
	v.SelectedParentChainSlice = v.SelectedParentChainSlice[:len(v.SelectedParentChainSlice)-removeCount]

	// Reverse nodesToAdd, since we collected them in reverse order
	for left, right := 0, len(nodesToAdd)-1; left < right; left, right = left+1, right-1 {
		nodesToAdd[left], nodesToAdd[right] = nodesToAdd[right], nodesToAdd[left]
	}
	// Add the nodes to the set and to the slice
	// Also, save the hashes of the added blocks to addedChainBlockHashes
	var addedChainBlockHashes []*daghash.Hash
	for _, node := range nodesToAdd {
		v.SelectedParentChainSet.Add(node)
		addedChainBlockHashes = append(addedChainBlockHashes, node.Hash)
	}
	v.SelectedParentChainSlice = append(v.SelectedParentChainSlice, nodesToAdd...)

	return &selectedParentChainUpdates{
		removedChainBlockHashes: removedChainBlockHashes,
		addedChainBlockHashes:   addedChainBlockHashes,
	}
}
