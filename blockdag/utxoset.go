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
func (u *fullUTXOSet) diffFrom(other utxoSet) (*utxoDiff, error) {
	o, ok := other.(*diffUTXOSet)
	if !ok {
		return nil, errors.New("can't diffFrom two fullUTXOSets")
	}

	if o.base != u {
		return nil, errors.New("can diffFrom only with diffUTXOSet where this fullUTXOSet is the base")
	}

	return o.utxoDiff, nil
}

// withDiff returns a utxoSet which is a diff between this and other utxoSet
func (u *fullUTXOSet) withDiff(utxoDiff *utxoDiff) (utxoSet, error) {
	return newDiffUTXOSet(u, utxoDiff.clone()), nil
}

// addTx adds a transaction to this utxoSet and returns true iff it's valid in this UTXO's context
func (u *fullUTXOSet) addTx(tx *wire.MsgTx) bool {
	if !u.verifyTx(tx) {
		return false
	}

	for _, txIn := range tx.TxIn {
		outPoint := *wire.NewOutPoint(&txIn.PreviousOutPoint.Hash, txIn.PreviousOutPoint.Index)
		delete(u.utxoCollection, outPoint)
	}

	for i, txOut := range tx.TxOut {
		hash := tx.TxHash()
		outPoint := *wire.NewOutPoint(&hash, uint32(i))
		entry := newUTXOEntry(txOut)

		u.utxoCollection[outPoint] = entry
	}

	return true
}

func (u *fullUTXOSet) verifyTx(tx *wire.MsgTx) bool {
	for _, txIn := range tx.TxIn {
		outPoint := *wire.NewOutPoint(&txIn.PreviousOutPoint.Hash, txIn.PreviousOutPoint.Index)
		if _, ok := u.utxoCollection[outPoint]; !ok {
			return false
		}
	}

	return true
}

// collection returns a collection of all UTXOs in this set
func (u *fullUTXOSet) collection() utxoCollection {
	return u.utxoCollection.clone()
}

// clone returns a clone of this utxoSet
func (u *fullUTXOSet) clone() utxoSet {
	return &fullUTXOSet{utxoCollection: u.utxoCollection.clone()}
}

// iterate returns an iterator for a fullUTXOSet
func (u *fullUTXOSet) iterate() utxoIterator {
	iterator := make(chan utxoIteratorOutput)

	go func() {
		for outPoint, entry := range u.utxoCollection {
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
func (u *diffUTXOSet) diffFrom(other utxoSet) (*utxoDiff, error) {
	o, ok := other.(*diffUTXOSet)
	if !ok {
		return nil, errors.New("can't diffFrom diffUTXOSet with fullUTXOSet")
	}

	if o.base != u.base {
		return nil, errors.New("can't diffFrom with another diffUTXOSet with a different base")
	}

	return u.utxoDiff.diffFrom(o.utxoDiff)
}

// withDiff return a new utxoSet which is a diffFrom between this and other utxoSet
func (u *diffUTXOSet) withDiff(utxoDiff *utxoDiff) (utxoSet, error) {
	diff, err := u.utxoDiff.withDiff(utxoDiff)
	if err != nil {
		return nil, err
	}

	return newDiffUTXOSet(u.base, diff), nil
}

// addTx adds a transaction to this utxoSet and returns true iff it's valid in this UTXO's context
func (u *diffUTXOSet) addTx(tx *wire.MsgTx) bool {
	if !u.verifyTx(tx) {
		return false
	}

	for _, txIn := range tx.TxIn {
		outPoint := *wire.NewOutPoint(&txIn.PreviousOutPoint.Hash, txIn.PreviousOutPoint.Index)
		if _, ok := u.utxoDiff.toAdd[outPoint]; ok {
			delete(u.utxoDiff.toAdd, outPoint)
		} else {
			prevUTXOEntry := u.base.utxoCollection[outPoint]
			u.utxoDiff.toRemove[outPoint] = prevUTXOEntry
		}
	}

	for i, txOut := range tx.TxOut {
		hash := tx.TxHash()
		outPoint := *wire.NewOutPoint(&hash, uint32(i))
		entry := newUTXOEntry(txOut)

		if _, ok := u.utxoDiff.toRemove[outPoint]; ok {
			delete(u.utxoDiff.toRemove, outPoint)
		} else {
			u.utxoDiff.toAdd[outPoint] = entry
		}
	}

	return true
}

func (u *diffUTXOSet) verifyTx(tx *wire.MsgTx) bool {
	for _, txIn := range tx.TxIn {
		outPoint := *wire.NewOutPoint(&txIn.PreviousOutPoint.Hash, txIn.PreviousOutPoint.Index)
		_, isInBase := u.base.utxoCollection[outPoint]
		_, isInDiffToAdd := u.utxoDiff.toAdd[outPoint]
		_, isInDiffToRemove := u.utxoDiff.toRemove[outPoint]
		if (!isInBase && !isInDiffToAdd) || isInDiffToRemove {
			return false
		}
	}

	return true
}

// meldToBase updates the base fullUTXOSet with all changes in diff
func (u *diffUTXOSet) meldToBase() {
	for outPoint := range u.utxoDiff.toRemove {
		delete(u.base.utxoCollection, outPoint)
	}

	for outPoint, utxoEntry := range u.utxoDiff.toAdd {
		u.base.utxoCollection[outPoint] = utxoEntry
	}

	u.utxoDiff = newUTXODiff()
}

func (u *diffUTXOSet) String() string {
	return fmt.Sprintf("{Base: %s, To Add: %s, To Remove: %s}", u.base, u.utxoDiff.toAdd, u.utxoDiff.toRemove)
}

// collection returns a collection of all UTXOs in this set
func (u *diffUTXOSet) collection() utxoCollection {
	clone := u.clone().(*diffUTXOSet)
	clone.meldToBase()

	return clone.base.collection()
}

// clone returns a clone of this UTXO Set
func (u *diffUTXOSet) clone() utxoSet {
	return newDiffUTXOSet(u.base.clone().(*fullUTXOSet), u.utxoDiff.clone())
}

func newUTXOEntry(txOut *wire.TxOut) *UtxoEntry {
	entry := new(UtxoEntry)
	entry.amount = txOut.Value
	entry.pkScript = txOut.PkScript

	return entry
}

// iterate returns an iterator for a diffUTXOSet
func (u *diffUTXOSet) iterate() utxoIterator {
	iterator := make(chan utxoIteratorOutput)

	go func() {
		for outPoint, entry := range u.base.utxoCollection {
			if _, ok := u.utxoDiff.toRemove[outPoint]; !ok {
				iterator <- utxoIteratorOutput{outPoint: outPoint, entry: entry}
			}
		}

		for outPoint, entry := range u.utxoDiff.toAdd {
			iterator <- utxoIteratorOutput{outPoint: outPoint, entry: entry}
		}
		close(iterator)
	}()

	return iterator
}
