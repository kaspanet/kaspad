package validator

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/processes/validator/txscript"
	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
	"github.com/kaspanet/kaspad/domain/consensus/utils/hashserialization"
	"math"
)

// ValidateBodyInContext validates block bodies in the context of the current
// consensus state
func (v *validator) ValidateBodyInContext(block *model.DomainBlock) error {
	return v.checkBlockTransactionsFinalized(block)
}

func (v *validator) checkBlockTransactionsFinalized(block *model.DomainBlock) error {
	blockTime := block.Header.TimeInMilliseconds

	hash := hashserialization.HeaderHash(block.Header)
	ghostdagData, err := v.ghostdagManager.BlockData(hash)
	if err != nil {
		return err
	}

	// If it's not genesis
	if len(block.Header.ParentHashes) != 0 {
		selectedParentGHOSTDAGData, err := v.ghostdagManager.BlockData(ghostdagData.SelectedParent)
		if err != nil {
			return err
		}

		blockTime, err = v.pastMedianTimeManager.PastMedianTime(selectedParentGHOSTDAGData)
		if err != nil {
			return err
		}
	}

	// Ensure all transactions in the block are finalized.
	for _, tx := range block.Transactions {
		if !v.isFinalizedTransaction(tx, ghostdagData.BlueScore, blockTime) {
			txID := hashserialization.TransactionID(tx)
			return ruleerrors.Errorf(ruleerrors.ErrUnfinalizedTx, "block contains unfinalized "+
				"transaction %s", txID)
		}
	}

	return nil
}

// IsFinalizedTransaction determines whether or not a transaction is finalized.
func (v *validator) isFinalizedTransaction(tx *model.DomainTransaction, blockBlueScore uint64, blockTime int64) bool {
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
