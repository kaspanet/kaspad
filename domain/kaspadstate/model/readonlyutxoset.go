package model

// ReadOnlyUTXOSet ...
type ReadOnlyUTXOSet interface {
	Iterator() ReadOnlyUTXOSetIterator
}

// ReadOnlyUTXOSetIterator ...
type ReadOnlyUTXOSetIterator interface {
	Next() bool
}
