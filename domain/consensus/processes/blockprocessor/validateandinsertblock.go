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

func (bp *blockProcessor) validateAndInsertBlock(block *externalapi.DomainBlock) error {
	blockHash := consensushashing.HeaderHash(block.Header)
	err := bp.validateBlock(block)
	if err != nil {
		bp.discardAllChanges()
		return err
	}

	isHeaderOnlyBlock := isHeaderOnlyBlock(block)
	if isHeaderOnlyBlock {
		bp.blockStatusStore.Stage(blockHash, externalapi.StatusHeaderOnly)
	} else {
		bp.blockStatusStore.Stage(blockHash, externalapi.StatusUTXOPendingVerification)
	}

	// Block validations passed, save whatever DAG data was
	// collected so far
	err = bp.commitAllChanges()
	if err != nil {
		return err
	}

	var oldHeadersSelectedTip *externalapi.DomainHash
	isGenesis := *blockHash != *bp.genesisHash
	if isGenesis {
		var err error
		oldHeadersSelectedTip, err = bp.headerTipsManager.SelectedTip()
		if err != nil {
			return err
		}
	}

	if isHeaderOnlyBlock {
		err = bp.headerTipsManager.AddHeaderTip(blockHash)
		if err != nil {
			return err
		}
	} else {
		// Attempt to add the block to the virtual
		err = bp.consensusStateManager.AddBlock(blockHash)
		if err != nil {
			return err
		}

		tips, err := bp.consensusStateStore.Tips(bp.databaseContext)
		if err != nil {
			return err
		}
		bp.headerTipsStore.Stage(tips)
	}

	if isGenesis {
		err := bp.updateReachabilityReindexRoot(oldHeadersSelectedTip)
		if err != nil {
			return err
		}
	}

	if !isHeaderOnlyBlock {
		// Trigger pruning, which will check if the pruning point changed and delete the data if it did.
		err = bp.pruningManager.UpdatePruningPointByVirtual()
		if err != nil {
			return err
		}
	}

	err = bp.commitAllChanges()
	if err != nil {
		return err
	}

	log.Debugf("Block %s validated and inserted", blockHash)

	var logClosureErr error
	log.Debugf("%s", logger.NewLogClosure(func() string {
		virtualGhostDAGData, err := bp.ghostdagDataStore.Get(bp.databaseContext, model.VirtualBlockHash)
		if err != nil {
			logClosureErr = err
			return fmt.Sprintf("Failed to get virtual GHOSTDAG data: %s", err)
		}
		syncInfo, err := bp.syncManager.GetSyncInfo()
		if err != nil {
			logClosureErr = err
			return fmt.Sprintf("Failed to get sync info: %s", err)
		}
		return fmt.Sprintf("New virtual's blue score: %d. Sync state: %s. Block count: %d. Header count: %d",
			virtualGhostDAGData.BlueScore(), syncInfo.State, syncInfo.BlockCount, syncInfo.HeaderCount)
	}))
	if logClosureErr != nil {
		return logClosureErr
	}

	return nil
}

func isHeaderOnlyBlock(block *externalapi.DomainBlock) bool {
	return len(block.Transactions) == 0
}

func (bp *blockProcessor) updateReachabilityReindexRoot(oldHeadersSelectedTip *externalapi.DomainHash) error {
	headersSelectedTip, err := bp.headerTipsManager.SelectedTip()
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

	isBlockBodyAfterBlockHeader := isHeaderOnlyBlock && status == externalapi.StatusHeaderOnly
	if !isBlockBodyAfterBlockHeader {
		return errors.Wrapf(ruleerrors.ErrDuplicateBlock, "block %s already exists", hash)
	}

	isDuplicateHeader := isHeaderOnlyBlock && status == externalapi.StatusHeaderOnly
	if isDuplicateHeader {
		return errors.Wrapf(ruleerrors.ErrDuplicateBlock, "block %s already exists", hash)
	}

	return nil
}

func (bp *blockProcessor) validateBlock(block *externalapi.DomainBlock) error {

	blockHash := consensushashing.HeaderHash(block.Header)
	log.Debugf("Validating block %s", blockHash)

	err := bp.checkBlockStatus(block)
	if err != nil {
		return err
	}

	hasValidatedHeader, err := bp.hasValidatedHeader(blockHash)
	if err != nil {
		return err
	}

	if !hasValidatedHeader {
		bp.blockHeaderStore.Stage(blockHash, block.Header)
	}

	// If any validation until (included) proof-of-work fails, simply
	// return an error without writing anything in the database.
	// This is to prevent spamming attacks.
	err = bp.validatePreProofOfWork(block)
	if err != nil {
		return err
	}

	if !hasValidatedHeader {
		err = bp.validatePruningPointViolationAndProofOfWorkAndDifficulty(block)
		if err != nil {
			return err
		}
	}

	// If in-context validations fail, discard all changes and store the
	// block with StatusInvalid.
	err = bp.validatePostProofOfWork(block)
	if err != nil {
		if errors.As(err, &ruleerrors.RuleError{}) {
			bp.discardAllChanges()
			hash := consensushashing.BlockHash(block)
			bp.blockStatusStore.Stage(hash, externalapi.StatusInvalid)
			commitErr := bp.commitAllChanges()
			if commitErr != nil {
				return commitErr
			}
		}
		return err
	}
	return nil
}

func (bp *blockProcessor) validatePreProofOfWork(block *externalapi.DomainBlock) error {
	blockHash := consensushashing.BlockHash(block)

	hasHeader, err := bp.hasValidatedHeader(blockHash)
	if err != nil {
		return err
	}

	if hasHeader {
		return nil
	}

	err = bp.blockValidator.ValidateHeaderInIsolation(blockHash)
	if err != nil {
		return err
	}
	return nil
}

func (bp *blockProcessor) validatePruningPointViolationAndProofOfWorkAndDifficulty(block *externalapi.DomainBlock) error {
	blockHash := consensushashing.HeaderHash(block.Header)
	if !isHeaderOnlyBlock(block) {
		// We stage the block here since we need it for parent validation
		bp.blockStore.Stage(blockHash, block)
	}
	return bp.blockValidator.ValidatePruningPointViolationAndProofOfWorkAndDifficulty(blockHash)
}

func (bp *blockProcessor) validatePostProofOfWork(block *externalapi.DomainBlock) error {
	blockHash := consensushashing.BlockHash(block)

	if !isHeaderOnlyBlock(block) {
		err := bp.blockValidator.ValidateBodyInIsolation(blockHash)
		if err != nil {
			return err
		}
	}

	hasHeader, err := bp.hasValidatedHeader(blockHash)
	if err != nil {
		return err
	}

	if !hasHeader {
		err = bp.blockValidator.ValidateHeaderInContext(blockHash)
		if err != nil {
			return err
		}
	}

	return nil
}

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

	return status == externalapi.StatusHeaderOnly, nil
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
