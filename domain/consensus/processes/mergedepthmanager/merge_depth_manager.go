package mergedepthmanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
	"github.com/kaspanet/kaspad/infrastructure/db/database"
	"github.com/pkg/errors"
)

type mergeDepthManager struct {
	databaseContext     model.DBReader
	dagTopologyManager  model.DAGTopologyManager
	dagTraversalManager model.DAGTraversalManager
	finalityManager     model.FinalityManager

	genesisHash *externalapi.DomainHash
	mergeDepth  uint64
	hf1DAAScore uint64

	ghostdagDataStore   model.GHOSTDAGDataStore
	mergeDepthRootStore model.MergeDepthRootStore
	daaBlocksStore      model.DAABlocksStore
	pruningStore        model.PruningStore
	finalityStore       model.FinalityStore
}

// New instantiates a new MergeDepthManager
func New(
	databaseContext model.DBReader,
	dagTopologyManager model.DAGTopologyManager,
	dagTraversalManager model.DAGTraversalManager,
	finalityManager model.FinalityManager,

	genesisHash *externalapi.DomainHash,
	mergeDepth uint64,
	hf1DAAScore uint64,

	ghostdagDataStore model.GHOSTDAGDataStore,
	mergeDepthRootStore model.MergeDepthRootStore,
	daaBlocksStore model.DAABlocksStore,
	pruningStore model.PruningStore,
	finalityStore model.FinalityStore) model.MergeDepthManager {

	return &mergeDepthManager{
		databaseContext:     databaseContext,
		dagTopologyManager:  dagTopologyManager,
		dagTraversalManager: dagTraversalManager,
		finalityManager:     finalityManager,
		genesisHash:         genesisHash,
		mergeDepth:          mergeDepth,
		hf1DAAScore:         hf1DAAScore,
		ghostdagDataStore:   ghostdagDataStore,
		mergeDepthRootStore: mergeDepthRootStore,
		daaBlocksStore:      daaBlocksStore,
		pruningStore:        pruningStore,
		finalityStore:       finalityStore,
	}

}

// CheckBoundedMergeDepth is used for validation, so must follow the HF1 DAA score for determining the correct depth to verify
func (mdm *mergeDepthManager) CheckBoundedMergeDepth(stagingArea *model.StagingArea, blockHash *externalapi.DomainHash, isBlockWithTrustedData bool) error {
	ghostdagData, err := mdm.ghostdagDataStore.Get(mdm.databaseContext, stagingArea, blockHash, false)
	if err != nil {
		return err
	}

	// Return nil on genesis
	if ghostdagData.SelectedParent() == nil {
		return nil
	}

	// For validation, we must follow the HF1 DAA score in order to determine the correct depth to verify
	mergeDepthRootByHF1, err := mdm.mergeDepthRootByHF1DAAScoreForValidationOnly(stagingArea, blockHash, isBlockWithTrustedData)
	if err != nil {
		return err
	}

	nonBoundedMergeDepthViolatingBlues, err := mdm.nonBoundedMergeDepthViolatingBlues(stagingArea, blockHash, mergeDepthRootByHF1)
	if err != nil {
		return err
	}

	for _, red := range ghostdagData.MergeSetReds() {
		doesRedHaveMergeRootInPast, err := mdm.dagTopologyManager.IsAncestorOf(stagingArea, mergeDepthRootByHF1, red)
		if err != nil {
			return err
		}

		if doesRedHaveMergeRootInPast {
			continue
		}

		isRedInPastOfAnyNonMergeDepthViolatingBlue, err :=
			mdm.dagTopologyManager.IsAncestorOfAny(stagingArea, red, nonBoundedMergeDepthViolatingBlues)
		if err != nil {
			return err
		}

		if !isRedInPastOfAnyNonMergeDepthViolatingBlue {
			return errors.Wrapf(ruleerrors.ErrViolatingBoundedMergeDepth, "block is violating bounded merge depth")
		}
	}

	return nil
}

// NonBoundedMergeDepthViolatingBlues is currently called only by the `Pick virtual parents` algorithm,
// hence, unlike validation, we can use the new more strict depth root
func (mdm *mergeDepthManager) NonBoundedMergeDepthViolatingBlues(stagingArea *model.StagingArea, blockHash *externalapi.DomainHash, isBlockWithTrustedData bool) ([]*externalapi.DomainHash, error) {
	mergeDepthRoot, err := mdm.MergeDepthRoot(stagingArea, blockHash, isBlockWithTrustedData)
	if err != nil {
		return nil, err
	}

	return mdm.nonBoundedMergeDepthViolatingBlues(stagingArea, blockHash, mergeDepthRoot)
}

func (mdm *mergeDepthManager) nonBoundedMergeDepthViolatingBlues(
	stagingArea *model.StagingArea, blockHash, mergeDepthRoot *externalapi.DomainHash) ([]*externalapi.DomainHash, error) {

	ghostdagData, err := mdm.ghostdagDataStore.Get(mdm.databaseContext, stagingArea, blockHash, false)
	if err != nil {
		return nil, err
	}

	nonBoundedMergeDepthViolatingBlues := make([]*externalapi.DomainHash, 0, len(ghostdagData.MergeSetBlues()))

	if err != nil {
		return nil, err
	}
	for _, blue := range ghostdagData.MergeSetBlues() {
		isMergeDepthRootInSelectedChainOfBlue, err := mdm.dagTopologyManager.IsInSelectedParentChainOf(stagingArea, mergeDepthRoot, blue)
		if err != nil {
			return nil, err
		}

		if isMergeDepthRootInSelectedChainOfBlue {
			nonBoundedMergeDepthViolatingBlues = append(nonBoundedMergeDepthViolatingBlues, blue)
		}
	}

	return nonBoundedMergeDepthViolatingBlues, nil
}

func (mdm *mergeDepthManager) mergeDepthRootByHF1DAAScoreForValidationOnly(
	stagingArea *model.StagingArea, blockHash *externalapi.DomainHash, isBlockWithTrustedData bool) (*externalapi.DomainHash, error) {
	daaScore, err := mdm.daaBlocksStore.DAAScore(mdm.databaseContext, stagingArea, blockHash)
	if err != nil {
		return nil, err
	}

	if daaScore >= mdm.hf1DAAScore {
		return mdm.MergeDepthRoot(stagingArea, blockHash, isBlockWithTrustedData)
	}

	// We fall back to the merge depth root before the HF, which was the finality point
	return mdm.finalityManager.FinalityPoint(stagingArea, blockHash, isBlockWithTrustedData)
}

func (mdm *mergeDepthManager) VirtualMergeDepthRoot(stagingArea *model.StagingArea) (*externalapi.DomainHash, error) {
	log.Tracef("VirtualMergeDepthRoot start")
	defer log.Tracef("VirtualMergeDepthRoot end")

	virtualMergeDepthRoot, err := mdm.calculateMergeDepthRoot(stagingArea, model.VirtualBlockHash, false)
	if err != nil {
		return nil, err
	}
	log.Debugf("The current virtual merge depth root is: %s", virtualMergeDepthRoot)

	return virtualMergeDepthRoot, nil
}

func (mdm *mergeDepthManager) MergeDepthRoot(stagingArea *model.StagingArea, blockHash *externalapi.DomainHash, isBlockWithTrustedData bool) (*externalapi.DomainHash, error) {
	log.Tracef("MergeDepthRoot start")
	defer log.Tracef("MergeDepthRoot end")
	if blockHash.Equal(model.VirtualBlockHash) {
		return mdm.VirtualMergeDepthRoot(stagingArea)
	}
	root, err := mdm.mergeDepthRootStore.MergeDepthRoot(mdm.databaseContext, stagingArea, blockHash)
	if err != nil {
		log.Debugf("%s merge depth root not found in store - calculating", blockHash)
		if errors.Is(err, database.ErrNotFound) {
			return mdm.calculateAndStageMergeDepthRoot(stagingArea, blockHash, isBlockWithTrustedData)
		}
		return nil, err
	}
	return root, nil
}

func (mdm *mergeDepthManager) calculateAndStageMergeDepthRoot(
	stagingArea *model.StagingArea, blockHash *externalapi.DomainHash, isBlockWithTrustedData bool) (*externalapi.DomainHash, error) {

	root, err := mdm.calculateMergeDepthRoot(stagingArea, blockHash, isBlockWithTrustedData)
	if err != nil {
		return nil, err
	}
	mdm.mergeDepthRootStore.StageMergeDepthRoot(stagingArea, blockHash, root)
	return root, nil
}

func (mdm *mergeDepthManager) calculateMergeDepthRoot(stagingArea *model.StagingArea, blockHash *externalapi.DomainHash, isBlockWithTrustedData bool) (
	*externalapi.DomainHash, error) {

	log.Tracef("calculateMergeDepthRoot start")
	defer log.Tracef("calculateMergeDepthRoot end")

	if isBlockWithTrustedData {
		return model.VirtualGenesisBlockHash, nil
	}

	ghostdagData, err := mdm.ghostdagDataStore.Get(mdm.databaseContext, stagingArea, blockHash, false)
	if err != nil {
		return nil, err
	}

	if ghostdagData.BlueScore() < mdm.mergeDepth {
		log.Debugf("%s blue score lower then merge depth - returning genesis as merge depth root", blockHash)
		return mdm.genesisHash, nil
	}

	pruningPoint, err := mdm.pruningStore.PruningPoint(mdm.databaseContext, stagingArea)
	if err != nil {
		return nil, err
	}
	pruningPointGhostdagData, err := mdm.ghostdagDataStore.Get(mdm.databaseContext, stagingArea, pruningPoint, false)
	if err != nil {
		return nil, err
	}
	if ghostdagData.BlueScore() < pruningPointGhostdagData.BlueScore()+mdm.mergeDepth {
		log.Debugf("%s blue score less than merge depth over pruning point - returning virtual genesis as merge depth root", blockHash)
		return model.VirtualGenesisBlockHash, nil
	}
	isPruningPointOnChain, err := mdm.dagTopologyManager.IsInSelectedParentChainOf(stagingArea, pruningPoint, blockHash)
	if err != nil {
		return nil, err
	}
	if !isPruningPointOnChain {
		log.Debugf("pruning point not in selected chain of %s - returning virtual genesis as merge depth root", blockHash)
		return model.VirtualGenesisBlockHash, nil
	}

	selectedParent := ghostdagData.SelectedParent()
	if selectedParent.Equal(mdm.genesisHash) {
		return mdm.genesisHash, nil
	}

	current, err := mdm.mergeDepthRootStore.MergeDepthRoot(mdm.databaseContext, stagingArea, ghostdagData.SelectedParent())
	if database.IsNotFoundError(err) {
		current, err = mdm.finalityStore.FinalityPoint(mdm.databaseContext, stagingArea, ghostdagData.SelectedParent())
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}
	// In this case we expect the pruning point or a block above it to be the merge depth root.
	// Note that above we already verified the chain and distance conditions for this
	if current.Equal(model.VirtualGenesisBlockHash) {
		current = pruningPoint
	}

	requiredBlueScore := ghostdagData.BlueScore() - mdm.mergeDepth
	log.Debugf("%s's merge depth root is the one having the highest blue score lower then %d", blockHash, requiredBlueScore)

	var next *externalapi.DomainHash
	for {
		next, err = mdm.dagTopologyManager.ChildInSelectedParentChainOf(stagingArea, current, blockHash)
		if err != nil {
			return nil, err
		}
		nextGHOSTDAGData, err := mdm.ghostdagDataStore.Get(mdm.databaseContext, stagingArea, next, false)
		if err != nil {
			return nil, err
		}
		if nextGHOSTDAGData.BlueScore() >= requiredBlueScore {
			log.Debugf("%s's merge depth root is %s", blockHash, current)
			return current, nil
		}

		current = next
	}
}
