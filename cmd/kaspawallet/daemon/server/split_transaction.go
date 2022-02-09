package server

import (
	"github.com/kaspanet/go-secp256k1"
	"github.com/kaspanet/kaspad/cmd/kaspawallet/libkaspawallet"
	"github.com/kaspanet/kaspad/cmd/kaspawallet/libkaspawallet/serialization"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
	"github.com/kaspanet/kaspad/domain/consensus/utils/constants"
	"github.com/kaspanet/kaspad/domain/consensus/utils/utxo"
	"github.com/kaspanet/kaspad/domain/miningmanager/mempool"
	"github.com/kaspanet/kaspad/util"
)

func (s *server) maybeSplitTransaction(transactionBytes []byte) ([][]byte, error) {
	transaction, err := serialization.DeserializePartiallySignedTransaction(transactionBytes)
	if err != nil {
		return nil, err
	}

	splitAddress, splitWalletAddress, err := s.changeAddress()
	if err != nil {
		return nil, err
	}

	splitTransactions, err := s.maybeSplitTransactionInner(transaction, splitAddress)
	if err != nil {
		return nil, err
	}
	if len(splitTransactions) > 1 {
		mergeTransaction, err := s.mergeTransaction(splitTransactions, transaction, splitAddress, splitWalletAddress)
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
	originalTransaction *serialization.PartiallySignedTransaction, splitAddress util.Address,
	splitWalletAddress *walletAddress) (*serialization.PartiallySignedTransaction, error) {

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

	mergeTransactionBytes, err := libkaspawallet.CreateUnsignedTransaction(s.keysFile.ExtendedPublicKeys,
		s.keysFile.MinimumSignatures,
		[]*libkaspawallet.Payment{{
			Address: targetAddress,
			Amount:  sentValue,
		}, {
			Address: changeAddress,
			Amount:  totalValue - sentValue,
		}}, utxos)

	return serialization.DeserializePartiallySignedTransaction(mergeTransactionBytes)
}

func (s *server) maybeSplitTransactionInner(transaction *serialization.PartiallySignedTransaction,
	splitAddress util.Address) ([]*serialization.PartiallySignedTransaction, error) {

	transactionMass := s.txMassCalculator.CalculateTransactionMass(transaction.Tx)
	transactionMass += s.estimateMassIncreaseForSignatures(transaction.Tx)

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
		selectedUTXOs[i-startIndex] = &libkaspawallet.UTXO{
			Outpoint:       &transaction.Tx.Inputs[i].PreviousOutpoint,
			UTXOEntry:      transaction.Tx.Inputs[i].UTXOEntry,
			DerivationPath: transaction.PartiallySignedInputs[i].DerivationPath,
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

func (s *server) estimateMassIncreaseForSignatures(transaction *externalapi.DomainTransaction) uint64 {
	var signatureSize uint64
	if s.keysFile.ECDSA {
		signatureSize = secp256k1.SerializedECDSASignatureSize
	} else {
		signatureSize = secp256k1.SerializedSchnorrSignatureSize
	}

	return uint64(len(transaction.Inputs)) *
		uint64(s.keysFile.MinimumSignatures) *
		signatureSize *
		s.txMassCalculator.MassPerTxByte()
	// TODO: Add increase per sigop after https://github.com/kaspanet/kaspad/issues/1874 is handled
}
