package dagtopologymanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/database"
	"github.com/kaspanet/kaspad/domain/consensus/model"
)

// dagTopologyManager exposes methods for querying relationships
// between blocks in the DAG
type dagTopologyManager struct {
	reachabilityTree   model.ReachabilityTree
	blockRelationStore model.BlockRelationStore
	databaseContext    *database.DomainDBContext
}

// New instantiates a new dagTopologyManager
func New(
	databaseContext *database.DomainDBContext,
	reachabilityTree model.ReachabilityTree,
	blockRelationStore model.BlockRelationStore) model.DAGTopologyManager {

	return &dagTopologyManager{
		databaseContext:    databaseContext,
		reachabilityTree:   reachabilityTree,
		blockRelationStore: blockRelationStore,
	}
}

// Parents returns the DAG parents of the given blockHash
func (dtm *dagTopologyManager) Parents(blockHash *model.DomainHash) []*model.DomainHash {
	return dtm.blockRelationStore.Get(dtm.databaseContext, blockHash).Parents
}

// Children returns the DAG children of the given blockHash
func (dtm *dagTopologyManager) Children(blockHash *model.DomainHash) []*model.DomainHash {
	return dtm.blockRelationStore.Get(dtm.databaseContext, blockHash).Children
}

// IsParentOf returns true if blockHashA is a direct DAG parent of blockHashB
func (dtm *dagTopologyManager) IsParentOf(blockHashA *model.DomainHash, blockHashB *model.DomainHash) bool {
	return isHashInSlice(blockHashA, dtm.blockRelationStore.Get(dtm.databaseContext, blockHashB).Parents)
}

// IsChildOf returns true if blockHashA is a direct DAG child of blockHashB
func (dtm *dagTopologyManager) IsChildOf(blockHashA *model.DomainHash, blockHashB *model.DomainHash) bool {
	return isHashInSlice(blockHashA, dtm.blockRelationStore.Get(dtm.databaseContext, blockHashB).Children)
}

// IsAncestorOf returns true if blockHashA is a DAG ancestor of blockHashB
func (dtm *dagTopologyManager) IsAncestorOf(blockHashA *model.DomainHash, blockHashB *model.DomainHash) bool {
	return dtm.reachabilityTree.IsDAGAncestorOf(blockHashA, blockHashB)
}

// IsDescendantOf returns true if blockHashA is a DAG descendant of blockHashB
func (dtm *dagTopologyManager) IsDescendantOf(blockHashA *model.DomainHash, blockHashB *model.DomainHash) bool {
	return dtm.reachabilityTree.IsDAGAncestorOf(blockHashB, blockHashA)
}

// IsAncestorOfAny returns true if `blockHash` is an ancestor of at least one of `potentialDescendants`
func (dtm *dagTopologyManager) IsAncestorOfAny(blockHash *model.DomainHash, potentialDescendants []*model.DomainHash) bool {
	return false
}

// IsInSelectedParentChainOf returns true if blockHashA is in the selected parent chain of blockHashB
func (dtm *dagTopologyManager) IsInSelectedParentChainOf(blockHashA *model.DomainHash, blockHashB *model.DomainHash) bool {
	return false
}

func isHashInSlice(hash *model.DomainHash, hashes []*model.DomainHash) bool {
	for _, h := range hashes {
		if *h == *hash {
			return true
		}
	}
	return false
}
