package dagtraversalmanager

import (
	"fmt"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// dagTraversalManager exposes methods for travering blocks
// in the DAG
type dagTraversalManager struct {
	databaseContext model.DBReader

	dagTopologyManager model.DAGTopologyManager
	ghostdagDataStore  model.GHOSTDAGDataStore
	ghostdagManager    model.GHOSTDAGManager
}

// selectedParentIterator implements the `model.SelectedParentIterator` API
type selectedParentIterator struct {
	databaseContext   model.DBReader
	ghostdagDataStore model.GHOSTDAGDataStore
	current           *externalapi.DomainHash
}

func (spi *selectedParentIterator) Next() bool {
	if spi.current == nil {
		return false
	}
	ghostdagData, err := spi.ghostdagDataStore.Get(spi.databaseContext, spi.current)
	if err != nil {
		panic(fmt.Sprintf("ghostdagDataStore is missing ghostdagData for: %v. '%s' ", spi.current, err))
	}
	spi.current = ghostdagData.SelectedParent
	return spi.current != nil
}

func (spi *selectedParentIterator) Get() *externalapi.DomainHash {
	return spi.current
}

// New instantiates a new DAGTraversalManager
func New(
	databaseContext model.DBReader,
	dagTopologyManager model.DAGTopologyManager,
	ghostdagDataStore model.GHOSTDAGDataStore,
	ghostdagManager model.GHOSTDAGManager) model.DAGTraversalManager {
	return &dagTraversalManager{
		databaseContext:    databaseContext,
		dagTopologyManager: dagTopologyManager,
		ghostdagDataStore:  ghostdagDataStore,
		ghostdagManager:    ghostdagManager,
	}
}

// SelectedParentIterator creates an iterator over the selected
// parent chain of the given highHash
func (dtm *dagTraversalManager) SelectedParentIterator(highHash *externalapi.DomainHash) model.SelectedParentIterator {
	return &selectedParentIterator{
		databaseContext:   dtm.databaseContext,
		ghostdagDataStore: dtm.ghostdagDataStore,
		current:           highHash,
	}
}

// HighestChainBlockBelowBlueScore returns the hash of the
// highest block with a blue score lower than the given
// blueScore in the block with the given highHash's selected
// parent chain
func (dtm *dagTraversalManager) HighestChainBlockBelowBlueScore(highHash *externalapi.DomainHash, blueScore uint64) (*externalapi.DomainHash, error) {
	blockHash := highHash
	chainBlock, err := dtm.ghostdagDataStore.Get(dtm.databaseContext, highHash)
	if err != nil {
		return nil, err
	}
	if chainBlock.BlueScore < blueScore { // will practically return genesis.
		blueScore = chainBlock.BlueScore
	}

	requiredBlueScore := chainBlock.BlueScore - blueScore

	// If we used `SelectedParentIterator` we'd need to do more calls to `ghostdagDataStore` so we can get the blueScore
	for chainBlock.BlueScore >= requiredBlueScore {
		if chainBlock.SelectedParent == nil { // genesis
			return blockHash, nil
		}
		blockHash = chainBlock.SelectedParent
		chainBlock, err = dtm.ghostdagDataStore.Get(dtm.databaseContext, highHash)
		if err != nil {
			return nil, err
		}
	}
	return blockHash, nil
}

func (dtm *dagTraversalManager) LowestChainBlockAboveOrEqualToBlueScore(highHash *externalapi.DomainHash, blueScore uint64) (*externalapi.DomainHash, error) {
	iterator := dtm.SelectedParentIterator(highHash)
	highBlockGHOSTDAGData, err := dtm.ghostdagDataStore.Get(dtm.databaseContext, highHash)
	if err != nil {
		return nil, err
	}
	currentBlueScore := highBlockGHOSTDAGData.BlueScore

	currentHash := highHash
	currentBlockGHOSTDAGData := highBlockGHOSTDAGData
	for iterator.Next() {
		if currentBlueScore < blueScore {
			break
		}

		currentHash = currentBlockGHOSTDAGData.SelectedParent
		currentBlockGHOSTDAGData, err = dtm.ghostdagDataStore.Get(dtm.databaseContext, currentBlockGHOSTDAGData.SelectedParent)
		if err != nil {
			return nil, err
		}
		currentBlueScore = currentBlockGHOSTDAGData.BlueScore
	}

	return currentHash, nil
}
