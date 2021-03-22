package blockprocessor

import (
	"fmt"

	"github.com/kaspanet/kaspad/util/difficulty"

	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
	"github.com/kaspanet/kaspad/infrastructure/logger"
	"github.com/pkg/errors"
)

func (bp *blockProcessor) setBlockStatusAfterBlockValidation(block *externalapi.DomainBlock, isPruningPoint bool) error {
	blockHash := consensushashing.BlockHash(block)

	exists, err := bp.blockStatusStore.Exists(bp.databaseContext, nil, blockHash)
	if err != nil {
		return err
	}
	if exists {
		status, err := bp.blockStatusStore.Get(bp.databaseContext, nil, blockHash)
		if err != nil {
			return err
		}

		if status == externalapi.StatusUTXOValid {
			// A block cannot have status StatusUTXOValid just after finishing bp.validateBlock, because
			// if it's the case it should have been rejected as duplicate block.
			// The only exception is the pruning point because its status is manually set before inserting
			// the block.
			if !isPruningPoint {
				return errors.Errorf("block %s that is not the pruning point is not expected to be valid "+
					"before adding to to the consensus state manager", blockHash)
			}
			log.Debugf("Block %s is the pruning point and has status %s, so leaving its status untouched",
				blockHash, status)
			return nil
		}
	}

	isHeaderOnlyBlock := isHeaderOnlyBlock(block)
	if isHeaderOnlyBlock {
		log.Debugf("Block %s is a header-only block so setting its status as %s",
			blockHash, externalapi.StatusHeaderOnly)
		bp.blockStatusStore.Stage(nil, blockHash, externalapi.StatusHeaderOnly)
	} else {
		log.Debugf("Block %s has body so setting its status as %s",
			blockHash, externalapi.StatusUTXOPendingVerification)
		bp.blockStatusStore.Stage(nil, blockHash, externalapi.StatusUTXOPendingVerification)
	}

	return nil
}

func (bp *blockProcessor) validateAndInsertBlock(block *externalapi.DomainBlock, isPruningPoint bool) (*externalapi.BlockInsertionResult, error) {
	stagingArea := model.NewStagingArea()

	blockHash := consensushashing.HeaderHash(block.Header)
	err := bp.validateBlock(stagingArea, block, isPruningPoint)
	if err != nil {
		return nil, err
	}

	err = bp.setBlockStatusAfterBlockValidation(block, isPruningPoint)
	if err != nil {
		return nil, err
	}

	var oldHeadersSelectedTip *externalapi.DomainHash
	isGenesis := blockHash.Equal(bp.genesisHash)
	if !isGenesis {
		var err error
		oldHeadersSelectedTip, err = bp.headersSelectedTipStore.HeadersSelectedTip(bp.databaseContext, nil)
		if err != nil {
			return nil, err
		}
	}

	err = bp.headerTipsManager.AddHeaderTip(blockHash)
	if err != nil {
		return nil, err
	}

	var selectedParentChainChanges *externalapi.SelectedChainPath
	var virtualUTXODiff externalapi.UTXODiff
	isHeaderOnlyBlock := isHeaderOnlyBlock(block)
	if !isHeaderOnlyBlock {
		// There's no need to update the consensus state manager when
		// processing the pruning point since it was already handled
		// in consensusStateManager.ImportPruningPoint
		if !isPruningPoint {
			// Attempt to add the block to the virtual
			selectedParentChainChanges, virtualUTXODiff, err = bp.consensusStateManager.AddBlock(blockHash)
			if err != nil {
				return nil, err
			}
		}
	}

	if !isGenesis {
		err := bp.updateReachabilityReindexRoot(oldHeadersSelectedTip)
		if err != nil {
			return nil, err
		}
	}

	if !isHeaderOnlyBlock {
		// Trigger pruning, which will check if the pruning point changed and delete the data if it did.
		err = bp.pruningManager.UpdatePruningPointByVirtual()
		if err != nil {
			return nil, err
		}
	}

	err = bp.commitAllChanges(nil)
	if err != nil {
		return nil, err
	}

	err = bp.pruningManager.UpdatePruningPointUTXOSetIfRequired()
	if err != nil {
		return nil, err
	}

	log.Debug(logger.NewLogClosure(func() string {
		hashrate := difficulty.GetHashrateString(difficulty.CompactToBig(block.Header.Bits()), bp.targetTimePerBlock)
		return fmt.Sprintf("Block %s validated and inserted, network hashrate: %s", blockHash, hashrate)
	}))

	var logClosureErr error
	log.Debug(logger.NewLogClosure(func() string {
		virtualGhostDAGData, err := bp.ghostdagDataStore.Get(bp.databaseContext, nil, model.VirtualBlockHash)
		if err != nil {
			logClosureErr = err
			return fmt.Sprintf("Failed to get virtual GHOSTDAG data: %s", err)
		}
		headerCount := bp.blockHeaderStore.Count(nil)
		blockCount := bp.blockStore.Count()
		return fmt.Sprintf("New virtual's blue score: %d. Block count: %d. Header count: %d",
			virtualGhostDAGData.BlueScore(), blockCount, headerCount)
	}))
	if logClosureErr != nil {
		return nil, logClosureErr
	}

	virtualParents, err := bp.dagTopologyManager.Parents(nil, model.VirtualBlockHash)
	if err != nil {
		return nil, err
	}

	bp.blockLogger.LogBlock(block)

	return &externalapi.BlockInsertionResult{
		VirtualSelectedParentChainChanges: selectedParentChainChanges,
		VirtualUTXODiff:                   virtualUTXODiff,
		VirtualParents:                    virtualParents,
	}, nil
}

func isHeaderOnlyBlock(block *externalapi.DomainBlock) bool {
	return len(block.Transactions) == 0
}

func (bp *blockProcessor) updateReachabilityReindexRoot(oldHeadersSelectedTip *externalapi.DomainHash) error {
	headersSelectedTip, err := bp.headersSelectedTipStore.HeadersSelectedTip(bp.databaseContext, nil)
	if err != nil {
		return err
	}

	if headersSelectedTip.Equal(oldHeadersSelectedTip) {
		return nil
	}

	return bp.reachabilityManager.UpdateReindexRoot(nil, headersSelectedTip)
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

func (bp *blockProcessor) validatePostProofOfWork(stagingArea *model.StagingArea, block *externalapi.DomainBlock, isPruningPoint bool) error {
	blockHash := consensushashing.BlockHash(block)

	isHeaderOnlyBlock := isHeaderOnlyBlock(block)
	if !isHeaderOnlyBlock {
		bp.blockStore.Stage(stagingArea, blockHash, block)
		err := bp.blockValidator.ValidateBodyInIsolation(blockHash)
		if err != nil {
			return err
		}
	}

	hasValidatedHeader, err := bp.hasValidatedHeader(nil, blockHash)
	if err != nil {
		return err
	}

	if !hasValidatedHeader {
		err = bp.blockValidator.ValidateHeaderInContext(blockHash)
		if err != nil {
			return err
		}
	}

	if !isHeaderOnlyBlock {
		err = bp.blockValidator.ValidateBodyInContext(blockHash, isPruningPoint)
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

func (bp *blockProcessor) commitAllChanges(stagingArea *model.StagingArea) error {
	onEnd := logger.LogAndMeasureExecutionTime(log, "commitAllChanges")
	defer onEnd()

	dbTx, err := bp.databaseContext.Begin()
	if err != nil {
		return err
	}

	return stagingArea.Commit(dbTx)
}
