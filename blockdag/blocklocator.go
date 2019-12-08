package blockdag

import (
	"github.com/daglabs/kaspad/util"
	"github.com/daglabs/kaspad/util/daghash"
)

// BlockLocator is used to help locate a specific block.  The algorithm for
// building the block locator is to add block hashes in reverse order on the
// block's selected parent chain until the desired stop block is reached.
// In order to keep the list of locator hashes to a reasonable number of entries,
// the step between each entry is doubled each loop iteration to exponentially
// decrease the number of hashes as a function of the distance from the block
// being located.
//
// For example, assume a selected parent chain with IDs as depicted below, and the
// stop block is genesis:
// 	genesis -> 1 -> 2 -> ... -> 15 -> 16  -> 17  -> 18
//
// The block locator for block 17 would be the hashes of blocks:
//  [17 16 14 11 7 2 genesis]
type BlockLocator []*daghash.Hash

// BlockLocatorFromHashes returns a block locator from start and stop hash.
// See BlockLocator for details on the algorithm used to create a block locator.
//
// In addition to the general algorithm referenced above, this function will
// return the block locator for the selected tip if the passed hash is not currently
// known.
//
// This function is safe for concurrent access.
func (dag *BlockDAG) BlockLocatorFromHashes(startHash, stopHash *daghash.Hash) BlockLocator {
	dag.dagLock.RLock()
	defer dag.dagLock.RUnlock()
	startNode := dag.index.LookupNode(startHash)
	var stopNode *blockNode
	if !stopHash.IsEqual(&daghash.ZeroHash) {
		stopNode = dag.index.LookupNode(stopHash)
	}
	return dag.blockLocator(startNode, stopNode)
}

// LatestBlockLocator returns a block locator for the current tips of the DAG.
//
// This function is safe for concurrent access.
func (dag *BlockDAG) LatestBlockLocator() BlockLocator {
	dag.dagLock.RLock()
	defer dag.dagLock.RUnlock()
	return dag.blockLocator(nil, nil)
}

// blockLocator returns a block locator for the passed start and stop nodes.
// The default value for the start node is the selected tip, and the default
// values of the stop node is the genesis block.
//
// See the BlockLocator type comments for more details.
//
// This function MUST be called with the DAG state lock held (for reads).
func (dag *BlockDAG) blockLocator(startNode, stopNode *blockNode) BlockLocator {
	// Use the selected tip if requested.
	if startNode == nil {
		startNode = dag.virtual.selectedParent
	}

	if stopNode == nil {
		stopNode = dag.genesis
	}

	// We use the selected parent of the start node, so the
	// block locator won't contain the start node.
	startNode = startNode.selectedParent

	// If the start node or the stop node are not in the
	// virtual's selected parent chain, we replace them with their
	// closest selected parent that is part of the virtual's
	// selected parent chain.
	for !dag.IsInSelectedParentChain(stopNode.hash) {
		stopNode = stopNode.selectedParent
	}

	for !dag.IsInSelectedParentChain(startNode.hash) {
		startNode = startNode.selectedParent
	}

	// Calculate the max number of entries that will ultimately be in the
	// block locator.  See the description of the algorithm for how these
	// numbers are derived.

	// startNode.hash + stopNode.hash.
	// Then floor(log2(startNode.chainHeight-stopNode.chainHeight)) entries for the skip portion.
	maxEntries := 2 + util.FastLog2Floor(startNode.chainHeight-stopNode.chainHeight)
	locator := make(BlockLocator, 0, maxEntries)

	step := uint64(1)
	for node := startNode; node != nil; {
		locator = append(locator, node.hash)

		// Nothing more to add once the stop node has been added.
		if node.chainHeight == stopNode.chainHeight {
			break
		}

		// Calculate chainHeight of previous node to include ensuring the
		// final node is stopNode.
		nextChainHeight := node.chainHeight - step
		if nextChainHeight < stopNode.chainHeight {
			nextChainHeight = stopNode.chainHeight
		}

		// walk backwards through the nodes to the correct ancestor.
		node = node.SelectedAncestor(nextChainHeight)

		// Double the distance between included hashes.
		step *= 2
	}

	return locator
}

// FindNextLocatorBoundaries returns the lowest unknown block locator, hash
// and the highest known block locator hash. This is used to create the
// next block locator to find the highest shared known chain block with the
// sync peer.
//
// This function MUST be called with the DAG state lock held (for reads).
func (dag *BlockDAG) FindNextLocatorBoundaries(locator BlockLocator) (startHash, stopHash *daghash.Hash) {
	// Find the most recent locator block hash in the DAG.  In the case none of
	// the hashes in the locator are in the DAG, fall back to the genesis block.
	stopNode := dag.genesis
	nextBlockLocatorIndex := int64(len(locator) - 1)
	for i, hash := range locator {
		node := dag.index.LookupNode(hash)
		if node != nil {
			stopNode = node
			nextBlockLocatorIndex = int64(i) - 1
			break
		}
	}
	if nextBlockLocatorIndex < 0 {
		return nil, stopNode.hash
	}
	return locator[nextBlockLocatorIndex], stopNode.hash
}
