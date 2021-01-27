package consensusstatemanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

func (csm *consensusStateManager) updateVirtual(newBlockHash *externalapi.DomainHash,
	tips []*externalapi.DomainHash) (*externalapi.SelectedChainPath, error) {

	log.Debugf("updateVirtual start for block %s", newBlockHash)
	defer log.Debugf("updateVirtual end for block %s", newBlockHash)

	log.Debugf("Saving a reference to the GHOSTDAG data of the old virtual")
	var oldVirtualSelectedParent *externalapi.DomainHash
	if !newBlockHash.Equal(csm.genesisHash) {
		oldVirtualGHOSTDAGData, err := csm.ghostdagDataStore.Get(csm.databaseContext, model.VirtualBlockHash)
		if err != nil {
			return nil, err
		}
		oldVirtualSelectedParent = oldVirtualGHOSTDAGData.SelectedParent()
	}

	log.Debugf("Picking virtual parents from tips len: %d", len(tips))
	virtualParents, err := csm.pickVirtualParents(tips)
	if err != nil {
		return nil, err
	}
	log.Debugf("Picked virtual parents: %s", virtualParents)

	err = csm.dagTopologyManager.SetParents(model.VirtualBlockHash, virtualParents)
	if err != nil {
		return nil, err
	}
	log.Debugf("Set new parents for the virtual block hash")

	err = csm.ghostdagManager.GHOSTDAG(model.VirtualBlockHash)
	if err != nil {
		return nil, err
	}

	log.Debugf("Calculating past UTXO, acceptance data, and multiset for the new virtual block")
	virtualUTXODiff, virtualAcceptanceData, virtualMultiset, err := csm.CalculatePastUTXOAndAcceptanceData(model.VirtualBlockHash)
	if err != nil {
		return nil, err
	}

	log.Debugf("Staging new acceptance data for the virtual block")
	csm.acceptanceDataStore.Stage(model.VirtualBlockHash, virtualAcceptanceData)

	log.Debugf("Staging new multiset for the virtual block")
	csm.multisetStore.Stage(model.VirtualBlockHash, virtualMultiset)

	log.Debugf("Staging new UTXO diff for the virtual block")
	csm.consensusStateStore.StageVirtualUTXODiff(virtualUTXODiff)

	log.Debugf("Updating the virtual diff parents after adding %s to the DAG", newBlockHash)
	err = csm.updateVirtualDiffParents(virtualUTXODiff)
	if err != nil {
		return nil, err
	}

	log.Debugf("Calculating selected parent chain changes")
	var selectedParentChainChanges *externalapi.SelectedChainPath
	if !newBlockHash.Equal(csm.genesisHash) {
		newVirtualGHOSTDAGData, err := csm.ghostdagDataStore.Get(csm.databaseContext, model.VirtualBlockHash)
		if err != nil {
			return nil, err
		}
		newVirtualSelectedParent := newVirtualGHOSTDAGData.SelectedParent()
		selectedParentChainChanges, err = csm.dagTraversalManager.
			CalculateChainPath(oldVirtualSelectedParent, newVirtualSelectedParent)
		if err != nil {
			return nil, err
		}
	}

	return selectedParentChainChanges, nil
}

func (csm *consensusStateManager) updateVirtualDiffParents(virtualUTXODiff model.UTXODiff) error {
	log.Debugf("updateVirtualDiffParents start")
	defer log.Debugf("updateVirtualDiffParents end")

	virtualDiffParents, err := csm.consensusStateStore.VirtualDiffParents(csm.databaseContext)
	if err != nil {
		return err
	}

	for _, virtualDiffParent := range virtualDiffParents {
		log.Debugf("Calculating new UTXO diff for virtual diff parent %s", virtualDiffParent)
		virtualDiffParentUTXODiff, err := csm.utxoDiffStore.UTXODiff(csm.databaseContext, virtualDiffParent)
		if err != nil {
			return err
		}
		newDiff, err := virtualUTXODiff.DiffFrom(virtualDiffParentUTXODiff)
		if err != nil {
			return err
		}

		log.Debugf("Staging new UTXO diff for virtual diff parent %s", virtualDiffParent)
		err = csm.stageDiff(virtualDiffParent, newDiff, nil)
		if err != nil {
			return err
		}
	}

	return nil
}
