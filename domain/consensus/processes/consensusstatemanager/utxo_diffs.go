package consensusstatemanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/pkg/errors"
)

func (csm *consensusStateManager) stageDiff(blockHash *externalapi.DomainHash,
	utxoDiff *model.UTXODiff, utxoDiffChild *externalapi.DomainHash) error {

	log.Tracef("stageDiff start for block %s", blockHash)
	defer log.Tracef("stageDiff end for block %s", blockHash)

	log.Tracef("Staging block %s as the diff child of %s", utxoDiffChild, blockHash)
	csm.utxoDiffStore.Stage(blockHash, utxoDiff, utxoDiffChild)

	if utxoDiffChild == nil {
		log.Tracef("Adding block %s to the virtual diff parents", blockHash)
		return csm.addToVirtualDiffParents(blockHash)
	}

	log.Tracef("Removing block %s from the virtual diff parents", blockHash)
	return csm.removeFromVirtualDiffParents(blockHash)
}

func (csm *consensusStateManager) addToVirtualDiffParents(blockHash *externalapi.DomainHash) error {
	log.Tracef("addToVirtualDiffParents start for block %s", blockHash)
	defer log.Tracef("addToVirtualDiffParents end for block %s", blockHash)

	var oldVirtualDiffParents []*externalapi.DomainHash
	if !blockHash.Equal(csm.genesisHash) {
		var err error
		oldVirtualDiffParents, err = csm.consensusStateStore.VirtualDiffParents(csm.databaseContext)
		if err != nil {
			return err
		}
	}

	isInVirtualDiffParents := false
	for _, diffParent := range oldVirtualDiffParents {
		if diffParent.Equal(blockHash) {
			isInVirtualDiffParents = true
			break
		}
	}

	if isInVirtualDiffParents {
		log.Tracef("Block %s is already a virtual diff parent, so there's no need to add it", blockHash)
		return nil
	}

	newVirtualDiffParents := append([]*externalapi.DomainHash{blockHash}, oldVirtualDiffParents...)
	log.Tracef("Staging virtual diff parents after adding %s to it", blockHash)
	csm.consensusStateStore.StageVirtualDiffParents(newVirtualDiffParents)
	return nil
}

func (csm *consensusStateManager) removeFromVirtualDiffParents(blockHash *externalapi.DomainHash) error {
	log.Tracef("removeFromVirtualDiffParents start for block %s", blockHash)
	defer log.Tracef("removeFromVirtualDiffParents end for block %s", blockHash)

	oldVirtualDiffParents, err := csm.consensusStateStore.VirtualDiffParents(csm.databaseContext)
	if err != nil {
		return err
	}

	newVirtualDiffParents := make([]*externalapi.DomainHash, 0, len(oldVirtualDiffParents)-1)
	for _, diffParent := range oldVirtualDiffParents {
		if !diffParent.Equal(blockHash) {
			newVirtualDiffParents = append(newVirtualDiffParents, diffParent)
		}
	}

	if len(newVirtualDiffParents) != len(oldVirtualDiffParents)-1 {
		return errors.Errorf("expected to remove one member from virtual diff parents and "+
			"have a length of %d but got length of %d", len(oldVirtualDiffParents)-1, len(newVirtualDiffParents))
	}

	log.Tracef("Staging virtual diff parents after removing %s from it", blockHash)
	csm.consensusStateStore.StageVirtualDiffParents(newVirtualDiffParents)
	return nil
}
