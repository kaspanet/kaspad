package dagtopologymanagerimpl

import (
	"github.com/kaspanet/kaspad/domain/state/algorithms/reachabilitytree"
	"github.com/kaspanet/kaspad/domain/state/datastructures/blockrelationstore"
)

type DAGTopologyManager struct {
	reachabilityTree   reachabilitytree.ReachabilityTree
	blockRelationStore blockrelationstore.BlockRelationStore
}

func New(
	reachabilityTree reachabilitytree.ReachabilityTree,
	blockRelationStore blockrelationstore.BlockRelationStore) *DAGTopologyManager {
	return &DAGTopologyManager{
		reachabilityTree:   reachabilityTree,
		blockRelationStore: blockRelationStore,
	}
}
