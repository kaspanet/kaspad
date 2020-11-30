package utxo

import (
	"fmt"
	"sort"
	"strings"

	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

type utxoCollection map[externalapi.DomainOutpoint]*externalapi.UTXOEntry

// Get returns the model.UTXOEntry represented by provided outpoint,
// and a boolean value indicating if said model.UTXOEntry is in the set or not
func (uc utxoCollection) Get(outpoint *externalapi.DomainOutpoint) (*externalapi.UTXOEntry, bool) {
	entry, ok := uc[*outpoint]
	return entry, ok
}

// Contains returns a boolean value indicating whether a UTXO entry is in the set
func (uc utxoCollection) Contains(outpoint *externalapi.DomainOutpoint) bool {
	_, ok := uc[*outpoint]
	return ok
}

func (uc utxoCollection) Clone() utxoCollection {
	if uc == nil {
		return nil
	}

	clone := make(utxoCollection, len(uc))
	for outpoint, entry := range uc {
		clone[outpoint] = entry.Clone()
	}

	return clone
}

func (uc utxoCollection) String() string {
	utxoStrings := make([]string, len(uc))

	i := 0
	for outpoint, utxoEntry := range uc {
		utxoStrings[i] = fmt.Sprintf("(%s, %d) => %d, blueScore: %d",
			outpoint.TransactionID, outpoint.Index, utxoEntry.Amount, utxoEntry.BlockBlueScore)
		i++
	}

	// Sort strings for determinism.
	sort.Strings(utxoStrings)

	return fmt.Sprintf("[ %s ]", strings.Join(utxoStrings, ", "))
}

// add adds a new UTXO entry to this collection
func (uc utxoCollection) add(outpoint *externalapi.DomainOutpoint, entry *externalapi.UTXOEntry) {
	uc[*outpoint] = entry
}

// addMultiple adds multiple UTXO entries to this collection
func (uc utxoCollection) addMultiple(collectionToAdd utxoCollection) {
	for outpoint, entry := range collectionToAdd {
		uc[outpoint] = entry
	}
}

// remove removes a UTXO entry from this collection if it exists
func (uc utxoCollection) remove(outpoint *externalapi.DomainOutpoint) {
	delete(uc, *outpoint)
}

// removeMultiple removes multiple UTXO entries from this collection if it exists
func (uc utxoCollection) removeMultiple(collectionToRemove utxoCollection) {
	for outpoint := range collectionToRemove {
		delete(uc, outpoint)
	}
}

// containsWithBlueScore returns a boolean value indicating whether a model.UTXOEntry
// is in the set and its blue score is equal to the given blue score.
func (uc utxoCollection) containsWithBlueScore(outpoint *externalapi.DomainOutpoint, blueScore uint64) bool {
	entry, ok := uc.Get(outpoint)
	return ok && entry.BlockBlueScore == blueScore
}
