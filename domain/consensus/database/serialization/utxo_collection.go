package serialization

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
)

func utxoCollectionToDBUTXOCollection(utxoCollection model.UTXOCollection) []*DbUtxoCollectionItem {
	items := make([]*DbUtxoCollectionItem, len(utxoCollection))
	i := 0
	for outpoint, entry := range utxoCollection {
		items[i] = &DbUtxoCollectionItem{
			Outpoint:  DomainOutpointToDbOutpoint(&outpoint),
			UtxoEntry: utxoEntryToDBUTXOEntry(entry),
		}
		i++
	}

	return items
}

func dbUTXOCollectionToUTXOCollection(items []*DbUtxoCollectionItem) (model.UTXOCollection, error) {
	collection := make(model.UTXOCollection)
	for _, item := range items {
		outpoint, err := DbOutpointToDomainOutpoint(item.Outpoint)
		if err != nil {
			return nil, err
		}

		collection[*outpoint] = dbUTXOEntryToUTXOEntry(item.UtxoEntry)
	}
	return collection, nil
}
