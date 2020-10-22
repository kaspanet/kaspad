package validator

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/processes/validator/txscript"
	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
	"github.com/kaspanet/kaspad/domain/consensus/utils/transactionhelper"
)

func (v *validator) checkTransactionInContext(tx *model.DomainTransaction, ghostdagData *model.BlockGHOSTDAGData,
	utxoEntries []*model.UTXOEntry, selectedParentMedianTime int64) (txFee uint64, err error) {

	err = v.checkTxCoinbaseMaturity(ghostdagData, tx, utxoEntries)
	if err != nil {
		return 0, nil
	}

	totalSompiIn, err := v.checkTxInputAmounts(utxoEntries)
	if err != nil {
		return 0, nil
	}

	totalSompiOut, err := v.checkTxOutputAmounts(tx, totalSompiIn)
	if err != nil {
		return 0, nil
	}

	txFee = totalSompiIn - totalSompiOut

	err = v.checkTxSequenceLock(ghostdagData, tx, utxoEntries, selectedParentMedianTime)
	if err != nil {
		return 0, nil
	}

	err = v.validateTransactionScripts(tx, utxoEntries)
	if err != nil {
		return 0, err
	}

	return txFee, nil
}

func (v *validator) checkTxCoinbaseMaturity(
	ghostdagData *model.BlockGHOSTDAGData, tx *model.DomainTransaction, utxoEntries []*model.UTXOEntry) error {
	txBlueScore := ghostdagData.BlueScore
	for i, txIn := range tx.Inputs {
		utxoEntry := utxoEntries[i]

		if utxoEntry.IsCoinbase {
			originBlueScore := utxoEntry.BlockBlueScore
			blueScoreSincePrev := txBlueScore - originBlueScore
			if blueScoreSincePrev < v.blockCoinbaseMaturity {

				return ruleerrors.Errorf(ruleerrors.ErrImmatureSpend, "tried to spend coinbase "+
					"transaction output %s from blue score %d "+
					"to blue score %d before required maturity "+
					"of %d", txIn.PreviousOutpoint,
					originBlueScore, txBlueScore,
					v.blockCoinbaseMaturity)
			}
		}
	}

	return nil
}

func (v *validator) checkTxInputAmounts(inputUTXOEntries []*model.UTXOEntry) (totalSompiIn uint64, err error) {

	totalSompiIn = 0

	for _, utxoEntry := range inputUTXOEntries {

		// Ensure the transaction amounts are in range. Each of the
		// output values of the input transactions must not be negative
		// or more than the max allowed per transaction. All amounts in
		// a transaction are in a unit value known as a sompi. One
		// kaspa is a quantity of sompi as defined by the
		// SompiPerKaspa constant.
		totalSompiIn, err = v.checkEntryAmounts(utxoEntry, totalSompiIn)
		if err != nil {
			return 0, err
		}
	}

	return totalSompiIn, nil
}

func (v *validator) checkEntryAmounts(entry *model.UTXOEntry, totalSompiInBefore uint64) (totalSompiInAfter uint64, err error) {
	// The total of all outputs must not be more than the max
	// allowed per transaction. Also, we could potentially overflow
	// the accumulator so check for overflow.
	originTxSompi := entry.Amount
	totalSompiInAfter = totalSompiInBefore + originTxSompi
	if totalSompiInAfter < totalSompiInBefore ||
		totalSompiInAfter > maxSompi {
		return 0, ruleerrors.Errorf(ruleerrors.ErrBadTxOutValue, "total value of all transaction "+
			"inputs is %d which is higher than max "+
			"allowed value of %d", totalSompiInBefore,
			maxSompi)
	}
	return totalSompiInAfter, nil
}

func (v *validator) checkTxOutputAmounts(tx *model.DomainTransaction, totalSompiIn uint64) (uint64, error) {
	totalSompiOut := uint64(0)
	// Calculate the total output amount for this transaction. It is safe
	// to ignore overflow and out of range errors here because those error
	// conditions would have already been caught by checkTransactionSanity.
	for _, output := range tx.Outputs {
		totalSompiOut += output.Value
	}

	// Ensure the transaction does not spend more than its inputs.
	if totalSompiIn < totalSompiOut {
		return 0, ruleerrors.Errorf(ruleerrors.ErrSpendTooHigh, "total value of all transaction inputs for "+
			"the transaction is %d which is less than the amount "+
			"spent of %d", totalSompiIn, totalSompiOut)
	}
	return totalSompiOut, nil
}

func (v *validator) checkTxSequenceLock(ghostdagData *model.BlockGHOSTDAGData, tx *model.DomainTransaction,
	referencedUTXOEntries []*model.UTXOEntry, medianTime int64) error {

	// A transaction can only be included within a block
	// once the sequence locks of *all* its inputs are
	// active.
	sequenceLock, err := v.calcTxSequenceLockFromReferencedUTXOEntries(ghostdagData, tx, referencedUTXOEntries)
	if err != nil {
		return err
	}

	if !v.sequenceLockActive(sequenceLock, ghostdagData.BlueScore, medianTime) {
		return ruleerrors.Errorf(ruleerrors.ErrUnfinalizedTx, "block contains "+
			"transaction whose input sequence "+
			"locks are not met")
	}

	return nil
}

func (v *validator) validateTransactionScripts(tx *model.DomainTransaction, utxoEntries []*model.UTXOEntry) error {
	for i, input := range tx.Inputs {
		// Create a new script engine for the script pair.
		sigScript := input.SignatureScript
		scriptPubKey := utxoEntries[i].ScriptPublicKey
		vm, err := txscript.NewEngine(scriptPubKey, tx,
			i, txscript.ScriptNoFlags, nil)
		if err != nil {
			return ruleerrors.Errorf(ruleerrors.ErrScriptMalformed, "failed to parse input "+
				"%d which references output %s - "+
				"%s (input script bytes %x, prev "+
				"output script bytes %x)",
				i,
				input.PreviousOutpoint, err, sigScript, scriptPubKey)
		}

		// Execute the script pair.
		if err := vm.Execute(); err != nil {
			return ruleerrors.Errorf(ruleerrors.ErrScriptValidation, "failed to validate input "+
				"%d which references output %s - "+
				"%s (input script bytes %x, prev output "+
				"script bytes %x)",
				i,
				input.PreviousOutpoint, err, sigScript, scriptPubKey)
		}
	}

	return nil
}

func (v *validator) calcTxSequenceLockFromReferencedUTXOEntries(
	ghostdagData *model.BlockGHOSTDAGData, tx *model.DomainTransaction, referencedUTXOEntries []*model.UTXOEntry) (*sequenceLock, error) {

	// A value of -1 for each relative lock type represents a relative time
	// lock value that will allow a transaction to be included in a block
	// at any given height or time.
	sequenceLock := &sequenceLock{Milliseconds: -1, BlockBlueScore: -1}

	// Sequence locks don't apply to coinbase transactions Therefore, we
	// return sequence lock values of -1 indicating that this transaction
	// can be included within a block at any given height or time.
	if transactionhelper.IsCoinBase(tx) {
		return sequenceLock, nil
	}

	for i, input := range tx.Inputs {
		utxoEntry := referencedUTXOEntries[i]

		// If the input blue score is set to the mempool blue score, then we
		// assume the transaction makes it into the next block when
		// evaluating its sequence blocks.
		inputBlueScore := utxoEntry.BlockBlueScore

		// Given a sequence number, we apply the relative time lock
		// mask in order to obtain the time lock delta required before
		// this input can be spent.
		sequenceNum := input.Sequence
		relativeLock := int64(sequenceNum & appmessage.SequenceLockTimeMask)

		switch {
		// Relative time locks are disabled for this input, so we can
		// skip any further calculation.
		case sequenceNum&appmessage.SequenceLockTimeDisabled == appmessage.SequenceLockTimeDisabled:
			continue
		case sequenceNum&appmessage.SequenceLockTimeIsSeconds == appmessage.SequenceLockTimeIsSeconds:
			// This input requires a relative time lock expressed
			// in seconds before it can be spent. Therefore, we
			// need to query for the block prior to the one in
			// which this input was accepted within so we can
			// compute the past median time for the block prior to
			// the one which accepted this referenced output.
			baseGHOSTDAGData := ghostdagData

			for {
				selectedParentGHOSTDAGData, err := v.ghostdagManager.BlockData(baseGHOSTDAGData.SelectedParent)
				if err != nil {
					return nil, err
				}

				if selectedParentGHOSTDAGData.BlueScore <= inputBlueScore {
					break
				}

				baseGHOSTDAGData = selectedParentGHOSTDAGData
			}

			medianTime, err := v.pastMedianTimeManager.PastMedianTime(baseGHOSTDAGData)
			if err != nil {
				return nil, err
			}

			// Time based relative time-locks have a time granularity of
			// appmessage.SequenceLockTimeGranularity, so we shift left by this
			// amount to convert to the proper relative time-lock. We also
			// subtract one from the relative lock to maintain the original
			// lockTime semantics.
			timeLockMilliseconds := (relativeLock << appmessage.SequenceLockTimeGranularity) - 1
			timeLock := medianTime + timeLockMilliseconds
			if timeLock > sequenceLock.Milliseconds {
				sequenceLock.Milliseconds = timeLock
			}
		default:
			// The relative lock-time for this input is expressed
			// in blocks so we calculate the relative offset from
			// the input's blue score as its converted absolute
			// lock-time. We subtract one from the relative lock in
			// order to maintain the original lockTime semantics.
			blockBlueScore := int64(inputBlueScore) + relativeLock - 1
			if blockBlueScore > sequenceLock.BlockBlueScore {
				sequenceLock.BlockBlueScore = blockBlueScore
			}
		}
	}

	return sequenceLock, nil
}

// sequenceLock represents the converted relative lock-time in seconds, and
// absolute block-blue-score for a transaction input's relative lock-times.
// According to sequenceLock, after the referenced input has been confirmed
// within a block, a transaction spending that input can be included into a
// block either after 'seconds' (according to past median time), or once the
// 'BlockBlueScore' has been reached.
type sequenceLock struct {
	Milliseconds   int64
	BlockBlueScore int64
}

// sequenceLockActive determines if a transaction's sequence locks have been
// met, meaning that all the inputs of a given transaction have reached a
// blue score or time sufficient for their relative lock-time maturity.
func (v *validator) sequenceLockActive(sequenceLock *sequenceLock, blockBlueScore uint64,
	medianTimePast int64) bool {

	// If either the milliseconds, or blue score relative-lock time has not yet
	// reached, then the transaction is not yet mature according to its
	// sequence locks.
	if sequenceLock.Milliseconds >= medianTimePast ||
		sequenceLock.BlockBlueScore >= int64(blockBlueScore) {
		return false
	}

	return true
}
