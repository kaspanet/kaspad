package dagtopologymanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/database"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// dagTopologyManager exposes methods for querying relationships
// between blocks in the DAG
type dagTopologyManager struct {
	reachabilityTree   model.ReachabilityTree
	blockRelationStore model.BlockRelationStore
	databaseContext    *database.DomainDBContext
}

// New instantiates a new DAGTopologyManager
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
func (dtm *dagTopologyManager) Parents(blockHash *externalapi.DomainHash) ([]*externalapi.DomainHash, error) {
	blockRelations, err := dtm.blockRelationStore.BlockRelation(dtm.databaseContext, blockHash)
	if err != nil {
		return nil, err
	}
	return blockRelations.Parents, nil
}

// Children returns the DAG children of the given blockHash
func (dtm *dagTopologyManager) Children(blockHash *externalapi.DomainHash) ([]*externalapi.DomainHash, error) {
	blockRelations, err := dtm.blockRelationStore.BlockRelation(dtm.databaseContext, blockHash)
	if err != nil {
		return nil, err
	}
	return blockRelations.Children, nil
}

// IsParentOf returns true if blockHashA is a direct DAG parent of blockHashB
func (dtm *dagTopologyManager) IsParentOf(blockHashA *externalapi.DomainHash, blockHashB *externalapi.DomainHash) (bool, error) {
	blockRelations, err := dtm.blockRelationStore.BlockRelation(dtm.databaseContext, blockHashB)
	if err != nil {
		return false, err
	}
	return isHashInSlice(blockHashA, blockRelations.Parents), nil
}

// IsChildOf returns true if blockHashA is a direct DAG child of blockHashB
func (dtm *dagTopologyManager) IsChildOf(blockHashA *externalapi.DomainHash, blockHashB *externalapi.DomainHash) (bool, error) {
	blockRelations, err := dtm.blockRelationStore.BlockRelation(dtm.databaseContext, blockHashB)
	if err != nil {
		return false, err
	}
	return isHashInSlice(blockHashA, blockRelations.Children), nil
}

// IsAncestorOf returns true if blockHashA is a DAG ancestor of blockHashB
func (dtm *dagTopologyManager) IsAncestorOf(blockHashA *externalapi.DomainHash, blockHashB *externalapi.DomainHash) (bool, error) {
	return dtm.reachabilityTree.IsDAGAncestorOf(blockHashA, blockHashB)
}

// IsDescendantOf returns true if blockHashA is a DAG descendant of blockHashB
func (dtm *dagTopologyManager) IsDescendantOf(blockHashA *externalapi.DomainHash, blockHashB *externalapi.DomainHash) (bool, error) {
	return dtm.reachabilityTree.IsDAGAncestorOf(blockHashB, blockHashA)
}

// IsAncestorOfAny returns true if `blockHash` is an ancestor of at least one of `potentialDescendants`
func (dtm *dagTopologyManager) IsAncestorOfAny(blockHash *externalapi.DomainHash, potentialDescendants []*externalapi.DomainHash) (bool, error) {
	for _, potentialDescendant := range potentialDescendants {
		isAncestorOf, err := dtm.IsAncestorOf(blockHash, potentialDescendant)
		if err != nil {
			return false, err
		}

		if isAncestorOf {
			return true, nil
		}
	}

	return false, nil
}

// IsInSelectedParentChainOf returns true if blockHashA is in the selected parent chain of blockHashB
func (dtm *dagTopologyManager) IsInSelectedParentChainOf(blockHashA *externalapi.DomainHash, blockHashB *externalapi.DomainHash) (bool, error) {
	return dtm.reachabilityTree.IsReachabilityTreeAncestorOf(blockHashA, blockHashB)
}

func isHashInSlice(hash *externalapi.DomainHash, hashes []*externalapi.DomainHash) bool {
	for _, h := range hashes {
		if *h == *hash {
			return true
		}
	}
	return false
}
