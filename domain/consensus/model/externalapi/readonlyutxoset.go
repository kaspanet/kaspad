package externalapi

// ReadOnlyUTXOSetIterator is an iterator over all entries in a
// ReadOnlyUTXOSet
type ReadOnlyUTXOSetIterator interface {
	First() bool
	Next() bool
	Get() (outpoint *DomainOutpoint, utxoEntry UTXOEntry, err error)
	Close() error
}
