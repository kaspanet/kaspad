package utxoindex

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

type UTXOIndex struct {
	store *utxoIndexStore
}

func New() *UTXOIndex {
	store := newUTXOIndexStore()
	return &UTXOIndex{
		store: store,
	}
}

func (ui *UTXOIndex) Update(chainChanges *externalapi.SelectedParentChainChanges) error {
	return nil
}
