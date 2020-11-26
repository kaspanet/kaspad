package consensusstatemanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
	"github.com/kaspanet/kaspad/domain/consensus/utils/utxo/utxoalgebra"
	"github.com/pkg/errors"
)

func (csm *consensusStateManager) resolveBlockStatus(blockHash *externalapi.DomainHash) (externalapi.BlockStatus, error) {
	log.Tracef("resolveBlockStatus start for block %s", blockHash)
	defer log.Tracef("resolveBlockStatus end for block %s", blockHash)

	log.Tracef("Getting a list of all blocks in the selected "+
		"parent chain of %s that have no yet resolved their status", blockHash)
	unverifiedBlocks, err := csm.getUnverifiedChainBlocks(blockHash)
	if err != nil {
		return 0, err
	}
	log.Tracef("Got %d unverified blocks in the selected parent "+
		"chain of %s: %s", len(unverifiedBlocks), blockHash, unverifiedBlocks)

	// If there's no unverified blocks in the given block's chain - this means the given block already has a
	// UTXO-verified status, and therefore it should be retrieved from the store and returned
	if len(unverifiedBlocks) == 0 {
		log.Tracef("There are not unverified blocks in %s's selected parent chain. "+
			"This means that the block already has a UTXO-verified status.", blockHash)
		status, err := csm.blockStatusStore.Get(csm.databaseContext, blockHash)
		if err != nil {
			return 0, err
		}
		log.Tracef("Block %s's status resolved to: %s", blockHash, status)
		return status, nil
	}

	log.Tracef("Finding the status of the selected parent of %s", blockHash)
	selectedParentStatus, err := csm.findSelectedParentStatus(unverifiedBlocks)
	if err != nil {
		return 0, err
	}
	log.Tracef("The status of the selected parent of %s is: %s", blockHash, selectedParentStatus)

	log.Tracef("Resolving the unverified blocks' status in reverse order (past to present)")
	var blockStatus externalapi.BlockStatus
	for i := len(unverifiedBlocks) - 1; i >= 0; i-- {
		unverifiedBlockHash := unverifiedBlocks[i]

		if selectedParentStatus == externalapi.StatusDisqualifiedFromChain {
			blockStatus = externalapi.StatusDisqualifiedFromChain
		} else {
			blockStatus, err = csm.resolveSingleBlockStatus(unverifiedBlockHash)
			if err != nil {
				return 0, err
			}
		}

		csm.blockStatusStore.Stage(unverifiedBlockHash, blockStatus)
		selectedParentStatus = blockStatus
		log.Debugf("Block %s status resolved to `%s`", unverifiedBlockHash, blockStatus)
	}

	return blockStatus, nil
}

// findSelectedParentStatus returns the status of the selectedParent of the last block in the unverifiedBlocks chain
func (csm *consensusStateManager) findSelectedParentStatus(unverifiedBlocks []*externalapi.DomainHash) (
	externalapi.BlockStatus, error) {

	log.Tracef("findSelectedParentStatus start")
	defer log.Tracef("findSelectedParentStatus end")

	lastUnverifiedBlock := unverifiedBlocks[len(unverifiedBlocks)-1]
	if lastUnverifiedBlock.Equal(csm.genesisHash) {
		log.Tracef("the most recent unverified block is the genesis block, "+
			"which by definition has status: %s", externalapi.StatusValid)
		return externalapi.StatusValid, nil
	}
	lastUnverifiedBlockGHOSTDAGData, err := csm.ghostdagDataStore.Get(csm.databaseContext, lastUnverifiedBlock)
	if err != nil {
		return 0, err
	}
	return csm.blockStatusStore.Get(csm.databaseContext, lastUnverifiedBlockGHOSTDAGData.SelectedParent)
}

func (csm *consensusStateManager) getUnverifiedChainBlocks(
	blockHash *externalapi.DomainHash) ([]*externalapi.DomainHash, error) {

	log.Tracef("getUnverifiedChainBlocks start for block %s", blockHash)
	defer log.Tracef("getUnverifiedChainBlocks end for block %s", blockHash)

	var unverifiedBlocks []*externalapi.DomainHash
	currentHash := blockHash
	for {
		log.Tracef("Getting status for block %s", currentHash)
		currentBlockStatus, err := csm.blockStatusStore.Get(csm.databaseContext, currentHash)
		if err != nil {
			return nil, err
		}
		if currentBlockStatus != externalapi.StatusUTXOPendingVerification {
			log.Tracef("Block %s has status %s. Returning all the "+
				"unverified blocks prior to it: %s", currentHash, currentBlockStatus, unverifiedBlocks)
			return unverifiedBlocks, nil
		}

		log.Tracef("Block %s is unverified. Adding it to the unverified block collection", currentHash)
		unverifiedBlocks = append(unverifiedBlocks, currentHash)

		currentBlockGHOSTDAGData, err := csm.ghostdagDataStore.Get(csm.databaseContext, currentHash)
		if err != nil {
			return nil, err
		}

		if currentBlockGHOSTDAGData.SelectedParent == nil {
			log.Tracef("Genesis block reached. Returning all the "+
				"unverified blocks prior to it: %s", unverifiedBlocks)
			return unverifiedBlocks, nil
		}

		currentHash = currentBlockGHOSTDAGData.SelectedParent
	}
}

func (csm *consensusStateManager) resolveSingleBlockStatus(blockHash *externalapi.DomainHash) (externalapi.BlockStatus, error) {
	log.Tracef("resolveSingleBlockStatus start for block %s", blockHash)
	defer log.Tracef("resolveSingleBlockStatus end for block %s", blockHash)

	log.Tracef("Calculating pastUTXO and acceptance data and multiset for block %s", blockHash)
	pastUTXODiff, acceptanceData, multiset, err := csm.CalculatePastUTXOAndAcceptanceData(blockHash)
	if err != nil {
		return 0, err
	}

	log.Tracef("Staging the calculated acceptance data of block %s", blockHash)
	csm.acceptanceDataStore.Stage(blockHash, acceptanceData)

	block, err := csm.blockStore.Block(csm.databaseContext, blockHash)
	if err != nil {
		return 0, err
	}

	log.Tracef("verifying the UTXO of block %s", blockHash)
	err = csm.verifyUTXO(block, blockHash, pastUTXODiff, acceptanceData, multiset)
	if err != nil {
		if errors.As(err, &ruleerrors.RuleError{}) {
			log.Debugf("UTXO verification for block %s failed: %s", blockHash, err)
			return externalapi.StatusDisqualifiedFromChain, nil
		}
		return 0, err
	}
	log.Debugf("UTXO verification for block %s passed", blockHash)

	log.Tracef("Staging the multiset of block %s", blockHash)
	csm.multisetStore.Stage(blockHash, multiset)

	log.Tracef("Staging the utxoDiff of block %s", blockHash)
	err = csm.stageDiff(blockHash, pastUTXODiff, nil)
	if err != nil {
		return 0, err
	}

	log.Tracef("Remove block ancestors from virtual diff parents and assign %s as their diff child", blockHash)
	err = csm.removeAncestorsFromVirtualDiffParentsAndAssignDiffChild(blockHash, pastUTXODiff)
	if err != nil {
		return 0, err
	}

	return externalapi.StatusValid, nil
}

func (csm *consensusStateManager) removeAncestorsFromVirtualDiffParentsAndAssignDiffChild(
	blockHash *externalapi.DomainHash, pastUTXODiff *model.UTXODiff) error {

	log.Tracef("removeAncestorsFromVirtualDiffParentsAndAssignDiffChild start for block %s", blockHash)
	defer log.Tracef("removeAncestorsFromVirtualDiffParentsAndAssignDiffChild end for block %s", blockHash)

	if blockHash.Equal(csm.genesisHash) {
		log.Tracef("Genesis block doesn't have ancestors to remove from the virtual diff parents")
		return nil
	}

	virtualDiffParents, err := csm.consensusStateStore.VirtualDiffParents(csm.databaseContext)
	if err != nil {
		return err
	}

	for _, virtualDiffParent := range virtualDiffParents {
		if virtualDiffParent.Equal(blockHash) {
			log.Tracef("Skipping updating virtual diff parent %s "+
				"because it was updated before.", virtualDiffParent)
			continue
		}

		isAncestorOfBlock, err := csm.dagTopologyManager.IsAncestorOf(virtualDiffParent, blockHash)
		if err != nil {
			return err
		}

		if !isAncestorOfBlock {
			log.Tracef("Skipping block %s because it's not an "+
				"ancestor of %s", virtualDiffParent, blockHash)
			continue
		}

		// parents that didn't have a utxo-diff child until now were actually virtual's diffParents.
		// Update them to have the new block as their utxo-diff child
		log.Tracef("Updating %s to be the diff child of %s", blockHash, virtualDiffParent)
		currentDiff, err := csm.utxoDiffStore.UTXODiff(csm.databaseContext, virtualDiffParent)
		if err != nil {
			return err
		}
		newDiff, err := utxoalgebra.DiffFrom(pastUTXODiff, currentDiff)
		if err != nil {
			return err
		}

		err = csm.stageDiff(virtualDiffParent, newDiff, blockHash)
		if err != nil {
			return err
		}
	}

	return nil
}
