package mempool

import (
	"fmt"

	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/utxo"
	"github.com/kaspanet/kaspad/domain/miningmanager/mempool/model"
)

type mempoolUTXOSet struct {
	mempool                        *mempool
	poolUnspentOutputs             model.OutpointToUTXOEntry
	transactionsByPreviousOutpoint model.OutpointToTransaction
}

func newMempoolUTXOSet(mp *mempool) *mempoolUTXOSet {
	return &mempoolUTXOSet{
		mempool:                        mp,
		poolUnspentOutputs:             model.OutpointToUTXOEntry{},
		transactionsByPreviousOutpoint: model.OutpointToTransaction{},
	}
}

func (mpus *mempoolUTXOSet) addTransaction(transaction *model.MempoolTransaction) {
	outpoint := &externalapi.DomainOutpoint{TransactionID: *transaction.TransactionID()}

	for i, input := range transaction.Transaction().Inputs {
		outpoint.Index = uint32(i)

		delete(mpus.poolUnspentOutputs, *outpoint)
		mpus.transactionsByPreviousOutpoint[input.PreviousOutpoint] = transaction
	}

	for i, output := range transaction.Transaction().Outputs {
		outpoint := externalapi.DomainOutpoint{TransactionID: *transaction.TransactionID(), Index: uint32(i)}

		mpus.poolUnspentOutputs[outpoint] =
			utxo.NewUTXOEntry(output.Value, output.ScriptPublicKey, false, model.UnacceptedDAAScore)
	}
}

func (mpus *mempoolUTXOSet) removeTransaction(transaction *model.MempoolTransaction) {
	for _, input := range transaction.Transaction().Inputs {
		mpus.poolUnspentOutputs[input.PreviousOutpoint] = input.UTXOEntry
		delete(mpus.transactionsByPreviousOutpoint, input.PreviousOutpoint)
	}

	outpoint := externalapi.DomainOutpoint{TransactionID: *transaction.TransactionID()}
	for i := range transaction.Transaction().Outputs {
		outpoint.Index = uint32(i)

		delete(mpus.poolUnspentOutputs, outpoint)
	}
}

func (mpus *mempoolUTXOSet) checkDoubleSpends(transaction *model.MempoolTransaction) error {
	outpoint := externalapi.DomainOutpoint{TransactionID: *transaction.TransactionID()}

	for i, input := range transaction.Transaction().Inputs {
		outpoint.Index = uint32(i)

		if existingTransaction, exists := mpus.transactionsByPreviousOutpoint[input.PreviousOutpoint]; exists {
			str := fmt.Sprintf("output %s already spent by transaction %s in the memory pool",
				input.PreviousOutpoint, existingTransaction.TransactionID())
			return txRuleError(RejectDuplicate, str)
		}
	}

	return nil
}
