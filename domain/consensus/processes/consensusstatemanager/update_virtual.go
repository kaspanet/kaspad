package consensusstatemanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

func (csm *consensusStateManager) updateVirtual(newBlockHash *externalapi.DomainHash,
	tips []*externalapi.DomainHash) (*externalapi.SelectedParentChainChanges, error) {

	log.Tracef("updateVirtual start for block %s", newBlockHash)
	defer log.Tracef("updateVirtual end for block %s", newBlockHash)

	log.Tracef("Saving a reference to the GHOSTDAG data of the old virtual")
	var oldVirtualSelectedParent *externalapi.DomainHash
	if *newBlockHash != *csm.genesisHash {
		oldVirtualGHOSTDAGData, err := csm.ghostdagDataStore.Get(csm.databaseContext, model.VirtualBlockHash)
		if err != nil {
			return nil, err
		}
		oldVirtualSelectedParent = oldVirtualGHOSTDAGData.SelectedParent()
	}

	log.Tracef("Picking virtual parents from the tips: %s", tips)
	virtualParents, err := csm.pickVirtualParents(tips)
	if err != nil {
		return nil, err
	}
	log.Tracef("Picked virtual parents: %s", virtualParents)

	err = csm.dagTopologyManager.SetParents(model.VirtualBlockHash, virtualParents)
	if err != nil {
		return nil, err
	}
	log.Tracef("Set new parents for the virtual block hash")

	err = csm.ghostdagManager.GHOSTDAG(model.VirtualBlockHash)
	if err != nil {
		return nil, err
	}

	log.Tracef("Calculating past UTXO, acceptance data, and multiset for the new virtual block")
	virtualUTXODiff, virtualAcceptanceData, virtualMultiset, err := csm.CalculatePastUTXOAndAcceptanceData(model.VirtualBlockHash)
	if err != nil {
		return nil, err
	}

	log.Tracef("Staging new acceptance data for the virtual block")
	csm.acceptanceDataStore.Stage(model.VirtualBlockHash, virtualAcceptanceData)

	log.Tracef("Staging new multiset for the virtual block")
	csm.multisetStore.Stage(model.VirtualBlockHash, virtualMultiset)

	log.Tracef("Staging new UTXO diff for the virtual block")
	err = csm.consensusStateStore.StageVirtualUTXODiff(virtualUTXODiff)
	if err != nil {
		return nil, err
	}

	log.Tracef("Updating the virtual diff parents after adding %s to the DAG", newBlockHash)
	err = csm.updateVirtualDiffParents(virtualUTXODiff)
	if err != nil {
		return nil, err
	}

	log.Tracef("Calculating selected parent chain changes")
	var selectedParentChainChanges *externalapi.SelectedParentChainChanges
	if *newBlockHash != *csm.genesisHash {
		newVirtualGHOSTDAGData, err := csm.ghostdagDataStore.Get(csm.databaseContext, model.VirtualBlockHash)
		if err != nil {
			return nil, err
		}
		newVirtualSelectedParent := newVirtualGHOSTDAGData.SelectedParent()
		selectedParentChainChanges, err = csm.calculateSelectedParentChainChanges(oldVirtualSelectedParent, newVirtualSelectedParent)
		if err != nil {
			return nil, err
		}
	}

	return selectedParentChainChanges, nil
}

func (csm *consensusStateManager) updateVirtualDiffParents(virtualUTXODiff model.UTXODiff) error {
	log.Tracef("updateVirtualDiffParents start")
	defer log.Tracef("updateVirtualDiffParents end")

	virtualDiffParents, err := csm.consensusStateStore.VirtualDiffParents(csm.databaseContext)
	if err != nil {
		return err
	}

	for _, virtualDiffParent := range virtualDiffParents {
		log.Tracef("Calculating new UTXO diff for virtual diff parent %s", virtualDiffParent)
		virtualDiffParentUTXODiff, err := csm.utxoDiffStore.UTXODiff(csm.databaseContext, virtualDiffParent)
		if err != nil {
			return err
		}
		newDiff, err := virtualUTXODiff.DiffFrom(virtualDiffParentUTXODiff)
		if err != nil {
			return err
		}

		log.Tracef("Staging new UTXO diff for virtual diff parent %s: %s", virtualDiffParent, newDiff)
		err = csm.stageDiff(virtualDiffParent, newDiff, nil)
		if err != nil {
			return err
		}
	}

	return nil
}
