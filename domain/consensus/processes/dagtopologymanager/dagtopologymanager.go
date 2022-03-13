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
func (dtm *dagTopologyManager) Parents(stagingArea *model.StagingArea, blockHash *externalapi.DomainHash) ([]*externalapi.DomainHash, error) {
	blockRelations, err := dtm.blockRelationStore.BlockRelation(dtm.databaseContext, stagingArea, blockHash)
	if err != nil {
		return nil, err
	}
	return blockRelations.Parents, nil
}

// Children returns the DAG children of the given blockHash
func (dtm *dagTopologyManager) Children(stagingArea *model.StagingArea, blockHash *externalapi.DomainHash) ([]*externalapi.DomainHash, error) {
	blockRelations, err := dtm.blockRelationStore.BlockRelation(dtm.databaseContext, stagingArea, blockHash)
	if err != nil {
		return nil, err
	}
	return blockRelations.Children, nil
}

// IsParentOf returns true if blockHashA is a direct DAG parent of blockHashB
func (dtm *dagTopologyManager) IsParentOf(stagingArea *model.StagingArea, blockHashA *externalapi.DomainHash, blockHashB *externalapi.DomainHash) (bool, error) {
	blockRelations, err := dtm.blockRelationStore.BlockRelation(dtm.databaseContext, stagingArea, blockHashB)
	if err != nil {
		return false, err
	}
	return isHashInSlice(blockHashA, blockRelations.Parents), nil
}

// IsChildOf returns true if blockHashA is a direct DAG child of blockHashB
func (dtm *dagTopologyManager) IsChildOf(stagingArea *model.StagingArea, blockHashA *externalapi.DomainHash, blockHashB *externalapi.DomainHash) (bool, error) {
	blockRelations, err := dtm.blockRelationStore.BlockRelation(dtm.databaseContext, stagingArea, blockHashB)
	if err != nil {
		return false, err
	}
	return isHashInSlice(blockHashA, blockRelations.Children), nil
}

// IsAncestorOf returns true if blockHashA is a DAG ancestor of blockHashB
func (dtm *dagTopologyManager) IsAncestorOf(stagingArea *model.StagingArea, blockHashA *externalapi.DomainHash, blockHashB *externalapi.DomainHash) (bool, error) {
	return dtm.reachabilityManager.IsDAGAncestorOf(stagingArea, blockHashA, blockHashB)
}

// IsAncestorOfAny returns true if `blockHash` is an ancestor of at least one of `potentialDescendants`
func (dtm *dagTopologyManager) IsAncestorOfAny(stagingArea *model.StagingArea, blockHash *externalapi.DomainHash, potentialDescendants []*externalapi.DomainHash) (bool, error) {
	for _, potentialDescendant := range potentialDescendants {
		isAncestorOf, err := dtm.IsAncestorOf(stagingArea, blockHash, potentialDescendant)
		if err != nil {
			return false, err
		}

		if isAncestorOf {
			return true, nil
		}
	}

	return false, nil
}

// IsAnyAncestorOf returns true if at least one of `potentialAncestors` is an ancestor of `blockHash`
func (dtm *dagTopologyManager) IsAnyAncestorOf(stagingArea *model.StagingArea, potentialAncestors []*externalapi.DomainHash, blockHash *externalapi.DomainHash) (bool, error) {
	for _, potentialAncestor := range potentialAncestors {
		isAncestorOf, err := dtm.IsAncestorOf(stagingArea, potentialAncestor, blockHash)
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
func (dtm *dagTopologyManager) IsInSelectedParentChainOf(stagingArea *model.StagingArea, blockHashA *externalapi.DomainHash, blockHashB *externalapi.DomainHash) (bool, error) {

	// Virtual doesn't have reachability data, therefore, it should be treated as a special case -
	// use its selected parent as blockHashB.
	if blockHashB == model.VirtualBlockHash {
		ghostdagData, err := dtm.ghostdagStore.Get(dtm.databaseContext, stagingArea, blockHashB, false)
		if err != nil {
			return false, err
		}
		blockHashB = ghostdagData.SelectedParent()
	}

	return dtm.reachabilityManager.IsReachabilityTreeAncestorOf(stagingArea, blockHashA, blockHashB)
}

func isHashInSlice(hash *externalapi.DomainHash, hashes []*externalapi.DomainHash) bool {
	for _, h := range hashes {
		if h.Equal(hash) {
			return true
		}
	}
	return false
}

func (dtm *dagTopologyManager) SetParents(stagingArea *model.StagingArea, blockHash *externalapi.DomainHash, parentHashes []*externalapi.DomainHash) error {
	hasRelations, err := dtm.blockRelationStore.Has(dtm.databaseContext, stagingArea, blockHash)
	if err != nil {
		return err
	}

	if hasRelations {
		// Go over the block's current relations (if they exist), and remove the block from all its current parents
		// Note: In theory we should also remove the block from all its children, however, in practice no block
		// ever has its relations updated after getting any children, therefore we skip this step

		currentRelations, err := dtm.blockRelationStore.BlockRelation(dtm.databaseContext, stagingArea, blockHash)
		if err != nil {
			return err
		}

		for _, currentParent := range currentRelations.Parents {
			parentRelations, err := dtm.blockRelationStore.BlockRelation(dtm.databaseContext, stagingArea, currentParent)
			if err != nil {
				return err
			}
			for i, parentChild := range parentRelations.Children {
				if parentChild.Equal(blockHash) {
					parentRelations.Children = append(parentRelations.Children[:i], parentRelations.Children[i+1:]...)
					dtm.blockRelationStore.StageBlockRelation(stagingArea, currentParent, parentRelations)
					break
				}
			}
		}
	}

	// Go over all new parents and add block as their child
	for _, parent := range parentHashes {
		parentRelations, err := dtm.blockRelationStore.BlockRelation(dtm.databaseContext, stagingArea, parent)
		if err != nil {
			return err
		}
		isBlockAlreadyInChildren := false
		for _, parentChild := range parentRelations.Children {
			if parentChild.Equal(blockHash) {
				isBlockAlreadyInChildren = true
				break
			}
		}
		if !isBlockAlreadyInChildren {
			parentRelations.Children = append(parentRelations.Children, blockHash)
			dtm.blockRelationStore.StageBlockRelation(stagingArea, parent, parentRelations)
		}
	}

	// Finally - create the relations for the block itself
	dtm.blockRelationStore.StageBlockRelation(stagingArea, blockHash, &model.BlockRelations{
		Parents:  parentHashes,
		Children: []*externalapi.DomainHash{},
	})

	return nil
}

// ChildInSelectedParentChainOf returns the child of `lowHash` that is in the selected-parent-chain of `highHash`
func (dtm *dagTopologyManager) ChildInSelectedParentChainOf(stagingArea *model.StagingArea, lowHash, highHash *externalapi.DomainHash) (*externalapi.DomainHash, error) {

	// Virtual doesn't have reachability data, therefore, it should be treated as a special case -
	// use its selected parent as highHash.
	specifiedHighHash := highHash
	if highHash == model.VirtualBlockHash {
		ghostdagData, err := dtm.ghostdagStore.Get(dtm.databaseContext, stagingArea, highHash, false)
		if err != nil {
			return nil, err
		}
		selectedParent := ghostdagData.SelectedParent()

		// In case where `context` is an immediate parent of `highHash`
		if lowHash.Equal(selectedParent) {
			return highHash, nil
		}
		highHash = selectedParent
	}

	isInSelectedParentChain, err := dtm.IsInSelectedParentChainOf(stagingArea, lowHash, highHash)
	if err != nil {
		return nil, err
	}
	if !isInSelectedParentChain {
		return nil, errors.Errorf("Claimed chain ancestor (%s) is not in the selected-parent-chain of highHash (%s)",
			lowHash, specifiedHighHash)
	}

	return dtm.reachabilityManager.FindNextAncestor(stagingArea, highHash, lowHash)
}
