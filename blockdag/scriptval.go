// Copyright (c) 2013-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package blockdag

import (
	"fmt"
	"math"
	"runtime"
	"time"

	"github.com/daglabs/btcd/txscript"
	"github.com/daglabs/btcd/util"
	"github.com/daglabs/btcd/wire"
)

// txValidateItem holds a transaction along with which input to validate.
type txValidateItem struct {
	txInIndex int
	txIn      *wire.TxIn
	tx        *util.Tx
}

// txValidator provides a type which asynchronously validates transaction
// inputs.  It provides several channels for communication and a processing
// function that is intended to be in run multiple goroutines.
type txValidator struct {
	validateChan chan *txValidateItem
	quitChan     chan struct{}
	resultChan   chan error
	utxoSet      UTXOSet
	flags        txscript.ScriptFlags
	sigCache     *txscript.SigCache
}

// sendResult sends the result of a script pair validation on the internal
// result channel while respecting the quit channel.  This allows orderly
// shutdown when the validation process is aborted early due to a validation
// error in one of the other goroutines.
func (v *txValidator) sendResult(result error) {
	select {
	case v.resultChan <- result:
	case <-v.quitChan:
	}
}

// validateHandler consumes items to validate from the internal validate channel
// and returns the result of the validation on the internal result channel. It
// must be run as a goroutine.
func (v *txValidator) validateHandler() {
out:
	for {
		select {
		case txVI := <-v.validateChan:
			// Ensure the referenced input utxo is available.
			txIn := txVI.txIn
			entry, ok := v.utxoSet.Get(txIn.PreviousOutPoint)
			if !ok {
				str := fmt.Sprintf("unable to find unspent "+
					"output %v referenced from "+
					"transaction %s:%d",
					txIn.PreviousOutPoint, txVI.tx.ID(),
					txVI.txInIndex)
				err := ruleError(ErrMissingTxOut, str)
				v.sendResult(err)
				break out
			}

			// Create a new script engine for the script pair.
			sigScript := txIn.SignatureScript
			pkScript := entry.PkScript()
			vm, err := txscript.NewEngine(pkScript, txVI.tx.MsgTx(),
				txVI.txInIndex, v.flags, v.sigCache)
			if err != nil {
				str := fmt.Sprintf("failed to parse input "+
					"%s:%d which references output %v - "+
					"%v (input script bytes %x, prev "+
					"output script bytes %x)",
					txVI.tx.ID(), txVI.txInIndex,
					txIn.PreviousOutPoint, err, sigScript, pkScript)
				err := ruleError(ErrScriptMalformed, str)
				v.sendResult(err)
				break out
			}

			// Execute the script pair.
			if err := vm.Execute(); err != nil {
				str := fmt.Sprintf("failed to validate input "+
					"%s:%d which references output %v - "+
					"%v (input script bytes %x, prev output "+
					"script bytes %x)",
					txVI.tx.ID(), txVI.txInIndex,
					txIn.PreviousOutPoint, err, sigScript, pkScript)
				err := ruleError(ErrScriptValidation, str)
				v.sendResult(err)
				break out
			}

			// Validation succeeded.
			v.sendResult(nil)

		case <-v.quitChan:
			break out
		}
	}
}

// Validate validates the scripts for all of the passed transaction inputs using
// multiple goroutines.
func (v *txValidator) Validate(items []*txValidateItem) error {
	if len(items) == 0 {
		return nil
	}

	// Limit the number of goroutines to do script validation based on the
	// number of processor cores.  This helps ensure the system stays
	// reasonably responsive under heavy load.
	maxGoRoutines := runtime.NumCPU() * 3
	if maxGoRoutines <= 0 {
		maxGoRoutines = 1
	}
	if maxGoRoutines > len(items) {
		maxGoRoutines = len(items)
	}

	// Start up validation handlers that are used to asynchronously
	// validate each transaction input.
	for i := 0; i < maxGoRoutines; i++ {
		go v.validateHandler()
	}

	// Validate each of the inputs.  The quit channel is closed when any
	// errors occur so all processing goroutines exit regardless of which
	// input had the validation error.
	numInputs := len(items)
	currentItem := 0
	processedItems := 0
	for processedItems < numInputs {
		// Only send items while there are still items that need to
		// be processed.  The select statement will never select a nil
		// channel.
		var validateChan chan *txValidateItem
		var item *txValidateItem
		if currentItem < numInputs {
			validateChan = v.validateChan
			item = items[currentItem]
		}

		select {
		case validateChan <- item:
			currentItem++

		case err := <-v.resultChan:
			processedItems++
			if err != nil {
				close(v.quitChan)
				return err
			}
		}
	}

	close(v.quitChan)
	return nil
}

// newTxValidator returns a new instance of txValidator to be used for
// validating transaction scripts asynchronously.
func newTxValidator(utxoSet UTXOSet, flags txscript.ScriptFlags, sigCache *txscript.SigCache) *txValidator {
	return &txValidator{
		validateChan: make(chan *txValidateItem),
		quitChan:     make(chan struct{}),
		resultChan:   make(chan error),
		utxoSet:      utxoSet,
		sigCache:     sigCache,
		flags:        flags,
	}
}

// ValidateTransactionScripts validates the scripts for the passed transaction
// using multiple goroutines.
func ValidateTransactionScripts(tx *util.Tx, utxoSet UTXOSet, flags txscript.ScriptFlags, sigCache *txscript.SigCache) error {
	// Collect all of the transaction inputs and required information for
	// validation.
	txIns := tx.MsgTx().TxIn
	txValItems := make([]*txValidateItem, 0, len(txIns))
	for txInIdx, txIn := range txIns {
		// Skip block reward transactions.
		if txIn.PreviousOutPoint.Index == math.MaxUint32 {
			continue
		}

		txVI := &txValidateItem{
			txInIndex: txInIdx,
			txIn:      txIn,
			tx:        tx,
		}
		txValItems = append(txValItems, txVI)
	}

	// Validate all of the inputs.
	validator := newTxValidator(utxoSet, flags, sigCache)
	return validator.Validate(txValItems)
}

// checkBlockScripts executes and validates the scripts for all transactions in
// the passed block using multiple goroutines.
func checkBlockScripts(block *blockNode, utxoSet UTXOSet, transactions []*util.Tx, scriptFlags txscript.ScriptFlags, sigCache *txscript.SigCache) error {
	// Collect all of the transaction inputs and required information for
	// validation for all transactions in the block into a single slice.
	numInputs := 0
	for _, tx := range transactions {
		numInputs += len(tx.MsgTx().TxIn)
	}
	txValItems := make([]*txValidateItem, 0, numInputs)
	for _, tx := range transactions {
		for txInIdx, txIn := range tx.MsgTx().TxIn {
			// Skip block reward transactions.
			if txIn.PreviousOutPoint.Index == math.MaxUint32 {
				continue
			}

			txVI := &txValidateItem{
				txInIndex: txInIdx,
				txIn:      txIn,
				tx:        tx,
			}
			txValItems = append(txValItems, txVI)
		}
	}

	// Validate all of the inputs.
	validator := newTxValidator(utxoSet, scriptFlags, sigCache)
	start := time.Now()
	if err := validator.Validate(txValItems); err != nil {
		return err
	}
	elapsed := time.Since(start)

	log.Tracef("block %v took %v to verify", block.hash, elapsed)

	return nil
}
