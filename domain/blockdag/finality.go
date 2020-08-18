package blockdag

import (
	"fmt"

	"github.com/kaspanet/kaspad/util/daghash"
)

// LastFinalityPointHash returns the hash of the last finality point
func (dag *BlockDAG) LastFinalityPointHash() *daghash.Hash {
	if dag.lastFinalityPoint == nil {
		return nil
	}
	return dag.lastFinalityPoint.hash
}

// FinalityInterval is the interval that determines the finality window of the DAG.
func (dag *BlockDAG) FinalityInterval() uint64 {
	return uint64(dag.Params.FinalityDuration / dag.Params.TargetTimePerBlock)
}

// checkFinalityViolation checks the new block does not violate the finality rules
// specifically - the new block selectedParent chain should contain the old finality point.
func (dag *BlockDAG) checkFinalityViolation(newNode *blockNode) error {
	// the genesis block can not violate finality rules
	if newNode.isGenesis() {
		return nil
	}

	// Because newNode doesn't have reachability data we
	// need to check if the last finality point is in the
	// selected parent chain of newNode.selectedParent, so
	// we explicitly check if newNode.selectedParent is
	// the finality point.
	if dag.lastFinalityPoint == newNode.selectedParent {
		return nil
	}

	isInSelectedChain, err := dag.isInSelectedParentChainOf(dag.lastFinalityPoint, newNode.selectedParent)
	if err != nil {
		return err
	}

	if !isInSelectedChain {
		return ruleError(ErrFinality, "the last finality point is not in the selected parent chain of this block")
	}
	return nil
}

// updateFinalityPoint updates the dag's last finality point if necessary.
func (dag *BlockDAG) updateFinalityPoint() {
	selectedTip := dag.selectedTip()
	// if the selected tip is the genesis block - it should be the new finality point
	if selectedTip.isGenesis() {
		dag.lastFinalityPoint = selectedTip
		return
	}
	// We are looking for a new finality point only if the new block's finality score is higher
	// by 2 than the existing finality point's
	if selectedTip.finalityScore() < dag.lastFinalityPoint.finalityScore()+2 {
		return
	}

	var currentNode *blockNode
	for currentNode = selectedTip.selectedParent; ; currentNode = currentNode.selectedParent {
		// We look for the first node in the selected parent chain that has a higher finality score than the last finality point.
		if currentNode.selectedParent.finalityScore() == dag.lastFinalityPoint.finalityScore() {
			break
		}
	}
	dag.lastFinalityPoint = currentNode
	spawn("dag.finalizeNodesBelowFinalityPoint", func() {
		dag.finalizeNodesBelowFinalityPoint(true)
	})
}

func (dag *BlockDAG) finalizeNodesBelowFinalityPoint(deleteDiffData bool) {
	queue := make([]*blockNode, 0, len(dag.lastFinalityPoint.parents))
	for parent := range dag.lastFinalityPoint.parents {
		queue = append(queue, parent)
	}
	var nodesToDelete []*blockNode
	if deleteDiffData {
		nodesToDelete = make([]*blockNode, 0, dag.FinalityInterval())
	}
	for len(queue) > 0 {
		var current *blockNode
		current, queue = queue[0], queue[1:]
		if !current.isFinalized {
			current.isFinalized = true
			if deleteDiffData {
				nodesToDelete = append(nodesToDelete, current)
			}
			for parent := range current.parents {
				queue = append(queue, parent)
			}
		}
	}
	if deleteDiffData {
		err := dag.utxoDiffStore.removeBlocksDiffData(dag.databaseContext, nodesToDelete)
		if err != nil {
			panic(fmt.Sprintf("Error removing diff data from utxoDiffStore: %s", err))
		}
	}
}

// IsKnownFinalizedBlock returns whether the block is below the finality point.
// IsKnownFinalizedBlock might be false-negative because node finality status is
// updated in a separate goroutine. To get a definite answer if a block
// is finalized or not, use dag.checkFinalityViolation.
func (dag *BlockDAG) IsKnownFinalizedBlock(blockHash *daghash.Hash) bool {
	node, ok := dag.index.LookupNode(blockHash)
	return ok && node.isFinalized
}
