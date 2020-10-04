package reachabilitytree

import (
	"github.com/kaspanet/kaspad/domain/kaspadstate/datastructures"
	"github.com/kaspanet/kaspad/infrastructure/db/dbaccess"
	"github.com/kaspanet/kaspad/util/daghash"
)

// ReachabilityTree ...
type ReachabilityTree struct {
	blockRelationStore    datastructures.BlockRelationStore
	reachabilityDataStore datastructures.ReachabilityDataStore
}

// New ...
func New(
	blockRelationStore datastructures.BlockRelationStore,
	reachabilityDataStore datastructures.ReachabilityDataStore) *ReachabilityTree {
	return &ReachabilityTree{
		blockRelationStore:    blockRelationStore,
		reachabilityDataStore: reachabilityDataStore,
	}
}

// AddNode ...
func (rt *ReachabilityTree) AddNode(dbTx *dbaccess.TxContext, blockHash *daghash.Hash) {

}

// IsReachabilityAncestorOf ...
func (rt *ReachabilityTree) IsReachabilityAncestorOf(blockHashA *daghash.Hash, blockHashB *daghash.Hash) bool {
	return false
}

// IsDAGAncestorOf ...
func (rt *ReachabilityTree) IsDAGAncestorOf(blockHashA *daghash.Hash, blockHashB *daghash.Hash) bool {
	return false
}
