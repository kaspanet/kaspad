package mempool

import (
	"fmt"

	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"

	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/utxo"
	"github.com/kaspanet/kaspad/domain/miningmanager/mempool/model"
)

type mempoolUTXOSet struct {
	mempool                       *mempool
	poolUnspentOutputs            model.OutpointsToUTXOEntries
	transactionByPreviousOutpoint model.OutpointsToTransactions
}

func newMempoolUTXOSet(mp *mempool) *mempoolUTXOSet {
	return &mempoolUTXOSet{
		mempool:                       mp,
		poolUnspentOutputs:            model.OutpointsToUTXOEntries{},
		transactionByPreviousOutpoint: model.OutpointsToTransactions{},
	}
}

// this function MUST be called with the mempool mutex locked for writes
func (mpus *mempoolUTXOSet) addTransaction(transaction *model.MempoolTransaction) {
	outpoint := &externalapi.DomainOutpoint{TransactionID: *transaction.TransactionID()}

	for i, input := range transaction.Transaction().Inputs {
		outpoint.Index = uint32(i)

		delete(mpus.poolUnspentOutputs, *outpoint)
		mpus.transactionByPreviousOutpoint[input.PreviousOutpoint] = transaction
	}

	for i, output := range transaction.Transaction().Outputs {
		outpoint := externalapi.DomainOutpoint{TransactionID: *transaction.TransactionID(), Index: uint32(i)}

		mpus.poolUnspentOutputs[outpoint] =
			utxo.NewUTXOEntry(output.Value, output.ScriptPublicKey, false, model.UnacceptedDAAScore)
	}
}

// this function MUST be called with the mempool mutex locked for writes
func (mpus *mempoolUTXOSet) removeTransaction(transaction *model.MempoolTransaction) {
	for _, input := range transaction.Transaction().Inputs {
		// If the transaction creating the output spent by this input is in the mempool - restore it's UTXO
		if _, ok := mpus.mempool.transactionsPool.getTransaction(&input.PreviousOutpoint.TransactionID); ok {
			mpus.poolUnspentOutputs[input.PreviousOutpoint] = input.UTXOEntry
		}
		delete(mpus.transactionByPreviousOutpoint, input.PreviousOutpoint)
	}

	outpoint := externalapi.DomainOutpoint{TransactionID: *transaction.TransactionID()}
	for i := range transaction.Transaction().Outputs {
		outpoint.Index = uint32(i)

		delete(mpus.poolUnspentOutputs, outpoint)
	}
}

// this function MUST be called with the mempool mutex locked for reads
func (mpus *mempoolUTXOSet) checkDoubleSpends(transaction *externalapi.DomainTransaction) error {
	outpoint := externalapi.DomainOutpoint{TransactionID: *consensushashing.TransactionID(transaction)}

	for i, input := range transaction.Inputs {
		outpoint.Index = uint32(i)

		if existingTransaction, exists := mpus.transactionByPreviousOutpoint[input.PreviousOutpoint]; exists {
			str := fmt.Sprintf("output %s already spent by transaction %s in the memory pool",
				input.PreviousOutpoint, existingTransaction.TransactionID())
			return transactionRuleError(RejectDuplicate, str)
		}
	}

	return nil
}
