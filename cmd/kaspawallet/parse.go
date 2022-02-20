package main

import (
	"encoding/hex"
	"fmt"
	"github.com/kaspanet/kaspad/cmd/kaspawallet/libkaspawallet/serialization"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
	"github.com/kaspanet/kaspad/domain/consensus/utils/constants"
	"github.com/kaspanet/kaspad/domain/consensus/utils/txscript"
	"github.com/pkg/errors"
	"io/ioutil"
	"strings"
)

func parse(conf *parseConfig) error {
	if conf.Transaction == "" && conf.TransactionFile == "" {
		return errors.Errorf("Either --transaction or --transaction-file is required")
	}
	if conf.Transaction != "" && conf.TransactionFile != "" {
		return errors.Errorf("Both --transaction and --transaction-file cannot be passed at the same time")
	}

	transactionHex := conf.Transaction
	if conf.TransactionFile != "" {
		transactionHexBytes, err := ioutil.ReadFile(conf.TransactionFile)
		if err != nil {
			return errors.Wrapf(err, "Could not read hex from %s", conf.TransactionFile)
		}
		transactionHex = strings.TrimSpace(string(transactionHexBytes))
	}

	transaction, err := hex.DecodeString(transactionHex)
	if err != nil {
		return err
	}

	partiallySignedTransaction, err := serialization.DeserializePartiallySignedTransaction(transaction)
	if err != nil {
		return err
	}

	fmt.Printf("Transaction ID: \t%s\n", consensushashing.TransactionID(partiallySignedTransaction.Tx))
	fmt.Println()

	allInputSompi := uint64(0)
	for index, input := range partiallySignedTransaction.Tx.Inputs {
		partiallySignedInput := partiallySignedTransaction.PartiallySignedInputs[index]
		fmt.Printf("Input %d: \tOutpoint: %s:%d \tAmount: %.2f Kaspa\n", index, input.PreviousOutpoint.TransactionID,
			input.PreviousOutpoint.Index, float64(partiallySignedInput.PrevOutput.Value)/float64(constants.SompiPerKaspa))

		allInputSompi += partiallySignedInput.PrevOutput.Value
	}
	fmt.Println()

	allOutputSompi := uint64(0)
	for index, output := range partiallySignedTransaction.Tx.Outputs {
		scriptPublicKeyType, scriptPublicKeyAddress, err := txscript.ExtractScriptPubKeyAddress(output.ScriptPublicKey, conf.ActiveNetParams)
		if err != nil {
			return err
		}

		addressString := scriptPublicKeyAddress.EncodeAddress()
		if scriptPublicKeyType == txscript.NonStandardTy {
			addressString = hex.EncodeToString(output.ScriptPublicKey.Script)
		}

		fmt.Printf("Output %d: \tRecipient: %s \tAmount: %.2f Kaspa\n",
			index, addressString, float64(output.Value)/float64(constants.SompiPerKaspa))

		allOutputSompi += output.Value
	}
	fmt.Println()

	fmt.Printf("Fee:\t%d Sompi\n", allInputSompi-allOutputSompi)

	return nil
}
