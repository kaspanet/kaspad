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
// In addition to the general algorithm referenced above, this function will
// return the block locator for the selected tip if the passed hash is not currently
// known.
//
// This function is safe for concurrent access.
func (dag *BlockDAG) BlockLocatorFromHashes(startHash, stopHash *daghash.Hash) (BlockLocator, error) {
	dag.dagLock.RLock()
	defer dag.dagLock.RUnlock()
	highNode := dag.index.LookupNode(startHash)
	var lowNode *blockNode
	if !stopHash.IsEqual(&daghash.ZeroHash) {
		lowNode = dag.index.LookupNode(stopHash)
	}
	return dag.blockLocator(highNode, lowNode)
}

// blockLocator returns a block locator for the passed high and low nodes.
// The default value for the high node is the selected tip, and the default
// values of the low node is the genesis block.
//
// See the BlockLocator type comments for more details.
//
// This function MUST be called with the DAG state lock held (for reads).
func (dag *BlockDAG) blockLocator(highNode, lowNode *blockNode) (BlockLocator, error) {
	// Use the selected tip if requested.
	if highNode == nil {
		highNode = dag.virtual.selectedParent
	}

	if lowNode == nil {
		lowNode = dag.genesis
	}

	// We use the selected parent of the high node, so the
	// block locator won't contain the high node.
	highNode = highNode.selectedParent

	node := highNode
	step := uint64(1)
	locator := make(BlockLocator, 0)
	for node != nil {
		locator = append(locator, node.hash)

		// Nothing more to add once the stop node has been added.
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
func (dag *BlockDAG) FindNextLocatorBoundaries(locator BlockLocator) (highHash, stopHash *daghash.Hash) {
	// Find the most recent locator block hash in the DAG. In the case none of
	// the hashes in the locator are in the DAG, fall back to the genesis block.
	lowNode := dag.genesis
	nextBlockLocatorIndex := int64(len(locator) - 1)
	for i, hash := range locator {
		node := dag.index.LookupNode(hash)
		if node != nil {
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
