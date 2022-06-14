package blockprocessor

import (
	// we need to embed the utxoset of mainnet genesis here
	_ "embed"
	"fmt"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
	"github.com/kaspanet/kaspad/domain/consensus/utils/multiset"
	"github.com/kaspanet/kaspad/domain/consensus/utils/utxo"
	"github.com/kaspanet/kaspad/infrastructure/db/database"
	"github.com/kaspanet/kaspad/infrastructure/logger"
	"github.com/kaspanet/kaspad/util/difficulty"
	"github.com/kaspanet/kaspad/util/staging"
	"github.com/pkg/errors"
)

func (bp *blockProcessor) setBlockStatusAfterBlockValidation(
	stagingArea *model.StagingArea, block *externalapi.DomainBlock, isPruningPoint bool) (externalapi.BlockStatus, error) {

	blockHash := consensushashing.BlockHash(block)

	exists, err := bp.blockStatusStore.Exists(bp.databaseContext, stagingArea, blockHash)
	if err != nil {
		return externalapi.StatusInvalid, err
	}
	if exists {
		status, err := bp.blockStatusStore.Get(bp.databaseContext, stagingArea, blockHash)
		if err != nil {
			return externalapi.StatusInvalid, err
		}

		if status == externalapi.StatusUTXOValid {
			// A block cannot have status StatusUTXOValid just after finishing bp.validateBlock, because
			// if it's the case it should have been rejected as duplicate block.
			// The only exception is the pruning point because its status is manually set before inserting
			// the block.
			if !isPruningPoint {
				return externalapi.StatusInvalid, errors.Errorf("block %s that is not the pruning point is not expected to be valid "+
					"before adding to to the consensus state manager", blockHash)
			}
			log.Debugf("Block %s is the pruning point and has status %s, so leaving its status untouched",
				blockHash, status)
			return status, nil
		}
	}

	isHeaderOnlyBlock := isHeaderOnlyBlock(block)
	if isHeaderOnlyBlock {
		log.Debugf("Block %s is a header-only block so setting its status as %s",
			blockHash, externalapi.StatusHeaderOnly)
		bp.blockStatusStore.Stage(stagingArea, blockHash, externalapi.StatusHeaderOnly)
		return externalapi.StatusHeaderOnly, nil
	}

	log.Debugf("Block %s has body so setting its status as %s",
		blockHash, externalapi.StatusUTXOPendingVerification)
	bp.blockStatusStore.Stage(stagingArea, blockHash, externalapi.StatusUTXOPendingVerification)
	return externalapi.StatusUTXOPendingVerification, nil
}

func (bp *blockProcessor) updateVirtualAcceptanceDataAfterImportingPruningPoint(stagingArea *model.StagingArea) error {
	_, virtualAcceptanceData, virtualMultiset, err :=
		bp.consensusStateManager.CalculatePastUTXOAndAcceptanceData(stagingArea, model.VirtualBlockHash)
	if err != nil {
		return err
	}

	log.Debugf("Staging virtual acceptance data after importing the pruning point")
	bp.acceptanceDataStore.Stage(stagingArea, model.VirtualBlockHash, virtualAcceptanceData)

	log.Debugf("Staging virtual multiset after importing the pruning point")
	bp.multisetStore.Stage(stagingArea, model.VirtualBlockHash, virtualMultiset)
	return nil
}

func (bp *blockProcessor) validateAndInsertBlock(stagingArea *model.StagingArea, block *externalapi.DomainBlock,
	isPruningPoint bool, shouldValidateAgainstUTXO bool, isBlockWithTrustedData bool) (*externalapi.VirtualChangeSet, externalapi.BlockStatus, error) {

	blockHash := consensushashing.HeaderHash(block.Header)
	err := bp.validateBlock(stagingArea, block, isBlockWithTrustedData)
	if err != nil {
		return nil, externalapi.StatusInvalid, err
	}

	status, err := bp.setBlockStatusAfterBlockValidation(stagingArea, block, isPruningPoint)
	if err != nil {
		return nil, externalapi.StatusInvalid, err
	}

	var oldHeadersSelectedTip *externalapi.DomainHash
	hasHeaderSelectedTip, err := bp.headersSelectedTipStore.Has(bp.databaseContext, stagingArea)
	if err != nil {
		return nil, externalapi.StatusInvalid, err
	}
	if hasHeaderSelectedTip {
		var err error
		oldHeadersSelectedTip, err = bp.headersSelectedTipStore.HeadersSelectedTip(bp.databaseContext, stagingArea)
		if err != nil {
			return nil, externalapi.StatusInvalid, err
		}
	}

	shouldAddHeaderSelectedTip := false
	if !hasHeaderSelectedTip {
		shouldAddHeaderSelectedTip = true
	} else {
		pruningPoint, err := bp.pruningStore.PruningPoint(bp.databaseContext, stagingArea)
		if err != nil {
			return nil, externalapi.StatusInvalid, err
		}

		isInSelectedChainOfPruningPoint, err := bp.dagTopologyManager.IsInSelectedParentChainOf(stagingArea, pruningPoint, blockHash)
		if err != nil {
			return nil, externalapi.StatusInvalid, err
		}

		// Don't set blocks in the anticone of the pruning point as header selected tip.
		shouldAddHeaderSelectedTip = isInSelectedChainOfPruningPoint
	}

	if shouldAddHeaderSelectedTip {
		// Don't set blocks in the anticone of the pruning point as header selected tip.
		err = bp.headerTipsManager.AddHeaderTip(stagingArea, blockHash)
		if err != nil {
			return nil, externalapi.StatusInvalid, err
		}
	}

	bp.loadUTXODataForGenesis(stagingArea, block)
	var selectedParentChainChanges *externalapi.SelectedChainPath
	var virtualUTXODiff externalapi.UTXODiff
	var reversalData *model.UTXODiffReversalData
	isHeaderOnlyBlock := isHeaderOnlyBlock(block)
	if !isHeaderOnlyBlock {
		// Attempt to add the block to the virtual
		selectedParentChainChanges, virtualUTXODiff, reversalData, err = bp.consensusStateManager.AddBlock(stagingArea, blockHash, shouldValidateAgainstUTXO)
		if err != nil {
			return nil, externalapi.StatusInvalid, err
		}
	}

	if hasHeaderSelectedTip {
		err := bp.updateReachabilityReindexRoot(stagingArea, oldHeadersSelectedTip)
		if err != nil {
			return nil, externalapi.StatusInvalid, err
		}
	}

	if !isHeaderOnlyBlock && shouldValidateAgainstUTXO {
		// Trigger pruning, which will check if the pruning point changed and delete the data if it did.
		err = bp.pruningManager.UpdatePruningPointByVirtual(stagingArea)
		if err != nil {
			return nil, externalapi.StatusInvalid, err
		}
	}

	err = staging.CommitAllChanges(bp.databaseContext, stagingArea)
	if err != nil {
		return nil, externalapi.StatusInvalid, err
	}

	if reversalData != nil {
		err = bp.consensusStateManager.ReverseUTXODiffs(blockHash, reversalData)
		if err != nil {
			return nil, externalapi.StatusInvalid, err
		}
	}

	err = bp.pruningManager.UpdatePruningPointIfRequired()
	if err != nil {
		return nil, externalapi.StatusInvalid, err
	}

	log.Debug(logger.NewLogClosure(func() string {
		hashrate := difficulty.GetHashrateString(difficulty.CompactToBig(block.Header.Bits()), bp.targetTimePerBlock)
		return fmt.Sprintf("Block %s validated and inserted, network hashrate: %s", blockHash, hashrate)
	}))

	var logClosureErr error
	log.Debug(logger.NewLogClosure(func() string {
		virtualGhostDAGData, err := bp.ghostdagDataStore.Get(bp.databaseContext, stagingArea, model.VirtualBlockHash, false)
		if database.IsNotFoundError(err) {
			return fmt.Sprintf("Cannot log data for non-existent virtual")
		}

		if err != nil {
			logClosureErr = err
			return fmt.Sprintf("Failed to get virtual GHOSTDAG data: %s", err)
		}
		headerCount := bp.blockHeaderStore.Count(stagingArea)
		blockCount := bp.blockStore.Count(stagingArea)
		return fmt.Sprintf("New virtual's blue score: %d. Block count: %d. Header count: %d",
			virtualGhostDAGData.BlueScore(), blockCount, headerCount)
	}))
	if logClosureErr != nil {
		return nil, externalapi.StatusInvalid, logClosureErr
	}

	virtualParents, err := bp.dagTopologyManager.Parents(stagingArea, model.VirtualBlockHash)
	if database.IsNotFoundError(err) {
		virtualParents = nil
	} else if err != nil {
		return nil, externalapi.StatusInvalid, err
	}

	bp.pastMedianTimeManager.InvalidateVirtualPastMedianTimeCache()

	bp.blockLogger.LogBlock(block)

	return &externalapi.VirtualChangeSet{
		VirtualSelectedParentChainChanges: selectedParentChainChanges,
		VirtualUTXODiff:                   virtualUTXODiff,
		VirtualParents:                    virtualParents,
	}, status, nil
}

func (bp *blockProcessor) loadUTXODataForGenesis(stagingArea *model.StagingArea, block *externalapi.DomainBlock) {
	isGenesis := len(block.Header.DirectParents()) == 0
	if !isGenesis {
		return
	}
	blockHash := consensushashing.BlockHash(block)
	// Note: The applied UTXO set and multiset do not satisfy the UTXO commitment
	// of Mainnet's genesis. This is why any block that will be built on top of genesis
	// will have a wrong UTXO commitment as well, and will not be able to get to a consensus
	// with the rest of the network.
	// This is why getting direct blocks on top of genesis is forbidden, and the only way to
	// get a newer state for a node with genesis only is by requesting a proof for a recent
	// pruning point.
	// The actual UTXO set that fits Mainnet's genesis' UTXO commitment was removed from the codebase in order
	// to make reduce the consensus initialization time and the compiled binary size, but can be still
	// found here for anyone to verify: https://github.com/kaspanet/kaspad/blob/dbf18d8052f000ba0079be9e79b2d6f5a98b74ca/domain/consensus/processes/blockprocessor/resources/utxos.gz
	bp.consensusStateStore.StageVirtualUTXODiff(stagingArea, utxo.NewUTXODiff())
	bp.utxoDiffStore.Stage(stagingArea, blockHash, utxo.NewUTXODiff(), nil)
	bp.multisetStore.Stage(stagingArea, blockHash, multiset.New())
}

func isHeaderOnlyBlock(block *externalapi.DomainBlock) bool {
	return len(block.Transactions) == 0
}

func (bp *blockProcessor) updateReachabilityReindexRoot(stagingArea *model.StagingArea,
	oldHeadersSelectedTip *externalapi.DomainHash) error {

	headersSelectedTip, err := bp.headersSelectedTipStore.HeadersSelectedTip(bp.databaseContext, stagingArea)
	if err != nil {
		return err
	}

	if headersSelectedTip.Equal(oldHeadersSelectedTip) {
		return nil
	}

	return bp.reachabilityManager.UpdateReindexRoot(stagingArea, headersSelectedTip)
}

func (bp *blockProcessor) checkBlockStatus(stagingArea *model.StagingArea, block *externalapi.DomainBlock) error {
	hash := consensushashing.BlockHash(block)
	isHeaderOnlyBlock := isHeaderOnlyBlock(block)
	exists, err := bp.blockStatusStore.Exists(bp.databaseContext, stagingArea, hash)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}

	status, err := bp.blockStatusStore.Get(bp.databaseContext, stagingArea, hash)
	if err != nil {
		return err
	}

	if status == externalapi.StatusInvalid {
		return errors.Wrapf(ruleerrors.ErrKnownInvalid, "block %s is a known invalid block", hash)
	}

	if !isHeaderOnlyBlock {
		hasBlock, err := bp.blockStore.HasBlock(bp.databaseContext, stagingArea, hash)
		if err != nil {
			return err
		}
		if hasBlock {
			return errors.Wrapf(ruleerrors.ErrDuplicateBlock, "block %s already exists", hash)
		}
	} else {
		hasHeader, err := bp.blockHeaderStore.HasBlockHeader(bp.databaseContext, stagingArea, hash)
		if err != nil {
			return err
		}
		if hasHeader {
			return errors.Wrapf(ruleerrors.ErrDuplicateBlock, "block %s header already exists", hash)
		}
	}

	return nil
}

func (bp *blockProcessor) validatePreProofOfWork(stagingArea *model.StagingArea, block *externalapi.DomainBlock) error {
	blockHash := consensushashing.BlockHash(block)

	hasValidatedHeader, err := bp.hasValidatedHeader(stagingArea, blockHash)
	if err != nil {
		return err
	}

	if hasValidatedHeader {
		log.Debugf("Block %s header was already validated, so skip the rest of validatePreProofOfWork", blockHash)
		return nil
	}

	err = bp.blockValidator.ValidateHeaderInIsolation(stagingArea, blockHash)
	if err != nil {
		return err
	}
	return nil
}

func (bp *blockProcessor) validatePostProofOfWork(stagingArea *model.StagingArea, block *externalapi.DomainBlock, isBlockWithTrustedData bool) error {
	blockHash := consensushashing.BlockHash(block)

	isHeaderOnlyBlock := isHeaderOnlyBlock(block)
	if !isHeaderOnlyBlock {
		bp.blockStore.Stage(stagingArea, blockHash, block)
		err := bp.blockValidator.ValidateBodyInIsolation(stagingArea, blockHash)
		if err != nil {
			return err
		}
	}

	hasValidatedHeader, err := bp.hasValidatedHeader(stagingArea, blockHash)
	if err != nil {
		return err
	}

	if !hasValidatedHeader {
		err = bp.blockValidator.ValidateHeaderInContext(stagingArea, blockHash, isBlockWithTrustedData)
		if err != nil {
			return err
		}
	}

	if !isHeaderOnlyBlock {
		err = bp.blockValidator.ValidateBodyInContext(stagingArea, blockHash, isBlockWithTrustedData)
		if err != nil {
			return err
		}
	} else {
		log.Debugf("Skipping ValidateBodyInContext for block %s because it's header only", blockHash)
	}

	return nil
}

// hasValidatedHeader returns whether the block header was validated. It returns
// true in any case the block header was validated, whether it was validated as a
// header-only block or as a block with body.
func (bp *blockProcessor) hasValidatedHeader(stagingArea *model.StagingArea, blockHash *externalapi.DomainHash) (bool, error) {
	exists, err := bp.blockStatusStore.Exists(bp.databaseContext, stagingArea, blockHash)
	if err != nil {
		return false, err
	}

	if !exists {
		return false, nil
	}

	status, err := bp.blockStatusStore.Get(bp.databaseContext, stagingArea, blockHash)
	if err != nil {
		return false, err
	}

	return status != externalapi.StatusInvalid, nil
}
