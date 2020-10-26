package transactionvalidator

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
	"github.com/kaspanet/kaspad/domain/consensus/utils/stringers"
	"github.com/kaspanet/kaspad/domain/consensus/utils/transactionhelper"
	"github.com/kaspanet/kaspad/domain/consensus/utils/txscript"
	"github.com/pkg/errors"
)

// ValidateTransactionInContextAndPopulateMassAndFee validates the transaction against its referenced UTXO, and
// populates its mass and fee fields.
//
// Note: if the function fails, there's no guarantee that the transaction mass and fee fields will remain unaffected.
func (v *transactionValidator) ValidateTransactionInContextAndPopulateMassAndFee(tx *externalapi.DomainTransaction,
	povBlockHash *externalapi.DomainHash, selectedParentMedianTime int64) error {

	err := v.checkTransactionCoinbaseMaturity(povBlockHash, tx)
	if err != nil {
		return nil
	}

	totalSompiIn, err := v.checkTransactionInputAmounts(tx)
	if err != nil {
		return nil
	}

	totalSompiOut, err := v.checkTransactionOutputAmounts(tx, totalSompiIn)
	if err != nil {
		return nil
	}

	tx.Fee = totalSompiIn - totalSompiOut

	err = v.checkTransactionSequenceLock(povBlockHash, tx, selectedParentMedianTime)
	if err != nil {
		return nil
	}

	err = v.validateTransactionScripts(tx)
	if err != nil {
		return err
	}

	tx.Mass, err = v.transactionMass(tx)
	if err != nil {
		return err
	}

	return nil
}

func (v *transactionValidator) checkTransactionCoinbaseMaturity(
	povBlockHash *externalapi.DomainHash, tx *externalapi.DomainTransaction) error {

	ghostdagData, err := v.ghostdagDataStore.Get(v.databaseContext, povBlockHash)
	if err != nil {
		return err
	}

	txBlueScore := ghostdagData.BlueScore
	for _, txIn := range tx.Inputs {
		utxoEntry := txIn.UTXOEntry
		if utxoEntry == nil {
			return errors.Wrapf(ruleerrors.ErrMissingTxOut, "outpoint %s "+
				"either does not exist or "+
				"has already been spent", stringers.Outpoint(&txIn.PreviousOutpoint))
		}

		if utxoEntry.IsCoinbase {
			originBlueScore := utxoEntry.BlockBlueScore
			blueScoreSincePrev := txBlueScore - originBlueScore
			if blueScoreSincePrev < v.blockCoinbaseMaturity {
				return errors.Wrapf(ruleerrors.ErrImmatureSpend, "tried to spend coinbase "+
					"transaction output %s from blue score %d "+
					"to blue score %d before required maturity "+
					"of %d", stringers.Outpoint(&txIn.PreviousOutpoint),
					originBlueScore, txBlueScore,
					v.blockCoinbaseMaturity)
			}
		}
	}

	return nil
}

func (v *transactionValidator) checkTransactionInputAmounts(tx *externalapi.DomainTransaction) (totalSompiIn uint64, err error) {

	totalSompiIn = 0

	for _, input := range tx.Inputs {
		utxoEntry := input.UTXOEntry
		if utxoEntry == nil {
			return 0, errors.Wrapf(ruleerrors.ErrMissingTxOut, "output %s "+
				"either does not exist or "+
				"has already been spent", stringers.Outpoint(&input.PreviousOutpoint))
		}

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

func (v *transactionValidator) checkEntryAmounts(entry *externalapi.UTXOEntry, totalSompiInBefore uint64) (totalSompiInAfter uint64, err error) {
	// The total of all outputs must not be more than the max
	// allowed per transaction. Also, we could potentially overflow
	// the accumulator so check for overflow.
	originTxSompi := entry.Amount
	totalSompiInAfter = totalSompiInBefore + originTxSompi
	if totalSompiInAfter < totalSompiInBefore ||
		totalSompiInAfter > maxSompi {
		return 0, errors.Wrapf(ruleerrors.ErrBadTxOutValue, "total value of all transaction "+
			"inputs is %d which is higher than max "+
			"allowed value of %d", totalSompiInBefore,
			maxSompi)
	}
	return totalSompiInAfter, nil
}

func (v *transactionValidator) checkTransactionOutputAmounts(tx *externalapi.DomainTransaction, totalSompiIn uint64) (uint64, error) {
	totalSompiOut := uint64(0)
	// Calculate the total output amount for this transaction. It is safe
	// to ignore overflow and out of range errors here because those error
	// conditions would have already been caught by checkTransactionSanity.
	for _, output := range tx.Outputs {
		totalSompiOut += output.Value
	}

	// Ensure the transaction does not spend more than its inputs.
	if totalSompiIn < totalSompiOut {
		return 0, errors.Wrapf(ruleerrors.ErrSpendTooHigh, "total value of all transaction inputs for "+
			"the transaction is %d which is less than the amount "+
			"spent of %d", totalSompiIn, totalSompiOut)
	}
	return totalSompiOut, nil
}

func (v *transactionValidator) checkTransactionSequenceLock(povBlockHash *externalapi.DomainHash,
	tx *externalapi.DomainTransaction, medianTime int64) error {

	// A transaction can only be included within a block
	// once the sequence locks of *all* its inputs are
	// active.
	sequenceLock, err := v.calcTxSequenceLockFromReferencedUTXOEntries(povBlockHash, tx)
	if err != nil {
		return err
	}

	ghostdagData, err := v.ghostdagDataStore.Get(v.databaseContext, povBlockHash)
	if err != nil {
		return err
	}

	if !v.sequenceLockActive(sequenceLock, ghostdagData.BlueScore, medianTime) {
		return errors.Wrapf(ruleerrors.ErrUnfinalizedTx, "block contains "+
			"transaction whose input sequence "+
			"locks are not met")
	}

	return nil
}

func (v *transactionValidator) validateTransactionScripts(tx *externalapi.DomainTransaction) error {
	for i, input := range tx.Inputs {
		// Create a new script engine for the script pair.
		sigScript := input.SignatureScript
		utxoEntry := input.UTXOEntry
		if utxoEntry == nil {
			return errors.Wrapf(ruleerrors.ErrMissingTxOut, "output %s "+
				"either does not exist or "+
				"has already been spent", stringers.Outpoint(&input.PreviousOutpoint))
		}

		scriptPubKey := utxoEntry.ScriptPublicKey
		vm, err := txscript.NewEngine(scriptPubKey, tx,
			i, txscript.ScriptNoFlags, nil)
		if err != nil {
			return errors.Wrapf(ruleerrors.ErrScriptMalformed, "failed to parse input "+
				"%d which references output %s - "+
				"%s (input script bytes %x, prev "+
				"output script bytes %x)",
				i,
				stringers.Outpoint(&input.PreviousOutpoint), err, sigScript, scriptPubKey)
		}

		// Execute the script pair.
		if err := vm.Execute(); err != nil {
			return errors.Wrapf(ruleerrors.ErrScriptValidation, "failed to validate input "+
				"%d which references output %s - "+
				"%s (input script bytes %x, prev output "+
				"script bytes %x)",
				i,
				stringers.Outpoint(&input.PreviousOutpoint), err, sigScript, scriptPubKey)
		}
	}

	return nil
}

func (v *transactionValidator) calcTxSequenceLockFromReferencedUTXOEntries(
	povBlockHash *externalapi.DomainHash, tx *externalapi.DomainTransaction) (*sequenceLock, error) {

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

	for _, input := range tx.Inputs {
		utxoEntry := input.UTXOEntry
		if utxoEntry == nil {
			return nil, errors.Wrapf(ruleerrors.ErrMissingTxOut, "output %s "+
				"either does not exist or "+
				"has already been spent", stringers.Outpoint(&input.PreviousOutpoint))
		}

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
			baseGHOSTDAGData, err := v.ghostdagDataStore.Get(v.databaseContext, povBlockHash)
			if err != nil {
				return nil, err
			}

			baseHash := povBlockHash

			for {
				selectedParentGHOSTDAGData, err := v.ghostdagDataStore.Get(v.databaseContext,
					baseGHOSTDAGData.SelectedParent)
				if err != nil {
					return nil, err
				}

				if selectedParentGHOSTDAGData.BlueScore <= inputBlueScore {
					break
				}

				baseHash = baseGHOSTDAGData.SelectedParent
				baseGHOSTDAGData = selectedParentGHOSTDAGData
			}

			medianTime, err := v.pastMedianTimeManager.PastMedianTime(baseHash)
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
func (v *transactionValidator) sequenceLockActive(sequenceLock *sequenceLock, blockBlueScore uint64,
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
