package utxoalgebra

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// add adds a new UTXO entry to this collection
func collectionAdd(collection model.UTXOCollection, outpoint *externalapi.DomainOutpoint, entry *externalapi.UTXOEntry) {
	collection[*outpoint] = entry
}

// addMultiple adds multiple UTXO entries to this collection
func collectionAddMultiple(collection model.UTXOCollection, collectionToAdd model.UTXOCollection) {
	for outpoint, entry := range collectionToAdd {
		collection[outpoint] = entry
	}
}

// remove removes a UTXO entry from this collection if it exists
func collectionRemove(collection model.UTXOCollection, outpoint *externalapi.DomainOutpoint) {
	delete(collection, *outpoint)
}

// removeMultiple removes multiple UTXO entries from this collection if it exists
func collectionRemoveMultiple(collection model.UTXOCollection, collectionToRemove model.UTXOCollection) {
	for outpoint := range collectionToRemove {
		delete(collection, outpoint)
	}
}

// CollectionGet returns the model.UTXOEntry represented by provided outpoint,
// and a boolean value indicating if said model.UTXOEntry is in the set or not
func CollectionGet(collection model.UTXOCollection, outpoint *externalapi.DomainOutpoint) (*externalapi.UTXOEntry, bool) {
	entry, ok := collection[*outpoint]
	return entry, ok
}

// CollectionContains returns a boolean value indicating whether a UTXO entry is in the set
func CollectionContains(collection model.UTXOCollection, outpoint *externalapi.DomainOutpoint) bool {
	_, ok := collection[*outpoint]
	return ok
}

// CollectionContainsWithBlueScore returns a boolean value indicating whether a model.UTXOEntry
// is in the set and its blue score is equal to the given blue score.
func CollectionContainsWithBlueScore(collection model.UTXOCollection, outpoint *externalapi.DomainOutpoint, blueScore uint64) bool {
	entry, ok := CollectionGet(collection, outpoint)
	return ok && entry.BlockBlueScore == blueScore
}
