package consensusstatemanager

import (
	"fmt"

	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
	"github.com/kaspanet/kaspad/infrastructure/logger"
	"github.com/pkg/errors"
)

func (csm *consensusStateManager) resolveBlockStatus(blockHash *externalapi.DomainHash) (externalapi.BlockStatus, error) {
	onEnd := logger.LogAndMeasureExecutionTime(log, fmt.Sprintf("resolveBlockStatus for %s", blockHash))
	defer onEnd()

	log.Debugf("Getting a list of all blocks in the selected "+
		"parent chain of %s that have no yet resolved their status", blockHash)
	unverifiedBlocks, err := csm.getUnverifiedChainBlocks(blockHash)
	if err != nil {
		return 0, err
	}
	log.Debugf("Got %d unverified blocks in the selected parent "+
		"chain of %s: %s", len(unverifiedBlocks), blockHash, unverifiedBlocks)

	// If there's no unverified blocks in the given block's chain - this means the given block already has a
	// UTXO-verified status, and therefore it should be retrieved from the store and returned
	if len(unverifiedBlocks) == 0 {
		log.Debugf("There are not unverified blocks in %s's selected parent chain. "+
			"This means that the block already has a UTXO-verified status.", blockHash)
		status, err := csm.blockStatusStore.Get(csm.databaseContext, nil, blockHash)
		if err != nil {
			return 0, err
		}
		log.Debugf("Block %s's status resolved to: %s", blockHash, status)
		return status, nil
	}

	log.Debugf("Finding the status of the selected parent of %s", blockHash)
	selectedParentStatus, err := csm.findSelectedParentStatus(unverifiedBlocks)
	if err != nil {
		return 0, err
	}
	log.Debugf("The status of the selected parent of %s is: %s", blockHash, selectedParentStatus)

	log.Debugf("Resolving the unverified blocks' status in reverse order (past to present)")
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

		csm.blockStatusStore.Stage(nil, unverifiedBlockHash, blockStatus)
		selectedParentStatus = blockStatus
		log.Debugf("Block %s status resolved to `%s`, finished %d/%d of unverified blocks",
			unverifiedBlockHash, blockStatus, len(unverifiedBlocks)-i, len(unverifiedBlocks))
	}

	return blockStatus, nil
}

// findSelectedParentStatus returns the status of the selectedParent of the last block in the unverifiedBlocks chain
func (csm *consensusStateManager) findSelectedParentStatus(unverifiedBlocks []*externalapi.DomainHash) (
	externalapi.BlockStatus, error) {

	log.Debugf("findSelectedParentStatus start")
	defer log.Debugf("findSelectedParentStatus end")

	lastUnverifiedBlock := unverifiedBlocks[len(unverifiedBlocks)-1]
	if lastUnverifiedBlock.Equal(csm.genesisHash) {
		log.Debugf("the most recent unverified block is the genesis block, "+
			"which by definition has status: %s", externalapi.StatusUTXOValid)
		return externalapi.StatusUTXOValid, nil
	}
	lastUnverifiedBlockGHOSTDAGData, err := csm.ghostdagDataStore.Get(csm.databaseContext, nil, lastUnverifiedBlock)
	if err != nil {
		return 0, err
	}
	return csm.blockStatusStore.Get(csm.databaseContext, nil, lastUnverifiedBlockGHOSTDAGData.SelectedParent())
}

func (csm *consensusStateManager) getUnverifiedChainBlocks(
	blockHash *externalapi.DomainHash) ([]*externalapi.DomainHash, error) {

	log.Debugf("getUnverifiedChainBlocks start for block %s", blockHash)
	defer log.Debugf("getUnverifiedChainBlocks end for block %s", blockHash)

	var unverifiedBlocks []*externalapi.DomainHash
	currentHash := blockHash
	for {
		log.Debugf("Getting status for block %s", currentHash)
		currentBlockStatus, err := csm.blockStatusStore.Get(csm.databaseContext, nil, currentHash)
		if err != nil {
			return nil, err
		}
		if currentBlockStatus != externalapi.StatusUTXOPendingVerification {
			log.Debugf("Block %s has status %s. Returning all the "+
				"unverified blocks prior to it: %s", currentHash, currentBlockStatus, unverifiedBlocks)
			return unverifiedBlocks, nil
		}

		log.Debugf("Block %s is unverified. Adding it to the unverified block collection", currentHash)
		unverifiedBlocks = append(unverifiedBlocks, currentHash)

		currentBlockGHOSTDAGData, err := csm.ghostdagDataStore.Get(csm.databaseContext, nil, currentHash)
		if err != nil {
			return nil, err
		}

		if currentBlockGHOSTDAGData.SelectedParent() == nil {
			log.Debugf("Genesis block reached. Returning all the "+
				"unverified blocks prior to it: %s", unverifiedBlocks)
			return unverifiedBlocks, nil
		}

		currentHash = currentBlockGHOSTDAGData.SelectedParent()
	}
}

func (csm *consensusStateManager) resolveSingleBlockStatus(blockHash *externalapi.DomainHash) (externalapi.BlockStatus, error) {
	onEnd := logger.LogAndMeasureExecutionTime(log, fmt.Sprintf("resolveSingleBlockStatus for %s", blockHash))
	defer onEnd()

	log.Tracef("Calculating pastUTXO and acceptance data and multiset for block %s", blockHash)
	pastUTXODiff, acceptanceData, multiset, err := csm.CalculatePastUTXOAndAcceptanceData(blockHash)
	if err != nil {
		return 0, err
	}

	log.Tracef("Staging the calculated acceptance data of block %s", blockHash)
	csm.acceptanceDataStore.Stage(nil, blockHash, acceptanceData)

	block, err := csm.blockStore.Block(csm.databaseContext,, blockHash)
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

	if csm.genesisHash.Equal(blockHash) {
		log.Tracef("Staging the utxoDiff of genesis")
		csm.stageDiff(blockHash, pastUTXODiff, nil)
		return externalapi.StatusUTXOValid, nil
	}

	oldSelectedTip, err := csm.selectedTip()
	if err != nil {
		return 0, err
	}

	isNewSelectedTip, err := csm.isNewSelectedTip(blockHash, oldSelectedTip)
	if err != nil {
		return 0, err
	}
	oldSelectedTipUTXOSet, err := csm.restorePastUTXO(oldSelectedTip)
	if err != nil {
		return 0, err
	}
	if isNewSelectedTip {
		log.Debugf("Block %s is the new SelectedTip, therefore setting it as old selectedTip's diffChild", blockHash)
		oldSelectedTipUTXOSet, err := pastUTXODiff.DiffFrom(oldSelectedTipUTXOSet.ToImmutable())
		if err != nil {
			return 0, err
		}
		csm.stageDiff(oldSelectedTip, oldSelectedTipUTXOSet, blockHash)

		log.Tracef("Staging the utxoDiff of block %s", blockHash)
		csm.stageDiff(blockHash, pastUTXODiff, nil)
	} else {
		log.Debugf("Block %s is not the new SelectedTip, therefore setting old selectedTip as it's diffChild", blockHash)
		pastUTXODiff, err = oldSelectedTipUTXOSet.DiffFrom(pastUTXODiff)
		if err != nil {
			return 0, err
		}

		log.Tracef("Staging the utxoDiff of block %s", blockHash)
		csm.stageDiff(blockHash, pastUTXODiff, oldSelectedTip)
	}

	return externalapi.StatusUTXOValid, nil
}

func (csm *consensusStateManager) isNewSelectedTip(blockHash, oldSelectedTip *externalapi.DomainHash) (bool, error) {
	newSelectedTip, err := csm.ghostdagManager.ChooseSelectedParent(blockHash, oldSelectedTip)
	if err != nil {
		return false, err
	}

	return blockHash.Equal(newSelectedTip), nil
}

func (csm *consensusStateManager) selectedTip() (*externalapi.DomainHash, error) {
	virtualGHOSTDAGData, err := csm.ghostdagDataStore.Get(csm.databaseContext, nil, model.VirtualBlockHash)
	if err != nil {
		return nil, err
	}

	return virtualGHOSTDAGData.SelectedParent(), nil
}
