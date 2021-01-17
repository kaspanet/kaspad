package utxo

import (
	"fmt"
	"sort"
	"strings"

	"github.com/kaspanet/kaspad/domain/consensus/model"

	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// Collection is a map between outpoints and utxo entries.
type Collection map[externalapi.DomainOutpoint]externalapi.UTXOEntry

// NewUTXOCollection creates a UTXO-Collection from the given map from outpoint to UTXOEntry
func NewUTXOCollection(utxoMap map[externalapi.DomainOutpoint]externalapi.UTXOEntry) model.UTXOCollection {
	return Collection(utxoMap)
}

// Get returns the model.UTXOEntry represented by provided outpoint,
// and a boolean value indicating if said model.UTXOEntry is in the set or not
func (uc Collection) Get(outpoint *externalapi.DomainOutpoint) (externalapi.UTXOEntry, bool) {
	entry, ok := uc[*outpoint]
	return entry, ok
}

// Contains returns a boolean value indicating whether a UTXO entry is in the set
func (uc Collection) Contains(outpoint *externalapi.DomainOutpoint) bool {
	_, ok := uc[*outpoint]
	return ok
}

// Len returns the amount of entries in the collection.
func (uc Collection) Len() int {
	return len(uc)
}

// Clone clones the collection to a new one.
func (uc Collection) Clone() Collection {
	if uc == nil {
		return nil
	}

	clone := make(Collection, len(uc))
	for outpoint, entry := range uc {
		clone[outpoint] = entry
	}

	return clone
}

func (uc Collection) String() string {
	utxoStrings := make([]string, len(uc))

	i := 0
	for outpoint, utxoEntry := range uc {
		utxoStrings[i] = fmt.Sprintf("(%s, %d) => %d, blueScore: %d",
			outpoint.TransactionID, outpoint.Index, utxoEntry.Amount(), utxoEntry.BlockBlueScore())
		i++
	}

	// Sort strings for determinism.
	sort.Strings(utxoStrings)

	return fmt.Sprintf("[ %s ]", strings.Join(utxoStrings, ", "))
}

// add adds a new UTXO entry to this collection
func (uc Collection) add(outpoint *externalapi.DomainOutpoint, entry externalapi.UTXOEntry) {
	uc[*outpoint] = entry
}

// addMultiple adds multiple UTXO entries to this collection
func (uc Collection) addMultiple(collectionToAdd Collection) {
	for outpoint, entry := range collectionToAdd {
		uc[outpoint] = entry
	}
}

// remove removes a UTXO entry from this collection if it exists
func (uc Collection) remove(outpoint *externalapi.DomainOutpoint) {
	delete(uc, *outpoint)
}

// removeMultiple removes multiple UTXO entries from this collection if it exists
func (uc Collection) removeMultiple(collectionToRemove Collection) {
	for outpoint := range collectionToRemove {
		delete(uc, outpoint)
	}
}

// containsWithBlueScore returns a boolean value indicating whether a model.UTXOEntry
// is in the set and its blue score is equal to the given blue score.
func (uc Collection) containsWithBlueScore(outpoint *externalapi.DomainOutpoint, blueScore uint64) bool {
	entry, ok := uc.Get(outpoint)
	return ok && entry.BlockBlueScore() == blueScore
}
