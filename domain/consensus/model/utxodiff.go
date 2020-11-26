package model

import (
	"fmt"
	"sort"
	"strings"

	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// UTXOCollection represents a set of UTXOs indexed by their outpoints
type UTXOCollection map[externalapi.DomainOutpoint]*externalapi.UTXOEntry

// If this doesn't compile, it means the type definition has been changed, so it's
// an indication to update Equal and Clone accordingly.
var _ UTXOCollection = map[externalapi.DomainOutpoint]*externalapi.UTXOEntry{}

// Equal returns whether uc equals to other
func (uc UTXOCollection) Equal(other UTXOCollection) bool {
	if len(uc) != len(other) {
		return false
	}

	for outpoint, entry := range uc {
		otherEntry, exists := other[outpoint]
		if !exists {
			return false
		}

		if !entry.Equal(otherEntry) {
			return false
		}
	}

	return true
}

// Clone returns a clone of UTXOCollection
func (uc UTXOCollection) Clone() UTXOCollection {
	clone := make(UTXOCollection, len(uc))
	for outpoint, entry := range uc {
		clone[outpoint] = entry.Clone()
	}

	return clone
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

// UTXODiff represents a diff between two UTXO Sets.
type UTXODiff struct {
	ToAdd    UTXOCollection
	ToRemove UTXOCollection
}

// Clone returns a clone of UTXODiff
func (d *UTXODiff) Clone() *UTXODiff {
	return &UTXODiff{
		ToAdd:    d.ToAdd.Clone(),
		ToRemove: d.ToRemove.Clone(),
	}
}

func (d UTXODiff) String() string {
	return fmt.Sprintf("ToAdd: %s; ToRemove: %s", d.ToAdd, d.ToRemove)
}

// NewUTXODiff instantiates an empty UTXODiff
func NewUTXODiff() *UTXODiff {
	return &UTXODiff{
		ToAdd:    UTXOCollection{},
		ToRemove: UTXOCollection{},
	}
}
