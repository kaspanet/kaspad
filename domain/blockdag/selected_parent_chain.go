package blockdag

import (
	"fmt"

	"github.com/kaspanet/kaspad/domain/blocknode"
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
		blockHash = dag.genesis.Hash
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

		node, ok := dag.Index.LookupNode(blockHash)
		if !ok {
			return nil, nil, errors.Errorf("block %s does not exist in the DAG", blockHash)
		}
		blockHash = node.SelectedParent.Hash

		isBlockInSelectedParentChain, err = dag.IsInSelectedParentChain(blockHash)
		if err != nil {
			return nil, nil, err
		}
	}

	// Find the Index of the blockHash in the SelectedParentChainSlice
	blockHashIndex := len(dag.virtual.SelectedParentChainSlice) - 1
	for blockHashIndex >= 0 {
		node := dag.virtual.SelectedParentChainSlice[blockHashIndex]
		if node.Hash.IsEqual(blockHash) {
			break
		}
		blockHashIndex--
	}

	// Copy all the addedChainHashes starting from blockHashIndex (exclusive)
	addedChainHashes := make([]*daghash.Hash, len(dag.virtual.SelectedParentChainSlice)-blockHashIndex-1)
	for i, node := range dag.virtual.SelectedParentChainSlice[blockHashIndex+1:] {
		addedChainHashes[i] = node.Hash
	}

	return removedChainHashes, addedChainHashes, nil
}

// IsInSelectedParentChain returns whether or not a block hash is found in the selected
// parent chain. Note that this method returns an error if the given blockHash does not
// exist within the block Index.
//
// This method MUST be called with the DAG lock held
func (dag *BlockDAG) IsInSelectedParentChain(blockHash *daghash.Hash) (bool, error) {
	blockNode, ok := dag.Index.LookupNode(blockHash)
	if !ok {
		str := fmt.Sprintf("block %s is not in the DAG", blockHash)
		return false, ErrNotInDAG(str)
	}
	return dag.virtual.SelectedParentChainSet.Contains(blockNode), nil
}

// isInSelectedParentChainOf returns whether `node` is in the selected parent chain of `other`.
//
// Note: this method will return true if node == other
func (dag *BlockDAG) isInSelectedParentChainOf(node *blocknode.Node, other *blocknode.Node) (bool, error) {
	return dag.reachabilityTree.isReachabilityTreeAncestorOf(node, other)
}

// isInSelectedParentChainOfAll returns true if `node` is in the selected parent chain of all `others`
func (dag *BlockDAG) isInSelectedParentChainOfAll(node *blocknode.Node, others blocknode.Set) (bool, error) {
	for other := range others {
		isInSelectedParentChain, err := dag.isInSelectedParentChainOf(node, other)
		if err != nil {
			return false, err
		}
		if !isInSelectedParentChain {
			return false, nil
		}
	}
	return true, nil
}
