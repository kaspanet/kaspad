package blockdag

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/pkg/errors"

	"github.com/kaspanet/go-secp256k1"
	"github.com/kaspanet/kaspad/network/domainmessage"
)

const (
	// UnacceptedBlueScore is the blue score used for the "block" blueScore
	// field of the contextual transaction information provided in a
	// transaction store when it has not yet been accepted by a block.
	UnacceptedBlueScore uint64 = math.MaxUint64
)

// UTXOEntry houses details about an individual transaction output in a utxo
// set such as whether or not it was contained in a coinbase tx, the blue
// score of the block that accepts the tx, its public key script, and how
// much it pays.
type UTXOEntry struct {
	// NOTE: Additions, deletions, or modifications to the order of the
	// definitions in this struct should not be changed without considering
	// how it affects alignment on 64-bit platforms. The current order is
	// specifically crafted to result in minimal padding. There will be a
	// lot of these in memory, so a few extra bytes of padding adds up.

	amount         uint64
	scriptPubKey   []byte // The public key script for the output.
	blockBlueScore uint64 // Blue score of the block accepting the tx.

	// packedFlags contains additional info about output such as whether it
	// is a coinbase, and whether it has been modified
	// since it was loaded. This approach is used in order to reduce memory
	// usage since there will be a lot of these in memory.
	packedFlags txoFlags
}

// IsCoinbase returns whether or not the output was contained in a block
// reward transaction.
func (entry *UTXOEntry) IsCoinbase() bool {
	return entry.packedFlags&tfCoinbase == tfCoinbase
}

// BlockBlueScore returns the blue score of the block accepting the output.
func (entry *UTXOEntry) BlockBlueScore() uint64 {
	return entry.blockBlueScore
}

// Amount returns the amount of the output.
func (entry *UTXOEntry) Amount() uint64 {
	return entry.amount
}

// ScriptPubKey returns the public key script for the output.
func (entry *UTXOEntry) ScriptPubKey() []byte {
	return entry.scriptPubKey
}

// IsUnaccepted returns true iff this UTXOEntry has been included in a block
// but has not yet been accepted by any block.
func (entry *UTXOEntry) IsUnaccepted() bool {
	return entry.blockBlueScore == UnacceptedBlueScore
}

// txoFlags is a bitmask defining additional information and state for a
// transaction output in a UTXO set.
type txoFlags uint8

const (
	// tfCoinbase indicates that a txout was contained in a coinbase tx.
	tfCoinbase txoFlags = 1 << iota
)

// NewUTXOEntry creates a new utxoEntry representing the given txOut
func NewUTXOEntry(txOut *domainmessage.TxOut, isCoinbase bool, blockBlueScore uint64) *UTXOEntry {
	entry := &UTXOEntry{
		amount:         txOut.Value,
		scriptPubKey:   txOut.ScriptPubKey,
		blockBlueScore: blockBlueScore,
	}

	if isCoinbase {
		entry.packedFlags |= tfCoinbase
	}

	return entry
}

// utxoCollection represents a set of UTXOs indexed by their outpoints
type utxoCollection map[domainmessage.Outpoint]*UTXOEntry

func (uc utxoCollection) String() string {
	utxoStrings := make([]string, len(uc))

	i := 0
	for outpoint, utxoEntry := range uc {
		utxoStrings[i] = fmt.Sprintf("(%s, %d) => %d, blueScore: %d",
			outpoint.TxID, outpoint.Index, utxoEntry.amount, utxoEntry.blockBlueScore)
		i++
	}

	// Sort strings for determinism.
	sort.Strings(utxoStrings)

	return fmt.Sprintf("[ %s ]", strings.Join(utxoStrings, ", "))
}

// add adds a new UTXO entry to this collection
func (uc utxoCollection) add(outpoint domainmessage.Outpoint, entry *UTXOEntry) {
	uc[outpoint] = entry
}

// addMultiple adds multiple UTXO entries to this collection
func (uc utxoCollection) addMultiple(collectionToAdd utxoCollection) {
	for outpoint, entry := range collectionToAdd {
		uc[outpoint] = entry
	}
}

// remove removes a UTXO entry from this collection if it exists
func (uc utxoCollection) remove(outpoint domainmessage.Outpoint) {
	delete(uc, outpoint)
}

// removeMultiple removes multiple UTXO entries from this collection if it exists
func (uc utxoCollection) removeMultiple(collectionToRemove utxoCollection) {
	for outpoint := range collectionToRemove {
		delete(uc, outpoint)
	}
}

// get returns the UTXOEntry represented by provided outpoint,
// and a boolean value indicating if said UTXOEntry is in the set or not
func (uc utxoCollection) get(outpoint domainmessage.Outpoint) (*UTXOEntry, bool) {
	entry, ok := uc[outpoint]
	return entry, ok
}

// contains returns a boolean value indicating whether a UTXO entry is in the set
func (uc utxoCollection) contains(outpoint domainmessage.Outpoint) bool {
	_, ok := uc[outpoint]
	return ok
}

// containsWithBlueScore returns a boolean value indicating whether a UTXOEntry
// is in the set and its blue score is equal to the given blue score.
func (uc utxoCollection) containsWithBlueScore(outpoint domainmessage.Outpoint, blueScore uint64) bool {
	entry, ok := uc.get(outpoint)
	return ok && entry.blockBlueScore == blueScore
}

// clone returns a clone of this collection
func (uc utxoCollection) clone() utxoCollection {
	clone := utxoCollection{}
	for outpoint, entry := range uc {
		clone.add(outpoint, entry)
	}

	return clone
}

// UTXODiff represents a diff between two UTXO Sets.
type UTXODiff struct {
	toAdd    utxoCollection
	toRemove utxoCollection
}

// NewUTXODiff creates a new, empty utxoDiff
// without a multiset.
func NewUTXODiff() *UTXODiff {
	return &UTXODiff{
		toAdd:    utxoCollection{},
		toRemove: utxoCollection{},
	}
}

// checkIntersection checks if there is an intersection between two utxoCollections
func checkIntersection(collection1 utxoCollection, collection2 utxoCollection) bool {
	for outpoint := range collection1 {
		if collection2.contains(outpoint) {
			return true
		}
	}

	return false
}

// checkIntersectionWithRule checks if there is an intersection between two utxoCollections satisfying arbitrary rule
func checkIntersectionWithRule(collection1 utxoCollection, collection2 utxoCollection, extraRule func(domainmessage.Outpoint, *UTXOEntry, *UTXOEntry) bool) bool {
	for outpoint, utxoEntry := range collection1 {
		if diffEntry, ok := collection2.get(outpoint); ok {
			if extraRule(outpoint, utxoEntry, diffEntry) {
				return true
			}
		}
	}

	return false
}

// minInt returns the smaller of x or y integer values
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// intersectionWithRemainderHavingBlueScoreInPlace calculates an intersection between two utxoCollections
// having same blue score, returns the result and the remainder from collection1
func intersectionWithRemainderHavingBlueScore(collection1, collection2 utxoCollection) (result, remainder utxoCollection) {
	result = make(utxoCollection, minInt(len(collection1), len(collection2)))
	remainder = make(utxoCollection, len(collection1))
	intersectionWithRemainderHavingBlueScoreInPlace(collection1, collection2, result, remainder)
	return
}

// intersectionWithRemainderHavingBlueScoreInPlace calculates an intersection between two utxoCollections
// having same blue score, puts it into result and into remainder from collection1
func intersectionWithRemainderHavingBlueScoreInPlace(collection1, collection2, result, remainder utxoCollection) {
	for outpoint, utxoEntry := range collection1 {
		if collection2.containsWithBlueScore(outpoint, utxoEntry.blockBlueScore) {
			result.add(outpoint, utxoEntry)
		} else {
			remainder.add(outpoint, utxoEntry)
		}
	}
}

// substractionHavingBlueScore calculates a substraction between collection1 and collection2
// having same blue score, returns the result
func substractionHavingBlueScore(collection1, collection2 utxoCollection) (result utxoCollection) {
	result = make(utxoCollection, len(collection1))

	substractionHavingBlueScoreInPlace(collection1, collection2, result)
	return
}

// substractionHavingBlueScoreInPlace calculates a substraction between collection1 and collection2
// having same blue score, puts it into result
func substractionHavingBlueScoreInPlace(collection1, collection2, result utxoCollection) {
	for outpoint, utxoEntry := range collection1 {
		if !collection2.containsWithBlueScore(outpoint, utxoEntry.blockBlueScore) {
			result.add(outpoint, utxoEntry)
		}
	}
}

// substractionWithRemainderHavingBlueScore calculates a substraction between collection1 and collection2
// having same blue score, returns the result and the remainder from collection1
func substractionWithRemainderHavingBlueScore(collection1, collection2 utxoCollection) (result, remainder utxoCollection) {
	result = make(utxoCollection, len(collection1))
	remainder = make(utxoCollection, len(collection1))

	substractionWithRemainderHavingBlueScoreInPlace(collection1, collection2, result, remainder)
	return
}

// substractionWithRemainderHavingBlueScoreInPlace calculates a substraction between collection1 and collection2
// having same blue score, puts it into result and into remainder from collection1
func substractionWithRemainderHavingBlueScoreInPlace(collection1, collection2, result, remainder utxoCollection) {
	for outpoint, utxoEntry := range collection1 {
		if !collection2.containsWithBlueScore(outpoint, utxoEntry.blockBlueScore) {
			result.add(outpoint, utxoEntry)
		} else {
			remainder.add(outpoint, utxoEntry)
		}
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
	// Note that the following cases are not accounted for, as they are impossible
	// as long as the base utxoSet is the same:
	// - if utxoEntry is in d.toAdd and other.toRemove
	// - if utxoEntry is in d.toRemove and other.toAdd

	// check that NOT (entries with unequal blue scores AND utxoEntry is in d.toAdd and/or other.toRemove) -> Error
	isNotAddedOutputRemovedWithBlueScore := func(outpoint domainmessage.Outpoint, utxoEntry, diffEntry *UTXOEntry) bool {
		return !(diffEntry.blockBlueScore != utxoEntry.blockBlueScore &&
			(d.toAdd.containsWithBlueScore(outpoint, diffEntry.blockBlueScore) ||
				other.toRemove.containsWithBlueScore(outpoint, utxoEntry.blockBlueScore)))
	}

	if checkIntersectionWithRule(d.toRemove, other.toAdd, isNotAddedOutputRemovedWithBlueScore) {
		return nil, errors.New("diffFrom: outpoint both in d.toAdd and in other.toRemove")
	}

	//check that NOT (entries with unequal blue score AND utxoEntry is in d.toRemove and/or other.toAdd) -> Error
	isNotRemovedOutputAddedWithBlueScore := func(outpoint domainmessage.Outpoint, utxoEntry, diffEntry *UTXOEntry) bool {
		return !(diffEntry.blockBlueScore != utxoEntry.blockBlueScore &&
			(d.toRemove.containsWithBlueScore(outpoint, diffEntry.blockBlueScore) ||
				other.toAdd.containsWithBlueScore(outpoint, utxoEntry.blockBlueScore)))
	}

	if checkIntersectionWithRule(d.toAdd, other.toRemove, isNotRemovedOutputAddedWithBlueScore) {
		return nil, errors.New("diffFrom: outpoint both in d.toRemove and in other.toAdd")
	}

	// if have the same entry in d.toRemove and other.toRemove
	// and existing entry is with different blue score, in this case - this is an error
	if checkIntersectionWithRule(d.toRemove, other.toRemove,
		func(outpoint domainmessage.Outpoint, utxoEntry, diffEntry *UTXOEntry) bool {
			return utxoEntry.blockBlueScore != diffEntry.blockBlueScore
		}) {
		return nil, errors.New("diffFrom: outpoint both in d.toRemove and other.toRemove with different " +
			"blue scores, with no corresponding entry in d.toAdd")
	}

	result := UTXODiff{
		toAdd:    make(utxoCollection, len(d.toRemove)+len(other.toAdd)),
		toRemove: make(utxoCollection, len(d.toAdd)+len(other.toRemove)),
	}

	// All transactions in d.toAdd:
	// If they are not in other.toAdd - should be added in result.toRemove
	inBothToAdd := make(utxoCollection, len(d.toAdd))
	substractionWithRemainderHavingBlueScoreInPlace(d.toAdd, other.toAdd, result.toRemove, inBothToAdd)
	// If they are in other.toRemove - base utxoSet is not the same
	if checkIntersection(inBothToAdd, d.toRemove) != checkIntersection(inBothToAdd, other.toRemove) {
		return nil, errors.New(
			"diffFrom: outpoint both in d.toAdd, other.toAdd, and only one of d.toRemove and other.toRemove")
	}

	// All transactions in other.toRemove:
	// If they are not in d.toRemove - should be added in result.toRemove
	substractionHavingBlueScoreInPlace(other.toRemove, d.toRemove, result.toRemove)

	// All transactions in d.toRemove:
	// If they are not in other.toRemove - should be added in result.toAdd
	substractionHavingBlueScoreInPlace(d.toRemove, other.toRemove, result.toAdd)

	// All transactions in other.toAdd:
	// If they are not in d.toAdd - should be added in result.toAdd
	substractionHavingBlueScoreInPlace(other.toAdd, d.toAdd, result.toAdd)

	return &result, nil
}

// withDiffInPlace applies provided diff to this diff in-place, that would be the result if
// first d, and than diff were applied to the same base
func (d *UTXODiff) withDiffInPlace(diff *UTXODiff) error {
	if checkIntersectionWithRule(diff.toRemove, d.toRemove,
		func(outpoint domainmessage.Outpoint, entryToAdd, existingEntry *UTXOEntry) bool {
			return !d.toAdd.containsWithBlueScore(outpoint, entryToAdd.blockBlueScore)

		}) {
		return errors.New(
			"withDiffInPlace: outpoint both in d.toRemove and in diff.toRemove")
	}

	if checkIntersectionWithRule(diff.toAdd, d.toAdd,
		func(outpoint domainmessage.Outpoint, entryToAdd, existingEntry *UTXOEntry) bool {
			return !diff.toRemove.containsWithBlueScore(outpoint, existingEntry.blockBlueScore)
		}) {
		return errors.New(
			"withDiffInPlace: outpoint both in d.toAdd and in diff.toAdd")
	}

	intersection := make(utxoCollection, minInt(len(diff.toRemove), len(d.toAdd)))
	// If not exists neither in toAdd nor in toRemove - add to toRemove
	intersectionWithRemainderHavingBlueScoreInPlace(diff.toRemove, d.toAdd, intersection, d.toRemove)
	// If already exists in toAdd with the same blueScore - remove from toAdd
	d.toAdd.removeMultiple(intersection)

	intersection = make(utxoCollection, minInt(len(diff.toAdd), len(d.toRemove)))
	// If not exists neither in toAdd nor in toRemove, or exists in toRemove with different blueScore - add to toAdd
	intersectionWithRemainderHavingBlueScoreInPlace(diff.toAdd, d.toRemove, intersection, d.toAdd)
	// If already exists in toRemove with the same blueScore - remove from toRemove
	d.toRemove.removeMultiple(intersection)

	return nil
}

// WithDiff applies provided diff to this diff, creating a new utxoDiff, that would be the result if
// first d, and than diff were applied to some base
func (d *UTXODiff) WithDiff(diff *UTXODiff) (*UTXODiff, error) {
	clone := d.clone()

	err := clone.withDiffInPlace(diff)
	if err != nil {
		return nil, err
	}

	return clone, nil
}

// clone returns a clone of this utxoDiff
func (d *UTXODiff) clone() *UTXODiff {
	clone := &UTXODiff{
		toAdd:    d.toAdd.clone(),
		toRemove: d.toRemove.clone(),
	}
	return clone
}

// AddEntry adds a UTXOEntry to the diff
//
// If d.useMultiset is true, this function MUST be
// called with the DAG lock held.
func (d *UTXODiff) AddEntry(outpoint domainmessage.Outpoint, entry *UTXOEntry) error {
	if d.toRemove.containsWithBlueScore(outpoint, entry.blockBlueScore) {
		d.toRemove.remove(outpoint)
	} else if _, exists := d.toAdd[outpoint]; exists {
		return errors.Errorf("AddEntry: Cannot add outpoint %s twice", outpoint)
	} else {
		d.toAdd.add(outpoint, entry)
	}
	return nil
}

// RemoveEntry removes a UTXOEntry from the diff.
//
// If d.useMultiset is true, this function MUST be
// called with the DAG lock held.
func (d *UTXODiff) RemoveEntry(outpoint domainmessage.Outpoint, entry *UTXOEntry) error {
	if d.toAdd.containsWithBlueScore(outpoint, entry.blockBlueScore) {
		d.toAdd.remove(outpoint)
	} else if _, exists := d.toRemove[outpoint]; exists {
		return errors.Errorf("removeEntry: Cannot remove outpoint %s twice", outpoint)
	} else {
		d.toRemove.add(outpoint, entry)
	}
	return nil
}

func (d UTXODiff) String() string {
	return fmt.Sprintf("toAdd: %s; toRemove: %s", d.toAdd, d.toRemove)
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
	AddTx(tx *domainmessage.MsgTx, blockBlueScore uint64) (ok bool, err error)
	clone() UTXOSet
	Get(outpoint domainmessage.Outpoint) (*UTXOEntry, bool)
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

// newFullUTXOSetFromUTXOCollection converts a utxoCollection to a FullUTXOSet
func newFullUTXOSetFromUTXOCollection(collection utxoCollection) (*FullUTXOSet, error) {
	var err error
	multiset := secp256k1.NewMultiset()
	for outpoint, utxoEntry := range collection {
		multiset, err = addUTXOToMultiset(multiset, utxoEntry, &outpoint)
		if err != nil {
			return nil, err
		}
	}
	return &FullUTXOSet{
		utxoCollection: collection,
	}, nil
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

// AddTx adds a transaction to this utxoSet and returns isAccepted=true iff it's valid in this UTXO's context.
// It returns error if something unexpected happens, such as serialization error (isAccepted=false doesn't
// necessarily means there's an error).
//
// This function MUST be called with the DAG lock held.
func (fus *FullUTXOSet) AddTx(tx *domainmessage.MsgTx, blueScore uint64) (isAccepted bool, err error) {
	if !fus.containsInputs(tx) {
		return false, nil
	}

	for _, txIn := range tx.TxIn {
		fus.remove(txIn.PreviousOutpoint)
	}

	isCoinbase := tx.IsCoinBase()
	for i, txOut := range tx.TxOut {
		outpoint := *domainmessage.NewOutpoint(tx.TxID(), uint32(i))
		entry := NewUTXOEntry(txOut, isCoinbase, blueScore)
		fus.add(outpoint, entry)
	}

	return true, nil
}

func (fus *FullUTXOSet) containsInputs(tx *domainmessage.MsgTx) bool {
	for _, txIn := range tx.TxIn {
		outpoint := *domainmessage.NewOutpoint(&txIn.PreviousOutpoint.TxID, txIn.PreviousOutpoint.Index)
		if !fus.contains(outpoint) {
			return false
		}
	}

	return true
}

// clone returns a clone of this utxoSet
func (fus *FullUTXOSet) clone() UTXOSet {
	return &FullUTXOSet{utxoCollection: fus.utxoCollection.clone()}
}

// Get returns the UTXOEntry associated with the given Outpoint, and a boolean indicating if such entry was found
func (fus *FullUTXOSet) Get(outpoint domainmessage.Outpoint) (*UTXOEntry, bool) {
	utxoEntry, ok := fus.utxoCollection[outpoint]
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

// AddTx adds a transaction to this utxoSet and returns true iff it's valid in this UTXO's context.
//
// If dus.UTXODiff.useMultiset is true, this function MUST be
// called with the DAG lock held.
func (dus *DiffUTXOSet) AddTx(tx *domainmessage.MsgTx, blockBlueScore uint64) (bool, error) {
	if !dus.containsInputs(tx) {
		return false, nil
	}

	err := dus.appendTx(tx, blockBlueScore)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (dus *DiffUTXOSet) appendTx(tx *domainmessage.MsgTx, blockBlueScore uint64) error {
	for _, txIn := range tx.TxIn {
		entry, ok := dus.Get(txIn.PreviousOutpoint)
		if !ok {
			return errors.Errorf("couldn't find entry for outpoint %s", txIn.PreviousOutpoint)
		}
		err := dus.UTXODiff.RemoveEntry(txIn.PreviousOutpoint, entry)
		if err != nil {
			return err
		}
	}

	isCoinbase := tx.IsCoinBase()
	for i, txOut := range tx.TxOut {
		outpoint := *domainmessage.NewOutpoint(tx.TxID(), uint32(i))
		entry := NewUTXOEntry(txOut, isCoinbase, blockBlueScore)

		err := dus.UTXODiff.AddEntry(outpoint, entry)
		if err != nil {
			return err
		}
	}
	return nil
}

func (dus *DiffUTXOSet) containsInputs(tx *domainmessage.MsgTx) bool {
	for _, txIn := range tx.TxIn {
		outpoint := *domainmessage.NewOutpoint(&txIn.PreviousOutpoint.TxID, txIn.PreviousOutpoint.Index)
		isInBase := dus.base.contains(outpoint)
		isInDiffToAdd := dus.UTXODiff.toAdd.contains(outpoint)
		isInDiffToRemove := dus.UTXODiff.toRemove.contains(outpoint)
		if (!isInBase && !isInDiffToAdd) || (isInDiffToRemove && !(isInBase && isInDiffToAdd)) {
			return false
		}
	}

	return true
}

// meldToBase updates the base fullUTXOSet with all changes in diff
func (dus *DiffUTXOSet) meldToBase() error {
	for outpoint := range dus.UTXODiff.toRemove {
		if _, ok := dus.base.Get(outpoint); ok {
			dus.base.remove(outpoint)
		} else {
			return errors.Errorf("Couldn't remove outpoint %s because it doesn't exist in the DiffUTXOSet base", outpoint)
		}
	}

	for outpoint, utxoEntry := range dus.UTXODiff.toAdd {
		dus.base.add(outpoint, utxoEntry)
	}
	dus.UTXODiff = NewUTXODiff()
	return nil
}

func (dus *DiffUTXOSet) String() string {
	return fmt.Sprintf("{Base: %s, To Add: %s, To Remove: %s}", dus.base, dus.UTXODiff.toAdd, dus.UTXODiff.toRemove)
}

// clone returns a clone of this UTXO Set
func (dus *DiffUTXOSet) clone() UTXOSet {
	return NewDiffUTXOSet(dus.base.clone().(*FullUTXOSet), dus.UTXODiff.clone())
}

// cloneWithoutBase returns a *DiffUTXOSet with same
// base as this *DiffUTXOSet and a cloned diff.
func (dus *DiffUTXOSet) cloneWithoutBase() UTXOSet {
	return NewDiffUTXOSet(dus.base, dus.UTXODiff.clone())
}

// Get returns the UTXOEntry associated with provided outpoint in this UTXOSet.
// Returns false in second output if this UTXOEntry was not found
func (dus *DiffUTXOSet) Get(outpoint domainmessage.Outpoint) (*UTXOEntry, bool) {
	if toRemoveEntry, ok := dus.UTXODiff.toRemove.get(outpoint); ok {
		// An exception is made for entries with unequal blue scores
		// These are just "updates" to accepted blue score
		if toAddEntry, ok := dus.UTXODiff.toAdd.get(outpoint); ok && toAddEntry.blockBlueScore != toRemoveEntry.blockBlueScore {
			return toAddEntry, true
		}
		return nil, false
	}
	if txOut, ok := dus.base.get(outpoint); ok {
		return txOut, true
	}
	txOut, ok := dus.UTXODiff.toAdd.get(outpoint)
	return txOut, ok
}
