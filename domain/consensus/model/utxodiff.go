package model

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// UTXOCollection represents a collection of UTXO entries, indexed by their outpoint
type UTXOCollection interface {
	Iterator() ReadOnlyUTXOSetIterator
	Get(outpoint *externalapi.DomainOutpoint) (externalapi.UTXOEntry, bool)
	Contains(outpoint *externalapi.DomainOutpoint) bool
	Len() int
}

// UTXODiff represents the diff between two UTXO sets
type UTXODiff interface {
	ToAdd() UTXOCollection
	ToRemove() UTXOCollection
	WithDiff(other UTXODiff) (UTXODiff, error)
	DiffFrom(other UTXODiff) (UTXODiff, error)
	Clone() UTXODiff
	CloneMutable() MutableUTXODiff
}

// MutableUTXODiff represents a UTXO-Diff that can be mutated
type MutableUTXODiff interface {
	ToUnmutable() UTXODiff

	WithDiff(other UTXODiff) (UTXODiff, error)
	DiffFrom(other UTXODiff) (UTXODiff, error)
	ToAdd() UTXOCollection
	ToRemove() UTXOCollection

	WithDiffInPlace(other UTXODiff) error
	AddTransaction(transaction *externalapi.DomainTransaction, blockBlueScore uint64) error
	Clone() MutableUTXODiff
}
