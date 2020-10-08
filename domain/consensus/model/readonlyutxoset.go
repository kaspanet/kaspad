package model

// ReadOnlyUTXOSet represents a UTXOSet that can only be read from
type ReadOnlyUTXOSet interface {
	Iterator() ReadOnlyUTXOSetIterator
	Entry(outpoint *DomainOutpoint) *UTXOEntry
}

// ReadOnlyUTXOSetIterator is an iterator over all entries in a
// ReadOnlyUTXOSet
type ReadOnlyUTXOSetIterator interface {
	Next() bool
	Get() (outpoint *DomainOutpoint, utxoEntry *UTXOEntry)
}
