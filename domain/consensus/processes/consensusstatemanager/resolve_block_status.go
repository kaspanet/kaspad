package consensusstatemanager

import (
	"fmt"

	"github.com/kaspanet/kaspad/domain/consensus/utils/utxo"

	"github.com/kaspanet/kaspad/util/staging"

	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
	"github.com/kaspanet/kaspad/infrastructure/logger"
	"github.com/pkg/errors"
)

func (csm *consensusStateManager) resolveBlockStatus(stagingArea *model.StagingArea, blockHash *externalapi.DomainHash,
	useSeparateStagingAreaPerBlock bool) (externalapi.BlockStatus, error) {

	onEnd := logger.LogAndMeasureExecutionTime(log, fmt.Sprintf("resolveBlockStatus for %s", blockHash))
	defer onEnd()

	log.Debugf("Getting a list of all blocks in the selected "+
		"parent chain of %s that have no yet resolved their status", blockHash)
	unverifiedBlocks, err := csm.getUnverifiedChainBlocks(stagingArea, blockHash)
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
		status, err := csm.blockStatusStore.Get(csm.databaseContext, stagingArea, blockHash)
		if err != nil {
			return 0, err
		}
		log.Debugf("Block %s's status resolved to: %s", blockHash, status)
		return status, nil
	}

	log.Debugf("Finding the status of the selected parent of %s", blockHash)
	selectedParentHash, selectedParentStatus, selectedParentUTXOSet, err := csm.selectedParentInfo(stagingArea, unverifiedBlocks)
	if err != nil {
		return 0, err
	}
	log.Debugf("The status of the selected parent of %s is: %s", blockHash, selectedParentStatus)

	log.Debugf("Resolving the unverified blocks' status in reverse order (past to present)")
	var blockStatus externalapi.BlockStatus

	previousBlockHash := selectedParentHash
	previousBlockUTXOSet := selectedParentUTXOSet
	var lastResolvedBlockUTXOSet externalapi.UTXODiff

	for i := len(unverifiedBlocks) - 1; i >= 0; i-- {
		unverifiedBlockHash := unverifiedBlocks[i]

		stagingAreaForCurrentBlock := stagingArea
		isResolveTip := i == 0
		useSeparateStagingArea := useSeparateStagingAreaPerBlock && !isResolveTip
		if useSeparateStagingArea {
			stagingAreaForCurrentBlock = model.NewStagingArea()
		}

		if selectedParentStatus == externalapi.StatusDisqualifiedFromChain {
			blockStatus = externalapi.StatusDisqualifiedFromChain
		} else {
			lastResolvedBlockUTXOSet = previousBlockUTXOSet

			blockStatus, previousBlockUTXOSet, err = csm.resolveSingleBlockStatus(
				stagingAreaForCurrentBlock, unverifiedBlockHash, previousBlockHash, previousBlockUTXOSet, isResolveTip)
			if err != nil {
				return 0, err
			}
		}

		csm.blockStatusStore.Stage(stagingAreaForCurrentBlock, unverifiedBlockHash, blockStatus)
		selectedParentStatus = blockStatus
		log.Debugf("Block %s status resolved to `%s`, finished %d/%d of unverified blocks",
			unverifiedBlockHash, blockStatus, len(unverifiedBlocks)-i, len(unverifiedBlocks))

		if useSeparateStagingArea {
			err := staging.CommitAllChanges(csm.databaseContext, stagingAreaForCurrentBlock)
			if err != nil {
				return 0, err
			}
		}
		previousBlockHash = unverifiedBlockHash
	}

	tipUTXOSet := previousBlockUTXOSet
	if blockStatus == externalapi.StatusUTXOValid {
		log.Debugf("Reversing UTXODiffs of blocks below the tip")
		// During resolveSingleBlockStatus, all unverifiedBlocks (excluding the tip) were assigned their selectedParent
		// as their UTXODiffChild.
		// Now that the whole chain has been resolved - we can reverse the UTXODiffs, to create shorter UTXODiffChild paths.
		err = csm.reverseUTXODiffs(stagingArea, unverifiedBlocks, tipUTXOSet, lastResolvedBlockUTXOSet, useSeparateStagingAreaPerBlock)
		if err != nil {
			return 0, err
		}
	}

	return blockStatus, nil
}

// selectedParentInfo returns the hash and status of the selectedParent of the last block in the unverifiedBlocks
// chain, in addition, if the status is UTXOValid, it return it's pastUTXOSet
func (csm *consensusStateManager) selectedParentInfo(
	stagingArea *model.StagingArea, unverifiedBlocks []*externalapi.DomainHash) (
	*externalapi.DomainHash, externalapi.BlockStatus, externalapi.UTXODiff, error) {

	log.Debugf("findSelectedParentStatus start")
	defer log.Debugf("findSelectedParentStatus end")

	lastUnverifiedBlock := unverifiedBlocks[len(unverifiedBlocks)-1]
	if lastUnverifiedBlock.Equal(csm.genesisHash) {
		log.Debugf("the most recent unverified block is the genesis block, "+
			"which by definition has status: %s", externalapi.StatusUTXOValid)
		return lastUnverifiedBlock, externalapi.StatusUTXOValid, utxo.NewUTXODiff(), nil
	}
	lastUnverifiedBlockGHOSTDAGData, err := csm.ghostdagDataStore.Get(csm.databaseContext, stagingArea, lastUnverifiedBlock)
	if err != nil {
		return nil, 0, nil, err
	}
	selectedParent := lastUnverifiedBlockGHOSTDAGData.SelectedParent()
	selectedParentStatus, err := csm.blockStatusStore.Get(csm.databaseContext, stagingArea, selectedParent)
	if err != nil {
		return nil, 0, nil, err
	}
	if selectedParentStatus != externalapi.StatusUTXOValid {
		return selectedParent, selectedParentStatus, nil, nil
	}

	selectedParentUTXOSet, err := csm.restorePastUTXO(stagingArea, selectedParent)
	if err != nil {
		return nil, 0, nil, err
	}
	return selectedParent, selectedParentStatus, selectedParentUTXOSet, nil
}

func (csm *consensusStateManager) getUnverifiedChainBlocks(stagingArea *model.StagingArea,
	blockHash *externalapi.DomainHash) ([]*externalapi.DomainHash, error) {

	log.Debugf("getUnverifiedChainBlocks start for block %s", blockHash)
	defer log.Debugf("getUnverifiedChainBlocks end for block %s", blockHash)

	var unverifiedBlocks []*externalapi.DomainHash
	currentHash := blockHash
	for {
		log.Debugf("Getting status for block %s", currentHash)
		currentBlockStatus, err := csm.blockStatusStore.Get(csm.databaseContext, stagingArea, currentHash)
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

		currentBlockGHOSTDAGData, err := csm.ghostdagDataStore.Get(csm.databaseContext, stagingArea, currentHash)
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

func (csm *consensusStateManager) resolveSingleBlockStatus(stagingArea *model.StagingArea,
	blockHash, selectedParentHash *externalapi.DomainHash, selectedParentPastUTXOSet externalapi.UTXODiff, isResolveTip bool) (
	externalapi.BlockStatus, externalapi.UTXODiff, error) {

	onEnd := logger.LogAndMeasureExecutionTime(log, fmt.Sprintf("resolveSingleBlockStatus for %s", blockHash))
	defer onEnd()

	log.Tracef("Calculating pastUTXO and acceptance data and multiset for block %s", blockHash)
	pastUTXOSet, acceptanceData, multiset, err := csm.calculatePastUTXOAndAcceptanceDataWithSelectedParentUTXO(
		stagingArea, blockHash, selectedParentPastUTXOSet)
	if err != nil {
		return 0, nil, err
	}

	log.Tracef("Staging the calculated acceptance data of block %s", blockHash)
	csm.acceptanceDataStore.Stage(stagingArea, blockHash, acceptanceData)

	block, err := csm.blockStore.Block(csm.databaseContext, stagingArea, blockHash)
	if err != nil {
		return 0, nil, err
	}

	log.Tracef("verifying the UTXO of block %s", blockHash)
	err = csm.verifyUTXO(stagingArea, block, blockHash, pastUTXOSet, acceptanceData, multiset)
	if err != nil {
		if errors.As(err, &ruleerrors.RuleError{}) {
			log.Debugf("UTXO verification for block %s failed: %s", blockHash, err)
			return externalapi.StatusDisqualifiedFromChain, nil, nil
		}
		return 0, nil, err
	}
	log.Debugf("UTXO verification for block %s passed", blockHash)

	log.Tracef("Staging the multiset of block %s", blockHash)
	csm.multisetStore.Stage(stagingArea, blockHash, multiset)

	if csm.genesisHash.Equal(blockHash) {
		log.Tracef("Staging the utxoDiff of genesis")
		csm.stageDiff(stagingArea, blockHash, pastUTXOSet, nil)
		return externalapi.StatusUTXOValid, nil, nil
	}

	oldSelectedTip, err := csm.selectedTip(stagingArea)
	if err != nil {
		return 0, nil, err
	}

	if isResolveTip {
		oldSelectedTipUTXOSet, err := csm.restorePastUTXO(stagingArea, oldSelectedTip)
		if err != nil {
			return 0, nil, err
		}
		isNewSelectedTip, err := csm.isNewSelectedTip(stagingArea, blockHash, oldSelectedTip)
		if err != nil {
			return 0, nil, err
		}
		if isNewSelectedTip {
			log.Debugf("Block %s is the new selected tip, therefore setting it as old selected tip's diffChild", blockHash)

			updatedOldSelectedTipUTXOSet, err := pastUTXOSet.DiffFrom(oldSelectedTipUTXOSet)
			if err != nil {
				return 0, nil, err
			}
			log.Debugf("Setting the old selected tip's (%s) diffChild to be the new selected tip (%s)",
				oldSelectedTip, blockHash)
			csm.stageDiff(stagingArea, oldSelectedTip, updatedOldSelectedTipUTXOSet, blockHash)

			log.Tracef("Staging the utxoDiff of block %s, with virtual as diffChild", blockHash)
			csm.stageDiff(stagingArea, blockHash, pastUTXOSet, nil)
		} else {
			log.Debugf("Block %s is the tip of currently resolved chain, but not the new selected tip,"+
				"therefore setting it's utxoDiffChild to be the current selectedTip %s", blockHash, oldSelectedTip)
			utxoDiff, err := oldSelectedTipUTXOSet.DiffFrom(pastUTXOSet)
			if err != nil {
				return 0, nil, err
			}
			csm.stageDiff(stagingArea, blockHash, utxoDiff, oldSelectedTip)
		}
	} else {
		// If the block is not the tip of the currently resolved chain, we set it's diffChild to be the selectedParent,
		// this is a temporary measure to ensure there's a restore path to all blocks at all times.
		// Later down the process, the diff will be reversed in reverseUTXODiffs.
		log.Debugf("Block %s is not the new selected tip, and is not the tip of the currently verified chain, "+
			"therefore temporarily setting selectedParent as it's diffChild", blockHash)
		utxoDiff, err := selectedParentPastUTXOSet.DiffFrom(pastUTXOSet)
		if err != nil {
			return 0, nil, err
		}

		csm.stageDiff(stagingArea, blockHash, utxoDiff, selectedParentHash)
	}

	return externalapi.StatusUTXOValid, pastUTXOSet, nil
}

func (csm *consensusStateManager) isNewSelectedTip(stagingArea *model.StagingArea,
	blockHash, oldSelectedTip *externalapi.DomainHash) (bool, error) {

	newSelectedTip, err := csm.ghostdagManager.ChooseSelectedParent(stagingArea, blockHash, oldSelectedTip)
	if err != nil {
		return false, err
	}

	return blockHash.Equal(newSelectedTip), nil
}

func (csm *consensusStateManager) selectedTip(stagingArea *model.StagingArea) (*externalapi.DomainHash, error) {
	virtualGHOSTDAGData, err := csm.ghostdagDataStore.Get(csm.databaseContext, stagingArea, model.VirtualBlockHash)
	if err != nil {
		return nil, err
	}

	return virtualGHOSTDAGData.SelectedParent(), nil
}

func (csm *consensusStateManager) reverseUTXODiffs(stagingArea *model.StagingArea,
	recentlyVerifiedBlocks []*externalapi.DomainHash, tipUTXOSet, oneBlockBeforeTipUTXOSet externalapi.UTXODiff,
	useSeparateStagingAreaPerBlock bool) error {

	onEnd := logger.LogAndMeasureExecutionTime(log, "reverseUTXODiffs")
	defer onEnd()

	// During the process of resolving a chain of blocks, we temporarily set all blocks' (except the tip)
	// UTXODiffChild to be the selected parent.
	// Once the process is complete, we can reverse said chain, to now go directly to virtual through the relevant tip

	if len(recentlyVerifiedBlocks) == 1 {
		return nil
	}

	tip, oneBlockBeforeTip, restOfBlocks :=
		recentlyVerifiedBlocks[0], recentlyVerifiedBlocks[1], recentlyVerifiedBlocks[2:]

	log.Debugf("Reversing %d utxoDiffs", len(restOfBlocks)+1)

	// Set previousUTXODiff and previousBlock to oneBlockBeforeTip before we start touching them,
	// since oneBlockBeforeTip's UTXOSet is going to be over-written in the next step
	previousBlock := oneBlockBeforeTip
	previousUTXODiff, err := csm.utxoDiffStore.UTXODiff(csm.databaseContext, stagingArea, oneBlockBeforeTip)
	if err != nil {
		return err
	}

	// oneBlockBeforeTip is special in the sense that we don't have it's diff available in reverse, however, we can
	// calculate it from tipUTXOSet and oneBlockBeforeTipUTXOSet.
	// Therefore - we treat it separately
	oneBlockBeforeTipDiff, err := tipUTXOSet.DiffFrom(oneBlockBeforeTipUTXOSet)
	if err != nil {
		return err
	}
	if useSeparateStagingAreaPerBlock {
		err = csm.commitUTXODiffInSeparateStagingArea(oneBlockBeforeTip, oneBlockBeforeTipDiff, tip)
		if err != nil {
			return err
		}
	} else {
		csm.utxoDiffStore.Stage(stagingArea, oneBlockBeforeTip, oneBlockBeforeTipDiff, tip)
	}

	log.Trace("Reversed 1 utxoDiff")

	// Now go over the rest of the blocks and assign for every block Bi.UTXODiff = Bi+1.UTXODiff.Reversed()
	for i, currentBlock := range restOfBlocks {
		currentUTXODiff := previousUTXODiff.Reversed()

		// retrieve current utxoDiff for Bi, to be used by next block
		previousUTXODiff, err = csm.utxoDiffStore.UTXODiff(csm.databaseContext, stagingArea, currentBlock)
		if err != nil {
			return err
		}

		if useSeparateStagingAreaPerBlock {
			err = csm.commitUTXODiffInSeparateStagingArea(currentBlock, currentUTXODiff, previousBlock)
			if err != nil {
				return err
			}
		} else {
			csm.utxoDiffStore.Stage(stagingArea, currentBlock, currentUTXODiff, previousBlock)
		}

		previousBlock = currentBlock

		log.Tracef("Reversed %d utxoDiffs", i+1) // +1 because the first block is separate
	}

	return nil
}

func (csm *consensusStateManager) commitUTXODiffInSeparateStagingArea(
	blockHash *externalapi.DomainHash, utxoDiff externalapi.UTXODiff, utxoDiffChild *externalapi.DomainHash) error {

	stagingAreaForCurrentBlock := model.NewStagingArea()

	csm.utxoDiffStore.Stage(stagingAreaForCurrentBlock, blockHash, utxoDiff, utxoDiffChild)

	return staging.CommitAllChanges(csm.databaseContext, stagingAreaForCurrentBlock)
}
