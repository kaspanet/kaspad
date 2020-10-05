package reachabilitytree

import (
	"github.com/kaspanet/kaspad/domain/consensus/datastructures"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/util/daghash"
)

// ReachabilityTree maintains a structure that allows to answer
// reachability queries in sub-linear time
type ReachabilityTree struct {
	blockRelationStore    datastructures.BlockRelationStore
	reachabilityDataStore datastructures.ReachabilityDataStore
}

// New instantiates a new ReachabilityTree
func New(
	blockRelationStore datastructures.BlockRelationStore,
	reachabilityDataStore datastructures.ReachabilityDataStore) *ReachabilityTree {
	return &ReachabilityTree{
		blockRelationStore:    blockRelationStore,
		reachabilityDataStore: reachabilityDataStore,
	}
}

// IsReachabilityAncestorOf returns true if blockHashA is an ancestor
// of blockHashB in the reachability tree. Note that this does not
// necessarily mean that it isn't its ancestor in the DAG.
func (rt *ReachabilityTree) IsReachabilityTreeAncestorOf(blockHashA *daghash.Hash, blockHashB *daghash.Hash) bool {
	return false
}

// IsDAGAncestorOf returns true if blockHashA is an ancestor of
// blockHashB in the DAG.
func (rt *ReachabilityTree) IsDAGAncestorOf(blockHashA *daghash.Hash, blockHashB *daghash.Hash) bool {
	return false
}

// ReachabilityChangeset returns a set of changes that need to occur
// in order to add the given blockHash into the reachability tree.
func (rt *ReachabilityTree) ReachabilityChangeset(blockHash *daghash.Hash,
	blockGHOSTDAGData *model.BlockGHOSTDAGData) *model.ReachabilityChangeset {

	return nil
}
