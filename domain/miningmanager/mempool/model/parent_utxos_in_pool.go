package model

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// ParentUTXOsInPool represent the utxos a transaction spends out of the mempool.
// The utxos are indexed by transaction output index, for convenient access.
type ParentUTXOsInPool map[int]externalapi.UTXOEntry

func (pip ParentUTXOsInPool) Get(index int) (externalapi.UTXOEntry, bool) {
	utxoEntry, ok := pip[index]
	return utxoEntry, ok
}

func (pip ParentUTXOsInPool) Set(index int, utxoEntry externalapi.UTXOEntry) {
	pip[index] = utxoEntry
}
