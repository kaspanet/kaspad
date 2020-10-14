package validator

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
)

func (bv *Validator) checkBoundedMergeDepth(ghostdagData *model.BlockGHOSTDAGData) error {
	nonBoundedMergeDepthViolatingBlues, err := bv.nonBoundedMergeDepthViolatingBlues(ghostdagData)
	if err != nil {
		return err
	}

	finalityPoint := bv.finalityPoint(ghostdagData)
	for _, red := range ghostdagData.MergeSetReds {
		doesRedHaveFinalityPointInPast := bv.dagTopologyManager.IsAncestorOf(finalityPoint, red)
		if doesRedHaveFinalityPointInPast {
			continue
		}

		isRedInPastOfAnyNonFinalityViolatingBlue := bv.dagTopologyManager.IsAncestorOfAny(red,
			nonBoundedMergeDepthViolatingBlues)
		if !isRedInPastOfAnyNonFinalityViolatingBlue {
			return ruleerrors.Errorf(ruleerrors.ErrViolatingBoundedMergeDepth, "block is violating bounded merge depth")
		}
	}

	return nil
}

func (bv *Validator) nonBoundedMergeDepthViolatingBlues(ghostdagData *model.BlockGHOSTDAGData) ([]*model.DomainHash, error) {
	nonBoundedMergeDepthViolatingBlues := make([]*model.DomainHash, 0, len(ghostdagData.MergeSetBlues))

	for _, blue := range ghostdagData.MergeSetBlues {
		notViolatingFinality := bv.hasFinalityPointInOthersSelectedChain(ghostdagData, blue)
		if notViolatingFinality {
			nonBoundedMergeDepthViolatingBlues = append(nonBoundedMergeDepthViolatingBlues, blue)
		}
	}

	return nonBoundedMergeDepthViolatingBlues, nil
}

func (bv *Validator) hasFinalityPointInOthersSelectedChain(ghostdagData *model.BlockGHOSTDAGData, other *model.DomainHash) bool {
	finalityPoint := bv.finalityPoint(ghostdagData)
	return bv.dagTopologyManager.IsInSelectedParentChainOf(finalityPoint, other)
}

func (bv *Validator) finalityPoint(ghostdagData *model.BlockGHOSTDAGData) *model.DomainHash {
	panic("unimplemented")
}
