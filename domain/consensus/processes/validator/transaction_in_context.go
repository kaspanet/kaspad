package validator

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/processes/validator/txscript"
	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/util/mstime"
)

func (v *validator) checkTransactionInContext(tx *model.DomainTransaction, ghostdagData *model.BlockGHOSTDAGData,
	utxoEntries []*model.UTXOEntry) (txFee uint64, err error) {

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

	err = v.checkTxSequenceLock(node, tx, utxoEntries, selectedParentMedianTime)
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

func (v *validator) checkTxSequenceLock(node *blockNode, tx *util.Tx,
	referencedUTXOEntries []*UTXOEntry, medianTime mstime.Time) error {

	// A transaction can only be included within a block
	// once the sequence locks of *all* its inputs are
	// active.
	sequenceLock, err := dag.calcTxSequenceLockFromReferencedUTXOEntries(node, tx, referencedUTXOEntries)
	if err != nil {
		return err
	}
	if !SequenceLockActive(sequenceLock, node.blueScore, medianTime) {
		str := fmt.Sprintf("block contains " +
			"transaction whose input sequence " +
			"locks are not met")
		return ruleError(ErrUnfinalizedTx, str)
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
