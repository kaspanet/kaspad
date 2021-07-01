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

func (mdm *mergeDepthManager) CheckBoundedMergeDepth(stagingArea *model.StagingArea, blockHash *externalapi.DomainHash, isBlockWithPrefilledData bool) error {
	nonBoundedMergeDepthViolatingBlues, err := mdm.NonBoundedMergeDepthViolatingBlues(stagingArea, blockHash, isBlockWithPrefilledData)
	if err != nil {
		return err
	}

	ghostdagData, err := mdm.ghostdagDataStore.Get(mdm.databaseContext, stagingArea, blockHash)
	if err != nil {
		return err
	}

	// Return nil on genesis
	if ghostdagData.SelectedParent() == nil {
		return nil
	}

	finalityPoint, err := mdm.finalityManager.FinalityPoint(stagingArea, blockHash, isBlockWithPrefilledData)
	if err != nil {
		return err
	}

	for _, red := range ghostdagData.MergeSetReds() {
		doesRedHaveFinalityPointInPast, err := mdm.dagTopologyManager.IsAncestorOf(stagingArea, finalityPoint, red)
		if err != nil {
			return err
		}

		if doesRedHaveFinalityPointInPast {
			continue
		}

		isRedInPastOfAnyNonFinalityViolatingBlue, err :=
			mdm.dagTopologyManager.IsAncestorOfAny(stagingArea, red, nonBoundedMergeDepthViolatingBlues)
		if err != nil {
			return err
		}

		if !isRedInPastOfAnyNonFinalityViolatingBlue {
			return errors.Wrapf(ruleerrors.ErrViolatingBoundedMergeDepth, "block is violating bounded merge depth")
		}
	}

	return nil
}

func (mdm *mergeDepthManager) NonBoundedMergeDepthViolatingBlues(stagingArea *model.StagingArea, blockHash *externalapi.DomainHash, isBlockWithPrefilledData bool) ([]*externalapi.DomainHash, error) {
	ghostdagData, err := mdm.ghostdagDataStore.Get(mdm.databaseContext, stagingArea, blockHash)
	if err != nil {
		return nil, err
	}

	nonBoundedMergeDepthViolatingBlues := make([]*externalapi.DomainHash, 0, len(ghostdagData.MergeSetBlues()))

	finalityPoint, err := mdm.finalityManager.FinalityPoint(stagingArea, blockHash, isBlockWithPrefilledData)
	if err != nil {
		return nil, err
	}
	for _, blue := range ghostdagData.MergeSetBlues() {
		notViolatingFinality, err := mdm.dagTopologyManager.IsInSelectedParentChainOf(stagingArea, finalityPoint, blue)
		if err != nil {
			return nil, err
		}

		if notViolatingFinality {
			nonBoundedMergeDepthViolatingBlues = append(nonBoundedMergeDepthViolatingBlues, blue)
		}
	}

	return nonBoundedMergeDepthViolatingBlues, nil
}
