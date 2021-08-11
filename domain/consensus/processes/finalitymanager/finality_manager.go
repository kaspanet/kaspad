package finalitymanager

import (
	"errors"
	"fmt"

	"github.com/kaspanet/kaspad/domain/consensus/database"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
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
		fmt.Printf("block is with trusted data, its finality point must be fefefefe\n")
		return model.VirtualGenesisBlockHash, nil
	}
	fmt.Printf("block is NOT with trusted data\n")

	ghostdagData, err := fm.ghostdagDataStore.Get(fm.databaseContext, stagingArea, blockHash, false)
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

	current, err := fm.finalityStore.FinalityPoint(fm.databaseContext, stagingArea, ghostdagData.SelectedParent())
	if err != nil {
		return nil, err
	}
	requiredBlueScore := ghostdagData.BlueScore() - fm.finalityDepth
	log.Debugf("%s's finality point is the one having the highest blue score lower then %d", blockHash, requiredBlueScore)
	fmt.Printf("%s's finality point is the one having the highest blue score lower then %d\n", blockHash, requiredBlueScore)

	var next *externalapi.DomainHash
	for {
		fmt.Printf("current: %s\n", current)
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
			fmt.Printf("%s's finality point is %s\n", blockHash, current)
			return current, nil
		}

		current = next
	}
}
