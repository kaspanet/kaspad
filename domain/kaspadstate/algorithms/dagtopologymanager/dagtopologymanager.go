package dagtopologymanager

import (
	"github.com/kaspanet/kaspad/domain/kaspadstate/algorithms"
	"github.com/kaspanet/kaspad/domain/kaspadstate/datastructures"
	"github.com/kaspanet/kaspad/util/daghash"
)

type DAGTopologyManager struct {
	reachabilityTree   algorithms.ReachabilityTree
	blockRelationStore datastructures.BlockRelationStore
}

func New(
	reachabilityTree algorithms.ReachabilityTree,
	blockRelationStore datastructures.BlockRelationStore) *DAGTopologyManager {
	return &DAGTopologyManager{
		reachabilityTree:   reachabilityTree,
		blockRelationStore: blockRelationStore,
	}
}

func (dtm *DAGTopologyManager) Parents(blockHash *daghash.Hash) []*daghash.Hash {
	return nil
}

func (dtm *DAGTopologyManager) Children(blockHash *daghash.Hash) []*daghash.Hash {
	return nil
}

func (dtm *DAGTopologyManager) IsParentOf(blockHashA *daghash.Hash, blockHashB *daghash.Hash) bool {
	return false
}

func (dtm *DAGTopologyManager) IsChildOf(blockHashA *daghash.Hash, blockHashB *daghash.Hash) bool {
	return false
}

func (dtm *DAGTopologyManager) IsAncestorOf(blockHashA *daghash.Hash, blockHashB *daghash.Hash) bool {
	return false
}

func (dtm *DAGTopologyManager) IsDescendantOf(blockHashA *daghash.Hash, blockHashB *daghash.Hash) bool {
	return false
}
