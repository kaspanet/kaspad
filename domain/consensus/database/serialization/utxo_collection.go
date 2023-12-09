package serialization

import (
	"github.com/zoomy-network/zoomyd/domain/consensus/model/externalapi"
	"github.com/zoomy-network/zoomyd/domain/consensus/utils/utxo"
)

func utxoCollectionToDBUTXOCollection(utxoCollection externalapi.UTXOCollection) ([]*DbUtxoCollectionItem, error) {
	items := make([]*DbUtxoCollectionItem, utxoCollection.Len())
	i := 0
	utxoIterator := utxoCollection.Iterator()
	defer utxoIterator.Close()
	for ok := utxoIterator.First(); ok; ok = utxoIterator.Next() {
		outpoint, entry, err := utxoIterator.Get()
		if err != nil {
			return nil, err
		}

		items[i] = &DbUtxoCollectionItem{
			Outpoint:  DomainOutpointToDbOutpoint(outpoint),
			UtxoEntry: UTXOEntryToDBUTXOEntry(entry),
		}
		i++
	}

	return items, nil
}

func dbUTXOCollectionToUTXOCollection(items []*DbUtxoCollectionItem) (externalapi.UTXOCollection, error) {
	utxoMap := make(map[externalapi.DomainOutpoint]externalapi.UTXOEntry, len(items))
	for _, item := range items {
		outpoint, err := DbOutpointToDomainOutpoint(item.Outpoint)
		if err != nil {
			return nil, err
		}
		utxoEntry, err := DBUTXOEntryToUTXOEntry(item.UtxoEntry)
		if err != nil {
			return nil, err
		}
		utxoMap[*outpoint] = utxoEntry
	}
	return utxo.NewUTXOCollection(utxoMap), nil
}
