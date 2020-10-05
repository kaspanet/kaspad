package model

import "github.com/kaspanet/kaspad/app/appmessage"

// UTXODiff represents a diff between two UTXO Sets.
type UTXODiff struct {
	ToAdd    utxoCollection
	ToRemove utxoCollection
}

// utxoCollection represents a set of UTXOs indexed by their outpoints
type utxoCollection map[appmessage.Outpoint]*UTXOEntry
