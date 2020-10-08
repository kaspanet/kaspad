package model

// UTXODiff represents a diff between two UTXO Sets.
type UTXODiff struct {
	ToAdd    UTXOCollection
	ToRemove UTXOCollection
}

// UTXOCollection represents a set of UTXOs indexed by their outpoints
type UTXOCollection map[DomainOutpoint]*UTXOEntry
