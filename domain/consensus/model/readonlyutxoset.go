package model

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// ReadOnlyUTXOSet represents a UTXOSet that can only be read from
type ReadOnlyUTXOSet interface {
	Iterator() ReadOnlyUTXOSetIterator
	Entry(outpoint *externalapi.DomainOutpoint) externalapi.UTXOEntry
}

// ReadOnlyUTXOSetIterator is an iterator over all entries in a
// ReadOnlyUTXOSet
type ReadOnlyUTXOSetIterator interface {
	First() bool
	Next() bool
	Get() (outpoint *externalapi.DomainOutpoint, utxoEntry externalapi.UTXOEntry, err error)
}
