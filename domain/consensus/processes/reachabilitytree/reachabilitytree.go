package reachabilitytree

import (
	"github.com/kaspanet/kaspad/domain/consensus/database"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// reachabilityTreeManager maintains a structure that allows to answer
// reachability queries in sub-linear time
type reachabilityTreeManager struct {
	databaseContext       *database.DomainDBContext
	blockRelationStore    model.BlockRelationStore
	reachabilityDataStore model.ReachabilityDataStore
	ghostdagDataStore     model.GHOSTDAGDataStore
}

// New instantiates a new reachabilityTreeManager
func New(
	blockRelationStore model.BlockRelationStore,
	reachabilityDataStore model.ReachabilityDataStore) model.ReachabilityTree {
	return &reachabilityTreeManager{
		blockRelationStore:    blockRelationStore,
		reachabilityDataStore: reachabilityDataStore,
	}
}

// IsReachabilityTreeAncestorOf returns true if blockHashA is an
// ancestor of blockHashB in the reachability tree. Note that this
// does not necessarily mean that it isn't its ancestor in the DAG.
func (rt *reachabilityTreeManager) IsReachabilityTreeAncestorOf(blockHashA *externalapi.DomainHash, blockHashB *externalapi.DomainHash) (bool, error) {
	return false, nil
}

// ReachabilityChangeset returns a set of changes that need to occur
// in order to add the given blockHash into the reachability tree.
func (rt *reachabilityTreeManager) ReachabilityChangeset(blockHash *externalapi.DomainHash,
	blockGHOSTDAGData *model.BlockGHOSTDAGData) (*model.ReachabilityChangeset, error) {

	return nil, nil
}

func (rt *reachabilityTreeManager) addBlock(blockHash *externalapi.DomainHash) error {
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
