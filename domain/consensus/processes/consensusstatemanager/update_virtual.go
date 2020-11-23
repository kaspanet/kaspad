package consensusstatemanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/processes/consensusstatemanager/utxoalgebra"
)

func (csm *consensusStateManager) updateVirtual(newBlockHash *externalapi.DomainHash, tips []*externalapi.DomainHash) error {
	log.Tracef("updateVirtual start for block %s", newBlockHash)
	defer log.Tracef("updateVirtual end for block %s", newBlockHash)

	log.Tracef("Picking virtual parents from the tips: %s", tips)
	virtualParents, err := csm.pickVirtualParents(tips)
	if err != nil {
		return err
	}
	log.Tracef("Picked virtual parents: %s", virtualParents)

	err = csm.dagTopologyManager.SetParents(model.VirtualBlockHash, virtualParents)
	if err != nil {
		return err
	}
	log.Tracef("Set new parents for the virtual block hash")

	err = csm.ghostdagManager.GHOSTDAG(model.VirtualBlockHash)
	if err != nil {
		return err
	}

	log.Tracef("Calculating past UTXO, acceptance data, and multiset for the new virtual block")
	virtualUTXODiff, virtualAcceptanceData, virtualMultiset, err := csm.CalculatePastUTXOAndAcceptanceData(model.VirtualBlockHash)
	if err != nil {
		return err
	}

	log.Tracef("Staging new acceptance data for the virtual block")
	err = csm.acceptanceDataStore.Stage(model.VirtualBlockHash, virtualAcceptanceData)
	if err != nil {
		return err
	}

	log.Tracef("Staging new multiset for the virtual block")
	csm.multisetStore.Stage(model.VirtualBlockHash, virtualMultiset)

	log.Tracef("Staging new UTXO diff for the virtual block")
	err = csm.consensusStateStore.StageVirtualUTXODiff(virtualUTXODiff)
	if err != nil {
		return err
	}

	log.Tracef("Updating the virtual diff parents after adding %s to the DAG", newBlockHash)
	err = csm.updateVirtualDiffParents(newBlockHash, virtualUTXODiff)
	if err != nil {
		return err
	}

	return nil
}

func (csm *consensusStateManager) updateVirtualDiffParents(
	newBlockHash *externalapi.DomainHash, virtualUTXODiff *model.UTXODiff) error {

	log.Tracef("updateVirtualDiffParents start for block %s", newBlockHash)
	defer log.Tracef("updateVirtualDiffParents end for block %s", newBlockHash)

	var newVirtualDiffParents []*externalapi.DomainHash
	if *newBlockHash == *csm.genesisHash {
		log.Tracef("Block %s is the genesis, so by definition "+
			"it is the only member of the new virtual diff parents set", newBlockHash)
		newVirtualDiffParents = []*externalapi.DomainHash{newBlockHash}
	} else {
		oldVirtualDiffParents, err := csm.consensusStateStore.VirtualDiffParents(csm.databaseContext)
		if err != nil {
			return err
		}
		log.Tracef("The old virtual's diff parents are: %s", oldVirtualDiffParents)

		// If the status of the new block is not `Valid` - virtualDiffParents didn't change
		status, err := csm.blockStatusStore.Get(csm.databaseContext, newBlockHash)
		if err != nil {
			return err
		}
		if status != externalapi.StatusValid {
			log.Tracef("The status of the new block %s is non-valid. "+
				"As such, don't change the diff parents of the virtual", newBlockHash)
			newVirtualDiffParents = oldVirtualDiffParents
		} else {
			log.Tracef("Block %s is valid. Updating the virtual diff parents", newBlockHash)
			newVirtualDiffParents = []*externalapi.DomainHash{newBlockHash}
			for _, virtualDiffParent := range oldVirtualDiffParents {
				isAncestorOfNewBlock, err := csm.dagTopologyManager.IsAncestorOf(virtualDiffParent, newBlockHash)
				if err != nil {
					return err
				}

				if !isAncestorOfNewBlock {
					newVirtualDiffParents = append(newVirtualDiffParents, virtualDiffParent)
				}
			}
		}
	}
	log.Tracef("The new virtual diff parents are: %s", newVirtualDiffParents)

	for _, virtualDiffParent := range newVirtualDiffParents {
		log.Tracef("Calculating new UTXO diff for virtual diff parent %s", virtualDiffParent)
		virtualDiffParentUTXODiff, err := csm.utxoDiffStore.UTXODiff(csm.databaseContext, virtualDiffParent)
		if err != nil {
			return err
		}
		newDiff, err := utxoalgebra.DiffFrom(virtualUTXODiff, virtualDiffParentUTXODiff)
		if err != nil {
			return err
		}
		log.Tracef("Staging new UTXO diff for virtual diff parent %s: %s", virtualDiffParent, newDiff)
		err = csm.utxoDiffStore.Stage(virtualDiffParent, newDiff, nil)
		if err != nil {
			return err
		}
	}

	log.Tracef("Staging the new virtual UTXO diff parents")
	return csm.consensusStateStore.StageVirtualDiffParents(newVirtualDiffParents)
}
