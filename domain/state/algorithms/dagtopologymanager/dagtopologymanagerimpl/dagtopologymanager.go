package dagtopologymanagerimpl

import (
	"github.com/kaspanet/kaspad/domain/state/algorithms/reachabilitytree"
	"github.com/kaspanet/kaspad/domain/state/datastructures/blockrelationstore"
	"github.com/kaspanet/kaspad/infrastructure/db/dbaccess"
	"github.com/kaspanet/kaspad/util/daghash"
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

func (dtm *DAGTopologyManager) AddBlock(dbTx *dbaccess.TxContext, blockHash *daghash.Hash) {

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
