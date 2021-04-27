package externalapi

// UTXOCollection represents a collection of UTXO entries, indexed by their outpoint
type UTXOCollection interface {
	Iterator() ReadOnlyUTXOSetIterator
	Get(outpoint *DomainOutpoint) (UTXOEntry, bool)
	Contains(outpoint *DomainOutpoint) bool
	Len() int
}

// UTXODiff represents the diff between two UTXO sets
type UTXODiff interface {
	ToAdd() UTXOCollection
	ToRemove() UTXOCollection
	WithDiff(other UTXODiff) (UTXODiff, error)
	DiffFrom(other UTXODiff) (UTXODiff, error)
	Reversed() UTXODiff
	CloneMutable() MutableUTXODiff
}

// MutableUTXODiff represents a UTXO-Diff that can be mutated
type MutableUTXODiff interface {
	ToImmutable() UTXODiff

	WithDiff(other UTXODiff) (UTXODiff, error)
	DiffFrom(other UTXODiff) (UTXODiff, error)
	ToAdd() UTXOCollection
	ToRemove() UTXOCollection

	WithDiffInPlace(other UTXODiff) error
	AddTransaction(transaction *DomainTransaction, blockDAAScore uint64) error
}
