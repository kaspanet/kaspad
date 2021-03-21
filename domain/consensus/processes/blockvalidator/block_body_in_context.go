package blockvalidator

import (
	"github.com/kaspanet/kaspad/infrastructure/logger"
	"math"

	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
	"github.com/kaspanet/kaspad/domain/consensus/utils/txscript"
	"github.com/pkg/errors"
)

// ValidateBodyInContext validates block bodies in the context of the current
// consensus state
func (v *blockValidator) ValidateBodyInContext(blockHash *externalapi.DomainHash, isPruningPoint bool) error {
	onEnd := logger.LogAndMeasureExecutionTime(log, "ValidateBodyInContext")
	defer onEnd()

	err := v.checkBlockIsNotPruned(blockHash)
	if err != nil {
		return err
	}

	err = v.checkBlockTransactionsFinalized(blockHash)
	if err != nil {
		return err
	}

	if !isPruningPoint {
		err := v.checkParentBlockBodiesExist(blockHash)
		if err != nil {
			return err
		}
	}
	return nil
}

// checkBlockIsNotPruned Checks we don't add block bodies to pruned blocks
func (v *blockValidator) checkBlockIsNotPruned(blockHash *externalapi.DomainHash) error {
	hasValidatedHeader, err := v.hasValidatedHeader(blockHash)
	if err != nil {
		return err
	}

	// If we don't add block body to a header only block it can't be in the past
	// of the tips, because it'll be a new tip.
	if !hasValidatedHeader {
		return nil
	}

	tips, err := v.consensusStateStore.Tips(nil, v.databaseContext)
	if err != nil {
		return err
	}

	isAncestorOfSomeTips, err := v.dagTopologyManager.IsAncestorOfAny(blockHash, tips)
	if err != nil {
		return err
	}

	// A header only block in the past of one of the tips has to be pruned
	if isAncestorOfSomeTips {
		return errors.Wrapf(ruleerrors.ErrPrunedBlock, "cannot add block body to a pruned block %s", blockHash)
	}

	return nil
}

func (v *blockValidator) checkParentBlockBodiesExist(blockHash *externalapi.DomainHash) error {
	missingParentHashes := []*externalapi.DomainHash{}
	header, err := v.blockHeaderStore.BlockHeader(v.databaseContext, nil, blockHash)
	if err != nil {
		return err
	}
	for _, parent := range header.ParentHashes() {
		hasBlock, err := v.blockStore.HasBlock(v.databaseContext,, parent)
		if err != nil {
			return err
		}

		if !hasBlock {
			pruningPoint, err := v.pruningStore.PruningPoint(v.databaseContext, nil)
			if err != nil {
				return err
			}

			isInPastOfPruningPoint, err := v.dagTopologyManager.IsAncestorOf(parent, pruningPoint)
			if err != nil {
				return err
			}

			// If a block parent is in the past of the pruning point
			// it means its body will never be used, so it's ok if
			// it's missing.
			// This will usually happen during IBD when getting the blocks
			// in the pruning point anticone.
			if isInPastOfPruningPoint {
				log.Debugf("Block %s parent %s is missing a body, but is in the past of the pruning point",
					blockHash, parent)
				continue
			}

			log.Debugf("Block %s parent %s is missing a body", blockHash, parent)

			missingParentHashes = append(missingParentHashes, parent)
		}
	}

	if len(missingParentHashes) > 0 {
		return ruleerrors.NewErrMissingParents(missingParentHashes)
	}

	return nil
}

func (v *blockValidator) checkBlockTransactionsFinalized(blockHash *externalapi.DomainHash) error {
	block, err := v.blockStore.Block(v.databaseContext,, blockHash)
	if err != nil {
		return err
	}

	ghostdagData, err := v.ghostdagDataStore.Get(v.databaseContext, nil, blockHash)
	if err != nil {
		return err
	}

	blockTime, err := v.pastMedianTimeManager.PastMedianTime(blockHash)
	if err != nil {
		return err
	}

	// Ensure all transactions in the block are finalized.
	for _, tx := range block.Transactions {
		if !v.isFinalizedTransaction(tx, ghostdagData.BlueScore(), blockTime) {
			txID := consensushashing.TransactionID(tx)
			return errors.Wrapf(ruleerrors.ErrUnfinalizedTx, "block contains unfinalized "+
				"transaction %s", txID)
		}
	}

	return nil
}

// IsFinalizedTransaction determines whether or not a transaction is finalized.
func (v *blockValidator) isFinalizedTransaction(tx *externalapi.DomainTransaction, blockBlueScore uint64, blockTime int64) bool {
	// Lock time of zero means the transaction is finalized.
	lockTime := tx.LockTime
	if lockTime == 0 {
		return true
	}

	// The lock time field of a transaction is either a block blue score at
	// which the transaction is finalized or a timestamp depending on if the
	// value is before the txscript.LockTimeThreshold. When it is under the
	// threshold it is a block blue score.
	blockTimeOrBlueScore := uint64(0)
	if lockTime < txscript.LockTimeThreshold {
		blockTimeOrBlueScore = blockBlueScore
	} else {
		blockTimeOrBlueScore = uint64(blockTime)
	}
	if lockTime < blockTimeOrBlueScore {
		return true
	}

	// At this point, the transaction's lock time hasn't occurred yet, but
	// the transaction might still be finalized if the sequence number
	// for all transaction inputs is maxed out.
	for _, input := range tx.Inputs {
		if input.Sequence != math.MaxUint64 {
			return false
		}
	}
	return true
}
