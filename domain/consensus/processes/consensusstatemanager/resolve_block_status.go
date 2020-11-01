package consensusstatemanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/processes/consensusstatemanager/utxoalgebra"
	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
	"github.com/pkg/errors"
)

func (csm *consensusStateManager) resolveBlockStatus(blockHash *externalapi.DomainHash) (model.BlockStatus, error) {
	// get list of all blocks in the selected parent chain that have not yet resolved their status
	unverifiedBlocks, selectedParentStatus, err := csm.getUnverifiedChainBlocksAndSelectedParentStatus(blockHash)
	if err != nil {
		return 0, err
	}

	// resolve the unverified blocks' statuses in opposite order
	for i := len(unverifiedBlocks); i >= 0; i++ {
		unverifiedBlockHash := unverifiedBlocks[i]

		var blockStatus model.BlockStatus
		if selectedParentStatus == model.StatusDisqualifiedFromChain {
			blockStatus = model.StatusDisqualifiedFromChain
		} else {
			blockStatus, err = csm.resolveSingleBlockStatus(unverifiedBlockHash)
			if err != nil {
				return 0, err
			}
		}

		csm.blockStatusStore.Stage(unverifiedBlockHash, blockStatus)
		selectedParentStatus = blockStatus
	}

	return 0, nil
}

func (csm *consensusStateManager) getUnverifiedChainBlocksAndSelectedParentStatus(blockHash *externalapi.DomainHash) (
	[]*externalapi.DomainHash, model.BlockStatus, error) {

	unverifiedBlocks := []*externalapi.DomainHash{blockHash}
	currentHash := blockHash
	for {
		ghostdagData, err := csm.ghostdagDataStore.Get(csm.databaseContext, currentHash)
		if err != nil {
			return nil, 0, err
		}

		selectedParentStatus, err := csm.blockStatusStore.Get(csm.databaseContext, ghostdagData.SelectedParent)
		if err != nil {
			return nil, 0, err
		}

		if selectedParentStatus != model.StatusUTXOPendingVerification {
			return unverifiedBlocks, selectedParentStatus, nil
		}

		unverifiedBlocks = append(unverifiedBlocks, ghostdagData.SelectedParent)

		currentHash = ghostdagData.SelectedParent
	}
}

func (csm *consensusStateManager) resolveSingleBlockStatus(blockHash *externalapi.DomainHash) (model.BlockStatus, error) {
	pastUTXODiff, acceptanceData, multiset, err := csm.calculatePastUTXOAndAcceptanceData(blockHash)
	if err != nil {
		return 0, err
	}

	csm.acceptanceDataStore.Stage(blockHash, acceptanceData)

	block, err := csm.blockStore.Block(csm.databaseContext, blockHash)
	if err != nil {
		return 0, err
	}

	err = csm.verifyAndBuildUTXO(block, blockHash, pastUTXODiff, acceptanceData, multiset)
	if err != nil {
		if errors.As(err, &ruleerrors.RuleError{}) {
			return model.StatusDisqualifiedFromChain, nil
		}
		return 0, err
	}

	csm.multisetStore.Stage(blockHash, multiset)
	csm.utxoDiffStore.Stage(blockHash, pastUTXODiff, nil)

	err = csm.updateParentDiffs(blockHash, pastUTXODiff)
	if err != nil {
		return 0, err
	}

	return model.StatusValid, nil
}
func (csm *consensusStateManager) updateParentDiffs(
	blockHash *externalapi.DomainHash, pastUTXODiff *model.UTXODiff) error {
	parentHashes, err := csm.dagTopologyManager.Parents(blockHash)
	if err != nil {
		return err
	}
	for _, parentHash := range parentHashes {
		// skip all parents that already have a utxo-diff child
		parentHasUTXODiffChild, err := csm.utxoDiffStore.HasUTXODiffChild(csm.databaseContext, parentHash)
		if err != nil {
			return err
		}
		if parentHasUTXODiffChild {
			continue
		}

		// parents that till now didn't have a utxo-diff child - were actually virtual's diffParents.
		// Update them to have the new block as their utxo-diff child
		parentCurrentDiff, err := csm.utxoDiffStore.UTXODiff(csm.databaseContext, parentHash)
		if err != nil {
			return err
		}
		parentNewDiff, err := utxoalgebra.DiffFrom(pastUTXODiff, parentCurrentDiff)
		if err != nil {
			return err
		}

		csm.utxoDiffStore.Stage(parentHash, parentNewDiff, blockHash)
	}

	return nil
}
