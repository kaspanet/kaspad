package model

import (
	"fmt"
	"sort"
	"strings"

	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// UTXOCollection represents a set of UTXOs indexed by their outpoints
type UTXOCollection map[externalapi.DomainOutpoint]*externalapi.UTXOEntry

// UTXODiff represents a diff between two UTXO Sets.
type UTXODiff struct {
	ToAdd    UTXOCollection
	ToRemove UTXOCollection
}

// NewUTXODiff instantiates an empty UTXODiff
func NewUTXODiff() *UTXODiff {
	return &UTXODiff{
		ToAdd:    UTXOCollection{},
		ToRemove: UTXOCollection{},
	}
}

func (uc UTXOCollection) String() string {
	utxoStrings := make([]string, len(uc))

	i := 0
	for outpoint, utxoEntry := range uc {
		utxoStrings[i] = fmt.Sprintf("(%s, %d) => %d, blueScore: %d",
			outpoint.TransactionID, outpoint.Index, utxoEntry.Amount, utxoEntry.BlockBlueScore)
		i++
	}

	// Sort strings for determinism.
	sort.Strings(utxoStrings)

	return fmt.Sprintf("[ %s ]", strings.Join(utxoStrings, ", "))
}

func (d UTXODiff) String() string {
	return fmt.Sprintf("ToAdd: %s; ToRemove: %s", d.ToAdd, d.ToRemove)
}

func (d *UTXODiff) Clone() *UTXODiff {
	clone := &UTXODiff{
		ToAdd:    make(UTXOCollection, len(d.ToAdd)),
		ToRemove: make(UTXOCollection, len(d.ToRemove)),
	}

	for outpoint, entry := range d.ToAdd {
		clone.ToAdd[outpoint] = cloneUTXOEntry(entry)
	}

	for outpoint, entry := range d.ToRemove {
		clone.ToRemove[outpoint] = cloneUTXOEntry(entry)
	}

	return clone
}

func cloneUTXOEntry(entry *externalapi.UTXOEntry) *externalapi.UTXOEntry {
	scriptPublicKeyClone := make([]byte, len(entry.ScriptPublicKey))
	copy(scriptPublicKeyClone, entry.ScriptPublicKey)

	return &externalapi.UTXOEntry{
		Amount:          entry.Amount,
		ScriptPublicKey: scriptPublicKeyClone,
		BlockBlueScore:  entry.BlockBlueScore,
		IsCoinbase:      entry.IsCoinbase,
	}
}
