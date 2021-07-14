package transactionvalidator

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
	"github.com/kaspanet/kaspad/domain/consensus/utils/constants"
	"github.com/kaspanet/kaspad/domain/consensus/utils/transactionhelper"
	"github.com/kaspanet/kaspad/domain/consensus/utils/txscript"
	"github.com/pkg/errors"
)

// ValidateTransactionInContextAndPopulateMassAndFee validates the transaction against its referenced UTXO, and
// populates its mass and fee fields.
//
// Note: if the function fails, there's no guarantee that the transaction mass and fee fields will remain unaffected.
func (v *transactionValidator) ValidateTransactionInContextAndPopulateMassAndFee(stagingArea *model.StagingArea,
	tx *externalapi.DomainTransaction, povBlockHash *externalapi.DomainHash, selectedParentMedianTime int64) error {

	err := v.checkTransactionCoinbaseMaturity(stagingArea, povBlockHash, tx)
	if err != nil {
		return err
	}

	totalSompiIn, err := v.checkTransactionInputAmounts(tx)
	if err != nil {
		return err
	}

	totalSompiOut, err := v.checkTransactionOutputAmounts(tx, totalSompiIn)
	if err != nil {
		return err
	}

	tx.Fee = totalSompiIn - totalSompiOut

	err = v.checkTransactionSequenceLock(stagingArea, povBlockHash, tx, selectedParentMedianTime)
	if err != nil {
		return err
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

func (v *transactionValidator) checkTransactionCoinbaseMaturity(stagingArea *model.StagingArea,
	povBlockHash *externalapi.DomainHash, tx *externalapi.DomainTransaction) error {

	povDAAScore, err := v.daaBlocksStore.DAAScore(v.databaseContext, stagingArea, povBlockHash)
	if err != nil {
		return err
	}

	var missingOutpoints []*externalapi.DomainOutpoint
	for _, input := range tx.Inputs {
		utxoEntry := input.UTXOEntry
		if utxoEntry == nil {
			missingOutpoints = append(missingOutpoints, &input.PreviousOutpoint)
		} else if utxoEntry.IsCoinbase() {
			originDAAScore := utxoEntry.BlockDAAScore()
			if originDAAScore+v.blockCoinbaseMaturity > povDAAScore {
				return errors.Wrapf(ruleerrors.ErrImmatureSpend, "tried to spend coinbase "+
					"transaction output %s from DAA score %d "+
					"to DAA score %d before required maturity "+
					"of %d", input.PreviousOutpoint,
					originDAAScore, povDAAScore,
					v.blockCoinbaseMaturity)
			}
		}
	}
	if len(missingOutpoints) > 0 {
		return ruleerrors.NewErrMissingTxOut(missingOutpoints)
	}

	return nil
}

func (v *transactionValidator) checkTransactionInputAmounts(tx *externalapi.DomainTransaction) (totalSompiIn uint64, err error) {
	totalSompiIn = 0

	var missingOutpoints []*externalapi.DomainOutpoint
	for _, input := range tx.Inputs {
		utxoEntry := input.UTXOEntry
		if utxoEntry == nil {
			missingOutpoints = append(missingOutpoints, &input.PreviousOutpoint)
			continue
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

	if len(missingOutpoints) > 0 {
		return 0, ruleerrors.NewErrMissingTxOut(missingOutpoints)
	}

	return totalSompiIn, nil
}

func (v *transactionValidator) checkEntryAmounts(entry externalapi.UTXOEntry, totalSompiInBefore uint64) (totalSompiInAfter uint64, err error) {
	// The total of all outputs must not be more than the max
	// allowed per transaction. Also, we could potentially overflow
	// the accumulator so check for overflow.
	originTxSompi := entry.Amount()
	totalSompiInAfter = totalSompiInBefore + originTxSompi
	if totalSompiInAfter < totalSompiInBefore ||
		totalSompiInAfter > constants.MaxSompi {
		return 0, errors.Wrapf(ruleerrors.ErrBadTxOutValue, "total value of all transaction "+
			"inputs is %d which is higher than max "+
			"allowed value of %d", totalSompiInBefore,
			constants.MaxSompi)
	}
	return totalSompiInAfter, nil
}

func (v *transactionValidator) checkTransactionOutputAmounts(tx *externalapi.DomainTransaction, totalSompiIn uint64) (uint64, error) {
	totalSompiOut := uint64(0)
	// Calculate the total output amount for this transaction. It is safe
	// to ignore overflow and out of range errors here because those error
	// conditions would have already been caught by checkTransactionAmountRanges.
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

func (v *transactionValidator) checkTransactionSequenceLock(stagingArea *model.StagingArea,
	povBlockHash *externalapi.DomainHash, tx *externalapi.DomainTransaction, medianTime int64) error {

	// A transaction can only be included within a block
	// once the sequence locks of *all* its inputs are
	// active.
	sequenceLock, err := v.calcTxSequenceLockFromReferencedUTXOEntries(stagingArea, povBlockHash, tx)
	if err != nil {
		return err
	}

	daaScore, err := v.daaBlocksStore.DAAScore(v.databaseContext, stagingArea, povBlockHash)
	if err != nil {
		return err
	}

	if !v.sequenceLockActive(sequenceLock, daaScore, medianTime) {
		return errors.Wrapf(ruleerrors.ErrUnfinalizedTx, "block contains "+
			"transaction whose input sequence "+
			"locks are not met")
	}

	return nil
}

func (v *transactionValidator) validateTransactionScripts(tx *externalapi.DomainTransaction) error {
	var missingOutpoints []*externalapi.DomainOutpoint
	sighashReusedValues := &consensushashing.SighashReusedValues{}

	for i, input := range tx.Inputs {
		// Create a new script engine for the script pair.
		sigScript := input.SignatureScript
		utxoEntry := input.UTXOEntry
		if utxoEntry == nil {
			missingOutpoints = append(missingOutpoints, &input.PreviousOutpoint)
			continue
		}

		scriptPubKey := utxoEntry.ScriptPublicKey()
		vm, err := txscript.NewEngine(scriptPubKey, tx, i, txscript.ScriptNoFlags, v.sigCache, v.sigCacheECDSA, sighashReusedValues)
		if err != nil {
			return errors.Wrapf(ruleerrors.ErrScriptMalformed, "failed to parse input "+
				"%d which references output %s - "+
				"%s (input script bytes %x, prev "+
				"output script bytes %x)",
				i,
				input.PreviousOutpoint, err, sigScript, scriptPubKey)
		}

		// Execute the script pair.
		if err := vm.Execute(); err != nil {
			return errors.Wrapf(ruleerrors.ErrScriptValidation, "failed to validate input "+
				"%d which references output %s - "+
				"%s (input script bytes %x, prev output "+
				"script bytes %x)",
				i,
				input.PreviousOutpoint, err, sigScript, scriptPubKey)
		}
	}
	if len(missingOutpoints) > 0 {
		return ruleerrors.NewErrMissingTxOut(missingOutpoints)
	}
	return nil
}

func (v *transactionValidator) calcTxSequenceLockFromReferencedUTXOEntries(stagingArea *model.StagingArea,
	povBlockHash *externalapi.DomainHash, tx *externalapi.DomainTransaction) (*sequenceLock, error) {

	// A value of -1 for each relative lock type represents a relative time
	// lock value that will allow a transaction to be included in a block
	// at any given height or time.
	sequenceLock := &sequenceLock{Milliseconds: -1, BlockDAAScore: -1}

	// Sequence locks don't apply to coinbase transactions Therefore, we
	// return sequence lock values of -1 indicating that this transaction
	// can be included within a block at any given height or time.
	if transactionhelper.IsCoinBase(tx) {
		return sequenceLock, nil
	}

	var missingOutpoints []*externalapi.DomainOutpoint
	for _, input := range tx.Inputs {
		utxoEntry := input.UTXOEntry
		if utxoEntry == nil {
			missingOutpoints = append(missingOutpoints, &input.PreviousOutpoint)
			continue
		}

		inputDAAScore := utxoEntry.BlockDAAScore()

		// Given a sequence number, we apply the relative time lock
		// mask in order to obtain the time lock delta required before
		// this input can be spent.
		sequenceNum := input.Sequence
		relativeLock := int64(sequenceNum & constants.SequenceLockTimeMask)

		// Relative time locks are disabled for this input, so we can
		// skip any further calculation.
		if sequenceNum&constants.SequenceLockTimeDisabled == constants.SequenceLockTimeDisabled {
			continue
		}
		// The relative lock-time for this input is expressed
		// in blocks so we calculate the relative offset from
		// the input's DAA score as its converted absolute
		// lock-time. We subtract one from the relative lock in
		// order to maintain the original lockTime semantics.
		blockDAAScore := int64(inputDAAScore) + relativeLock - 1
		if blockDAAScore > sequenceLock.BlockDAAScore {
			sequenceLock.BlockDAAScore = blockDAAScore
		}
	}
	if len(missingOutpoints) > 0 {
		return nil, ruleerrors.NewErrMissingTxOut(missingOutpoints)
	}

	return sequenceLock, nil
}

// sequenceLock represents the converted relative lock-time in seconds, and
// absolute block-daa-score for a transaction input's relative lock-times.
// According to sequenceLock, after the referenced input has been confirmed
// within a block, a transaction spending that input can be included into a
// block either after 'seconds' (according to past median time), or once the
// 'BlockDAAScore' has been reached.
type sequenceLock struct {
	Milliseconds  int64
	BlockDAAScore int64
}

// sequenceLockActive determines if a transaction's sequence locks have been
// met, meaning that all the inputs of a given transaction have reached a
// DAA score or time sufficient for their relative lock-time maturity.
func (v *transactionValidator) sequenceLockActive(sequenceLock *sequenceLock, blockDAAScore uint64,
	medianTimePast int64) bool {

	// If either the milliseconds, or DAA score relative-lock time has not yet
	// reached, then the transaction is not yet mature according to its
	// sequence locks.
	if sequenceLock.Milliseconds >= medianTimePast ||
		sequenceLock.BlockDAAScore >= int64(blockDAAScore) {
		return false
	}

	return true
}
