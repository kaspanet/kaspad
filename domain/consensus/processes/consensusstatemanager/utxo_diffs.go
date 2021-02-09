package consensusstatemanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/pkg/errors"
)

func (csm *consensusStateManager) stageDiff(blockHash *externalapi.DomainHash,
	utxoDiff model.UTXODiff, utxoDiffChild *externalapi.DomainHash) error {

	log.Debugf("stageDiff start for block %s", blockHash)
	defer log.Debugf("stageDiff end for block %s", blockHash)

	log.Debugf("Staging block %s as the diff child of %s", utxoDiffChild, blockHash)
	csm.utxoDiffStore.Stage(blockHash, utxoDiff, utxoDiffChild)

	if utxoDiffChild == nil {
		log.Debugf("Adding block %s to the virtual diff parents", blockHash)
		return csm.addToVirtualDiffParents(blockHash)
	}

	log.Debugf("Removing block %s from the virtual diff parents", blockHash)
	return csm.removeFromVirtualDiffParents(blockHash)
}

func (csm *consensusStateManager) addToVirtualDiffParents(blockHash *externalapi.DomainHash) error {
	log.Debugf("addToVirtualDiffParents start for block %s", blockHash)
	defer log.Debugf("addToVirtualDiffParents end for block %s", blockHash)

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
		log.Debugf("Block %s is already a virtual diff parent, so there's no need to add it", blockHash)
		return nil
	}

	newVirtualDiffParents := append([]*externalapi.DomainHash{blockHash}, oldVirtualDiffParents...)
	log.Debugf("Staging virtual diff parents after adding %s to it", blockHash)
	csm.consensusStateStore.StageVirtualDiffParents(newVirtualDiffParents)
	return nil
}

func (csm *consensusStateManager) removeFromVirtualDiffParents(blockHash *externalapi.DomainHash) error {
	log.Debugf("removeFromVirtualDiffParents start for block %s", blockHash)
	defer log.Debugf("removeFromVirtualDiffParents end for block %s", blockHash)

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

	log.Debugf("Staging virtual diff parents after removing %s from it", blockHash)
	csm.consensusStateStore.StageVirtualDiffParents(newVirtualDiffParents)
	return nil
}
