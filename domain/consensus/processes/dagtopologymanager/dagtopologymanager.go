package dagtopologymanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
)

// DAGTopologyManager exposes methods for querying relationships
// between blocks in the DAG
type DAGTopologyManager struct {
	reachabilityTree   model.ReachabilityTree
	blockRelationStore model.BlockRelationStore
}

// New instantiates a new DAGTopologyManager
func New(
	reachabilityTree model.ReachabilityTree,
	blockRelationStore model.BlockRelationStore) *DAGTopologyManager {
	return &DAGTopologyManager{
		reachabilityTree:   reachabilityTree,
		blockRelationStore: blockRelationStore,
	}
}

// Parents returns the DAG parents of the given blockHash
func (dtm *DAGTopologyManager) Parents(blockHash *model.DomainHash) []*model.DomainHash {
	return nil
}

// Children returns the DAG children of the given blockHash
func (dtm *DAGTopologyManager) Children(blockHash *model.DomainHash) []*model.DomainHash {
	return nil
}

// IsParentOf returns true if blockHashA is a direct DAG parent of blockHashB
func (dtm *DAGTopologyManager) IsParentOf(blockHashA *model.DomainHash, blockHashB *model.DomainHash) bool {
	return false
}

// IsChildOf returns true if blockHashA is a direct DAG child of blockHashB
func (dtm *DAGTopologyManager) IsChildOf(blockHashA *model.DomainHash, blockHashB *model.DomainHash) bool {
	return false
}

// IsAncestorOf returns true if blockHashA is a DAG ancestor of blockHashB
func (dtm *DAGTopologyManager) IsAncestorOf(blockHashA *model.DomainHash, blockHashB *model.DomainHash) bool {
	return false
}

// IsDescendantOf returns true if blockHashA is a DAG descendant of blockHashB
func (dtm *DAGTopologyManager) IsDescendantOf(blockHashA *model.DomainHash, blockHashB *model.DomainHash) bool {
	return false
}
