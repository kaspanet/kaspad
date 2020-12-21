package blockprocessor

import (
	"fmt"

	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
	"github.com/kaspanet/kaspad/infrastructure/logger"
	"github.com/pkg/errors"
)

func (bp *blockProcessor) setBlockStatusAfterBlockValidation(block *externalapi.DomainBlock, isPruningPoint bool) error {
	blockHash := consensushashing.BlockHash(block)

	exists, err := bp.blockStatusStore.Exists(bp.databaseContext, blockHash)
	if err != nil {
		return err
	}
	if exists {
		status, err := bp.blockStatusStore.Get(bp.databaseContext, blockHash)
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
			log.Tracef("Block %s is the pruning point and has status %s, so leaving its status untouched",
				blockHash, status)
			return nil
		}
	}

	isHeaderOnlyBlock := isHeaderOnlyBlock(block)
	if isHeaderOnlyBlock {
		log.Tracef("Block %s is a header-only block so setting its status as %s",
			blockHash, externalapi.StatusHeaderOnly)
		bp.blockStatusStore.Stage(blockHash, externalapi.StatusHeaderOnly)
	} else {
		log.Tracef("Block %s has body so setting its status as %s",
			blockHash, externalapi.StatusUTXOPendingVerification)
		bp.blockStatusStore.Stage(blockHash, externalapi.StatusUTXOPendingVerification)
	}

	return nil
}

func (bp *blockProcessor) validateAndInsertBlock(block *externalapi.DomainBlock, isPruningPoint bool) (*externalapi.BlockInsertionResult, error) {
	blockHash := consensushashing.HeaderHash(block.Header)
	err := bp.validateBlock(block, isPruningPoint)
	if err != nil {
		bp.discardAllChanges()
		return nil, err
	}

	err = bp.setBlockStatusAfterBlockValidation(block, isPruningPoint)
	if err != nil {
		return nil, err
	}

	// Block validations passed, save whatever DAG data was
	// collected so far
	err = bp.commitAllChanges()
	if err != nil {
		return nil, err
	}

	var oldHeadersSelectedTip *externalapi.DomainHash
	isGenesis := *blockHash == *bp.genesisHash
	if !isGenesis {
		var err error
		oldHeadersSelectedTip, err = bp.headersSelectedTipStore.HeadersSelectedTip(bp.databaseContext)
		if err != nil {
			return nil, err
		}
	}

	err = bp.headerTipsManager.AddHeaderTip(blockHash)
	if err != nil {
		return nil, err
	}

	var selectedParentChainChanges *externalapi.SelectedParentChainChanges
	isHeaderOnlyBlock := isHeaderOnlyBlock(block)
	if !isHeaderOnlyBlock {
		// There's no need to update the consensus state manager when
		// processing the pruning point since it was already handled
		// in consensusStateManager.UpdatePruningPoint
		if !isPruningPoint {
			// Attempt to add the block to the virtual
			selectedParentChainChanges, err = bp.consensusStateManager.AddBlock(blockHash)
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

	err = bp.commitAllChanges()
	if err != nil {
		return nil, err
	}

	log.Debugf("Block %s validated and inserted", blockHash)

	var logClosureErr error
	log.Debugf("%s", logger.NewLogClosure(func() string {
		virtualGhostDAGData, err := bp.ghostdagDataStore.Get(bp.databaseContext, model.VirtualBlockHash)
		if err != nil {
			logClosureErr = err
			return fmt.Sprintf("Failed to get virtual GHOSTDAG data: %s", err)
		}
		headerCount := bp.blockHeaderStore.Count()
		blockCount := bp.blockStore.Count()
		return fmt.Sprintf("New virtual's blue score: %d. Block count: %d. Header count: %d",
			virtualGhostDAGData.BlueScore(), blockCount, headerCount)
	}))
	if logClosureErr != nil {
		return nil, logClosureErr
	}

	return &externalapi.BlockInsertionResult{
		VirtualSelectedParentChainChanges: selectedParentChainChanges,
	}, nil
}

func isHeaderOnlyBlock(block *externalapi.DomainBlock) bool {
	return len(block.Transactions) == 0
}

func (bp *blockProcessor) updateReachabilityReindexRoot(oldHeadersSelectedTip *externalapi.DomainHash) error {
	headersSelectedTip, err := bp.headersSelectedTipStore.HeadersSelectedTip(bp.databaseContext)
	if err != nil {
		return err
	}

	if *headersSelectedTip == *oldHeadersSelectedTip {
		return nil
	}

	return bp.reachabilityManager.UpdateReindexRoot(headersSelectedTip)
}

func (bp *blockProcessor) checkBlockStatus(block *externalapi.DomainBlock) error {
	hash := consensushashing.BlockHash(block)
	isHeaderOnlyBlock := isHeaderOnlyBlock(block)
	exists, err := bp.blockStatusStore.Exists(bp.databaseContext, hash)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}

	status, err := bp.blockStatusStore.Get(bp.databaseContext, hash)
	if err != nil {
		return err
	}

	if status == externalapi.StatusInvalid {
		return errors.Wrapf(ruleerrors.ErrKnownInvalid, "block %s is a known invalid block", hash)
	}

	if !isHeaderOnlyBlock {
		hasBlock, err := bp.blockStore.HasBlock(bp.databaseContext, hash)
		if err != nil {
			return err
		}
		if hasBlock {
			return errors.Wrapf(ruleerrors.ErrDuplicateBlock, "block %s already exists", hash)
		}
	} else {
		hasHeader, err := bp.blockHeaderStore.HasBlockHeader(bp.databaseContext, hash)
		if err != nil {
			return err
		}
		if hasHeader {
			return errors.Wrapf(ruleerrors.ErrDuplicateBlock, "block %s header already exists", hash)
		}
	}

	return nil
}

func (bp *blockProcessor) validatePreProofOfWork(block *externalapi.DomainBlock) error {
	blockHash := consensushashing.BlockHash(block)

	hasValidatedHeader, err := bp.hasValidatedHeader(blockHash)
	if err != nil {
		return err
	}

	if hasValidatedHeader {
		log.Debugf("Block %s header was already validated, so skip the rest of validatePreProofOfWork", blockHash)
		return nil
	}

	err = bp.blockValidator.ValidateHeaderInIsolation(blockHash)
	if err != nil {
		return err
	}
	return nil
}

func (bp *blockProcessor) validatePostProofOfWork(block *externalapi.DomainBlock, isPruningPoint bool) error {
	blockHash := consensushashing.BlockHash(block)

	isHeaderOnlyBlock := isHeaderOnlyBlock(block)
	if !isHeaderOnlyBlock {
		bp.blockStore.Stage(blockHash, block)
		err := bp.blockValidator.ValidateBodyInIsolation(blockHash)
		if err != nil {
			return err
		}
	}

	hasValidatedHeader, err := bp.hasValidatedHeader(blockHash)
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
		log.Tracef("Skipping ValidateBodyInContext for block %s because it's header only", blockHash)
	}

	return nil
}

// hasValidatedHeader returns whether the block header was validated. It returns
// true in any case the block header was validated, whether it was validated as a
// header-only block or as a block with body.
func (bp *blockProcessor) hasValidatedHeader(blockHash *externalapi.DomainHash) (bool, error) {
	exists, err := bp.blockStatusStore.Exists(bp.databaseContext, blockHash)
	if err != nil {
		return false, err
	}

	if !exists {
		return false, nil
	}

	status, err := bp.blockStatusStore.Get(bp.databaseContext, blockHash)
	if err != nil {
		return false, err
	}

	return status != externalapi.StatusInvalid, nil
}

func (bp *blockProcessor) discardAllChanges() {
	for _, store := range bp.stores {
		store.Discard()
	}
}

func (bp *blockProcessor) commitAllChanges() error {
	dbTx, err := bp.databaseContext.Begin()
	if err != nil {
		return err
	}

	for _, store := range bp.stores {
		err = store.Commit(dbTx)
		if err != nil {
			return err
		}
	}

	return dbTx.Commit()
}
