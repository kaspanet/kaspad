package utxoindex

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

type UTXOIndex struct {
	consensus externalapi.Consensus
	store     *utxoIndexStore
}

func New(consensus externalapi.Consensus) *UTXOIndex {
	store := newUTXOIndexStore()
	return &UTXOIndex{
		consensus: consensus,
		store:     store,
	}
}

func (ui *UTXOIndex) Update(chainChanges *externalapi.SelectedParentChainChanges) error {
	return nil
}
