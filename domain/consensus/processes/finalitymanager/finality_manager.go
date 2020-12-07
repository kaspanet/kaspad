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

func (fm *finalityManager) IsViolatingFinality(blockHash *externalapi.DomainHash) (bool, error) {
	if *blockHash == *fm.genesisHash {
		log.Tracef("Block %s is the genesis block, "+
			"and does not violate finality by definition", blockHash)
		return false, nil
	}
	log.Tracef("isViolatingFinality start for block %s", blockHash)
	defer log.Tracef("isViolatingFinality end for block %s", blockHash)

	virtualFinalityPoint, err := fm.VirtualFinalityPoint()
	if err != nil {
		return false, err
	}
	log.Tracef("The virtual finality point is: %s", virtualFinalityPoint)

	isInSelectedParentChain, err := fm.dagTopologyManager.IsInSelectedParentChainOf(virtualFinalityPoint, blockHash)
	if err != nil {
		return false, err
	}
	log.Tracef("Is the virtual finality point %s "+
		"in the selected parent chain of %s: %t", virtualFinalityPoint, blockHash, isInSelectedParentChain)

	return !isInSelectedParentChain, nil
}

func (fm *finalityManager) VirtualFinalityPoint() (*externalapi.DomainHash, error) {
	log.Tracef("virtualFinalityPoint start")
	defer log.Tracef("virtualFinalityPoint end")

	virtualFinalityPoint, err := fm.calculateFinalityPoint(model.VirtualBlockHash)
	if err != nil {
		return nil, err
	}
	log.Tracef("The current virtual finality block is: %s", virtualFinalityPoint)

	return virtualFinalityPoint, nil
}

func (fm *finalityManager) FinalityPoint(blockHash *externalapi.DomainHash) (*externalapi.DomainHash, error) {
	if *blockHash == *model.VirtualBlockHash {
		return fm.VirtualFinalityPoint()
	}
	finalityPoint, err := fm.finalityStore.FinalityPoint(fm.databaseContext, blockHash)
	if err != nil {
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
	ghostdagData, err := fm.ghostdagDataStore.Get(fm.databaseContext, blockHash)
	if err != nil {
		return nil, err
	}

	if ghostdagData.BlueScore() < fm.finalityDepth {
		return fm.genesisHash, nil
	}

	selectedParent := ghostdagData.SelectedParent()
	if *selectedParent == *fm.genesisHash {
		return fm.genesisHash, nil
	}

	current, err := fm.finalityStore.FinalityPoint(fm.databaseContext, ghostdagData.SelectedParent())
	if err != nil {
		return nil, err
	}
	requiredBlueScore := ghostdagData.BlueScore() - fm.finalityDepth

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
			return current, nil
		}

		current = next
	}
}
