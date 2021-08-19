package ghost

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/processes/reachabilitymanager"
	"github.com/kaspanet/kaspad/domain/consensus/utils/hashset"
	"github.com/pkg/errors"
)

type ghostReachabilityManager struct {
	ghostdagDataStore     *ghostdagDataStore
	reachabilityDataStore *reachabilityDataStore
	reachabilityManager   model.ReachabilityManager
}

type ghostdagDataStore struct {
	blockGHOSTDAGData map[externalapi.DomainHash]*externalapi.BlockGHOSTDAGData
}

func newGHOSTDAGDataStore() *ghostdagDataStore {
	return &ghostdagDataStore{
		blockGHOSTDAGData: map[externalapi.DomainHash]*externalapi.BlockGHOSTDAGData{},
	}
}

func (gds *ghostdagDataStore) Stage(_ *model.StagingArea, blockHash *externalapi.DomainHash,
	blockGHOSTDAGData *externalapi.BlockGHOSTDAGData, _ bool) {

	gds.blockGHOSTDAGData[*blockHash] = blockGHOSTDAGData
}

func (gds *ghostdagDataStore) IsStaged(_ *model.StagingArea) bool {
	return true
}

func (gds *ghostdagDataStore) Get(_ model.DBReader, _ *model.StagingArea,
	blockHash *externalapi.DomainHash, _ bool) (*externalapi.BlockGHOSTDAGData, error) {

	blockGHOSTDAGData, ok := gds.blockGHOSTDAGData[*blockHash]
	if !ok {
		return nil, errors.Errorf("ghostdag data not found for block hash %s", blockHash)
	}
	return blockGHOSTDAGData, nil
}

type reachabilityDataStore struct {
	reachabilityData        map[externalapi.DomainHash]model.ReachabilityData
	reachabilityReindexRoot *externalapi.DomainHash
}

func newReachabilityDataStore() *reachabilityDataStore {
	return &reachabilityDataStore{
		reachabilityData:        map[externalapi.DomainHash]model.ReachabilityData{},
		reachabilityReindexRoot: nil,
	}
}

func (rds *reachabilityDataStore) StageReachabilityData(_ *model.StagingArea,
	blockHash *externalapi.DomainHash, reachabilityData model.ReachabilityData) {

	rds.reachabilityData[*blockHash] = reachabilityData
}

func (rds *reachabilityDataStore) StageReachabilityReindexRoot(_ *model.StagingArea,
	reachabilityReindexRoot *externalapi.DomainHash) {

	rds.reachabilityReindexRoot = reachabilityReindexRoot
}

func (rds *reachabilityDataStore) IsStaged(_ *model.StagingArea) bool {
	return true
}

func (rds *reachabilityDataStore) ReachabilityData(_ model.DBReader, _ *model.StagingArea,
	blockHash *externalapi.DomainHash) (model.ReachabilityData, error) {

	reachabilityData, ok := rds.reachabilityData[*blockHash]
	if !ok {
		return nil, errors.Errorf("reachability data not found for block hash %s", blockHash)
	}
	return reachabilityData, nil
}

func (rds *reachabilityDataStore) HasReachabilityData(_ model.DBReader, _ *model.StagingArea,
	blockHash *externalapi.DomainHash) (bool, error) {

	_, ok := rds.reachabilityData[*blockHash]
	return ok, nil
}

func (rds *reachabilityDataStore) ReachabilityReindexRoot(_ model.DBReader,
	_ *model.StagingArea) (*externalapi.DomainHash, error) {

	return rds.reachabilityReindexRoot, nil
}

func newGHOSTReachabilityManager(subDAG *model.SubDAG) (*ghostReachabilityManager, error) {
	ghostdagDataStore := newGHOSTDAGDataStore()
	reachabilityDataStore := newReachabilityDataStore()
	reachabilityManager := reachabilitymanager.New(nil, ghostdagDataStore, reachabilityDataStore)

	ghostReachabilityManager := &ghostReachabilityManager{
		ghostdagDataStore:     ghostdagDataStore,
		reachabilityDataStore: reachabilityDataStore,
		reachabilityManager:   reachabilityManager,
	}
	err := ghostReachabilityManager.initialize(subDAG)
	if err != nil {
		return nil, err
	}
	return ghostReachabilityManager, nil
}

func (grm *ghostReachabilityManager) initialize(subDAG *model.SubDAG) error {
	blockHashToHeightMap, heightToBlockHashesMap, maxHeight := grm.buildHeightMaps(subDAG)

	for blockHash, block := range subDAG.Blocks {
		blockHeight := blockHashToHeightMap[blockHash]
		selectedParent := model.VirtualGenesisBlockHash
		if len(block.ParentHashes) > 0 {
			selectedParent = block.ParentHashes[0]
		}
		blockGHOSTDAGData := externalapi.NewBlockGHOSTDAGData(blockHeight, nil, selectedParent, nil, nil, nil)
		grm.ghostdagDataStore.Stage(nil, &blockHash, blockGHOSTDAGData, false)
	}

	err := grm.reachabilityManager.Init(nil)
	if err != nil {
		return err
	}

	for height := uint64(0); height <= maxHeight; height++ {
		for _, blockHash := range heightToBlockHashesMap[height] {
			err := grm.reachabilityManager.AddBlock(nil, blockHash)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (grm *ghostReachabilityManager) buildHeightMaps(subDAG *model.SubDAG) (
	map[externalapi.DomainHash]uint64, map[uint64][]*externalapi.DomainHash, uint64) {

	blockHashToHeightMap := make(map[externalapi.DomainHash]uint64, len(subDAG.Blocks))
	heightToBlockHashesMap := make(map[uint64][]*externalapi.DomainHash)
	maxHeight := uint64(0)

	queue := append([]*externalapi.DomainHash{}, subDAG.RootHashes...)
	addedToQueue := hashset.NewFromSlice(subDAG.RootHashes...)
	for len(queue) > 0 {
		var currentBlockHash *externalapi.DomainHash
		currentBlockHash, queue = queue[0], queue[1:]

		// Send the block to the back of the queue if one or more of its parents had not been processed yet
		currentBlock := subDAG.Blocks[*currentBlockHash]
		hasMissingParentData := false
		for _, parentHash := range currentBlock.ParentHashes {
			if _, ok := blockHashToHeightMap[*parentHash]; !ok {
				hasMissingParentData = true
				continue
			}
		}
		if hasMissingParentData {
			queue = append(queue, currentBlockHash)
			continue
		}

		for _, childHash := range currentBlock.ChildHashes {
			if addedToQueue.Contains(childHash) {
				continue
			}
			queue = append(queue, childHash)
			addedToQueue.Add(childHash)
		}

		currentBlockHeight := uint64(0)
		if len(currentBlock.ParentHashes) > 0 {
			highestParentHeight := uint64(0)
			for _, parentHash := range currentBlock.ParentHashes {
				parentHeight := blockHashToHeightMap[*parentHash]
				if parentHeight > highestParentHeight {
					highestParentHeight = parentHeight
				}
			}
			currentBlockHeight = highestParentHeight + 1
		}
		blockHashToHeightMap[*currentBlockHash] = currentBlockHeight

		if _, ok := heightToBlockHashesMap[currentBlockHeight]; !ok {
			heightToBlockHashesMap[currentBlockHeight] = []*externalapi.DomainHash{}
		}
		heightToBlockHashesMap[currentBlockHeight] = append(heightToBlockHashesMap[currentBlockHeight], currentBlockHash)

		if currentBlockHeight > maxHeight {
			maxHeight = currentBlockHeight
		}
	}
	return blockHashToHeightMap, heightToBlockHashesMap, maxHeight
}

func (grm *ghostReachabilityManager) isDescendantOf(blockAHash *externalapi.DomainHash, blockBHash *externalapi.DomainHash) (bool, error) {
	return grm.reachabilityManager.IsDAGAncestorOf(nil, blockBHash, blockAHash)
}
