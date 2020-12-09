package consensusstatemanager

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

func (csm *consensusStateManager) findSelectedParentChainChanges(
	oldVirtualSelectedParent, newVirtualSelectedParent *externalapi.DomainHash) (*externalapi.SelectedParentChainChanges, error) {

	// Walk down from the old virtual until we reach the common selected
	// parent chain ancestor of oldVirtualSelectedParent and
	// newVirtualSelectedParent. Note that this slice will be empty if
	// oldVirtualSelectedParent is the selected parent of
	// newVirtualSelectedParent
	var removed []*externalapi.DomainHash
	current := oldVirtualSelectedParent
	for {
		isCurrentInTheSelectedParentChainOfNewVirtualSelectedParent, err := csm.dagTopologyManager.IsInSelectedParentChainOf(current, newVirtualSelectedParent)
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

	// Walk down from the new virtual down to the common ancestor
	var added []*externalapi.DomainHash
	current = newVirtualSelectedParent
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
