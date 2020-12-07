package dagtopologymanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/pkg/errors"
)

// dagTopologyManager exposes methods for querying relationships
// between blocks in the DAG
type dagTopologyManager struct {
	reachabilityManager model.ReachabilityManager
	blockRelationStore  model.BlockRelationStore
	ghostdagStore       model.GHOSTDAGDataStore
	databaseContext     model.DBReader
}

// New instantiates a new DAGTopologyManager
func New(
	databaseContext model.DBReader,
	reachabilityManager model.ReachabilityManager,
	blockRelationStore model.BlockRelationStore,
	ghostdagStore model.GHOSTDAGDataStore) model.DAGTopologyManager {

	return &dagTopologyManager{
		databaseContext:     databaseContext,
		reachabilityManager: reachabilityManager,
		blockRelationStore:  blockRelationStore,
		ghostdagStore:       ghostdagStore,
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
	return dtm.reachabilityManager.IsDAGAncestorOf(blockHashA, blockHashB)
}

// IsDescendantOf returns true if blockHashA is a DAG descendant of blockHashB
func (dtm *dagTopologyManager) IsDescendantOf(blockHashA *externalapi.DomainHash, blockHashB *externalapi.DomainHash) (bool, error) {
	return dtm.reachabilityManager.IsDAGAncestorOf(blockHashB, blockHashA)
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
	return dtm.reachabilityManager.IsReachabilityTreeAncestorOf(blockHashA, blockHashB)
}

func isHashInSlice(hash *externalapi.DomainHash, hashes []*externalapi.DomainHash) bool {
	for _, h := range hashes {
		if *h == *hash {
			return true
		}
	}
	return false
}

func (dtm *dagTopologyManager) SetParents(blockHash *externalapi.DomainHash, parentHashes []*externalapi.DomainHash) error {
	hasRelations, err := dtm.blockRelationStore.Has(dtm.databaseContext, blockHash)
	if err != nil {
		return err
	}

	if hasRelations {
		// Go over the block's current relations (if they exist), and remove the block from all its current parents
		// Note: In theory we should also remove the block from all its children, however, in practice no block
		// ever has its relations updated after getting any children, therefore we skip this step

		currentRelations, err := dtm.blockRelationStore.BlockRelation(dtm.databaseContext, blockHash)
		if err != nil {
			return err
		}

		for _, currentParent := range currentRelations.Parents {
			parentRelations, err := dtm.blockRelationStore.BlockRelation(dtm.databaseContext, currentParent)
			if err != nil {
				return err
			}
			for i, parentChild := range parentRelations.Children {
				if *parentChild == *blockHash {
					parentRelations.Children = append(parentRelations.Children[:i], parentRelations.Children[i+1:]...)
					dtm.blockRelationStore.StageBlockRelation(currentParent, parentRelations)
					break
				}
			}
		}
	}

	// Go over all new parents and add block as their child
	for _, parent := range parentHashes {
		parentRelations, err := dtm.blockRelationStore.BlockRelation(dtm.databaseContext, parent)
		if err != nil {
			return err
		}
		isBlockAlreadyInChildren := false
		for _, parentChild := range parentRelations.Children {
			if *parentChild == *blockHash {
				isBlockAlreadyInChildren = true
				break
			}
		}
		if !isBlockAlreadyInChildren {
			parentRelations.Children = append(parentRelations.Children, blockHash)
			dtm.blockRelationStore.StageBlockRelation(parent, parentRelations)
		}
	}

	// Finally - create the relations for the block itself
	dtm.blockRelationStore.StageBlockRelation(blockHash, &model.BlockRelations{
		Parents:  parentHashes,
		Children: []*externalapi.DomainHash{},
	})

	return nil
}

// ChildInSelectedParentChainOf returns the child of `context` that is in the selected-parent-chain of `highHash`
func (dtm *dagTopologyManager) ChildInSelectedParentChainOf(
	blockHash, highHash *externalapi.DomainHash) (*externalapi.DomainHash, error) {

	// Virtual doesn't have reachability data, therefore, it should be treated as a special case -
	// use it's selected parent as highHash.
	specifiedHighHash := highHash
	if highHash == model.VirtualBlockHash {
		ghostdagData, err := dtm.ghostdagStore.Get(dtm.databaseContext, highHash)
		if err != nil {
			return nil, err
		}
		selectedParent := ghostdagData.SelectedParent()

		// In case where `blockHash` is an immediate parent of `highHash`
		if *blockHash == *selectedParent {
			return highHash, nil
		}
		highHash = selectedParent
	}

	isInSelectedParentChain, err := dtm.IsInSelectedParentChainOf(blockHash, highHash)
	if err != nil {
		return nil, err
	}
	if !isInSelectedParentChain {
		return nil, errors.Errorf("blockHash(%s) is not in the selected-parent-chain of highHash(%s)",
			blockHash, specifiedHighHash)
	}

	return dtm.reachabilityManager.FindAncestorOfThisAmongChildrenOfOther(highHash, blockHash)
}
