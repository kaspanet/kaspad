package model

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

type UTXOCollection interface {
	Iterator() ReadOnlyUTXOSetIterator
	Get(outpoint *externalapi.DomainOutpoint) (*externalapi.UTXOEntry, bool)
	Contains(outpoint *externalapi.DomainOutpoint) bool
}

type UTXODiff interface {
	WithDiff(other UTXODiff) (UTXODiff, error)
	DiffFrom(other UTXODiff) (UTXODiff, error)
	ToAdd() UTXOCollection
	ToRemove() UTXOCollection
	Clone() UTXODiff
	CloneMutable() MutableUTXODiff
}

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
