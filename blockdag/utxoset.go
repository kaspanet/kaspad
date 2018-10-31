package blockdag

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/daglabs/btcd/wire"
)

// UTXOEntry houses details about an individual transaction output in a utxo
// set such as whether or not it was contained in a coinbase tx, the height of
// the block that contains the tx, whether or not it is spent, its public key
// script, and how much it pays.
type UTXOEntry struct {
	// NOTE: Additions, deletions, or modifications to the order of the
	// definitions in this struct should not be changed without considering
	// how it affects alignment on 64-bit platforms.  The current order is
	// specifically crafted to result in minimal padding.  There will be a
	// lot of these in memory, so a few extra bytes of padding adds up.

	amount      uint64
	pkScript    []byte // The public key script for the output.
	blockHeight int32  // Height of block containing tx.

	// packedFlags contains additional info about output such as whether it
	// is a coinbase, whether it is spent, and whether it has been modified
	// since it was loaded.  This approach is used in order to reduce memory
	// usage since there will be a lot of these in memory.
	packedFlags txoFlags
}

// IsCoinBase returns whether or not the output was contained in a coinbase
// transaction.
func (entry *UTXOEntry) IsCoinBase() bool {
	return entry.packedFlags&tfCoinBase == tfCoinBase
}

// BlockHeight returns the height of the block containing the output.
func (entry *UTXOEntry) BlockHeight() int32 {
	return entry.blockHeight
}

// Amount returns the amount of the output.
func (entry *UTXOEntry) Amount() uint64 {
	return entry.amount
}

// PkScript returns the public key script for the output.
func (entry *UTXOEntry) PkScript() []byte {
	return entry.pkScript
}

// Clone returns a shallow copy of the utxo entry.
func (entry *UTXOEntry) Clone() *UTXOEntry {
	if entry == nil {
		return nil
	}

	return &UTXOEntry{
		amount:      entry.amount,
		pkScript:    entry.pkScript,
		blockHeight: entry.blockHeight,
		packedFlags: entry.packedFlags,
	}
}

// txoFlags is a bitmask defining additional information and state for a
// transaction output in a UTXO set.
type txoFlags uint8

const (
	// tfCoinBase indicates that a txout was contained in a coinbase tx.
	tfCoinBase txoFlags = 1 << iota
)

// utxoCollection represents a set of UTXOs indexed by their outPoints
type utxoCollection map[wire.OutPoint]*UTXOEntry

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

// add adds a new UTXO entry to this collection
func (uc utxoCollection) add(outPoint wire.OutPoint, entry *UTXOEntry) {
	uc[outPoint] = entry
}

// remove removes a UTXO entry from this collection if it exists
func (uc utxoCollection) remove(outPoint wire.OutPoint) {
	delete(uc, outPoint)
}

// get returns the UTXOEntry represented by provided outPoint,
// and a boolean value indicating if said UTXOEntry is in the set or not
func (uc utxoCollection) get(outPoint wire.OutPoint) (*UTXOEntry, bool) {
	entry, ok := uc[outPoint]
	return entry, ok
}

// contains returns a boolean value indicating whether a UTXO entry is in the set
func (uc utxoCollection) contains(outPoint wire.OutPoint) bool {
	_, ok := uc[outPoint]
	return ok
}

// clone returns a clone of this collection
func (uc utxoCollection) clone() utxoCollection {
	clone := utxoCollection{}
	for outPoint, entry := range uc {
		clone.add(outPoint, entry)
	}

	return clone
}

// UTXODiff represents a diff between two UTXO Sets.
type UTXODiff struct {
	toAdd    utxoCollection
	toRemove utxoCollection
}

// NewUTXODiff creates a new, empty utxoDiff
func NewUTXODiff() *UTXODiff {
	return &UTXODiff{
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
func (d *UTXODiff) diffFrom(other *UTXODiff) (*UTXODiff, error) {
	result := NewUTXODiff()

	// Note that the following cases are not accounted for, as they are impossible
	// as long as the base utxoSet is the same:
	// - if utxoEntry is in d.toAdd and other.toRemove
	// - if utxoEntry is in d.toRemove and other.toAdd

	// All transactions in d.toAdd:
	// If they are not in other.toAdd - should be added in result.toRemove
	// If they are in other.toRemove - base utxoSet is not the same
	for outPoint, utxoEntry := range d.toAdd {
		if !other.toAdd.contains(outPoint) {
			result.toRemove.add(outPoint, utxoEntry)
		}
		if other.toRemove.contains(outPoint) {
			return nil, errors.New("diffFrom: transaction both in d.toAdd and in other.toRemove")
		}
	}

	// All transactions in d.toRemove:
	// If they are not in other.toRemove - should be added in result.toAdd
	// If they are in other.toAdd - base utxoSet is not the same
	for outPoint, utxoEntry := range d.toRemove {
		if !other.toRemove.contains(outPoint) {
			result.toAdd.add(outPoint, utxoEntry)
		}
		if other.toAdd.contains(outPoint) {
			return nil, errors.New("diffFrom: transaction both in d.toRemove and in other.toAdd")
		}
	}

	// All transactions in other.toAdd:
	// If they are not in d.toAdd - should be added in result.toAdd
	for outPoint, utxoEntry := range other.toAdd {
		if !d.toAdd.contains(outPoint) {
			result.toAdd.add(outPoint, utxoEntry)
		}
	}

	// All transactions in other.toRemove:
	// If they are not in d.toRemove - should be added in result.toRemove
	for outPoint, utxoEntry := range other.toRemove {
		if !d.toRemove.contains(outPoint) {
			result.toRemove.add(outPoint, utxoEntry)
		}
	}

	return result, nil
}

// WithDiff applies provided diff to this diff, creating a new utxoDiff, that would be the result if
// first d, and than diff were applied to the same base
//
// WithDiff follows a set of rules represented by the following 3 by 3 table:
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
//    WithDiff results in nothing being added
// 2. This diff contains a UTXO in toRemove, and the other diff does not contain it
//    WithDiff results in the UTXO being added to toRemove
func (d *UTXODiff) WithDiff(diff *UTXODiff) (*UTXODiff, error) {
	result := NewUTXODiff()

	// All transactions in d.toAdd:
	// If they are not in diff.toRemove - should be added in result.toAdd
	// If they are in diff.toAdd - should throw an error
	// Otherwise - should be ignored
	for outPoint, utxoEntry := range d.toAdd {
		if !diff.toRemove.contains(outPoint) {
			result.toAdd.add(outPoint, utxoEntry)
		}
		if diff.toAdd.contains(outPoint) {
			return nil, ruleError(ErrWithDiff, "WithDiff: transaction both in d.toAdd and in other.toAdd")
		}
	}

	// All transactions in d.toRemove:
	// If they are not in diff.toAdd - should be added in result.toRemove
	// If they are in diff.toRemove - should throw an error
	// Otherwise - should be ignored
	for outPoint, utxoEntry := range d.toRemove {
		if !diff.toAdd.contains(outPoint) {
			result.toRemove.add(outPoint, utxoEntry)
		}
		if diff.toRemove.contains(outPoint) {
			return nil, ruleError(ErrWithDiff, "WithDiff: transaction both in d.toRemove and in other.toRemove")
		}
	}

	// All transactions in diff.toAdd:
	// If they are not in d.toRemove - should be added in result.toAdd
	for outPoint, utxoEntry := range diff.toAdd {
		if !d.toRemove.contains(outPoint) {
			result.toAdd.add(outPoint, utxoEntry)
		}
	}

	// All transactions in diff.toRemove:
	// If they are not in d.toAdd - should be added in result.toRemove
	for outPoint, utxoEntry := range diff.toRemove {
		if !d.toAdd.contains(outPoint) {
			result.toRemove.add(outPoint, utxoEntry)
		}
	}

	return result, nil
}

// clone returns a clone of this utxoDiff
func (d *UTXODiff) clone() *UTXODiff {
	return &UTXODiff{
		toAdd:    d.toAdd.clone(),
		toRemove: d.toRemove.clone(),
	}
}

//RemoveTxOuts marks the transaction's outputs to removal
func (d *UTXODiff) RemoveTxOuts(tx *wire.MsgTx) {
	for idx := range tx.TxOut {
		hash := tx.TxHash()
		d.toRemove.add(*wire.NewOutPoint(&hash, uint32(idx)), nil)
	}
}

//AddEntry adds an UTXOEntry to the diff
func (d *UTXODiff) AddEntry(outpoint wire.OutPoint, entry *UTXOEntry) {
	d.toAdd.add(outpoint, entry)
}

func (d UTXODiff) String() string {
	return fmt.Sprintf("toAdd: %s; toRemove: %s", d.toAdd, d.toRemove)
}

// NewUTXOEntry creates a new utxoEntry representing the given txOut
func NewUTXOEntry(txOut *wire.TxOut, isCoinbase bool, blockHeight int32) *UTXOEntry {
	entry := &UTXOEntry{
		amount:      txOut.Value,
		pkScript:    txOut.PkScript,
		blockHeight: blockHeight,
	}

	if isCoinbase {
		entry.packedFlags |= tfCoinBase
	}

	return entry
}

// UTXOSet represents a set of unspent transaction outputs
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
type UTXOSet interface {
	fmt.Stringer
	diffFrom(other UTXOSet) (*UTXODiff, error)
	WithDiff(utxoDiff *UTXODiff) (UTXOSet, error)
	diffFromTx(tx *wire.MsgTx, node *blockNode) (*UTXODiff, error)
	AddTx(tx *wire.MsgTx, blockHeight int32) (ok bool)
	clone() UTXOSet
	Get(outPoint wire.OutPoint) (*UTXOEntry, bool)
}

// diffFromTx is a common implementation for diffFromTx, that works
// for both diff-based and full UTXO sets
// Returns a diff that is equivalent to provided transaction,
// or an error if provided transaction is not valid in the context of this UTXOSet
func diffFromTx(u UTXOSet, tx *wire.MsgTx, containingNode *blockNode) (*UTXODiff, error) {
	diff := NewUTXODiff()
	isCoinbase := IsCoinBaseTx(tx)
	if !isCoinbase {
		for _, txIn := range tx.TxIn {
			if entry, ok := u.Get(txIn.PreviousOutPoint); ok {
				diff.toRemove.add(txIn.PreviousOutPoint, entry)
			} else {
				return nil, ruleError(ErrMissingTxOut, fmt.Sprintf(
					"Transaction %s is invalid because spends outpoint %s that is not in utxo set",
					tx.TxHash(), txIn.PreviousOutPoint))
			}
		}
	}
	for i, txOut := range tx.TxOut {
		hash := tx.TxHash()
		entry := NewUTXOEntry(txOut, isCoinbase, containingNode.height)
		outPoint := *wire.NewOutPoint(&hash, uint32(i))
		diff.toAdd.add(outPoint, entry)
	}
	return diff, nil
}

// FullUTXOSet represents a full list of transaction outputs and their values
type FullUTXOSet struct {
	utxoCollection
}

// NewFullUTXOSet creates a new utxoSet with full list of transaction outputs and their values
func NewFullUTXOSet() *FullUTXOSet {
	return &FullUTXOSet{
		utxoCollection: utxoCollection{},
	}
}

// diffFrom returns the difference between this utxoSet and another
// diffFrom can only work when other is a diffUTXOSet, and its base utxoSet is this.
func (fus *FullUTXOSet) diffFrom(other UTXOSet) (*UTXODiff, error) {
	otherDiffSet, ok := other.(*DiffUTXOSet)
	if !ok {
		return nil, errors.New("can't diffFrom two fullUTXOSets")
	}

	if otherDiffSet.base != fus {
		return nil, errors.New("can diffFrom only with diffUTXOSet where this fullUTXOSet is the base")
	}

	return otherDiffSet.UTXODiff, nil
}

// WithDiff returns a utxoSet which is a diff between this and another utxoSet
func (fus *FullUTXOSet) WithDiff(other *UTXODiff) (UTXOSet, error) {
	return NewDiffUTXOSet(fus, other.clone()), nil
}

// AddTx adds a transaction to this utxoSet and returns true iff it's valid in this UTXO's context
func (fus *FullUTXOSet) AddTx(tx *wire.MsgTx, blockHeight int32) bool {
	isCoinbase := IsCoinBaseTx(tx)
	if !isCoinbase {
		if !fus.containsInputs(tx) {
			return false
		}

		for _, txIn := range tx.TxIn {
			outPoint := *wire.NewOutPoint(&txIn.PreviousOutPoint.Hash, txIn.PreviousOutPoint.Index)
			fus.remove(outPoint)
		}
	}

	for i, txOut := range tx.TxOut {
		hash := tx.TxHash()
		outPoint := *wire.NewOutPoint(&hash, uint32(i))
		entry := NewUTXOEntry(txOut, isCoinbase, blockHeight)

		fus.add(outPoint, entry)
	}

	return true
}

// diffFromTx returns a diff that is equivalent to provided transaction,
// or an error if provided transaction is not valid in the context of this UTXOSet
func (fus *FullUTXOSet) diffFromTx(tx *wire.MsgTx, node *blockNode) (*UTXODiff, error) {
	return diffFromTx(fus, tx, node)
}

func (fus *FullUTXOSet) containsInputs(tx *wire.MsgTx) bool {
	for _, txIn := range tx.TxIn {
		outPoint := *wire.NewOutPoint(&txIn.PreviousOutPoint.Hash, txIn.PreviousOutPoint.Index)
		if !fus.contains(outPoint) {
			return false
		}
	}

	return true
}

// clone returns a clone of this utxoSet
func (fus *FullUTXOSet) clone() UTXOSet {
	return &FullUTXOSet{utxoCollection: fus.utxoCollection.clone()}
}

func (fus *FullUTXOSet) Get(outPoint wire.OutPoint) (*UTXOEntry, bool) {
	utxoEntry, ok := fus.utxoCollection[outPoint]
	return utxoEntry, ok
}

// DiffUTXOSet represents a utxoSet with a base fullUTXOSet and a UTXODiff
type DiffUTXOSet struct {
	base     *FullUTXOSet
	UTXODiff *UTXODiff
}

// NewDiffUTXOSet Creates a new utxoSet based on a base fullUTXOSet and a UTXODiff
func NewDiffUTXOSet(base *FullUTXOSet, diff *UTXODiff) *DiffUTXOSet {
	return &DiffUTXOSet{
		base:     base,
		UTXODiff: diff,
	}
}

// diffFrom returns the difference between this utxoSet and another.
// diffFrom can work if other is this's base fullUTXOSet, or a diffUTXOSet with the same base as this
func (dus *DiffUTXOSet) diffFrom(other UTXOSet) (*UTXODiff, error) {
	otherDiffSet, ok := other.(*DiffUTXOSet)
	if !ok {
		return nil, errors.New("can't diffFrom diffUTXOSet with fullUTXOSet")
	}

	if otherDiffSet.base != dus.base {
		return nil, errors.New("can't diffFrom with another diffUTXOSet with a different base")
	}

	return dus.UTXODiff.diffFrom(otherDiffSet.UTXODiff)
}

// WithDiff return a new utxoSet which is a diffFrom between this and another utxoSet
func (dus *DiffUTXOSet) WithDiff(other *UTXODiff) (UTXOSet, error) {
	diff, err := dus.UTXODiff.WithDiff(other)
	if err != nil {
		return nil, err
	}

	return NewDiffUTXOSet(dus.base, diff), nil
}

// AddTx adds a transaction to this utxoSet and returns true iff it's valid in this UTXO's context
func (dus *DiffUTXOSet) AddTx(tx *wire.MsgTx, blockHeight int32) bool {
	isCoinBase := IsCoinBaseTx(tx)
	if !isCoinBase && !dus.containsInputs(tx) {
		return false
	}

	dus.appendTx(tx, blockHeight, isCoinBase)

	return true
}

func (dus *DiffUTXOSet) appendTx(tx *wire.MsgTx, blockHeight int32, isCoinBase bool) {
	if !isCoinBase {

		for _, txIn := range tx.TxIn {
			outPoint := *wire.NewOutPoint(&txIn.PreviousOutPoint.Hash, txIn.PreviousOutPoint.Index)
			if dus.UTXODiff.toAdd.contains(outPoint) {
				dus.UTXODiff.toAdd.remove(outPoint)
			} else {
				prevUTXOEntry := dus.base.utxoCollection[outPoint]
				dus.UTXODiff.toRemove.add(outPoint, prevUTXOEntry)
			}
		}
	}

	for i, txOut := range tx.TxOut {
		hash := tx.TxHash()
		outPoint := *wire.NewOutPoint(&hash, uint32(i))
		entry := NewUTXOEntry(txOut, isCoinBase, blockHeight)

		if dus.UTXODiff.toRemove.contains(outPoint) {
			dus.UTXODiff.toRemove.remove(outPoint)
		} else {
			dus.UTXODiff.toAdd.add(outPoint, entry)
		}
	}
}

func (dus *DiffUTXOSet) containsInputs(tx *wire.MsgTx) bool {
	for _, txIn := range tx.TxIn {
		outPoint := *wire.NewOutPoint(&txIn.PreviousOutPoint.Hash, txIn.PreviousOutPoint.Index)
		isInBase := dus.base.contains(outPoint)
		isInDiffToAdd := dus.UTXODiff.toAdd.contains(outPoint)
		isInDiffToRemove := dus.UTXODiff.toRemove.contains(outPoint)
		if (!isInBase && !isInDiffToAdd) || isInDiffToRemove {
			return false
		}
	}

	return true
}

// meldToBase updates the base fullUTXOSet with all changes in diff
func (dus *DiffUTXOSet) meldToBase() {
	for outPoint := range dus.UTXODiff.toRemove {
		dus.base.remove(outPoint)
	}

	for outPoint, utxoEntry := range dus.UTXODiff.toAdd {
		dus.base.add(outPoint, utxoEntry)
	}

	dus.UTXODiff = NewUTXODiff()
}

// diffFromTx returns a diff that is equivalent to provided transaction,
// or an error if provided transaction is not valid in the context of this UTXOSet
func (dus *DiffUTXOSet) diffFromTx(tx *wire.MsgTx, node *blockNode) (*UTXODiff, error) {
	return diffFromTx(dus, tx, node)
}

func (dus *DiffUTXOSet) String() string {
	return fmt.Sprintf("{Base: %s, To Add: %s, To Remove: %s}", dus.base, dus.UTXODiff.toAdd, dus.UTXODiff.toRemove)
}

// clone returns a clone of this UTXO Set
func (dus *DiffUTXOSet) clone() UTXOSet {
	return NewDiffUTXOSet(dus.base.clone().(*FullUTXOSet), dus.UTXODiff.clone())
}

// Get returns the UTXOEntry associated with provided outPoint in this UTXOSet.
// Returns false in second output if this UTXOEntry was not found
func (dus *DiffUTXOSet) Get(outPoint wire.OutPoint) (*UTXOEntry, bool) {
	if dus.UTXODiff.toRemove.contains(outPoint) {
		return nil, false
	}
	if txOut, ok := dus.base.get(outPoint); ok {
		return txOut, true
	}
	txOut, ok := dus.UTXODiff.toAdd.get(outPoint)
	return txOut, ok
}
