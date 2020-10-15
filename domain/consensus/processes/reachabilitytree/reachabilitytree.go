package reachabilitytree

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
)

// reachabilityTree maintains a structure that allows to answer
// reachability queries in sub-linear time
type reachabilityTree struct {
	blockRelationStore    model.BlockRelationStore
	reachabilityDataStore model.ReachabilityDataStore
}

// New instantiates a new reachabilityTree
func New(
	blockRelationStore model.BlockRelationStore,
	reachabilityDataStore model.ReachabilityDataStore) model.ReachabilityTree {
	return &reachabilityTree{
		blockRelationStore:    blockRelationStore,
		reachabilityDataStore: reachabilityDataStore,
	}
}

// IsReachabilityTreeAncestorOf returns true if blockHashA is an
// ancestor of blockHashB in the reachability tree. Note that this
// does not necessarily mean that it isn't its ancestor in the DAG.
func (rt *reachabilityTree) IsReachabilityTreeAncestorOf(blockHashA *model.DomainHash, blockHashB *model.DomainHash) bool {
	return false
}

// IsDAGAncestorOf returns true if blockHashA is an ancestor of
// blockHashB in the DAG.
func (rt *reachabilityTree) IsDAGAncestorOf(blockHashA *model.DomainHash, blockHashB *model.DomainHash) bool {
	return false
}

// ReachabilityChangeset returns a set of changes that need to occur
// in order to add the given blockHash into the reachability tree.
func (rt *reachabilityTree) ReachabilityChangeset(blockHash *model.DomainHash,
	blockGHOSTDAGData *model.BlockGHOSTDAGData) *model.ReachabilityChangeset {

	return nil
}
