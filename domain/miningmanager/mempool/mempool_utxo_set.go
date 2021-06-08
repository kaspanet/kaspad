package mempool

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
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

func (mpus *mempoolUTXOSet) addTransaction(transaction *model.MempoolTransaction) error {
	panic("mempoolUTXOSet.addTransaction not implemented") // TODO (Mike)
}

func (mpus *mempoolUTXOSet) removeTransaction(transactionID *externalapi.DomainTransactionID) error {
	panic("mempoolUTXOSet.removeTransaction not implemented") // TODO (Mike)
}

func (mpus *mempoolUTXOSet) checkDoubleSpends(transaction *model.MempoolTransaction) error {
	panic("mempoolUTXOSet.checkDoubleSpends not implemented") // TODO (Mike)
}

func (mpus *mempoolUTXOSet) getOutpoint(outpoint *externalapi.DomainOutpoint) (externalapi.UTXOEntry, bool) {
	utxo, ok := mpus.poolUnspentOutputs[*outpoint]
	return utxo, ok
}
