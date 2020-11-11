package blockprocessor

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensusserialization"
	"github.com/pkg/errors"
)

func (bp *blockProcessor) validateAndInsertBlock(block *externalapi.DomainBlock) error {
	syncInfo, err := bp.syncManager.GetSyncInfo()
	if err != nil {
		return err
	}

	if len(block.Transactions) == 0 && syncInfo.State != externalapi.SyncStateRelay {
		syncInfo.State = externalapi.SyncStateHeadersFirst
	}

	hash := consensusserialization.HeaderHash(block.Header)
	if syncInfo.State == externalapi.SyncStateMissingUTXOSet {
		if isHeaderOnlyBlock(block) {
			// Allow processing headers while in state SyncStateMissingUTXOSet
			syncInfo.State = externalapi.SyncStateHeadersFirst
		} else {
			headerTipsPruningPoint, err := bp.consensusStateManager.HeaderTipsPruningPoint()
			if err != nil {
				return err
			}

			if *hash != *headerTipsPruningPoint {
				return errors.Errorf("cannot insert blocks other than the header pruning point "+
					"while in %s mode", syncInfo.State)
			}

			syncInfo.State = externalapi.SyncStateMissingBlockBodies
		}
	}

	if syncInfo.State == externalapi.SyncStateHeadersFirst && !isHeaderOnlyBlock(block) {
		syncInfo.State = externalapi.SyncStateRelay
		log.Warnf("block %s contains transactions while validating in header only mode", hash)
	}

	err = bp.checkBlockStatus(hash, syncInfo)
	if err != nil {
		return err
	}

	err = bp.validateBlock(block, syncInfo)
	if err != nil {
		bp.discardAllChanges()
		return err
	}

	hasHeader, err := bp.hasHeader(hash)
	if err != nil {
		return err
	}

	if !hasHeader {
		if syncInfo.State == externalapi.SyncStateMissingBlockBodies {
			return errors.Wrapf(ruleerrors.ErrMissingBlockHeaderInIBD, "no block header is stored for block %s. "+
				"Every block we get during %s mode should have a pre-stored header", syncInfo.State, hash)
		}
		err = bp.reachabilityManager.AddBlock(hash)
		if err != nil {
			return err
		}
	}

	if syncInfo.State == externalapi.SyncStateHeadersFirst {
		bp.blockStatusStore.Stage(hash, externalapi.StatusHeaderOnly)
	} else {
		bp.blockStatusStore.Stage(hash, externalapi.StatusUTXOPendingVerification)
	}

	// Block validations passed, save whatever DAG data was
	// collected so far
	err = bp.commitAllChanges()
	if err != nil {
		return err
	}

	hasTips, err := bp.headerTipsStore.HasTips(bp.databaseContext)
	if err != nil {
		return err
	}

	var oldHeadersSelectedTip *externalapi.DomainHash
	if hasTips {
		var err error
		oldHeadersSelectedTip, err = bp.headerTipsManager.SelectedTip()
		if err != nil {
			return err
		}
	}

	if syncInfo.State == externalapi.SyncStateHeadersFirst {
		err = bp.headerTipsManager.AddHeaderTip(hash)
		if err != nil {
			return err
		}
	} else if syncInfo.State == externalapi.SyncStateRelay || syncInfo.State == externalapi.SyncStateMissingGenesis {
		// Attempt to add the block to the virtual
		err = bp.consensusStateManager.AddBlockToVirtual(hash)
		if err != nil {
			return err
		}

		tips, err := bp.consensusStateStore.Tips(bp.databaseContext)
		if err != nil {
			return err
		}
		bp.headerTipsStore.Stage(tips)
	}

	if syncInfo.State != externalapi.SyncStateMissingGenesis {
		err = bp.updateReachabilityReindexRoot(oldHeadersSelectedTip)
		if err != nil {
			return err
		}
	}

	if syncInfo.State == externalapi.SyncStateRelay {
		// Trigger pruning, which will check if the pruning point changed and delete the data if it did.
		err = bp.pruningManager.FindNextPruningPoint()
		if err != nil {
			return err
		}
	}

	return bp.commitAllChanges()
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

func isHeaderOnlyBlock(block *externalapi.DomainBlock) bool {
	return len(block.Transactions) == 0
}

func (bp *blockProcessor) checkBlockStatus(hash *externalapi.DomainHash, mode *externalapi.SyncInfo) error {
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

	if mode.State == externalapi.SyncStateHeadersFirst || status != externalapi.StatusHeaderOnly {
		return errors.Wrapf(ruleerrors.ErrDuplicateBlock, "block %s already exists", hash)
	}

	return nil
}

func (bp *blockProcessor) validateBlock(block *externalapi.DomainBlock, mode *externalapi.SyncInfo) error {
	blockHash := consensusserialization.HeaderHash(block.Header)
	hasHeader, err := bp.hasHeader(blockHash)
	if err != nil {
		return err
	}

	if !hasHeader {
		bp.blockHeaderStore.Stage(blockHash, block.Header)
		err = bp.dagTopologyManager.SetParents(blockHash, block.Header.ParentHashes)
		if err != nil {
			return err
		}
	}

	// If any validation until (included) proof-of-work fails, simply
	// return an error without writing anything in the database.
	// This is to prevent spamming attacks.
	err = bp.validatePreProofOfWork(block)
	if err != nil {
		return err
	}

	err = bp.blockValidator.ValidateProofOfWorkAndDifficulty(blockHash)
	if err != nil {
		return err
	}

	// If in-context validations fail, discard all changes and store the
	// block with StatusInvalid.
	err = bp.validatePostProofOfWork(block, mode)
	if err != nil {
		if errors.As(err, &ruleerrors.RuleError{}) {
			bp.discardAllChanges()
			hash := consensusserialization.BlockHash(block)
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
	blockHash := consensusserialization.BlockHash(block)

	hasHeader, err := bp.hasHeader(blockHash)
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

func (bp *blockProcessor) validatePostProofOfWork(block *externalapi.DomainBlock, mode *externalapi.SyncInfo) error {
	blockHash := consensusserialization.BlockHash(block)

	if mode.State != externalapi.SyncStateHeadersFirst {
		bp.blockStore.Stage(blockHash, block)
		err := bp.blockValidator.ValidateBodyInIsolation(blockHash)
		if err != nil {
			return err
		}
	}

	hasHeader, err := bp.hasHeader(blockHash)
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

func (bp *blockProcessor) hasHeader(blockHash *externalapi.DomainHash) (bool, error) {
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
