package dagtopologymanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/infrastructure/db/dbaccess"
)

// DAGTopologyManager exposes methods for querying relationships
// between blocks in the DAG
type DAGTopologyManager struct {
	reachabilityTree   model.ReachabilityTree
	blockRelationStore model.BlockRelationStore
	databaseContext    *dbaccess.DatabaseContext
}

// New instantiates a new DAGTopologyManager
func New(
	databaseContext *dbaccess.DatabaseContext,
	reachabilityTree model.ReachabilityTree,
	blockRelationStore model.BlockRelationStore) *DAGTopologyManager {
	return &DAGTopologyManager{
		databaseContext:    databaseContext,
		reachabilityTree:   reachabilityTree,
		blockRelationStore: blockRelationStore,
	}
}

// Parents returns the DAG parents of the given blockHash
func (dtm *DAGTopologyManager) Parents(blockHash *model.DomainHash) []*model.DomainHash {
	return dtm.blockRelationStore.Get(dtm.databaseContext, blockHash).Parents
}

// Children returns the DAG children of the given blockHash
func (dtm *DAGTopologyManager) Children(blockHash *model.DomainHash) []*model.DomainHash {
	return dtm.blockRelationStore.Get(dtm.databaseContext, blockHash).Children
}

// IsParentOf returns true if blockHashA is a direct DAG parent of blockHashB
func (dtm *DAGTopologyManager) IsParentOf(blockHashA *model.DomainHash, blockHashB *model.DomainHash) bool {
	return isHashInSlice(blockHashA, dtm.blockRelationStore.Get(dtm.databaseContext, blockHashB).Parents)
}

// IsChildOf returns true if blockHashA is a direct DAG child of blockHashB
func (dtm *DAGTopologyManager) IsChildOf(blockHashA *model.DomainHash, blockHashB *model.DomainHash) bool {
	return isHashInSlice(blockHashA, dtm.blockRelationStore.Get(dtm.databaseContext, blockHashB).Children)
}

// IsAncestorOf returns true if blockHashA is a DAG ancestor of blockHashB
func (dtm *DAGTopologyManager) IsAncestorOf(blockHashA *model.DomainHash, blockHashB *model.DomainHash) bool {
	return dtm.reachabilityTree.IsDAGAncestorOf(blockHashA, blockHashB)
}

// IsDescendantOf returns true if blockHashA is a DAG descendant of blockHashB
func (dtm *DAGTopologyManager) IsDescendantOf(blockHashA *model.DomainHash, blockHashB *model.DomainHash) bool {
	return dtm.reachabilityTree.IsDAGAncestorOf(blockHashB, blockHashA)
}

func isHashInSlice(hash *daghash.Hash, hashes []*daghash.Hash) bool {
	for _, h := range hashes {
		if *h == *hash {
			return true
		}
	}
	return false
}
