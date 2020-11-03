package consensusstatemanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/processes/consensusstatemanager/utxoalgebra"
	"github.com/kaspanet/kaspad/domain/consensus/utils/hashset"
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

	virtualUTXODiff, _, _, err := csm.calculatePastUTXOAndAcceptanceData(model.VirtualBlockHash)
	if err != nil {
		return err
	}

	csm.consensusStateStore.StageVirtualUTXODiff(virtualUTXODiff)

	err = csm.updateVirtualDiffParents(newBlockHash, virtualUTXODiff)
	if err != nil {
		return err
	}

	return nil
}

func (csm *consensusStateManager) updateVirtualDiffParents(
	newBlockHash *externalapi.DomainHash, virtualUTXODiff *model.UTXODiff) error {

	virtualDiffParents, err := csm.consensusStateStore.VirtualDiffParents(csm.databaseContext)
	if err != nil {
		return err
	}

	newBlockParentsSlice, err := csm.dagTopologyManager.Parents(newBlockHash)
	if err != nil {
		return err
	}
	newBlockParents := hashset.NewFromSlice(newBlockParentsSlice...)

	newVirtualDiffParents := []*externalapi.DomainHash{newBlockHash}
	for _, virtualDiffParent := range virtualDiffParents {
		if newBlockParents.Contains(virtualDiffParent) {
			virtualDiffParentUTXODiff, err := csm.utxoDiffStore.UTXODiff(csm.databaseContext, virtualDiffParent)
			if err != nil {
				return err
			}
			newDiff, err := utxoalgebra.DiffFrom(virtualUTXODiff, virtualDiffParentUTXODiff)
			if err != nil {
				return err
			}
			csm.utxoDiffStore.Stage(virtualDiffParent, newDiff, newBlockHash)
		} else {
			newVirtualDiffParents = append(newVirtualDiffParents, virtualDiffParent)
		}
	}

	csm.consensusStateStore.StageVirtualDiffParents(newVirtualDiffParents)

	return nil
}
