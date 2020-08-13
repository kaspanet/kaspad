package blockdag

import (
	"fmt"
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/pkg/errors"
)

// SelectedParentChain returns the selected parent chain starting from blockHash (exclusive)
// up to the virtual (exclusive). If blockHash is nil then the genesis block is used. If
// blockHash is not within the select parent chain, go down its own selected parent chain,
// while collecting each block hash in removedChainHashes, until reaching a block within
// the main selected parent chain.
//
// This method MUST be called with the DAG lock held
func (dag *BlockDAG) SelectedParentChain(blockHash *daghash.Hash) ([]*daghash.Hash, []*daghash.Hash, error) {
	if blockHash == nil {
		blockHash = dag.genesis.hash
	}
	if !dag.IsInDAG(blockHash) {
		return nil, nil, errors.Errorf("blockHash %s does not exist in the DAG", blockHash)
	}

	// If blockHash is not in the selected parent chain, go down its selected parent chain
	// until we find a block that is in the main selected parent chain.
	var removedChainHashes []*daghash.Hash
	isBlockInSelectedParentChain, err := dag.IsInSelectedParentChain(blockHash)
	if err != nil {
		return nil, nil, err
	}
	for !isBlockInSelectedParentChain {
		removedChainHashes = append(removedChainHashes, blockHash)

		node, ok := dag.index.LookupNode(blockHash)
		if !ok {
			return nil, nil, errors.Errorf("block %s does not exist in the DAG", blockHash)
		}
		blockHash = node.selectedParent.hash

		isBlockInSelectedParentChain, err = dag.IsInSelectedParentChain(blockHash)
		if err != nil {
			return nil, nil, err
		}
	}

	// Find the index of the blockHash in the selectedParentChainSlice
	blockHashIndex := len(dag.virtual.selectedParentChainSlice) - 1
	for blockHashIndex >= 0 {
		node := dag.virtual.selectedParentChainSlice[blockHashIndex]
		if node.hash.IsEqual(blockHash) {
			break
		}
		blockHashIndex--
	}

	// Copy all the addedChainHashes starting from blockHashIndex (exclusive)
	addedChainHashes := make([]*daghash.Hash, len(dag.virtual.selectedParentChainSlice)-blockHashIndex-1)
	for i, node := range dag.virtual.selectedParentChainSlice[blockHashIndex+1:] {
		addedChainHashes[i] = node.hash
	}

	return removedChainHashes, addedChainHashes, nil
}

// IsInSelectedParentChain returns whether or not a block hash is found in the selected
// parent chain. Note that this method returns an error if the given blockHash does not
// exist within the block index.
//
// This method MUST be called with the DAG lock held
func (dag *BlockDAG) IsInSelectedParentChain(blockHash *daghash.Hash) (bool, error) {
	blockNode, ok := dag.index.LookupNode(blockHash)
	if !ok {
		str := fmt.Sprintf("block %s is not in the DAG", blockHash)
		return false, ErrNotInDAG(str)
	}
	return dag.virtual.selectedParentChainSet.contains(blockNode), nil
}

// isInSelectedParentChainOf returns whether `node` is in the selected parent chain of `other`.
func (dag *BlockDAG) isInSelectedParentChainOf(node *blockNode, other *blockNode) (bool, error) {
	// By definition, a node is not in the selected parent chain of itself.
	if node == other {
		return false, nil
	}

	return dag.reachabilityTree.isReachabilityTreeAncestorOf(node, other)
}
