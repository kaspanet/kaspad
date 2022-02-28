package server

import (
	"github.com/kaspanet/go-secp256k1"
	"github.com/pkg/errors"

	"github.com/kaspanet/kaspad/cmd/kaspawallet/libkaspawallet"
	"github.com/kaspanet/kaspad/cmd/kaspawallet/libkaspawallet/serialization"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
	"github.com/kaspanet/kaspad/domain/consensus/utils/constants"
	"github.com/kaspanet/kaspad/domain/consensus/utils/utxo"
	"github.com/kaspanet/kaspad/domain/miningmanager/mempool"
	"github.com/kaspanet/kaspad/util"
)

// maybeAutoCompoundTransaction checks if a transaction's mass is higher that what is allowed for a standard
// transaction.
// If it is - the transaction is split into multiple transactions, each with a portion of the inputs and a single output
// into a change address.
// An additional `mergeTransaction` is generated - which merges the outputs of the above splits into a single output
// paying to the original transaction's payee.
func (s *server) maybeAutoCompoundTransaction(transactionBytes []byte, toAddress util.Address,
	changeAddress util.Address, changeWalletAddress *walletAddress) ([][]byte, error) {
	transaction, err := serialization.DeserializePartiallySignedTransaction(transactionBytes)
	if err != nil {
		return nil, err
	}

	splitTransactions, err := s.maybeSplitTransaction(transaction, changeAddress)
	if err != nil {
		return nil, err
	}
	if len(splitTransactions) > 1 {
		mergeTransaction, err := s.mergeTransaction(splitTransactions, transaction, toAddress, changeAddress, changeWalletAddress)
		if err != nil {
			return nil, err
		}
		splitTransactions = append(splitTransactions, mergeTransaction)
	}

	splitTransactionsBytes := make([][]byte, len(splitTransactions))
	for i, splitTransaction := range splitTransactions {
		splitTransactionsBytes[i], err = serialization.SerializePartiallySignedTransaction(splitTransaction)
		if err != nil {
			return nil, err
		}
	}
	return splitTransactionsBytes, nil
}

func (s *server) mergeTransaction(
	splitTransactions []*serialization.PartiallySignedTransaction,
	originalTransaction *serialization.PartiallySignedTransaction,
	toAddress util.Address,
	changeAddress util.Address,
	changeWalletAddress *walletAddress,
) (*serialization.PartiallySignedTransaction, error) {
	numOutputs := len(originalTransaction.Tx.Outputs)
	if numOutputs > 2 || numOutputs == 0 {
		return nil, errors.Errorf("original transaction has %d outputs, while 1 or 2 are expected",
			len(originalTransaction.Tx.Outputs))
	}

	totalValue := uint64(0)
	sentValue := originalTransaction.Tx.Outputs[0].Value
	utxos := make([]*libkaspawallet.UTXO, len(splitTransactions))
	for i, splitTransaction := range splitTransactions {
		output := splitTransaction.Tx.Outputs[0]
		utxos[i] = &libkaspawallet.UTXO{
			Outpoint: &externalapi.DomainOutpoint{
				TransactionID: *consensushashing.TransactionID(splitTransaction.Tx),
				Index:         0,
			},
			UTXOEntry:      utxo.NewUTXOEntry(output.Value, output.ScriptPublicKey, false, constants.UnacceptedDAAScore),
			DerivationPath: s.walletAddressPath(changeWalletAddress),
		}
		totalValue += output.Value
		totalValue -= feePerInput
	}

	if totalValue < sentValue {
		// sometimes the fees from compound transactions make the total output higher than what's available from selected
		// utxos, in such cases - find one more UTXO and use it.
		oneMoreUTXO, err := s.oneMoreUTXOForMergeTransaction(utxos, sentValue-totalValue)
		if err != nil {
			return nil, err
		}
		utxos = append(utxos, oneMoreUTXO)
		totalValue += oneMoreUTXO.UTXOEntry.Amount()
	}

	payments := []*libkaspawallet.Payment{{
		Address: toAddress,
		Amount:  sentValue,
	}}
	if totalValue > sentValue {
		payments = append(payments, &libkaspawallet.Payment{
			Address: changeAddress,
			Amount:  totalValue - sentValue,
		})
	}

	mergeTransactionBytes, err := libkaspawallet.CreateUnsignedTransaction(s.keysFile.ExtendedPublicKeys,
		s.keysFile.MinimumSignatures, payments, utxos)
	if err != nil {
		return nil, err
	}

	return serialization.DeserializePartiallySignedTransaction(mergeTransactionBytes)
}

func (s *server) maybeSplitTransaction(transaction *serialization.PartiallySignedTransaction,
	changeAddress util.Address) ([]*serialization.PartiallySignedTransaction, error) {

	transactionMass, err := s.estimateMassAfterSignatures(transaction)
	if err != nil {
		return nil, err
	}

	if transactionMass < mempool.MaximumStandardTransactionMass {
		return []*serialization.PartiallySignedTransaction{transaction}, nil
	}

	splitCount := int(transactionMass / mempool.MaximumStandardTransactionMass)
	if transactionMass%mempool.MaximumStandardTransactionMass > 0 {
		splitCount++
	}
	inputCountPerSplit := len(transaction.Tx.Inputs) / splitCount
	if len(transaction.Tx.Inputs)%splitCount > 0 {
		// note we are incrementing splitCount, and not inputCountPerSplit, since incrementing inputCountPerSplit
		// might make the transaction mass too high
		splitCount++
	}

	splitTransactions := make([]*serialization.PartiallySignedTransaction, splitCount)
	for i := 0; i < splitCount; i++ {
		startIndex := i * inputCountPerSplit
		endIndex := startIndex + inputCountPerSplit
		var err error
		splitTransactions[i], err = s.createSplitTransaction(transaction, changeAddress, startIndex, endIndex)
		if err != nil {
			return nil, err
		}
	}

	return splitTransactions, nil
}

func (s *server) createSplitTransaction(transaction *serialization.PartiallySignedTransaction,
	changeAddress util.Address, startIndex int, endIndex int) (*serialization.PartiallySignedTransaction, error) {

	selectedUTXOs := make([]*libkaspawallet.UTXO, 0, endIndex-startIndex)
	totalSompi := uint64(0)

	for i := startIndex; i < endIndex && i < len(transaction.PartiallySignedInputs); i++ {
		partiallySignedInput := transaction.PartiallySignedInputs[i]
		selectedUTXOs = append(selectedUTXOs, &libkaspawallet.UTXO{
			Outpoint: &transaction.Tx.Inputs[i].PreviousOutpoint,
			UTXOEntry: utxo.NewUTXOEntry(
				partiallySignedInput.PrevOutput.Value, partiallySignedInput.PrevOutput.ScriptPublicKey,
				false, constants.UnacceptedDAAScore),
			DerivationPath: partiallySignedInput.DerivationPath,
		})

		totalSompi += selectedUTXOs[i-startIndex].UTXOEntry.Amount()
		totalSompi -= feePerInput
	}
	unsignedTransactionBytes, err := libkaspawallet.CreateUnsignedTransaction(s.keysFile.ExtendedPublicKeys,
		s.keysFile.MinimumSignatures,
		[]*libkaspawallet.Payment{{
			Address: changeAddress,
			Amount:  totalSompi,
		}}, selectedUTXOs)
	if err != nil {
		return nil, err
	}

	return serialization.DeserializePartiallySignedTransaction(unsignedTransactionBytes)
}

func (s *server) estimateMassAfterSignatures(transaction *serialization.PartiallySignedTransaction) (uint64, error) {
	transaction = transaction.Clone()
	var signatureSize uint64
	if s.keysFile.ECDSA {
		signatureSize = secp256k1.SerializedECDSASignatureSize
	} else {
		signatureSize = secp256k1.SerializedSchnorrSignatureSize
	}

	for i, input := range transaction.PartiallySignedInputs {
		for j, pubKeyPair := range input.PubKeySignaturePairs {
			if uint32(j) >= s.keysFile.MinimumSignatures {
				break
			}
			pubKeyPair.Signature = make([]byte, signatureSize+1) // +1 for SigHashType
		}
		transaction.Tx.Inputs[i].SigOpCount = byte(len(input.PubKeySignaturePairs))
	}

	transactionWithSignatures, err := libkaspawallet.ExtractTransactionDeserialized(transaction, s.keysFile.ECDSA)
	if err != nil {
		return 0, err
	}

	return s.txMassCalculator.CalculateTransactionMass(transactionWithSignatures), nil
}
