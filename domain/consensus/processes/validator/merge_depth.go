package blockvalidator

import "github.com/kaspanet/kaspad/domain/consensus/model"

func (bv *BlockValidator) checkBoundedMergeDepth(ghostdagData *model.BlockGHOSTDAGData) error {
	nonBoundedMergeDepthViolatingBlues, err := bv.nonBoundedMergeDepthViolatingBlues(ghostdagData)
	if err != nil {
		return err
	}

	finalityPoint := node.finalityPoint()
	for _, red := range node.reds {
		doesRedHaveFinalityPointInPast, err := node.dag.isInPast(finalityPoint, red)
		if err != nil {
			return err
		}

		isRedInPastOfAnyNonFinalityViolatingBlue, err := node.dag.isInPastOfAny(red, nonBoundedMergeDepthViolatingBlues)
		if err != nil {
			return err
		}

		if !doesRedHaveFinalityPointInPast && !isRedInPastOfAnyNonFinalityViolatingBlue {
			return ruleError(ErrViolatingBoundedMergeDepth, "block is violating bounded merge depth")
		}
	}

	return nil
}

func (bv *BlockValidator) nonBoundedMergeDepthViolatingBlues(ghostdagData *model.BlockGHOSTDAGData) ([]*model.DomainHash, error) {
	nonBoundedMergeDepthViolatingBlues := make([]*model.DomainHash, 0, len(ghostdagData.MergeSetBlues))

	for _, blueNode := range ghostdagData.MergeSetBlues {
		notViolatingFinality, err := node.hasFinalityPointInOthersSelectedChain(blueNode)
		if err != nil {
			return nil, err
		}
		if notViolatingFinality {
			nonBoundedMergeDepthViolatingBlues.add(blueNode)
		}
	}

	return nonBoundedMergeDepthViolatingBlues, nil
}

func (bv *BlockValidator) hasFinalityPointInOthersSelectedChain(ghostdagData *model.BlockGHOSTDAGData, other *blockNode) (bool, error) {
	finalityPoint := node.finalityPoint()
	return node.dag.isInSelectedParentChainOf(finalityPoint, other)
}
