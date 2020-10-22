package reachabilitytree

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// reachabilityTree maintains a structure that allows to answer
// reachability queries in sub-linear time
type reachabilityTree struct {
	blockRelationStore    model.BlockRelationStore
	reachabilityDataStore model.ReachabilityDataStore
}

// New instantiates a new ReachabilityTree
func New(
	blockRelationStore model.BlockRelationStore,
	reachabilityDataStore model.ReachabilityDataStore) model.ReachabilityTree {
	return &reachabilityTree{
		blockRelationStore:    blockRelationStore,
		reachabilityDataStore: reachabilityDataStore,
	}
}

// AddBlock adds the block with the given blockHash into the reachability tree.
func (rt *reachabilityTree) AddBlock(blockHash *externalapi.DomainHash) error {
	return nil
}

// IsReachabilityTreeAncestorOf returns true if blockHashA is an
// ancestor of blockHashB in the reachability tree. Note that this
// does not necessarily mean that it isn't its ancestor in the DAG.
func (rt *reachabilityTree) IsReachabilityTreeAncestorOf(blockHashA *externalapi.DomainHash, blockHashB *externalapi.DomainHash) (bool, error) {
	return false, nil
}

// IsDAGAncestorOf returns true if blockHashA is an ancestor of
// blockHashB in the DAG.
func (rt *reachabilityTree) IsDAGAncestorOf(blockHashA *externalapi.DomainHash, blockHashB *externalapi.DomainHash) (bool, error) {
	return false, nil
}
