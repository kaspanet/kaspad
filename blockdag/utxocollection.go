package blockdag

import (
	"fmt"
	"strings"
	"github.com/daglabs/btcd/dagconfig/daghash"
	"github.com/daglabs/btcd/wire"
	"sort"
)

// utxoCollection represents a set of UTXOs indexed by their previousHash and index
type utxoCollection map[daghash.Hash]map[int]*wire.TxOut

func (uc utxoCollection) String() string {
	utxoStrings := make([]string, uc.len())

	i := 0
	for utxo := range uc.Iterate() {
		utxoStrings[i] = fmt.Sprintf("(%s, %d) => %d", utxo.previousHash, utxo.index, utxo.txOut.Value)
		i++
	}

	// Sort strings for determinism.
	sort.Strings(utxoStrings)

	return fmt.Sprintf("[ %s ]", strings.Join(utxoStrings, ", "))
}

// len returns the the number of UTXOs in this utxoCollection
func (uc utxoCollection) len() int {
	counter := 0
	for _, txOuts := range uc {
		counter += len(txOuts)
	}

	return counter
}

// get returns the txOut represented by provided hash and index,
// and a boolean value indicating if said txOut is in the set or not
func (uc utxoCollection) get(previousHash daghash.Hash, index int) (*wire.TxOut, bool) {
	previous, ok := uc[previousHash]
	if !ok {
		return nil, false
	}
	txOut, ok := previous[index]
	return txOut, ok
}

// contains returns a boolean value indicating if represented by provided hash and index is in the set or not
func (uc utxoCollection) contains(previousHash daghash.Hash, index int) bool {
	previous, ok := uc[previousHash]
	if !ok {
		return false
	}
	_, ok = previous[index]
	return ok
}

// add adds a new UTXO to this collection
func (uc utxoCollection) add(previousHash daghash.Hash, index int, txOut *wire.TxOut) {
	_, ok := uc[previousHash]
	if !ok {
		uc[previousHash] = map[int]*wire.TxOut{}
	}

	uc[previousHash][index] = txOut
}

// remove removes a UTXO from this collection if exists
func (uc utxoCollection) remove(previousHash daghash.Hash, index int) {
	previous, ok := uc[previousHash]
	if !ok {
		return
	}
	delete(previous, index)
	if len(previous) == 0 {
		delete(uc, previousHash)
	}
}

// clone returns a clone of this collection
func (uc utxoCollection) clone() utxoCollection {
	clone := utxoCollection{}

	for previousID, txOuts := range uc {
		clone[previousID] = map[int]*wire.TxOut{}
		for index, txOut := range txOuts {
			clone[previousID][index] = txOut
		}
	}

	return clone
}
