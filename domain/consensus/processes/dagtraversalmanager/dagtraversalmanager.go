package dagtraversalmanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/pkg/errors"
)

// dagTraversalManager exposes methods for travering blocks
// in the DAG
type dagTraversalManager struct {
	databaseContext model.DBReader

	dagTopologyManager    model.DAGTopologyManager
	ghostdagManager       model.GHOSTDAGManager
	ghostdagDataStore     model.GHOSTDAGDataStore
	reachabilityDataStore model.ReachabilityDataStore
	consensusStateStore   model.ConsensusStateStore
	daaWindowStore        model.BlocksWithMetaDataDAAWindowStore
	genesisHash           *externalapi.DomainHash
}

// New instantiates a new DAGTraversalManager
func New(
	databaseContext model.DBReader,
	dagTopologyManager model.DAGTopologyManager,
	ghostdagDataStore model.GHOSTDAGDataStore,
	reachabilityDataStore model.ReachabilityDataStore,
	ghostdagManager model.GHOSTDAGManager,
	conssensusStateStore model.ConsensusStateStore,
	daaWindowStore model.BlocksWithMetaDataDAAWindowStore,
	genesisHash *externalapi.DomainHash) model.DAGTraversalManager {
	return &dagTraversalManager{
		databaseContext:       databaseContext,
		dagTopologyManager:    dagTopologyManager,
		ghostdagDataStore:     ghostdagDataStore,
		reachabilityDataStore: reachabilityDataStore,
		ghostdagManager:       ghostdagManager,
		consensusStateStore:   conssensusStateStore,
		daaWindowStore:        daaWindowStore,

		genesisHash: genesisHash,
	}
}

// BlockAtDepth returns the hash of the highest block with a blue score
// lower than (highHash.blueSore - depth) in the selected-parent-chain
// of the block with the given highHash's selected parent chain.
func (dtm *dagTraversalManager) BlockAtDepth(stagingArea *model.StagingArea, highHash *externalapi.DomainHash, depth uint64) (*externalapi.DomainHash, error) {
	currentBlockHash := highHash
	highBlockGHOSTDAGData, err := dtm.ghostdagDataStore.Get(dtm.databaseContext, stagingArea, highHash)
	if err != nil {
		return nil, err
	}

	requiredBlueScore := uint64(0)
	if highBlockGHOSTDAGData.BlueScore() > depth {
		requiredBlueScore = highBlockGHOSTDAGData.BlueScore() - depth
	}

	currentBlockGHOSTDAGData := highBlockGHOSTDAGData
	// If we used `BlockIterator` we'd need to do more calls to `ghostdagDataStore` so we can get the blueScore
	for currentBlockGHOSTDAGData.BlueScore() >= requiredBlueScore {
		if currentBlockGHOSTDAGData.SelectedParent() == nil { // genesis
			return currentBlockHash, nil
		}
		currentBlockHash = currentBlockGHOSTDAGData.SelectedParent()
		currentBlockGHOSTDAGData, err = dtm.ghostdagDataStore.Get(dtm.databaseContext, stagingArea, currentBlockHash)
		if err != nil {
			return nil, err
		}
	}
	return currentBlockHash, nil
}

func (dtm *dagTraversalManager) LowestChainBlockAboveOrEqualToBlueScore(stagingArea *model.StagingArea, highHash *externalapi.DomainHash, blueScore uint64) (*externalapi.DomainHash, error) {
	highBlockGHOSTDAGData, err := dtm.ghostdagDataStore.Get(dtm.databaseContext, stagingArea, highHash)
	if err != nil {
		return nil, err
	}

	if highBlockGHOSTDAGData.BlueScore() < blueScore {
		return nil, errors.Errorf("the given blue score %d is higher than block %s blue score of %d",
			blueScore, highHash, highBlockGHOSTDAGData.BlueScore())
	}

	currentHash := highHash
	currentBlockGHOSTDAGData := highBlockGHOSTDAGData

	for !currentHash.Equal(dtm.genesisHash) {
		selectedParentBlockGHOSTDAGData, err := dtm.ghostdagDataStore.Get(dtm.databaseContext, stagingArea,
			currentBlockGHOSTDAGData.SelectedParent())
		if err != nil {
			return nil, err
		}

		if selectedParentBlockGHOSTDAGData.BlueScore() < blueScore {
			break
		}
		currentHash = currentBlockGHOSTDAGData.SelectedParent()
		currentBlockGHOSTDAGData = selectedParentBlockGHOSTDAGData
	}

	return currentHash, nil
}

func (dtm *dagTraversalManager) CalculateChainPath(stagingArea *model.StagingArea,
	fromBlockHash, toBlockHash *externalapi.DomainHash) (*externalapi.SelectedChainPath, error) {

	// Walk down from fromBlockHash until we reach the common selected
	// parent chain ancestor of fromBlockHash and toBlockHash. Note
	// that this slice will be empty if fromBlockHash is the selected
	// parent of toBlockHash
	var removed []*externalapi.DomainHash
	current := fromBlockHash
	for {
		isCurrentInTheSelectedParentChainOfNewVirtualSelectedParent, err :=
			dtm.dagTopologyManager.IsInSelectedParentChainOf(stagingArea, current, toBlockHash)
		if err != nil {
			return nil, err
		}
		if isCurrentInTheSelectedParentChainOfNewVirtualSelectedParent {
			break
		}
		removed = append(removed, current)

		currentGHOSTDAGData, err := dtm.ghostdagDataStore.Get(dtm.databaseContext, stagingArea, current)
		if err != nil {
			return nil, err
		}
		current = currentGHOSTDAGData.SelectedParent()
	}
	commonAncestor := current

	// Walk down from the toBlockHash to the common ancestor
	var added []*externalapi.DomainHash
	current = toBlockHash
	for !current.Equal(commonAncestor) {
		added = append(added, current)
		currentGHOSTDAGData, err := dtm.ghostdagDataStore.Get(dtm.databaseContext, stagingArea, current)
		if err != nil {
			return nil, err
		}
		current = currentGHOSTDAGData.SelectedParent()
	}

	// Reverse the order of `added` so that it's sorted from low hash to high hash
	for i, j := 0, len(added)-1; i < j; i, j = i+1, j-1 {
		added[i], added[j] = added[j], added[i]
	}

	return &externalapi.SelectedChainPath{
		Added:   added,
		Removed: removed,
	}, nil
}
