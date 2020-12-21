package consensusstatemanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

func (csm *consensusStateManager) GetVirtualSelectedParentChainFromBlock(
	blockHash *externalapi.DomainHash) (*externalapi.SelectedParentChainChanges, error) {

	virtualGHOSTDAGData, err := csm.ghostdagDataStore.Get(csm.databaseContext, model.VirtualBlockHash)
	if err != nil {
		return nil, err
	}
	virtualSelectedParent := virtualGHOSTDAGData.SelectedParent()

	// Calculate chain changes between the given blockHash and the
	// virtual's selected parent. Note that we explicitly don't
	// do the calculation against the virtual itself so that we
	// won't later need to remove it from the result.
	return csm.calculateSelectedParentChainChanges(blockHash, virtualSelectedParent)
}

func (csm *consensusStateManager) calculateSelectedParentChainChanges(
	fromBlockHash, toBlockHash *externalapi.DomainHash) (*externalapi.SelectedParentChainChanges, error) {

	// Walk down from fromBlockHash until we reach the common selected
	// parent chain ancestor of fromBlockHash and toBlockHash. Note
	// that this slice will be empty if fromBlockHash is the selected
	// parent of toBlockHash
	var removed []*externalapi.DomainHash
	current := fromBlockHash
	for {
		isCurrentInTheSelectedParentChainOfNewVirtualSelectedParent, err := csm.dagTopologyManager.IsInSelectedParentChainOf(current, toBlockHash)
		if err != nil {
			return nil, err
		}
		if isCurrentInTheSelectedParentChainOfNewVirtualSelectedParent {
			break
		}
		removed = append(removed, current)

		currentGHOSTDAGData, err := csm.ghostdagDataStore.Get(csm.databaseContext, current)
		if err != nil {
			return nil, err
		}
		current = currentGHOSTDAGData.SelectedParent()
	}
	commonAncestor := current

	// Walk down from the toBlockHash to the common ancestor
	var added []*externalapi.DomainHash
	current = toBlockHash
	for *current != *commonAncestor {
		added = append(added, current)
		currentGHOSTDAGData, err := csm.ghostdagDataStore.Get(csm.databaseContext, current)
		if err != nil {
			return nil, err
		}
		current = currentGHOSTDAGData.SelectedParent()
	}

	// Reverse the order of `added` so that it's sorted from low hash to high hash
	for i, j := 0, len(added)-1; i < j; i, j = i+1, j-1 {
		added[i], added[j] = added[j], added[i]
	}

	return &externalapi.SelectedParentChainChanges{
		Added:   added,
		Removed: removed,
	}, nil
}
