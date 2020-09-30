package reachabilitytreeimpl

import (
	"github.com/kaspanet/kaspad/domain/kaspadstate/datastructures"
	"github.com/kaspanet/kaspad/infrastructure/db/dbaccess"
	"github.com/kaspanet/kaspad/util/daghash"
)

type ReachabilityTree struct {
	blockRelationStore    datastructures.BlockRelationStore
	reachabilityDataStore datastructures.ReachabilityDataStore
}

func New(
	blockRelationStore datastructures.BlockRelationStore,
	reachabilityDataStore datastructures.ReachabilityDataStore) *ReachabilityTree {
	return &ReachabilityTree{
		blockRelationStore:    blockRelationStore,
		reachabilityDataStore: reachabilityDataStore,
	}
}

func (rt *ReachabilityTree) AddNode(dbTx *dbaccess.TxContext, blockHash *daghash.Hash) {

}

func (rt *ReachabilityTree) IsInPastOf(blockHashA *daghash.Hash, blockHashB *daghash.Hash) bool {
	return false
}
