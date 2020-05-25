package blockdag

import (
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/pkg/errors"
)

// BlockLocator is used to help locate a specific block. The algorithm for
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

// BlockLocatorFromHashes returns a block locator from high and low hash.
// See BlockLocator for details on the algorithm used to create a block locator.
//
// This function is safe for concurrent access.
func (dag *BlockDAG) BlockLocatorFromHashes(highHash, lowHash *daghash.Hash) (BlockLocator, error) {
	dag.dagLock.RLock()
	defer dag.dagLock.RUnlock()

	highNode, ok := dag.index.LookupNode(highHash)
	if !ok {
		return nil, errors.Errorf("block %s is unknown", highHash)
	}

	lowNode, ok := dag.index.LookupNode(lowHash)
	if !ok {
		return nil, errors.Errorf("block %s is unknown", lowHash)
	}

	return dag.blockLocator(highNode, lowNode)
}

// blockLocator returns a block locator for the passed high and low nodes.
// See the BlockLocator type comments for more details.
//
// This function MUST be called with the DAG state lock held (for reads).
func (dag *BlockDAG) blockLocator(highNode, lowNode *blockNode) (BlockLocator, error) {
	// We use the selected parent of the high node, so the
	// block locator won't contain the high node.
	highNode = highNode.selectedParent

	node := highNode
	step := uint64(1)
	locator := make(BlockLocator, 0)
	for node != nil {
		locator = append(locator, node.hash)

		// Nothing more to add once the low node has been added.
		if node.blueScore <= lowNode.blueScore {
			if node != lowNode {
				return nil, errors.Errorf("highNode and lowNode are " +
					"not in the same selected parent chain.")
			}
			break
		}

		// Calculate blueScore of previous node to include ensuring the
		// final node is lowNode.
		nextBlueScore := node.blueScore - step
		if nextBlueScore < lowNode.blueScore {
			nextBlueScore = lowNode.blueScore
		}

		// walk backwards through the nodes to the correct ancestor.
		node = node.SelectedAncestor(nextBlueScore)

		// Double the distance between included hashes.
		step *= 2
	}

	return locator, nil
}

// FindNextLocatorBoundaries returns the lowest unknown block locator, hash
// and the highest known block locator hash. This is used to create the
// next block locator to find the highest shared known chain block with the
// sync peer.
//
// This function MUST be called with the DAG state lock held (for reads).
func (dag *BlockDAG) FindNextLocatorBoundaries(locator BlockLocator) (highHash, lowHash *daghash.Hash) {
	// Find the most recent locator block hash in the DAG. In the case none of
	// the hashes in the locator are in the DAG, fall back to the genesis block.
	lowNode := dag.genesis
	nextBlockLocatorIndex := int64(len(locator) - 1)
	for i, hash := range locator {
		node, ok := dag.index.LookupNode(hash)
		if ok {
			lowNode = node
			nextBlockLocatorIndex = int64(i) - 1
			break
		}
	}
	if nextBlockLocatorIndex < 0 {
		return nil, lowNode.hash
	}
	return locator[nextBlockLocatorIndex], lowNode.hash
}
