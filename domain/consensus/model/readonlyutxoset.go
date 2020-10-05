package model

import "github.com/kaspanet/kaspad/app/appmessage"

// ReadOnlyUTXOSet represents a UTXOSet that can only be read from
type ReadOnlyUTXOSet interface {
	Iterator() ReadOnlyUTXOSetIterator
	Entry(outpoint *appmessage.Outpoint) *UTXOEntry
}

// ReadOnlyUTXOSetIterator is an iterator over all entries in a
// ReadOnlyUTXOSet
type ReadOnlyUTXOSetIterator interface {
	Next() bool
	Get() (outpoint *appmessage.Outpoint, utxoEntry *UTXOEntry)
}
