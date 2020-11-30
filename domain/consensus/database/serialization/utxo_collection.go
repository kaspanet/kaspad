package serialization

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/utxo"
)

func utxoCollectionToDBUTXOCollection(utxoCollection model.UTXOCollection) ([]*DbUtxoCollectionItem, error) {
	items := make([]*DbUtxoCollectionItem, utxoCollection.Len())
	i := 0
	utxoIterator := utxoCollection.Iterator()
	for utxoIterator.Next() {
		outpoint, entry, err := utxoIterator.Get()
		if err != nil {
			return nil, err
		}

		outpointCopy := outpoint
		items[i] = &DbUtxoCollectionItem{
			Outpoint:  DomainOutpointToDbOutpoint(outpointCopy),
			UtxoEntry: UTXOEntryToDBUTXOEntry(entry),
		}
		i++
	}

	return items, nil
}

func dbUTXOCollectionToUTXOCollection(items []*DbUtxoCollectionItem) (model.UTXOCollection, error) {
	utxoMap := make(map[externalapi.DomainOutpoint]*externalapi.UTXOEntry, len(items))
	for _, item := range items {
		outpoint, err := DbOutpointToDomainOutpoint(item.Outpoint)
		if err != nil {
			return nil, err
		}

		utxoMap[*outpoint] = DBUTXOEntryToUTXOEntry(item.UtxoEntry)
	}
	return utxo.NewUTXOCollection(utxoMap), nil
}
