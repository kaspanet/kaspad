package mempool

import (
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

func (mpus *mempoolUTXOSet) getParentsInPool(transaction *model.MempoolTransaction) model.ParentUTXOsInPool {
	parentsInPool := model.ParentUTXOsInPool{}

	outpoint := &externalapi.DomainOutpoint{
		TransactionID: *transaction.TransactionID(),
	}
	for i := range transaction.Transaction.Inputs {
		outpoint.Index = uint32(i)
		utxo, ok := mpus.getOutpoint(outpoint)
		if ok {
			parentsInPool.Set(i, utxo)
		}
	}

	return parentsInPool
}

func (mpus *mempoolUTXOSet) addTransaction(transaction *model.MempoolTransaction) {
	outpoint := &externalapi.DomainOutpoint{TransactionID: *transaction.TransactionID()}

	for i, input := range transaction.Transaction.Inputs {
		outpoint.Index = uint32(i)

		delete(mpus.poolUnspentOutputs, *outpoint)
		mpus.transactionsByPreviousOutpoint[input.PreviousOutpoint] = transaction
	}

	for i, output := range transaction.Transaction.Outputs {
		outpoint := externalapi.DomainOutpoint{TransactionID: *transaction.TransactionID(), Index: uint32(i)}

		mpus.poolUnspentOutputs[outpoint] =
			utxo.NewUTXOEntry(output.Value, output.ScriptPublicKey, false, model.UnacceptedDAAScore)
	}
}

func (mpus *mempoolUTXOSet) removeTransaction(transaction *model.MempoolTransaction) {
	for _, input := range transaction.Transaction.Inputs {
		mpus.poolUnspentOutputs[input.PreviousOutpoint] = input.UTXOEntry
		delete(mpus.transactionsByPreviousOutpoint, input.PreviousOutpoint)
	}

	outpoint := &externalapi.DomainOutpoint{TransactionID: *transaction.TransactionID()}
	for i := range transaction.Transaction.Outputs {
		outpoint.Index = uint32(i)

		delete(mpus.poolUnspentOutputs, outpoint)
	}
}

func (mpus *mempoolUTXOSet) checkDoubleSpends(transaction *model.MempoolTransaction) error {
	panic("mempoolUTXOSet.checkDoubleSpends not implemented") // TODO (Mike)
}

func (mpus *mempoolUTXOSet) getOutpoint(outpoint *externalapi.DomainOutpoint) (externalapi.UTXOEntry, bool) {
	utxo, ok := mpus.poolUnspentOutputs[*outpoint]
	return utxo, ok
}
