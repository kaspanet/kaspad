package mergedepthmanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
	"github.com/pkg/errors"
)

type mergeDepthManager struct {
	finalityDepth uint64

	databaseContext     model.DBReader
	dagTopologyManager  model.DAGTopologyManager
	dagTraversalManager model.DAGTraversalManager

	ghostdagDataStore model.GHOSTDAGDataStore
}

func New(finalityDepth uint64,
	databaseContext model.DBReader,
	dagTopologyManager model.DAGTopologyManager,
	dagTraversalManager model.DAGTraversalManager,
	ghostdagDataStore model.GHOSTDAGDataStore) model.MergeDepthManager {

	return &mergeDepthManager{
		finalityDepth:       finalityDepth,
		databaseContext:     databaseContext,
		dagTopologyManager:  dagTopologyManager,
		dagTraversalManager: dagTraversalManager,
		ghostdagDataStore:   ghostdagDataStore,
	}

}

func (mdm *mergeDepthManager) CheckBoundedMergeDepth(blockHash *externalapi.DomainHash) error {
	nonBoundedMergeDepthViolatingBlues, err := mdm.NonBoundedMergeDepthViolatingBlues(blockHash)
	if err != nil {
		return err
	}

	ghostdagData, err := mdm.ghostdagDataStore.Get(mdm.databaseContext, blockHash)
	if err != nil {
		return err
	}

	finalityPoint, err := mdm.finalityPoint(ghostdagData)
	if err != nil {
		return err
	}

	for _, red := range ghostdagData.MergeSetReds {
		doesRedHaveFinalityPointInPast, err := mdm.dagTopologyManager.IsAncestorOf(finalityPoint, red)
		if err != nil {
			return err
		}

		if doesRedHaveFinalityPointInPast {
			continue
		}

		isRedInPastOfAnyNonFinalityViolatingBlue, err := mdm.dagTopologyManager.IsAncestorOfAny(red,
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

func (mdm mergeDepthManager) NonBoundedMergeDepthViolatingBlues(blockHash *externalapi.DomainHash) ([]*externalapi.DomainHash, error) {
	ghostdagData, err := mdm.ghostdagDataStore.Get(mdm.databaseContext, blockHash)
	if err != nil {
		return nil, err
	}

	nonBoundedMergeDepthViolatingBlues := make([]*externalapi.DomainHash, 0, len(ghostdagData.MergeSetBlues))

	for _, blue := range ghostdagData.MergeSetBlues {
		notViolatingFinality, err := mdm.hasFinalityPointInOthersSelectedChain(ghostdagData, blue)
		if err != nil {
			return nil, err
		}

		if notViolatingFinality {
			nonBoundedMergeDepthViolatingBlues = append(nonBoundedMergeDepthViolatingBlues, blue)
		}
	}

	return nonBoundedMergeDepthViolatingBlues, nil
}

func (mdm *mergeDepthManager) hasFinalityPointInOthersSelectedChain(ghostdagData *model.BlockGHOSTDAGData, other *externalapi.DomainHash) (bool, error) {
	finalityPoint, err := mdm.finalityPoint(ghostdagData)
	if err != nil {
		return false, err
	}

	return mdm.dagTopologyManager.IsInSelectedParentChainOf(finalityPoint, other)
}

func (mdm *mergeDepthManager) finalityPoint(ghostdagData *model.BlockGHOSTDAGData) (*externalapi.DomainHash, error) {
	return mdm.dagTraversalManager.HighestChainBlockBelowBlueScore(ghostdagData.SelectedParent, ghostdagData.BlueScore-mdm.finalityDepth)
}
