package ghost

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/processes/reachabilitymanager"
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

func newGHOSTReachabilityManager(subDAG *model.SubDAG, heightMaps *heightMaps) (*ghostReachabilityManager, error) {
	ghostdagDataStore := newGHOSTDAGDataStore()
	reachabilityDataStore := newReachabilityDataStore()
	reachabilityManager := reachabilitymanager.New(nil, ghostdagDataStore, reachabilityDataStore)

	ghostReachabilityManager := &ghostReachabilityManager{
		ghostdagDataStore:     ghostdagDataStore,
		reachabilityDataStore: reachabilityDataStore,
		reachabilityManager:   reachabilityManager,
	}
	err := ghostReachabilityManager.initialize(subDAG, heightMaps)
	if err != nil {
		return nil, err
	}
	return ghostReachabilityManager, nil
}

func (grm *ghostReachabilityManager) initialize(subDAG *model.SubDAG, heightMaps *heightMaps) error {
	for blockHash, block := range subDAG.Blocks {
		blockHeight := heightMaps.blockHashToHeightMap[blockHash]
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

	for height := uint64(0); height <= heightMaps.maxHeight; height++ {
		for _, blockHash := range heightMaps.heightToBlockHashesMap[height] {
			err := grm.reachabilityManager.AddBlock(nil, blockHash)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (grm *ghostReachabilityManager) isDescendantOf(blockAHash *externalapi.DomainHash, blockBHash *externalapi.DomainHash) (bool, error) {
	return grm.reachabilityManager.IsDAGAncestorOf(nil, blockBHash, blockAHash)
}
