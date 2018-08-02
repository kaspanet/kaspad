package blockdag

import (
	"fmt"
	"github.com/daglabs/btcd/wire"
	"errors"
)

// utxoIteratorOutput represents all fields of a single UTXO, to be returned by an iterator
type utxoIteratorOutput struct {
	outPoint wire.OutPoint
	entry    *UtxoEntry
}

// utxoIterator is used to iterate over a utxoSet
type utxoIterator <-chan utxoIteratorOutput

// newUTXOEntry creates a new utxoEntry representing the given txOut
func newUTXOEntry(txOut *wire.TxOut) *UtxoEntry {
	entry := new(UtxoEntry)
	entry.amount = txOut.Value
	entry.pkScript = txOut.PkScript

	return entry
}

// utxoSet represents a set of unspent transaction outputs
type utxoSet interface {
	fmt.Stringer
	diffFrom(other utxoSet) (*utxoDiff, error)
	withDiff(utxoDiff *utxoDiff) (utxoSet, error)
	addTx(tx *wire.MsgTx) (ok bool)
	iterate() utxoIterator
	clone() utxoSet
}

// fullUTXOSet represents a full list of transaction outputs and their values
type fullUTXOSet struct {
	utxoCollection
}

// newFullUTXOSet creates a new utxoSet with full list of transaction outputs and their values
func newFullUTXOSet() *fullUTXOSet {
	return &fullUTXOSet{
		utxoCollection: utxoCollection{},
	}
}

// diffFrom returns the difference between this utxoSet and another
// diffFrom can only work when other is a diffUTXOSet, and its base utxoSet is this.
func (fus *fullUTXOSet) diffFrom(other utxoSet) (*utxoDiff, error) {
	otherDiffSet, ok := other.(*diffUTXOSet)
	if !ok {
		return nil, errors.New("can't diffFrom two fullUTXOSets")
	}

	if otherDiffSet.base != fus {
		return nil, errors.New("can diffFrom only with diffUTXOSet where this fullUTXOSet is the base")
	}

	return otherDiffSet.utxoDiff, nil
}

// withDiff returns a utxoSet which is a diff between this and another utxoSet
func (fus *fullUTXOSet) withDiff(other *utxoDiff) (utxoSet, error) {
	return newDiffUTXOSet(fus, other.clone()), nil
}

// addTx adds a transaction to this utxoSet and returns true iff it's valid in this UTXO's context
func (fus *fullUTXOSet) addTx(tx *wire.MsgTx) bool {
	if !fus.areInputsInUTXO(tx) {
		return false
	}

	for _, txIn := range tx.TxIn {
		outPoint := *wire.NewOutPoint(&txIn.PreviousOutPoint.Hash, txIn.PreviousOutPoint.Index)
		delete(fus.utxoCollection, outPoint)
	}

	for i, txOut := range tx.TxOut {
		hash := tx.TxHash()
		outPoint := *wire.NewOutPoint(&hash, uint32(i))
		entry := newUTXOEntry(txOut)

		fus.utxoCollection[outPoint] = entry
	}

	return true
}

func (fus *fullUTXOSet) areInputsInUTXO(tx *wire.MsgTx) bool {
	for _, txIn := range tx.TxIn {
		outPoint := *wire.NewOutPoint(&txIn.PreviousOutPoint.Hash, txIn.PreviousOutPoint.Index)
		if _, ok := fus.utxoCollection[outPoint]; !ok {
			return false
		}
	}

	return true
}

// collection returns a collection of all UTXOs in this set
func (fus *fullUTXOSet) collection() utxoCollection {
	return fus.utxoCollection.clone()
}

// clone returns a clone of this utxoSet
func (fus *fullUTXOSet) clone() utxoSet {
	return &fullUTXOSet{utxoCollection: fus.utxoCollection.clone()}
}

// iterate returns an iterator for a fullUTXOSet
func (fus *fullUTXOSet) iterate() utxoIterator {
	iterator := make(chan utxoIteratorOutput)

	go func() {
		for outPoint, entry := range fus.utxoCollection {
			iterator <- utxoIteratorOutput{outPoint: outPoint, entry: entry}
		}

		close(iterator)
	}()

	return iterator
}

// diffUTXOSet represents a utxoSet with a base fullUTXOSet and a UTXODiff
type diffUTXOSet struct {
	base     *fullUTXOSet
	utxoDiff *utxoDiff
}

// newDiffUTXOSet Creates a new utxoSet based on a base fullUTXOSet and a UTXODiff
func newDiffUTXOSet(base *fullUTXOSet, diff *utxoDiff) *diffUTXOSet {
	return &diffUTXOSet{
		base:     base,
		utxoDiff: diff,
	}
}

// diffFrom returns the difference between this utxoSet and another.
// diffFrom can work if other is this's base fullUTXOSet, or a diffUTXOSet with the same base as this
func (dus *diffUTXOSet) diffFrom(other utxoSet) (*utxoDiff, error) {
	otherDiffSet, ok := other.(*diffUTXOSet)
	if !ok {
		return nil, errors.New("can't diffFrom diffUTXOSet with fullUTXOSet")
	}

	if otherDiffSet.base != dus.base {
		return nil, errors.New("can't diffFrom with another diffUTXOSet with a different base")
	}

	return dus.utxoDiff.diffFrom(otherDiffSet.utxoDiff)
}

// withDiff return a new utxoSet which is a diffFrom between this and another utxoSet
func (dus *diffUTXOSet) withDiff(other *utxoDiff) (utxoSet, error) {
	diff, err := dus.utxoDiff.withDiff(other)
	if err != nil {
		return nil, err
	}

	return newDiffUTXOSet(dus.base, diff), nil
}

// addTx adds a transaction to this utxoSet and returns true iff it's valid in this UTXO's context
func (dus *diffUTXOSet) addTx(tx *wire.MsgTx) bool {
	if !dus.areInputsInUTXO(tx) {
		return false
	}

	for _, txIn := range tx.TxIn {
		outPoint := *wire.NewOutPoint(&txIn.PreviousOutPoint.Hash, txIn.PreviousOutPoint.Index)
		if _, ok := dus.utxoDiff.toAdd[outPoint]; ok {
			delete(dus.utxoDiff.toAdd, outPoint)
		} else {
			prevUTXOEntry := dus.base.utxoCollection[outPoint]
			dus.utxoDiff.toRemove[outPoint] = prevUTXOEntry
		}
	}

	for i, txOut := range tx.TxOut {
		hash := tx.TxHash()
		outPoint := *wire.NewOutPoint(&hash, uint32(i))
		entry := newUTXOEntry(txOut)

		if _, ok := dus.utxoDiff.toRemove[outPoint]; ok {
			delete(dus.utxoDiff.toRemove, outPoint)
		} else {
			dus.utxoDiff.toAdd[outPoint] = entry
		}
	}

	return true
}

func (dus *diffUTXOSet) areInputsInUTXO(tx *wire.MsgTx) bool {
	for _, txIn := range tx.TxIn {
		outPoint := *wire.NewOutPoint(&txIn.PreviousOutPoint.Hash, txIn.PreviousOutPoint.Index)
		_, isInBase := dus.base.utxoCollection[outPoint]
		_, isInDiffToAdd := dus.utxoDiff.toAdd[outPoint]
		_, isInDiffToRemove := dus.utxoDiff.toRemove[outPoint]
		if (!isInBase && !isInDiffToAdd) || isInDiffToRemove {
			return false
		}
	}

	return true
}

// meldToBase updates the base fullUTXOSet with all changes in diff
func (dus *diffUTXOSet) meldToBase() {
	for outPoint := range dus.utxoDiff.toRemove {
		delete(dus.base.utxoCollection, outPoint)
	}

	for outPoint, utxoEntry := range dus.utxoDiff.toAdd {
		dus.base.utxoCollection[outPoint] = utxoEntry
	}

	dus.utxoDiff = newUTXODiff()
}

func (dus *diffUTXOSet) String() string {
	return fmt.Sprintf("{Base: %s, To Add: %s, To Remove: %s}", dus.base, dus.utxoDiff.toAdd, dus.utxoDiff.toRemove)
}

// collection returns a collection of all UTXOs in this set
func (dus *diffUTXOSet) collection() utxoCollection {
	clone := dus.clone().(*diffUTXOSet)
	clone.meldToBase()

	return clone.base.collection()
}

// clone returns a clone of this UTXO Set
func (dus *diffUTXOSet) clone() utxoSet {
	return newDiffUTXOSet(dus.base.clone().(*fullUTXOSet), dus.utxoDiff.clone())
}

// iterate returns an iterator for a diffUTXOSet
func (dus *diffUTXOSet) iterate() utxoIterator {
	iterator := make(chan utxoIteratorOutput)

	go func() {
		for outPoint, entry := range dus.base.utxoCollection {
			if _, ok := dus.utxoDiff.toRemove[outPoint]; !ok {
				iterator <- utxoIteratorOutput{outPoint: outPoint, entry: entry}
			}
		}

		for outPoint, entry := range dus.utxoDiff.toAdd {
			iterator <- utxoIteratorOutput{outPoint: outPoint, entry: entry}
		}
		close(iterator)
	}()

	return iterator
}
