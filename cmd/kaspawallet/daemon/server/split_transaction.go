package server

import (
	"github.com/kaspanet/go-secp256k1"
	"github.com/kaspanet/kaspad/cmd/kaspawallet/libkaspawallet/serialization"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

func (s *server) maybeSplitTransaction(partiallySignedTransactionBytes []byte) ([][]byte, error) {
	partiallySignedTransaction, err := serialization.DeserializePartiallySignedTransaction(partiallySignedTransactionBytes)
	if err != nil {
		return nil, err
	}

	partiallySignedTransactions := s.maybeSplitTransactionInner(partiallySignedTransaction)
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

func (s *server) maybeSplitTransactionInner(partiallySignedTransaction *serialization.PartiallySignedTransaction) []*serialization.PartiallySignedTransaction {
	transactionMass := s.txMassCalculator.CalculateTransactionMass(partiallySignedTransaction.Tx)
	transactionMass += s.estimateMassIncreaseForSignatures(partiallySignedTransaction.Tx)
}

func (s *server) estimateMassIncreaseForSignatures(transaction *externalapi.DomainTransaction) uint64 {
	var signatureSize uint64
	if s.keysFile.ECDSA {
		signatureSize = secp256k1.SerializedECDSASignatureSize
	} else {
		signatureSize = secp256k1.SerializedSchnorrSignatureSize
	}

	return uint64(s.keysFile.MinimumSignatures) * signatureSize * s.txMassCalculator.MassPerTxByte()
}
