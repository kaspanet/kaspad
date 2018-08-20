package blockdag

import (
	"fmt"
	"github.com/daglabs/btcd/wire"
	"errors"
	"sort"
	"strings"
)

// utxoCollection represents a set of UTXOs indexed by their outPoints
type utxoCollection map[wire.OutPoint]*UtxoEntry

func (uc utxoCollection) String() string {
	utxoStrings := make([]string, len(uc))

	i := 0
	for outPoint, utxoEntry := range uc {
		utxoStrings[i] = fmt.Sprintf("(%s, %d) => %d", outPoint.Hash, outPoint.Index, utxoEntry.amount)
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

// utxoDiff represents a diff between two UTXO Sets.
type utxoDiff struct {
	toAdd    utxoCollection
	toRemove utxoCollection
}

// newUTXODiff creates a new, empty utxoDiff
func newUTXODiff() *utxoDiff {
	return &utxoDiff{
		toAdd:    utxoCollection{},
		toRemove: utxoCollection{},
	}
}

// diffFrom returns a new utxoDiff with the difference between this utxoDiff and another
// Assumes that:
// Both utxoDiffs are from the same base
// If a txOut exists in both utxoDiffs, its underlying values would be the same
//
// diffFrom follows a set of rules represented by the following 3 by 3 table:
//
//          |           | this      |           |
// ---------+-----------+-----------+-----------+-----------
//          |           | toAdd     | toRemove  | None
// ---------+-----------+-----------+-----------+-----------
// other    | toAdd     | -         | X         | toAdd
// ---------+-----------+-----------+-----------+-----------
//          | toRemove  | X         | -         | toRemove
// ---------+-----------+-----------+-----------+-----------
//          | None      | toRemove  | toAdd     | -
//
// Key:
// -		Don't add anything to the result
// X		Return an error
// toAdd	Add the UTXO into the toAdd collection of the result
// toRemove	Add the UTXO into the toRemove collection of the result
//
// Examples:
// 1. This diff contains a UTXO in toAdd, and the other diff contains it in toRemove
//    diffFrom results in an error
// 2. This diff contains a UTXO in toRemove, and the other diff does not contain it
//    diffFrom results in the UTXO being added to toAdd
func (d *utxoDiff) diffFrom(other *utxoDiff) (*utxoDiff, error) {
	result := newUTXODiff()

	// Note that the following cases are not accounted for, as they are impossible
	// as long as the base utxoSet is the same:
	// - if utxoEntry is in d.toAdd and other.toRemove
	// - if utxoEntry is in d.toRemove and other.toAdd

	// All transactions in d.toAdd:
	// If they are not in other.toAdd - should be added in result.toRemove
	// If they are in other.toRemove - base utxoSet is not the same
	for outPoint, utxoEntry := range d.toAdd {
		if _, ok := other.toAdd[outPoint]; !ok {
			result.toRemove[outPoint] = utxoEntry
		}
		if _, ok := other.toRemove[outPoint]; ok {
			return nil, errors.New("diffFrom: transaction both in d.toAdd and in other.toRemove")
		}
	}

	// All transactions in d.toRemove:
	// If they are not in other.toRemove - should be added in result.toAdd
	// If they are in other.toAdd - base utxoSet is not the same
	for outPoint, utxoEntry := range d.toRemove {
		if _, ok := other.toRemove[outPoint]; !ok {
			result.toAdd[outPoint] = utxoEntry
		}
		if _, ok := other.toAdd[outPoint]; ok {
			return nil, errors.New("diffFrom: transaction both in d.toRemove and in other.toAdd")
		}
	}

	// All transactions in other.toAdd:
	// If they are not in d.toAdd - should be added in result.toAdd
	for outPoint, utxoEntry := range other.toAdd {
		if _, ok := d.toAdd[outPoint]; !ok {
			result.toAdd[outPoint] = utxoEntry
		}
	}

	// All transactions in other.toRemove:
	// If they are not in d.toRemove - should be added in result.toRemove
	for outPoint, utxoEntry := range other.toRemove {
		if _, ok := d.toRemove[outPoint]; !ok {
			result.toRemove[outPoint] = utxoEntry
		}
	}

	return result, nil
}

// withDiff applies provided diff to this diff, creating a new utxoDiff, that would be the result if
// first d, and than diff were applied to the same base
//
// withDiff follows a set of rules represented by the following 3 by 3 table:
//
//          |           | this      |           |
// ---------+-----------+-----------+-----------+-----------
//          |           | toAdd     | toRemove  | None
// ---------+-----------+-----------+-----------+-----------
// other    | toAdd     | X         | -         | toAdd
// ---------+-----------+-----------+-----------+-----------
//          | toRemove  | -         | X         | toRemove
// ---------+-----------+-----------+-----------+-----------
//          | None      | toAdd     | toRemove  | -
//
// Key:
// -		Don't add anything to the result
// X		Return an error
// toAdd	Add the UTXO into the toAdd collection of the result
// toRemove	Add the UTXO into the toRemove collection of the result
//
// Examples:
// 1. This diff contains a UTXO in toAdd, and the other diff contains it in toRemove
//    withDiff results in nothing being added
// 2. This diff contains a UTXO in toRemove, and the other diff does not contain it
//    withDiff results in the UTXO being added to toRemove
func (d *utxoDiff) withDiff(diff *utxoDiff) (*utxoDiff, error) {
	result := newUTXODiff()

	// All transactions in d.toAdd:
	// If they are not in diff.toRemove - should be added in result.toAdd
	// If they are in diff.toAdd - should throw an error
	// Otherwise - should be ignored
	for outPoint, utxoEntry := range d.toAdd {
		if _, ok := diff.toRemove[outPoint]; !ok {
			result.toAdd[outPoint] = utxoEntry
		}
		if _, ok := diff.toAdd[outPoint]; ok {
			return nil, errors.New("withDiff: transaction both in d.toAdd and in other.toAdd")
		}
	}

	// All transactions in d.toRemove:
	// If they are not in diff.toAdd - should be added in result.toRemove
	// If they are in diff.toRemove - should throw an error
	// Otherwise - should be ignored
	for outPoint, utxoEntry := range d.toRemove {
		if _, ok := diff.toAdd[outPoint]; !ok {
			result.toRemove[outPoint] = utxoEntry
		}
		if _, ok := diff.toRemove[outPoint]; ok {
			return nil, errors.New("withDiff: transaction both in d.toRemove and in other.toRemove")
		}
	}

	// All transactions in diff.toAdd:
	// If they are not in d.toRemove - should be added in result.toAdd
	for outPoint, utxoEntry := range diff.toAdd {
		if _, ok := d.toRemove[outPoint]; !ok {
			result.toAdd[outPoint] = utxoEntry
		}
	}

	// All transactions in diff.toRemove:
	// If they are not in d.toAdd - should be added in result.toRemove
	for outPoint, utxoEntry := range diff.toRemove {
		if _, ok := d.toAdd[outPoint]; !ok {
			result.toRemove[outPoint] = utxoEntry
		}
	}

	return result, nil
}

// clone returns a clone of this utxoDiff
func (d *utxoDiff) clone() *utxoDiff {
	return &utxoDiff{
		toAdd:    d.toAdd.clone(),
		toRemove: d.toRemove.clone(),
	}
}

func (d utxoDiff) String() string {
	return fmt.Sprintf("toAdd: %s; toRemove: %s", d.toAdd, d.toRemove)
}

// newUTXOEntry creates a new utxoEntry representing the given txOut
func newUTXOEntry(txOut *wire.TxOut, isCoinbase bool, blockHeight int32) *UtxoEntry {
	entry := new(UtxoEntry)
	entry.amount = txOut.Value
	entry.pkScript = txOut.PkScript
	entry.blockHeight = blockHeight

	entry.packedFlags = tfModified
	if isCoinbase {
		entry.packedFlags |= tfCoinBase
	}

	return entry
}

// utxoSet represents a set of unspent transaction outputs
// Every DAG has exactly one fullUTXOSet.
// When a new block arrives, it is validated and applied to the fullUTXOSet in the following manner:
// 1. Get the block's PastUTXO:
// 2. Add all the block's transactions to the block's PastUTXO
// 3. For each of the block's parents,
// 3.1. Rebuild their utxoDiff
// 3.2. Set the block as their diffChild
// 4. Create and initialize a new virtual block
// 5. Get the new virtual's PastUTXO
// 6. Rebuild the utxoDiff for all the tips
// 7. Convert (meld) the new virtual's diffUTXOSet into a fullUTXOSet. This updates the DAG's fullUTXOSet
type utxoSet interface {
	fmt.Stringer
	diffFrom(other utxoSet) (*utxoDiff, error)
	withDiff(utxoDiff *utxoDiff) (utxoSet, error)
	addTx(tx *wire.MsgTx, blockHeight int32) (ok bool)
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
func (fus *fullUTXOSet) addTx(tx *wire.MsgTx, blockHeight int32) bool {
	isCoinbase := IsCoinBaseTx(tx)
	if !isCoinbase {
		if !fus.containsInputs(tx) {
			return false
		}

		for _, txIn := range tx.TxIn {
			outPoint := *wire.NewOutPoint(&txIn.PreviousOutPoint.Hash, txIn.PreviousOutPoint.Index)
			delete(fus.utxoCollection, outPoint)
		}
	}

	for i, txOut := range tx.TxOut {
		hash := tx.TxHash()
		outPoint := *wire.NewOutPoint(&hash, uint32(i))
		entry := newUTXOEntry(txOut, isCoinbase, blockHeight)

		fus.utxoCollection[outPoint] = entry
	}

	return true
}

func (fus *fullUTXOSet) containsInputs(tx *wire.MsgTx) bool {
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

func (fus *fullUTXOSet) getUTXOEntry(outPoint wire.OutPoint) (*UtxoEntry, bool) {
	utxoEntry, ok := fus.utxoCollection[outPoint]
	return utxoEntry, ok
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
func (dus *diffUTXOSet) addTx(tx *wire.MsgTx, blockHeight int32) bool {
	isCoinbase := IsCoinBaseTx(tx)
	if !isCoinbase {
		if !dus.containsInputs(tx) {
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
	}

	for i, txOut := range tx.TxOut {
		hash := tx.TxHash()
		outPoint := *wire.NewOutPoint(&hash, uint32(i))
		entry := newUTXOEntry(txOut, isCoinbase, blockHeight)

		if _, ok := dus.utxoDiff.toRemove[outPoint]; ok {
			delete(dus.utxoDiff.toRemove, outPoint)
		} else {
			dus.utxoDiff.toAdd[outPoint] = entry
		}
	}

	return true
}

func (dus *diffUTXOSet) containsInputs(tx *wire.MsgTx) bool {
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
