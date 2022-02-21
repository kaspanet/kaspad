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

func (s *server) maybeAutoCompoundTransaction(transactionBytes []byte) ([][]byte, error) {
	transaction, err := serialization.DeserializePartiallySignedTransaction(transactionBytes)
	if err != nil {
		return nil, err
	}

	splitAddress, splitWalletAddress, err := s.changeAddress()
	if err != nil {
		return nil, err
	}

	splitTransactions, err := s.maybeSplitTransaction(transaction, splitAddress)
	if err != nil {
		return nil, err
	}
	if len(splitTransactions) > 1 {
		mergeTransaction, err := s.mergeTransaction(splitTransactions, transaction, splitWalletAddress)
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

func (s *server) mergeTransaction(splitTransactions []*serialization.PartiallySignedTransaction,
	originalTransaction *serialization.PartiallySignedTransaction, splitWalletAddress *walletAddress) (
	*serialization.PartiallySignedTransaction, error) {

	if len(originalTransaction.Tx.Outputs) != 2 {
		return nil, errors.Errorf("original transaction has %d outputs, while 2 are expected",
			len(originalTransaction.Tx.Outputs))
	}

	targetAddress, err := util.NewAddressScriptHash(originalTransaction.Tx.Outputs[0].ScriptPublicKey.Script, s.params.Prefix)
	if err != nil {
		return nil, err
	}
	changeAddress, err := util.NewAddressScriptHash(originalTransaction.Tx.Outputs[1].ScriptPublicKey.Script, s.params.Prefix)
	if err != nil {
		return nil, err
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
			DerivationPath: s.walletAddressPath(splitWalletAddress),
		}
		totalValue += output.Value
		totalValue -= feePerInput
	}

	var payments []*libkaspawallet.Payment
	if totalValue >= sentValue {
		payments = []*libkaspawallet.Payment{{
			Address: targetAddress,
			Amount:  sentValue,
		}, {
			Address: changeAddress,
			Amount:  totalValue - sentValue,
		}}
	} else {
		// sometimes the fees from compound transactions make the total output higher than what's available from selected
		// utxos, in such cases, the remaining fee will be deduced from the resulting amount
		payments = []*libkaspawallet.Payment{{
			Address: targetAddress,
			Amount:  totalValue,
		}}
	}
	mergeTransactionBytes, err := libkaspawallet.CreateUnsignedTransaction(s.keysFile.ExtendedPublicKeys,
		s.keysFile.MinimumSignatures, payments, utxos)
	if err != nil {
		return nil, err
	}

	return serialization.DeserializePartiallySignedTransaction(mergeTransactionBytes)
}

func (s *server) maybeSplitTransaction(transaction *serialization.PartiallySignedTransaction,
	splitAddress util.Address) ([]*serialization.PartiallySignedTransaction, error) {

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

	splitTransactions := make([]*serialization.PartiallySignedTransaction, splitCount)
	for i := 0; i < splitCount; i++ {
		startIndex := i * inputCountPerSplit
		endIndex := startIndex + inputCountPerSplit
		var err error
		splitTransactions[i], err = s.createSplitTransaction(transaction, splitAddress, startIndex, endIndex)
		if err != nil {
			return nil, err
		}
	}

	return splitTransactions, nil
}

func (s *server) createSplitTransaction(transaction *serialization.PartiallySignedTransaction,
	changeAddress util.Address, startIndex int, endIndex int) (*serialization.PartiallySignedTransaction, error) {

	selectedUTXOs := make([]*libkaspawallet.UTXO, endIndex-startIndex)
	totalSompi := uint64(0)

	for i := startIndex; i < endIndex; i++ {
		partiallySignedInput := transaction.PartiallySignedInputs[i]
		selectedUTXOs[i-startIndex] = &libkaspawallet.UTXO{
			Outpoint: &transaction.Tx.Inputs[i].PreviousOutpoint,
			UTXOEntry: utxo.NewUTXOEntry(
				partiallySignedInput.PrevOutput.Value, partiallySignedInput.PrevOutput.ScriptPublicKey,
				false, constants.UnacceptedDAAScore),
			DerivationPath: partiallySignedInput.DerivationPath,
		}

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
