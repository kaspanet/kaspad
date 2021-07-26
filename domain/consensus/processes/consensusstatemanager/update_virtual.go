package consensusstatemanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/infrastructure/logger"
)

func (csm *consensusStateManager) updateVirtual(stagingArea *model.StagingArea, newBlockHash *externalapi.DomainHash,
	tips []*externalapi.DomainHash) (*externalapi.SelectedChainPath, externalapi.UTXODiff, error) {

	onEnd := logger.LogAndMeasureExecutionTime(log, "updateVirtual")
	defer onEnd()

	log.Debugf("updateVirtual start for block %s", newBlockHash)

	log.Debugf("Saving a reference to the GHOSTDAG data of the old virtual")
	var oldVirtualSelectedParent *externalapi.DomainHash
	if !newBlockHash.Equal(csm.genesisHash) {
		oldVirtualGHOSTDAGData, err := csm.ghostdagDataStore.Get(csm.databaseContext, stagingArea, model.VirtualBlockHash, false)
		if err != nil {
			return nil, nil, err
		}
		oldVirtualSelectedParent = oldVirtualGHOSTDAGData.SelectedParent()
	}

	log.Debugf("Picking virtual parents from tips len: %d", len(tips))
	virtualParents, err := csm.pickVirtualParents(stagingArea, tips)
	if err != nil {
		return nil, nil, err
	}
	log.Debugf("Picked virtual parents: %s", virtualParents)

	err = csm.dagTopologyManager.SetParents(stagingArea, model.VirtualBlockHash, virtualParents)
	if err != nil {
		return nil, nil, err
	}
	log.Debugf("Set new parents for the virtual block hash")

	err = csm.ghostdagManager.GHOSTDAG(stagingArea, model.VirtualBlockHash)
	if err != nil {
		return nil, nil, err
	}

	// This is needed for `csm.CalculatePastUTXOAndAcceptanceData`
	_, err = csm.difficultyManager.StageDAADataAndReturnRequiredDifficulty(stagingArea, model.VirtualBlockHash, false)
	if err != nil {
		return nil, nil, err
	}

	log.Debugf("Calculating past UTXO, acceptance data, and multiset for the new virtual block")
	virtualUTXODiff, virtualAcceptanceData, virtualMultiset, err :=
		csm.CalculatePastUTXOAndAcceptanceData(stagingArea, model.VirtualBlockHash)
	if err != nil {
		return nil, nil, err
	}

	log.Debugf("Calculated the past UTXO of the new virtual. "+
		"Diff toAdd length: %d, toRemove length: %d",
		virtualUTXODiff.ToAdd().Len(), virtualUTXODiff.ToRemove().Len())

	log.Debugf("Staging new acceptance data for the virtual block")
	csm.acceptanceDataStore.Stage(stagingArea, model.VirtualBlockHash, virtualAcceptanceData)

	log.Debugf("Staging new multiset for the virtual block")
	csm.multisetStore.Stage(stagingArea, model.VirtualBlockHash, virtualMultiset)

	log.Debugf("Staging new UTXO diff for the virtual block")
	csm.consensusStateStore.StageVirtualUTXODiff(stagingArea, virtualUTXODiff)

	log.Debugf("Updating the selected tip's utxo-diff after adding %s to the DAG", newBlockHash)
	err = csm.updateSelectedTipUTXODiff(stagingArea, virtualUTXODiff)
	if err != nil {
		return nil, nil, err
	}

	log.Debugf("Calculating selected parent chain changes")
	var selectedParentChainChanges *externalapi.SelectedChainPath
	if !newBlockHash.Equal(csm.genesisHash) {
		newVirtualGHOSTDAGData, err := csm.ghostdagDataStore.Get(csm.databaseContext, stagingArea, model.VirtualBlockHash, false)
		if err != nil {
			return nil, nil, err
		}
		newVirtualSelectedParent := newVirtualGHOSTDAGData.SelectedParent()
		selectedParentChainChanges, err = csm.dagTraversalManager.
			CalculateChainPath(stagingArea, oldVirtualSelectedParent, newVirtualSelectedParent)
		if err != nil {
			return nil, nil, err
		}
		log.Debugf("Selected parent chain changes: %d blocks were removed and %d blocks were added",
			len(selectedParentChainChanges.Removed), len(selectedParentChainChanges.Added))
	}

	return selectedParentChainChanges, virtualUTXODiff, nil
}

func (csm *consensusStateManager) updateSelectedTipUTXODiff(
	stagingArea *model.StagingArea, virtualUTXODiff externalapi.UTXODiff) error {

	onEnd := logger.LogAndMeasureExecutionTime(log, "updateSelectedTipUTXODiff")
	defer onEnd()

	selectedTip, err := csm.selectedTip(stagingArea)
	if err != nil {
		return err
	}

	log.Debugf("Calculating new UTXO diff for virtual diff parent %s", selectedTip)
	selectedTipUTXODiff, err := csm.utxoDiffStore.UTXODiff(csm.databaseContext, stagingArea, selectedTip)
	if err != nil {
		return err
	}
	newDiff, err := virtualUTXODiff.DiffFrom(selectedTipUTXODiff)
	if err != nil {
		return err
	}

	log.Debugf("Staging new UTXO diff for virtual diff parent %s", selectedTip)
	csm.stageDiff(stagingArea, selectedTip, newDiff, nil)

	return nil
}
