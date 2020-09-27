package reachabilitytreeimpl

import (
	"github.com/kaspanet/kaspad/domain/kaspadstate/datastructures/blockrelationstore"
	"github.com/kaspanet/kaspad/domain/kaspadstate/datastructures/reachabilitydatastore"
	"github.com/kaspanet/kaspad/infrastructure/db/dbaccess"
	"github.com/kaspanet/kaspad/util/daghash"
)

type ReachabilityTree struct {
	blockRelationStore    blockrelationstore.BlockRelationStore
	reachabilityDataStore reachabilitydatastore.ReachabilityDataStore
}

func New(
	blockRelationStore blockrelationstore.BlockRelationStore,
	reachabilityDataStore reachabilitydatastore.ReachabilityDataStore) *ReachabilityTree {
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
