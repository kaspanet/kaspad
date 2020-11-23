package consensusstatemanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/processes/consensusstatemanager/utxoalgebra"
)

func (csm *consensusStateManager) updateVirtual(newBlockHash *externalapi.DomainHash, tips []*externalapi.DomainHash) error {
	virtualParents, err := csm.pickVirtualParents(tips)
	if err != nil {
		return err
	}

	err = csm.dagTopologyManager.SetParents(model.VirtualBlockHash, virtualParents)
	if err != nil {
		return err
	}

	err = csm.ghostdagManager.GHOSTDAG(model.VirtualBlockHash)
	if err != nil {
		return err
	}

	virtualUTXODiff, virtualAcceptanceData, virtualMultiset, err := csm.CalculatePastUTXOAndAcceptanceData(model.VirtualBlockHash)
	if err != nil {
		return err
	}

	err = csm.acceptanceDataStore.Stage(model.VirtualBlockHash, virtualAcceptanceData)
	if err != nil {
		return err
	}

	csm.multisetStore.Stage(model.VirtualBlockHash, virtualMultiset)

	err = csm.consensusStateStore.StageVirtualUTXODiff(virtualUTXODiff)
	if err != nil {
		return err
	}

	err = csm.updateVirtualDiffParents(virtualUTXODiff)
	if err != nil {
		return err
	}

	return nil
}

func (csm *consensusStateManager) updateVirtualDiffParents(virtualUTXODiff *model.UTXODiff) error {

	virtualDiffParents, err := csm.consensusStateStore.VirtualDiffParents(csm.databaseContext)
	if err != nil {
		return err
	}

	for _, virtualDiffParent := range virtualDiffParents {
		virtualDiffParentUTXODiff, err := csm.utxoDiffStore.UTXODiff(csm.databaseContext, virtualDiffParent)
		if err != nil {
			return err
		}
		newDiff, err := utxoalgebra.DiffFrom(virtualUTXODiff, virtualDiffParentUTXODiff)
		if err != nil {
			return err
		}
		err = csm.stageDiff(virtualDiffParent, newDiff, nil)
		if err != nil {
			return err
		}
	}

	return nil
}
