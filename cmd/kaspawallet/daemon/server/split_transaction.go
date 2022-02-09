package server

import (
	"github.com/kaspanet/go-secp256k1"
	"github.com/kaspanet/kaspad/cmd/kaspawallet/libkaspawallet"
	"github.com/kaspanet/kaspad/cmd/kaspawallet/libkaspawallet/serialization"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/miningmanager/mempool"
	"github.com/kaspanet/kaspad/util"
)

func (s *server) maybeSplitTransaction(partiallySignedTransactionBytes []byte) ([][]byte, error) {
	partiallySignedTransaction, err := serialization.DeserializePartiallySignedTransaction(partiallySignedTransactionBytes)
	if err != nil {
		return nil, err
	}

	partiallySignedTransactions, err := s.maybeSplitTransactionInner(partiallySignedTransaction)
	if err != nil {
		return nil, err
	}
	if len(partiallySignedTransactions) > 1 {
		partiallySignedTransactions = append(partiallySignedTransactions, mergeTransaction(partiallySignedTransactions))
	}

	partiallySignedTransactionsBytes := make([][]byte, len(partiallySignedTransactions))
	for i, partiallySignedTransaction := range partiallySignedTransactions {
		partiallySignedTransactionsBytes[i], err = serialization.SerializePartiallySignedTransaction(partiallySignedTransaction)
		if err != nil {
			return nil, err
		}
	}
	return partiallySignedTransactionsBytes, nil
}

func mergeTransaction(transactions []*serialization.PartiallySignedTransaction) *serialization.PartiallySignedTransaction {
	// TODO
}

func (s *server) maybeSplitTransactionInner(transaction *serialization.PartiallySignedTransaction) (
	[]*serialization.PartiallySignedTransaction, error) {

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

	changeAddress, err := s.changeAddress()
	if err != nil {
		return nil, err
	}

	splitTransactions := make([]*serialization.PartiallySignedTransaction, splitCount)
	for i := 0; i < splitCount; i++ {
		startIndex := i * inputCountPerSplit
		endIndex := startIndex + inputCountPerSplit
		splitTransactions[i], err = s.createSplitTransaction(transaction, changeAddress, startIndex, endIndex)
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
