package validator

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
)

func (v *validator) checkBoundedMergeDepth(ghostdagData *model.BlockGHOSTDAGData) error {
	nonBoundedMergeDepthViolatingBlues, err := v.nonBoundedMergeDepthViolatingBlues(ghostdagData)
	if err != nil {
		return err
	}

	finalityPoint, err := v.finalityPoint(ghostdagData)
	if err != nil {
		return err
	}

	for _, red := range ghostdagData.MergeSetReds {
		doesRedHaveFinalityPointInPast, err := v.dagTopologyManager.IsAncestorOf(finalityPoint, red)
		if err != nil {
			return err
		}

		if doesRedHaveFinalityPointInPast {
			continue
		}

		isRedInPastOfAnyNonFinalityViolatingBlue, err := v.dagTopologyManager.IsAncestorOfAny(red,
			nonBoundedMergeDepthViolatingBlues)
		if err != nil {
			return err
		}

		if !isRedInPastOfAnyNonFinalityViolatingBlue {
			return ruleerrors.Errorf(ruleerrors.ErrViolatingBoundedMergeDepth, "block is violating bounded merge depth")
		}
	}

	return nil
}

func (v *validator) nonBoundedMergeDepthViolatingBlues(ghostdagData *model.BlockGHOSTDAGData) ([]*model.DomainHash, error) {
	nonBoundedMergeDepthViolatingBlues := make([]*model.DomainHash, 0, len(ghostdagData.MergeSetBlues))

	for _, blue := range ghostdagData.MergeSetBlues {
		notViolatingFinality, err := v.hasFinalityPointInOthersSelectedChain(ghostdagData, blue)
		if err != nil {
			return nil, err
		}

		if notViolatingFinality {
			nonBoundedMergeDepthViolatingBlues = append(nonBoundedMergeDepthViolatingBlues, blue)
		}
	}

	return nonBoundedMergeDepthViolatingBlues, nil
}

func (v *validator) hasFinalityPointInOthersSelectedChain(ghostdagData *model.BlockGHOSTDAGData, other *model.DomainHash) (bool, error) {
	finalityPoint, err := v.finalityPoint(ghostdagData)
	if err != nil {
		return false, err
	}

	return v.dagTopologyManager.IsInSelectedParentChainOf(finalityPoint, other)
}

func (v *validator) finalityPoint(ghostdagData *model.BlockGHOSTDAGData) (*model.DomainHash, error) {
	return v.dagTraversalManager.HighestChainBlockBelowBlueScore(ghostdagData.SelectedParent, ghostdagData.BlueScore-v.finalityDepth)
}
