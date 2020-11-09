package blockvalidator

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensusserialization"
	"github.com/kaspanet/kaspad/domain/consensus/utils/txscript"
	"github.com/pkg/errors"
	"math"
)

// ValidateBodyInContext validates block bodies in the context of the current
// consensus state
func (v *blockValidator) ValidateBodyInContext(blockHash *externalapi.DomainHash) error {
	return v.checkBlockTransactionsFinalized(blockHash)
}

func (v *blockValidator) checkBlockTransactionsFinalized(blockHash *externalapi.DomainHash) error {
	block, err := v.blockStore.Block(v.databaseContext, blockHash)
	if err != nil {
		return err
	}

	blockTime := block.Header.TimeInMilliseconds

	ghostdagData, err := v.ghostdagDataStore.Get(v.databaseContext, blockHash)
	if err != nil {
		return err
	}

	// If it's not genesis
	if len(block.Header.ParentHashes) != 0 {

		blockTime, err = v.pastMedianTimeManager.PastMedianTime(ghostdagData.SelectedParent)
		if err != nil {
			return err
		}
	}

	// Ensure all transactions in the block are finalized.
	for _, tx := range block.Transactions {
		if !v.isFinalizedTransaction(tx, ghostdagData.BlueScore, blockTime) {
			txID := consensusserialization.TransactionID(tx)
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
