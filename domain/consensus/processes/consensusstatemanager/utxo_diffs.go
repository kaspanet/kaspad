package consensusstatemanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
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
	var virtualDiffParents []*externalapi.DomainHash
	if *blockHash != *csm.genesisHash {
		var err error
		virtualDiffParents, err = csm.consensusStateStore.VirtualDiffParents(csm.databaseContext)
		if err != nil {
			return err
		}
	}

	virtualDiffParents = append(virtualDiffParents, blockHash)
	return csm.consensusStateStore.StageVirtualDiffParents(virtualDiffParents)
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

	return csm.consensusStateStore.StageVirtualDiffParents(newVirtualDiffParents)
}
