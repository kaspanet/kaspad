package model

import "github.com/kaspanet/kaspad/app/appmessage"

// UTXODiff represents a diff between two UTXO Sets.
type UTXODiff struct {
	ToAdd    UTXOCollection
	ToRemove UTXOCollection
}

// UTXOCollection represents a set of UTXOs indexed by their outpoints
type UTXOCollection map[appmessage.Outpoint]*UTXOEntry
