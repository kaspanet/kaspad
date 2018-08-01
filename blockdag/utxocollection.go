package blockdag

import (
	"fmt"
	"strings"
	"github.com/daglabs/btcd/wire"
	"sort"
)

// utxoCollection represents a set of UTXOs indexed by their outPoints
type utxoCollection map[wire.OutPoint]*UtxoEntry

func (uc utxoCollection) String() string {
	utxoStrings := make([]string, len(uc))

	i := 0
	for utxo := range uc.iterate() {
		utxoStrings[i] = fmt.Sprintf("(%s, %d) => %d", utxo.outPoint.Hash, utxo.outPoint.Index, utxo.entry.amount)
		i++
	}

	// Sort strings for determinism.
	sort.Strings(utxoStrings)

	return fmt.Sprintf("[ %s ]", strings.Join(utxoStrings, ", "))
}

// clone returns a clone of this collection
func (uc utxoCollection) clone() utxoCollection {
	clone := utxoCollection{}
	for outPoint, entry := range uc {
		clone[outPoint] = entry
	}

	return clone
}
