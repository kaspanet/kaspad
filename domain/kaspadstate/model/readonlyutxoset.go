package model

import "github.com/kaspanet/kaspad/app/appmessage"

// ReadOnlyUTXOSet ...
type ReadOnlyUTXOSet interface {
	Iterator() ReadOnlyUTXOSetIterator
	Entry(outpoint *appmessage.Outpoint) *UTXOEntry
}

// ReadOnlyUTXOSetIterator ...
type ReadOnlyUTXOSetIterator interface {
	Next() bool
	Get() (outpoint *appmessage.Outpoint, utxoEntry *UTXOEntry)
}
