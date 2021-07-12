package reachabilitymanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// reachabilityManager maintains a structure that allows to answer
// reachability queries in sub-linear time
type reachabilityManager struct {
	databaseContext       model.DBReader
	reachabilityDataStore model.ReachabilityDataStore
	ghostdagDataStore     model.GHOSTDAGDataStore
	reindexSlack          uint64
	reindexWindow         uint64
}

// New instantiates a new reachabilityManager
func New(
	databaseContext model.DBReader,
	ghostdagDataStore model.GHOSTDAGDataStore,
	reachabilityDataStore model.ReachabilityDataStore,
) model.ReachabilityManager {
	return &reachabilityManager{
		databaseContext:       databaseContext,
		ghostdagDataStore:     ghostdagDataStore,
		reachabilityDataStore: reachabilityDataStore,
		reindexSlack:          defaultReindexSlack,
		reindexWindow:         defaultReindexWindow,
	}
}

// AddBlock adds the block with the given blockHash into the reachability tree.
func (rt *reachabilityManager) AddBlock(stagingArea *model.StagingArea, blockHash *externalapi.DomainHash) error {
	// Allocate a new reachability data
	newReachabilityData := newReachabilityTreeData()
	rt.stageData(stagingArea, blockHash, newReachabilityData)

	ghostdagData, err := rt.ghostdagDataStore.Get(rt.databaseContext, stagingArea, blockHash)
	if err != nil {
		return err
	}

	reindexRoot, err := rt.reindexRoot(stagingArea)
	if err != nil {
		return err
	}

	// Insert the node into the selected parent's reachability tree
	err = rt.addChild(stagingArea, ghostdagData.SelectedParent(), blockHash, reindexRoot)
	if err != nil {
		return err
	}

	// Add the block to the futureCoveringSets of all the blocks
	// in the merget set
	mergeSet := make([]*externalapi.DomainHash, len(ghostdagData.MergeSetBlues())+len(ghostdagData.MergeSetReds()))
	copy(mergeSet, ghostdagData.MergeSetBlues())
	copy(mergeSet[len(ghostdagData.MergeSetBlues()):], ghostdagData.MergeSetReds())

	for _, current := range mergeSet {
		err = rt.insertToFutureCoveringSet(stagingArea, current, blockHash)
		if err != nil {
			return err
		}
	}

	return nil
}

func (rt *reachabilityManager) Init(stagingArea *model.StagingArea) error {
	hasReachabilityData, err := rt.reachabilityDataStore.HasReachabilityData(rt.databaseContext, stagingArea, model.VirtualGenesisBlockHash)
	if err != nil {
		return err
	}

	if hasReachabilityData {
		return nil
	}

	newReachabilityData := newReachabilityTreeData()
	rt.stageData(stagingArea, model.VirtualGenesisBlockHash, newReachabilityData)
	rt.stageReindexRoot(stagingArea, model.VirtualGenesisBlockHash)

	return nil
}
