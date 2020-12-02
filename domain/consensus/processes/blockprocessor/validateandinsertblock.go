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

type insertMode uint8

const (
	insertModeGenesis insertMode = iota
	insertModeHeader
	insertModeBlockWithoutUpdatingVirtual
	insertModeBlock
)

func (bp *blockProcessor) validateAndInsertBlock(block *externalapi.DomainBlock) error {
	blockHash := consensushashing.HeaderHash(block.Header)
	log.Debugf("Validating block %s", blockHash)

	insertMode, err := bp.validateAgainstSyncStateAndResolveInsertMode(block)
	if err != nil {
		return err
	}

	err = bp.checkBlockStatus(blockHash, insertMode)
	if err != nil {
		return err
	}

	err = bp.validateBlock(block, insertMode)
	if err != nil {
		bp.discardAllChanges()
		return err
	}

	hasHeader, err := bp.hasHeader(blockHash)
	if err != nil {
		return err
	}

	if !hasHeader {
		err = bp.reachabilityManager.AddBlock(blockHash)
		if err != nil {
			return err
		}
	}

	if insertMode == insertModeHeader {
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
	if insertMode != insertModeGenesis {
		var err error
		oldHeadersSelectedTip, err = bp.headerTipsManager.SelectedTip()
		if err != nil {
			return err
		}
	}

	if insertMode == insertModeHeader {
		err = bp.headerTipsManager.AddHeaderTip(blockHash)
		if err != nil {
			return err
		}
	} else if insertMode == insertModeBlock || insertMode == insertModeGenesis {
		// Attempt to add the block to the virtual
		err = bp.consensusStateManager.AddBlockToVirtual(blockHash)
		if err != nil {
			return err
		}

		tips, err := bp.consensusStateStore.Tips(bp.databaseContext)
		if err != nil {
			return err
		}
		bp.headerTipsStore.Stage(tips)
	}

	if insertMode != insertModeGenesis {
		err := bp.updateReachabilityReindexRoot(oldHeadersSelectedTip)
		if err != nil {
			return err
		}
	}

	if insertMode == insertModeBlock {
		// Trigger pruning, which will check if the pruning point changed and delete the data if it did.
		err = bp.pruningManager.FindNextPruningPoint()
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
			virtualGhostDAGData.BlueScore, syncInfo.State, syncInfo.BlockCount, syncInfo.HeaderCount)
	}))
	if logClosureErr != nil {
		return logClosureErr
	}

	return nil
}

func (bp *blockProcessor) validateAgainstSyncStateAndResolveInsertMode(block *externalapi.DomainBlock) (insertMode, error) {
	syncInfo, err := bp.syncManager.GetSyncInfo()
	if err != nil {
		return 0, err
	}
	syncState := syncInfo.State

	isHeaderOnlyBlock := isHeaderOnlyBlock(block)
	blockHash := consensushashing.HeaderHash(block.Header)
	if syncState == externalapi.SyncStateAwaitingGenesis {
		if isHeaderOnlyBlock {
			return 0, errors.Errorf("Got a header-only block while awaiting genesis")
		}
		if *blockHash != *bp.genesisHash {
			return 0, errors.Errorf("Received a non-genesis block while awaiting genesis")
		}
		return insertModeGenesis, nil
	}

	if isHeaderOnlyBlock {
		return insertModeHeader, nil
	}

	if syncState == externalapi.SyncStateAwaitingUTXOSet {
		headerTipsPruningPoint, err := bp.consensusStateManager.HeaderTipsPruningPoint()
		if err != nil {
			return 0, err
		}
		if *blockHash != *headerTipsPruningPoint {
			return 0, errors.Errorf("cannot insert blocks other than the header pruning point " +
				"while awaiting the UTXO set")
		}
		return insertModeBlock, nil
	}

	if syncState == externalapi.SyncStateAwaitingBlockBodies {
		headerTips, err := bp.headerTipsStore.Tips(bp.databaseContext)
		if err != nil {
			return 0, err
		}
		selectedHeaderTip, err := bp.ghostdagManager.ChooseSelectedParent(headerTips...)
		if err != nil {
			return 0, err
		}
		if *selectedHeaderTip != *blockHash {
			return insertModeBlockWithoutUpdatingVirtual, nil
		}
	}

	return insertModeBlock, nil
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

func (bp *blockProcessor) checkBlockStatus(hash *externalapi.DomainHash, mode insertMode) error {
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

	isBlockBodyAfterBlockHeader := mode != insertModeHeader && status == externalapi.StatusHeaderOnly
	if !isBlockBodyAfterBlockHeader {
		return errors.Wrapf(ruleerrors.ErrDuplicateBlock, "block %s already exists", hash)
	}

	return nil
}

func (bp *blockProcessor) validateBlock(block *externalapi.DomainBlock, mode insertMode) error {
	blockHash := consensushashing.HeaderHash(block.Header)
	hasHeader, err := bp.hasHeader(blockHash)
	if err != nil {
		return err
	}

	if !hasHeader {
		bp.blockHeaderStore.Stage(blockHash, block.Header)
	}

	// If any validation until (included) proof-of-work fails, simply
	// return an error without writing anything in the database.
	// This is to prevent spamming attacks.
	err = bp.validatePreProofOfWork(block)
	if err != nil {
		return err
	}

	err = bp.blockValidator.ValidatePruningPointViolationAndProofOfWorkAndDifficulty(blockHash)
	if err != nil {
		return err
	}

	// If in-context validations fail, discard all changes and store the
	// block with StatusInvalid.
	err = bp.validatePostProofOfWork(block, mode)
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

func (bp *blockProcessor) validatePostProofOfWork(block *externalapi.DomainBlock, mode insertMode) error {
	blockHash := consensushashing.BlockHash(block)

	if mode != insertModeHeader {
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
