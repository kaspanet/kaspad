package utxo

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

type utxoOutpointEntryPair struct {
	outpoint externalapi.DomainOutpoint
	entry    *externalapi.UTXOEntry
}

type utxoCollectionIterator struct {
	index int
	pairs []utxoOutpointEntryPair
}

func CollectionIterator(collection model.UTXOCollection) model.ReadOnlyUTXOSetIterator {
	pairs := make([]utxoOutpointEntryPair, len(collection))
	i := 0
	for outpoint, entry := range collection {
		pairs[i] = utxoOutpointEntryPair{
			outpoint: outpoint,
			entry:    entry,
		}
		i++
	}
	return &utxoCollectionIterator{index: -1, pairs: pairs}
}

func (u *utxoCollectionIterator) Next() bool {
	u.index++
	return u.index < len(u.pairs)
}

func (u *utxoCollectionIterator) Get() (outpoint *externalapi.DomainOutpoint, utxoEntry *externalapi.UTXOEntry, err error) {
	pair := u.pairs[u.index]
	return &pair.outpoint, pair.entry, nil
}
