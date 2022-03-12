package finalitymanager

import (
	"errors"

	"github.com/kaspanet/kaspad/domain/consensus/database"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

type finalityManager struct {
	databaseContext    model.DBReader
	dagTopologyManager model.DAGTopologyManager
	finalityStore      model.FinalityStore
	ghostdagDataStore  model.GHOSTDAGDataStore
	pruningStore       model.PruningStore
	blockHeaderStore   model.BlockHeaderStore
	genesisHash        *externalapi.DomainHash
	finalityDepth      uint64
}

// New instantiates a new FinalityManager
func New(databaseContext model.DBReader,
	dagTopologyManager model.DAGTopologyManager,
	finalityStore model.FinalityStore,
	ghostdagDataStore model.GHOSTDAGDataStore,
	pruningStore model.PruningStore,
	blockHeaderStore model.BlockHeaderStore,
	genesisHash *externalapi.DomainHash,
	finalityDepth uint64) model.FinalityManager {

	return &finalityManager{
		databaseContext:    databaseContext,
		genesisHash:        genesisHash,
		dagTopologyManager: dagTopologyManager,
		finalityStore:      finalityStore,
		ghostdagDataStore:  ghostdagDataStore,
		pruningStore:       pruningStore,
		blockHeaderStore:   blockHeaderStore,
		finalityDepth:      finalityDepth,
	}
}

func (fm *finalityManager) VirtualFinalityPoint(stagingArea *model.StagingArea) (*externalapi.DomainHash, error) {
	log.Debugf("virtualFinalityPoint start")
	defer log.Debugf("virtualFinalityPoint end")

	virtualFinalityPoint, err := fm.calculateFinalityPoint(stagingArea, model.VirtualBlockHash, false)
	if err != nil {
		return nil, err
	}
	log.Debugf("The current virtual finality block is: %s", virtualFinalityPoint)

	return virtualFinalityPoint, nil
}

func (fm *finalityManager) FinalityPoint(stagingArea *model.StagingArea, blockHash *externalapi.DomainHash, isBlockWithTrustedData bool) (*externalapi.DomainHash, error) {
	log.Debugf("FinalityPoint start")
	defer log.Debugf("FinalityPoint end")
	if blockHash.Equal(model.VirtualBlockHash) {
		return fm.VirtualFinalityPoint(stagingArea)
	}
	finalityPoint, err := fm.finalityStore.FinalityPoint(fm.databaseContext, stagingArea, blockHash)
	if err != nil {
		log.Debugf("%s finality point not found in store - calculating", blockHash)
		if errors.Is(err, database.ErrNotFound) {
			return fm.calculateAndStageFinalityPoint(stagingArea, blockHash, isBlockWithTrustedData)
		}
		return nil, err
	}
	return finalityPoint, nil
}

func (fm *finalityManager) calculateAndStageFinalityPoint(
	stagingArea *model.StagingArea, blockHash *externalapi.DomainHash, isBlockWithTrustedData bool) (*externalapi.DomainHash, error) {

	finalityPoint, err := fm.calculateFinalityPoint(stagingArea, blockHash, isBlockWithTrustedData)
	if err != nil {
		return nil, err
	}
	fm.finalityStore.StageFinalityPoint(stagingArea, blockHash, finalityPoint)
	return finalityPoint, nil
}

func (fm *finalityManager) calculateFinalityPoint(stagingArea *model.StagingArea, blockHash *externalapi.DomainHash, isBlockWithTrustedData bool) (
	*externalapi.DomainHash, error) {

	log.Debugf("calculateFinalityPoint start")
	defer log.Debugf("calculateFinalityPoint end")

	if isBlockWithTrustedData {
		return model.VirtualGenesisBlockHash, nil
	}

	ghostdagData, err := fm.ghostdagDataStore.Get(fm.databaseContext, stagingArea, blockHash, false)
	if err != nil {
		return nil, err
	}

	if ghostdagData.BlueScore() < fm.finalityDepth {
		log.Debugf("%s blue score lower then finality depth - returning genesis as finality point", blockHash)
		return fm.genesisHash, nil
	}

	pruningPoint, err := fm.pruningStore.PruningPoint(fm.databaseContext, stagingArea)
	if err != nil {
		return nil, err
	}
	pruningPointHeader, err := fm.blockHeaderStore.BlockHeader(fm.databaseContext, stagingArea, pruningPoint)
	if err != nil {
		return nil, err
	}
	if ghostdagData.BlueScore() < pruningPointHeader.BlueScore()+fm.finalityDepth {
		log.Debugf("%s blue score less than finality distance over pruning point - returning virtual genesis as finality point", blockHash)
		return model.VirtualGenesisBlockHash, nil
	}
	isPruningPointOnChain, err := fm.dagTopologyManager.IsInSelectedParentChainOf(stagingArea, pruningPoint, blockHash)
	if err != nil {
		return nil, err
	}
	if !isPruningPointOnChain {
		log.Debugf("pruning point not in selected chain of %s - returning virtual genesis as finality point", blockHash)
		return model.VirtualGenesisBlockHash, nil
	}

	selectedParent := ghostdagData.SelectedParent()
	if selectedParent.Equal(fm.genesisHash) {
		return fm.genesisHash, nil
	}

	current, err := fm.finalityStore.FinalityPoint(fm.databaseContext, stagingArea, ghostdagData.SelectedParent())
	if err != nil {
		return nil, err
	}
	// In this case we expect the pruning point or a block above it to be the finality point.
	// Note that above we already verified the chain and distance conditions for this
	if current.Equal(model.VirtualGenesisBlockHash) {
		current = pruningPoint
	}

	requiredBlueScore := ghostdagData.BlueScore() - fm.finalityDepth
	log.Debugf("%s's finality point is the one having the highest blue score lower then %d", blockHash, requiredBlueScore)

	var next *externalapi.DomainHash
	for {
		next, err = fm.dagTopologyManager.ChildInSelectedParentChainOf(stagingArea, current, blockHash)
		if err != nil {
			return nil, err
		}
		nextGHOSTDAGData, err := fm.ghostdagDataStore.Get(fm.databaseContext, stagingArea, next, false)
		if err != nil {
			return nil, err
		}
		if nextGHOSTDAGData.BlueScore() >= requiredBlueScore {
			log.Debugf("%s's finality point is %s", blockHash, current)
			return current, nil
		}

		current = next
	}
}
