package mergedepthmanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
	"github.com/pkg/errors"
)

type mergeDepthManager struct {
	databaseContext     model.DBReader
	dagTopologyManager  model.DAGTopologyManager
	dagTraversalManager model.DAGTraversalManager
	finalityManager     model.FinalityManager

	ghostdagDataStore model.GHOSTDAGDataStore
}

// New instantiates a new MergeDepthManager
func New(
	databaseContext model.DBReader,
	dagTopologyManager model.DAGTopologyManager,
	dagTraversalManager model.DAGTraversalManager,
	finalityManager model.FinalityManager,
	ghostdagDataStore model.GHOSTDAGDataStore) model.MergeDepthManager {

	return &mergeDepthManager{
		databaseContext:     databaseContext,
		dagTopologyManager:  dagTopologyManager,
		dagTraversalManager: dagTraversalManager,
		finalityManager:     finalityManager,
		ghostdagDataStore:   ghostdagDataStore,
	}

}

func (mdm *mergeDepthManager) CheckBoundedMergeDepth(blockHash *externalapi.DomainHash) error {
	nonBoundedMergeDepthViolatingBlues, err := mdm.NonBoundedMergeDepthViolatingBlues(blockHash)
	if err != nil {
		return err
	}

	ghostdagData, err := mdm.ghostdagDataStore.Get(mdm.databaseContext, nil, blockHash)
	if err != nil {
		return err
	}

	// Return nil on genesis
	if ghostdagData.SelectedParent() == nil {
		return nil
	}

	finalityPoint, err := mdm.finalityManager.FinalityPoint(blockHash)
	if err != nil {
		return err
	}

	for _, red := range ghostdagData.MergeSetReds() {
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

func (mdm *mergeDepthManager) NonBoundedMergeDepthViolatingBlues(blockHash *externalapi.DomainHash) ([]*externalapi.DomainHash, error) {
	ghostdagData, err := mdm.ghostdagDataStore.Get(mdm.databaseContext, nil, blockHash)
	if err != nil {
		return nil, err
	}

	nonBoundedMergeDepthViolatingBlues := make([]*externalapi.DomainHash, 0, len(ghostdagData.MergeSetBlues()))

	finalityPoint, err := mdm.finalityManager.FinalityPoint(blockHash)
	if err != nil {
		return nil, err
	}
	for _, blue := range ghostdagData.MergeSetBlues() {
		notViolatingFinality, err := mdm.dagTopologyManager.IsInSelectedParentChainOf(finalityPoint, blue)
		if err != nil {
			return nil, err
		}

		if notViolatingFinality {
			nonBoundedMergeDepthViolatingBlues = append(nonBoundedMergeDepthViolatingBlues, blue)
		}
	}

	return nonBoundedMergeDepthViolatingBlues, nil
}
