package reachabilitytreeimpl

import (
	"github.com/kaspanet/kaspad/domain/state/datastructures/blockrelationstore"
	"github.com/kaspanet/kaspad/domain/state/datastructures/reachabilitydatastore"
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
