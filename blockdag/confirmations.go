package blockdag

import (
	"github.com/kaspanet/kaspad/domainmessage"
	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/pkg/errors"
)

// BlockConfirmationsByHash returns the confirmations number for a block with the
// given hash. See blockConfirmations for further details.
//
// This function is safe for concurrent access
func (dag *BlockDAG) BlockConfirmationsByHash(hash *daghash.Hash) (uint64, error) {
	dag.dagLock.RLock()
	defer dag.dagLock.RUnlock()

	return dag.BlockConfirmationsByHashNoLock(hash)
}

// BlockConfirmationsByHashNoLock is lock free version of BlockConfirmationsByHash
//
// This function is unsafe for concurrent access.
func (dag *BlockDAG) BlockConfirmationsByHashNoLock(hash *daghash.Hash) (uint64, error) {
	if hash.IsEqual(&daghash.ZeroHash) {
		return 0, nil
	}

	node, ok := dag.index.LookupNode(hash)
	if !ok {
		return 0, errors.Errorf("block %s is unknown", hash)
	}

	return dag.blockConfirmations(node)
}

// UTXOConfirmations returns the confirmations for the given outpoint, if it exists
// in the DAG's UTXO set.
//
// This function is safe for concurrent access.
func (dag *BlockDAG) UTXOConfirmations(outpoint *domainmessage.Outpoint) (uint64, bool) {
	dag.dagLock.RLock()
	defer dag.dagLock.RUnlock()

	utxoEntry, ok := dag.GetUTXOEntry(*outpoint)
	if !ok {
		return 0, false
	}
	confirmations := dag.SelectedTipBlueScore() - utxoEntry.BlockBlueScore() + 1

	return confirmations, true
}

// blockConfirmations returns the current confirmations number of the given node
// The confirmations number is defined as follows:
// * If the node is in the selected tip red set	-> 0
// * If the node is the selected tip			-> 1
// * Otherwise									-> selectedTip.blueScore - acceptingBlock.blueScore + 2
func (dag *BlockDAG) blockConfirmations(node *blockNode) (uint64, error) {
	acceptingBlock, err := dag.acceptingBlock(node)
	if err != nil {
		return 0, err
	}

	// if acceptingBlock is nil, the node is red
	if acceptingBlock == nil {
		return 0, nil
	}

	return dag.selectedTip().blueScore - acceptingBlock.blueScore + 1, nil
}

// acceptingBlock finds the node in the selected-parent chain that had accepted
// the given node
func (dag *BlockDAG) acceptingBlock(node *blockNode) (*blockNode, error) {
	// Return an error if the node is the virtual block
	if node == &dag.virtual.blockNode {
		return nil, errors.New("cannot get acceptingBlock for virtual")
	}

	// If the node is a chain-block itself, the accepting block is its chain-child
	isNodeInSelectedParentChain, err := dag.IsInSelectedParentChain(node.hash)
	if err != nil {
		return nil, err
	}
	if isNodeInSelectedParentChain {
		if len(node.children) == 0 {
			// If the node is the selected tip, it doesn't have an accepting block
			return nil, nil
		}
		for child := range node.children {
			isChildInSelectedParentChain, err := dag.IsInSelectedParentChain(child.hash)
			if err != nil {
				return nil, err
			}
			if isChildInSelectedParentChain {
				return child, nil
			}
		}
		return nil, errors.Errorf("chain block %s does not have a chain child", node.hash)
	}

	// Find the only chain block that may contain the node in its blues
	candidateAcceptingBlock := dag.oldestChainBlockWithBlueScoreGreaterThan(node.blueScore)

	// if no candidate is found, it means that the node has same or more
	// blue score than the selected tip and is found in its anticone, so
	// it doesn't have an accepting block
	if candidateAcceptingBlock == nil {
		return nil, nil
	}

	// candidateAcceptingBlock is the accepting block only if it actually contains
	// the node in its blues
	for _, blue := range candidateAcceptingBlock.blues {
		if blue == node {
			return candidateAcceptingBlock, nil
		}
	}

	// Otherwise, the node is red or in the selected tip anticone, and
	// doesn't have an accepting block
	return nil, nil
}

// oldestChainBlockWithBlueScoreGreaterThan finds the oldest chain block with a blue score
// greater than blueScore. If no such block exists, this method returns nil
func (dag *BlockDAG) oldestChainBlockWithBlueScoreGreaterThan(blueScore uint64) *blockNode {
	chainBlockIndex, ok := util.SearchSlice(len(dag.virtual.selectedParentChainSlice), func(i int) bool {
		selectedPathNode := dag.virtual.selectedParentChainSlice[i]
		return selectedPathNode.blueScore > blueScore
	})
	if !ok {
		return nil
	}
	return dag.virtual.selectedParentChainSlice[chainBlockIndex]
}
