package mempool

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

type outpointToUTXOEntry map[externalapi.DomainOutpoint]externalapi.UTXOEntry

type mempoolUTXOSet struct {
	poolUnspentOutputs outpointToUTXOEntry
}

func (mpus *mempoolUTXOSet) getParentsInPool(transaction *mempoolTransaction) ([]*mempoolTransaction, error) {
	panic("mempoolUTXOSet.getParentsInPool not implemented") // TODO (Mike)
}

func (mpus *mempoolUTXOSet) addTransaction(transaction *mempoolTransaction) error {
	panic("mempoolUTXOSet.addTransaction not implemented") // TODO (Mike)
}
