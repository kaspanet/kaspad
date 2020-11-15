package addressindex

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// UTXOCollection represents a set of UTXOs outpoints
type UTXOCollection map[externalapi.DomainOutpoint]*externalapi.UTXOEntry

// Add adds a new UTXO to this collection
func (uc UTXOCollection) Add(outpoint *externalapi.DomainOutpoint, entry *externalapi.UTXOEntry) {
	uc[*outpoint] = entry
}

// Remove removes a UTXO from this collection if it exists
func (uc UTXOCollection) Remove(outpoint *externalapi.DomainOutpoint) {
	delete(uc, *outpoint)
}

// UTXOMap represents a set of UTXOs sets indexed by their addresses
type UTXOMap map[string]UTXOCollection

// Get returns the UTXOCollection represented by provided address,
// and a boolean value indicating if said address is in the set or not
func (um UTXOMap) Get(address string) (UTXOCollection, bool) {
	if collection, ok := um[address]; ok {
		return collection, true
	}

	return nil, false
}

// Add adds a new UTXO entry associated with an address to this UTXOMap
func (um UTXOMap) Add(address string, outpoint *externalapi.DomainOutpoint, entry *externalapi.UTXOEntry) {
	if collection, ok := um[address]; ok {
		collection[*outpoint] = entry
	}
}

// Remove removes an outpoint associated with an address from this UTXOMap if it exists
func (um UTXOMap) Remove(address string, outpoint *externalapi.DomainOutpoint) {
	if collection, ok := um[address]; ok {
		delete(collection, *outpoint)
	}
}

// Addresses returns all addresses
func (um UTXOMap) Addresses() []string {
	addresses := make([]string, 0, len(um))
	for key := range um {
		addresses = append(addresses, key)
	}
	return addresses
}
