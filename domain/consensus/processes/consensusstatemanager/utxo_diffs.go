package consensusstatemanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/pkg/errors"
)

func (csm *consensusStateManager) stageDiff(blockHash *externalapi.DomainHash,
	utxoDiff *model.UTXODiff, utxoDiffChild *externalapi.DomainHash) error {

	err := csm.utxoDiffStore.Stage(blockHash, utxoDiff, utxoDiffChild)
	if err != nil {
		return err
	}

	if utxoDiffChild == nil {
		return csm.addToVirtualDiffParents(blockHash)
	}

	return csm.removeFromVirtualDiffParents(blockHash)
}

func (csm *consensusStateManager) addToVirtualDiffParents(blockHash *externalapi.DomainHash) error {
	var oldVirtualDiffParents []*externalapi.DomainHash
	if *blockHash != *csm.genesisHash {
		var err error
		oldVirtualDiffParents, err = csm.consensusStateStore.VirtualDiffParents(csm.databaseContext)
		if err != nil {
			return err
		}
	}

	isInVirtualDiffParents := false
	for _, diffParent := range oldVirtualDiffParents {
		if *diffParent == *blockHash {
			isInVirtualDiffParents = true
			break
		}
	}

	if isInVirtualDiffParents {
		return nil
	}

	newVirtualDiffParents := append([]*externalapi.DomainHash{blockHash}, oldVirtualDiffParents...)
	return csm.consensusStateStore.StageVirtualDiffParents(newVirtualDiffParents)
}

func (csm *consensusStateManager) removeFromVirtualDiffParents(blockHash *externalapi.DomainHash) error {
	oldVirtualDiffParents, err := csm.consensusStateStore.VirtualDiffParents(csm.databaseContext)
	if err != nil {
		return err
	}

	newVirtualDiffParents := make([]*externalapi.DomainHash, 0, len(oldVirtualDiffParents)-1)
	for _, diffParent := range oldVirtualDiffParents {
		if *diffParent != *blockHash {
			newVirtualDiffParents = append(newVirtualDiffParents, diffParent)
		}
	}

	if len(newVirtualDiffParents) != len(oldVirtualDiffParents)-1 {
		return errors.Errorf("expected to remove one member from virtual diff parents and "+
			"have a length of %d but got length of %d", len(oldVirtualDiffParents)-1, len(newVirtualDiffParents))
	}

	return csm.consensusStateStore.StageVirtualDiffParents(newVirtualDiffParents)
}
