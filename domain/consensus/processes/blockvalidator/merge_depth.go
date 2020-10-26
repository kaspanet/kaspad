package blockvalidator

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
	"github.com/pkg/errors"
)

func (v *blockValidator) checkBoundedMergeDepth(ghostdagData *model.BlockGHOSTDAGData) error {
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
			return errors.Wrapf(ruleerrors.ErrViolatingBoundedMergeDepth, "block is violating bounded merge depth")
		}
	}

	return nil
}

func (v *blockValidator) nonBoundedMergeDepthViolatingBlues(ghostdagData *model.BlockGHOSTDAGData) ([]*externalapi.DomainHash, error) {
	nonBoundedMergeDepthViolatingBlues := make([]*externalapi.DomainHash, 0, len(ghostdagData.MergeSetBlues))

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

func (v *blockValidator) hasFinalityPointInOthersSelectedChain(ghostdagData *model.BlockGHOSTDAGData, other *externalapi.DomainHash) (bool, error) {
	finalityPoint, err := v.finalityPoint(ghostdagData)
	if err != nil {
		return false, err
	}

	return v.dagTopologyManager.IsInSelectedParentChainOf(finalityPoint, other)
}

func (v *blockValidator) finalityPoint(ghostdagData *model.BlockGHOSTDAGData) (*externalapi.DomainHash, error) {
	return v.dagTraversalManager.HighestChainBlockBelowBlueScore(ghostdagData.SelectedParent, ghostdagData.BlueScore-v.finalityDepth)
}
