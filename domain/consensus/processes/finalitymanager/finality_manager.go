package finalitymanager

import (
	"errors"

	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/infrastructure/db/database"
)

type finalityManager struct {
	databaseContext    model.DBReader
	dagTopologyManager model.DAGTopologyManager
	finalityStore      model.FinalityStore
	ghostdagDataStore  model.GHOSTDAGDataStore
	genesisHash        *externalapi.DomainHash
	finalityDepth      uint64
}

// New instantiates a new FinalityManager
func New(databaseContext model.DBReader,
	dagTopologyManager model.DAGTopologyManager,
	finalityStore model.FinalityStore,
	ghostdagDataStore model.GHOSTDAGDataStore,
	genesisHash *externalapi.DomainHash,
	finalityDepth uint64) model.FinalityManager {

	return &finalityManager{
		databaseContext:    databaseContext,
		genesisHash:        genesisHash,
		dagTopologyManager: dagTopologyManager,
		finalityStore:      finalityStore,
		ghostdagDataStore:  ghostdagDataStore,
		finalityDepth:      finalityDepth,
	}
}

func (fm *finalityManager) VirtualFinalityPoint() (*externalapi.DomainHash, error) {
	log.Debugf("virtualFinalityPoint start")
	defer log.Debugf("virtualFinalityPoint end")

	virtualFinalityPoint, err := fm.calculateFinalityPoint(model.VirtualBlockHash)
	if err != nil {
		return nil, err
	}
	log.Debugf("The current virtual finality block is: %s", virtualFinalityPoint)

	return virtualFinalityPoint, nil
}

func (fm *finalityManager) FinalityPoint(blockHash *externalapi.DomainHash) (*externalapi.DomainHash, error) {
	log.Debugf("FinalityPoint start")
	defer log.Debugf("FinalityPoint end")
	if blockHash.Equal(model.VirtualBlockHash) {
		return fm.VirtualFinalityPoint()
	}
	finalityPoint, err := fm.finalityStore.FinalityPoint(fm.databaseContext, blockHash)
	if err != nil {
		log.Debugf("%s finality point not found in store - calculating", blockHash)
		if errors.Is(err, database.ErrNotFound) {
			return fm.calculateAndStageFinalityPoint(blockHash)
		}
		return nil, err
	}
	return finalityPoint, nil
}

func (fm *finalityManager) calculateAndStageFinalityPoint(blockHash *externalapi.DomainHash) (*externalapi.DomainHash, error) {
	finalityPoint, err := fm.calculateFinalityPoint(blockHash)
	if err != nil {
		return nil, err
	}
	fm.finalityStore.StageFinalityPoint(blockHash, finalityPoint)
	return finalityPoint, nil
}

func (fm *finalityManager) calculateFinalityPoint(blockHash *externalapi.DomainHash) (*externalapi.DomainHash, error) {
	log.Debugf("calculateFinalityPoint start")
	defer log.Debugf("calculateFinalityPoint end")
	ghostdagData, err := fm.ghostdagDataStore.Get(fm.databaseContext, blockHash)
	if err != nil {
		return nil, err
	}

	if ghostdagData.BlueScore() < fm.finalityDepth {
		log.Debugf("%s blue score lower then finality depth - returning genesis as finality point", blockHash)
		return fm.genesisHash, nil
	}

	selectedParent := ghostdagData.SelectedParent()
	if selectedParent.Equal(fm.genesisHash) {
		return fm.genesisHash, nil
	}

	current, err := fm.finalityStore.FinalityPoint(fm.databaseContext, ghostdagData.SelectedParent())
	if err != nil {
		return nil, err
	}
	requiredBlueScore := ghostdagData.BlueScore() - fm.finalityDepth
	log.Debugf("%s's finality point is the one having the highest blue score lower then %d", blockHash, requiredBlueScore)

	var next *externalapi.DomainHash
	for {
		next, err = fm.dagTopologyManager.ChildInSelectedParentChainOf(current, blockHash)
		if err != nil {
			return nil, err
		}
		nextGHOSTDAGData, err := fm.ghostdagDataStore.Get(fm.databaseContext, next)
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
