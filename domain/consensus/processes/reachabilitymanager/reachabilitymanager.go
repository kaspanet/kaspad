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
	}
}

// AddBlock adds the block with the given blockHash into the reachability tree.
func (rt *reachabilityManager) AddBlock(blockHash *externalapi.DomainHash) error {
	// Allocate a new reachability tree node
	newTreeNode := newReachabilityTreeNode()
	err := rt.stageTreeNode(blockHash, newTreeNode)
	if err != nil {
		return err
	}

	ghostdagData, err := rt.ghostdagDataStore.Get(rt.databaseContext, blockHash)
	if err != nil {
		return err
	}

	// If this is the genesis node, simply initialize it and return
	if ghostdagData.SelectedParent == nil {
		rt.stageReindexRoot(blockHash)
		return nil
	}

	reindexRoot, err := rt.reindexRoot()
	if err != nil {
		return err
	}

	// Insert the node into the selected parent's reachability tree
	err = rt.addChild(ghostdagData.SelectedParent, blockHash, reindexRoot)
	if err != nil {
		return err
	}

	// Add the block to the futureCoveringSets of all the blocks
	// in the merget set
	mergeSet := make([]*externalapi.DomainHash, len(ghostdagData.MergeSetBlues)+len(ghostdagData.MergeSetReds))
	copy(mergeSet, ghostdagData.MergeSetBlues)
	copy(mergeSet[len(ghostdagData.MergeSetBlues):], ghostdagData.MergeSetReds)

	for _, current := range mergeSet {
		err = rt.insertToFutureCoveringSet(current, blockHash)
		if err != nil {
			return err
		}
	}

	return nil
}
