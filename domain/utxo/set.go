package utxo

import (
	"bytes"
	"fmt"
	"math"
	"sort"
	"strings"
	"unsafe"

	"github.com/kaspanet/kaspad/infrastructure/db/dbaccess"

	"github.com/pkg/errors"

	"github.com/kaspanet/kaspad/app/appmessage"
)

const (
	// UnacceptedBlueScore is the blue score used for the "block" blueScore
	// field of the contextual transaction information provided in a
	// transaction store when it has not yet been accepted by a block.
	UnacceptedBlueScore uint64 = math.MaxUint64
)

// Entry houses details about an individual transaction output in a utxo
// set such as whether or not it was contained in a coinbase tx, the blue
// score of the block that accepts the tx, its public key script, and how
// much it pays.
type Entry struct {
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
	// since it was Loaded. This approach is used in order to reduce memory
	// usage since there will be a lot of these in memory.
	packedFlags txoFlags
}

// IsCoinbase returns whether or not the output was contained in a block
// reward transaction.
func (entry *Entry) IsCoinbase() bool {
	return entry.packedFlags&tfCoinbase == tfCoinbase
}

// BlockBlueScore returns the blue score of the block accepting the output.
func (entry *Entry) BlockBlueScore() uint64 {
	return entry.blockBlueScore
}

// Amount returns the amount of the output.
func (entry *Entry) Amount() uint64 {
	return entry.amount
}

// ScriptPubKey returns the public key script for the output.
func (entry *Entry) ScriptPubKey() []byte {
	return entry.scriptPubKey
}

// IsUnaccepted returns true iff this Entry has been included in a block
// but has not yet been accepted by any block.
func (entry *Entry) IsUnaccepted() bool {
	return entry.blockBlueScore == UnacceptedBlueScore
}

// txoFlags is a bitmask defining additional information and state for a
// transaction output in a UTXO set.
type txoFlags uint8

const (
	// tfCoinbase indicates that a txout was contained in a coinbase tx.
	tfCoinbase txoFlags = 1 << iota
)

// NewEntry creates a new utxoEntry representing the given txOut
func NewEntry(txOut *appmessage.TxOut, isCoinbase bool, blockBlueScore uint64) *Entry {
	entry := &Entry{
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
type utxoCollection map[appmessage.Outpoint]*Entry

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

// Add adds a new UTXO entry to this collection
func (uc utxoCollection) Add(outpoint appmessage.Outpoint, entry *Entry) {
	uc[outpoint] = entry
}

// addMultiple adds multiple UTXO entries to this collection
func (uc utxoCollection) addMultiple(collectionToAdd utxoCollection) {
	for outpoint, entry := range collectionToAdd {
		uc[outpoint] = entry
	}
}

// remove removes a UTXO entry from this collection if it exists
func (uc utxoCollection) remove(outpoint appmessage.Outpoint) {
	delete(uc, outpoint)
}

// removeMultiple removes multiple UTXO entries from this collection if it exists
func (uc utxoCollection) removeMultiple(collectionToRemove utxoCollection) {
	for outpoint := range collectionToRemove {
		delete(uc, outpoint)
	}
}

// get returns the Entry represented by provided outpoint,
// and a boolean value indicating if said Entry is in the set or not
func (uc utxoCollection) get(outpoint appmessage.Outpoint) (*Entry, bool) {
	entry, ok := uc[outpoint]
	return entry, ok
}

// contains returns a boolean value indicating whether a UTXO entry is in the set
func (uc utxoCollection) contains(outpoint appmessage.Outpoint) bool {
	_, ok := uc[outpoint]
	return ok
}

// containsWithBlueScore returns a boolean value indicating whether a Entry
// is in the set and its blue score is equal to the given blue score.
func (uc utxoCollection) containsWithBlueScore(outpoint appmessage.Outpoint, blueScore uint64) bool {
	entry, ok := uc.get(outpoint)
	return ok && entry.blockBlueScore == blueScore
}

// clone returns a clone of this collection
func (uc utxoCollection) clone() utxoCollection {
	clone := make(utxoCollection, len(uc))
	for outpoint, entry := range uc {
		clone.Add(outpoint, entry)
	}

	return clone
}

// Diff represents a Diff between two UTXO Sets.
type Diff struct {
	ToAdd    utxoCollection
	ToRemove utxoCollection
}

// NewDiff creates a new, empty utxoDiff
// without a multiset.
func NewDiff() *Diff {
	return &Diff{
		ToAdd:    utxoCollection{},
		ToRemove: utxoCollection{},
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
func checkIntersectionWithRule(collection1 utxoCollection, collection2 utxoCollection, extraRule func(appmessage.Outpoint, *Entry, *Entry) bool) bool {
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

// intersectionWithRemainderHavingBlueScore calculates an intersection between two utxoCollections
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
			result.Add(outpoint, utxoEntry)
		} else {
			remainder.Add(outpoint, utxoEntry)
		}
	}
}

// subtractionHavingBlueScore calculates a subtraction between collection1 and collection2
// having same blue score, returns the result
func subtractionHavingBlueScore(collection1, collection2 utxoCollection) (result utxoCollection) {
	result = make(utxoCollection, len(collection1))

	subtractionHavingBlueScoreInPlace(collection1, collection2, result)
	return
}

// subtractionHavingBlueScoreInPlace calculates a subtraction between collection1 and collection2
// having same blue score, puts it into result
func subtractionHavingBlueScoreInPlace(collection1, collection2, result utxoCollection) {
	for outpoint, utxoEntry := range collection1 {
		if !collection2.containsWithBlueScore(outpoint, utxoEntry.blockBlueScore) {
			result.Add(outpoint, utxoEntry)
		}
	}
}

// subtractionWithRemainderHavingBlueScore calculates a subtraction between collection1 and collection2
// having same blue score, returns the result and the remainder from collection1
func subtractionWithRemainderHavingBlueScore(collection1, collection2 utxoCollection) (result, remainder utxoCollection) {
	result = make(utxoCollection, len(collection1))
	remainder = make(utxoCollection, len(collection1))

	subtractionWithRemainderHavingBlueScoreInPlace(collection1, collection2, result, remainder)
	return
}

// subtractionWithRemainderHavingBlueScoreInPlace calculates a subtraction between collection1 and collection2
// having same blue score, puts it into result and into remainder from collection1
func subtractionWithRemainderHavingBlueScoreInPlace(collection1, collection2, result, remainder utxoCollection) {
	for outpoint, utxoEntry := range collection1 {
		if !collection2.containsWithBlueScore(outpoint, utxoEntry.blockBlueScore) {
			result.Add(outpoint, utxoEntry)
		} else {
			remainder.Add(outpoint, utxoEntry)
		}
	}
}

// diffFrom returns a new utxoDiff with the difference between this utxoDiff and another
// Assumes that:
// Both utxoDiffs are from the same Base
// If a txOut exists in both utxoDiffs, its underlying values would be the same
//
// diffFrom follows a set of rules represented by the following 3 by 3 table:
//
//          |           | this      |           |
// ---------+-----------+-----------+-----------+-----------
//          |           | ToAdd     | ToRemove  | None
// ---------+-----------+-----------+-----------+-----------
// other    | ToAdd     | -         | X         | ToAdd
// ---------+-----------+-----------+-----------+-----------
//          | ToRemove  | X         | -         | ToRemove
// ---------+-----------+-----------+-----------+-----------
//          | None      | ToRemove  | ToAdd     | -
//
// Key:
// -		Don't Add anything to the result
// X		Return an error
// ToAdd	Add the UTXO into the ToAdd collection of the result
// ToRemove	Add the UTXO into the ToRemove collection of the result
//
// Examples:
// 1. This Diff contains a UTXO in ToAdd, and the other Diff contains it in ToRemove
//    diffFrom results in an error
// 2. This Diff contains a UTXO in ToRemove, and the other Diff does not contain it
//    diffFrom results in the UTXO being added to ToAdd
func (d *Diff) diffFrom(other *Diff) (*Diff, error) {
	// Note that the following cases are not accounted for, as they are impossible
	// as long as the Base utxoSet is the same:
	// - if utxoEntry is in d.ToAdd and other.ToRemove
	// - if utxoEntry is in d.ToRemove and other.ToAdd

	// check that NOT (entries with unequal blue scores AND utxoEntry is in d.ToAdd and/or other.ToRemove) -> Error
	isNotAddedOutputRemovedWithBlueScore := func(outpoint appmessage.Outpoint, utxoEntry, diffEntry *Entry) bool {
		return !(diffEntry.blockBlueScore != utxoEntry.blockBlueScore &&
			(d.ToAdd.containsWithBlueScore(outpoint, diffEntry.blockBlueScore) ||
				other.ToRemove.containsWithBlueScore(outpoint, utxoEntry.blockBlueScore)))
	}

	if checkIntersectionWithRule(d.ToRemove, other.ToAdd, isNotAddedOutputRemovedWithBlueScore) {
		return nil, errors.New("DiffFrom: outpoint both in d.ToAdd and in other.ToRemove")
	}

	//check that NOT (entries with unequal blue score AND utxoEntry is in d.ToRemove and/or other.ToAdd) -> Error
	isNotRemovedOutputAddedWithBlueScore := func(outpoint appmessage.Outpoint, utxoEntry, diffEntry *Entry) bool {
		return !(diffEntry.blockBlueScore != utxoEntry.blockBlueScore &&
			(d.ToRemove.containsWithBlueScore(outpoint, diffEntry.blockBlueScore) ||
				other.ToAdd.containsWithBlueScore(outpoint, utxoEntry.blockBlueScore)))
	}

	if checkIntersectionWithRule(d.ToAdd, other.ToRemove, isNotRemovedOutputAddedWithBlueScore) {
		return nil, errors.New("DiffFrom: outpoint both in d.ToRemove and in other.ToAdd")
	}

	// if have the same entry in d.ToRemove and other.ToRemove
	// and existing entry is with different blue score, in this case - this is an error
	if checkIntersectionWithRule(d.ToRemove, other.ToRemove,
		func(outpoint appmessage.Outpoint, utxoEntry, diffEntry *Entry) bool {
			return utxoEntry.blockBlueScore != diffEntry.blockBlueScore
		}) {
		return nil, errors.New("DiffFrom: outpoint both in d.ToRemove and other.ToRemove with different " +
			"blue scores, with no corresponding entry in d.ToAdd")
	}

	result := Diff{
		ToAdd:    make(utxoCollection, len(d.ToRemove)+len(other.ToAdd)),
		ToRemove: make(utxoCollection, len(d.ToAdd)+len(other.ToRemove)),
	}

	// All transactions in d.ToAdd:
	// If they are not in other.ToAdd - should be added in result.ToRemove
	inBothToAdd := make(utxoCollection, len(d.ToAdd))
	subtractionWithRemainderHavingBlueScoreInPlace(d.ToAdd, other.ToAdd, result.ToRemove, inBothToAdd)
	// If they are in other.ToRemove - Base utxoSet is not the same
	if checkIntersection(inBothToAdd, d.ToRemove) != checkIntersection(inBothToAdd, other.ToRemove) {
		return nil, errors.New(
			"DiffFrom: outpoint both in d.ToAdd, other.ToAdd, and only one of d.ToRemove and other.ToRemove")
	}

	// All transactions in other.ToRemove:
	// If they are not in d.ToRemove - should be added in result.ToRemove
	subtractionHavingBlueScoreInPlace(other.ToRemove, d.ToRemove, result.ToRemove)

	// All transactions in d.ToRemove:
	// If they are not in other.ToRemove - should be added in result.ToAdd
	subtractionHavingBlueScoreInPlace(d.ToRemove, other.ToRemove, result.ToAdd)

	// All transactions in other.ToAdd:
	// If they are not in d.ToAdd - should be added in result.ToAdd
	subtractionHavingBlueScoreInPlace(other.ToAdd, d.ToAdd, result.ToAdd)

	return &result, nil
}

// WithDiffInPlace applies provided Diff to this Diff in-place, that would be the result if
// first d, and than Diff were applied to the same Base
func (d *Diff) WithDiffInPlace(diff *Diff) error {
	if checkIntersectionWithRule(diff.ToRemove, d.ToRemove,
		func(outpoint appmessage.Outpoint, entryToAdd, existingEntry *Entry) bool {
			return !d.ToAdd.containsWithBlueScore(outpoint, entryToAdd.blockBlueScore)

		}) {
		return errors.New(
			"WithDiffInPlace: outpoint both in d.ToRemove and in Diff.ToRemove")
	}

	if checkIntersectionWithRule(diff.ToAdd, d.ToAdd,
		func(outpoint appmessage.Outpoint, entryToAdd, existingEntry *Entry) bool {
			return !diff.ToRemove.containsWithBlueScore(outpoint, existingEntry.blockBlueScore)
		}) {
		return errors.New(
			"WithDiffInPlace: outpoint both in d.ToAdd and in Diff.ToAdd")
	}

	intersection := make(utxoCollection, minInt(len(diff.ToRemove), len(d.ToAdd)))
	// If not exists neither in ToAdd nor in ToRemove - Add to ToRemove
	intersectionWithRemainderHavingBlueScoreInPlace(diff.ToRemove, d.ToAdd, intersection, d.ToRemove)
	// If already exists in ToAdd with the same blueScore - remove from ToAdd
	d.ToAdd.removeMultiple(intersection)

	intersection = make(utxoCollection, minInt(len(diff.ToAdd), len(d.ToRemove)))
	// If not exists neither in ToAdd nor in ToRemove, or exists in ToRemove with different blueScore - Add to ToAdd
	intersectionWithRemainderHavingBlueScoreInPlace(diff.ToAdd, d.ToRemove, intersection, d.ToAdd)
	// If already exists in ToRemove with the same blueScore - remove from ToRemove
	d.ToRemove.removeMultiple(intersection)

	return nil
}

// WithDiff applies provided Diff to this Diff, creating a new utxoDiff, that would be the result if
// first d, and than Diff were applied to some Base
func (d *Diff) WithDiff(diff *Diff) (*Diff, error) {
	clone := d.Clone()

	err := clone.WithDiffInPlace(diff)
	if err != nil {
		return nil, err
	}

	return clone, nil
}

// Clone returns a Clone of this utxoDiff
func (d *Diff) Clone() *Diff {
	clone := &Diff{
		ToAdd:    d.ToAdd.clone(),
		ToRemove: d.ToRemove.clone(),
	}
	return clone
}

// AddEntry adds a Entry to the Diff
//
// If d.useMultiset is true, this function MUST be
// called with the DAG lock held.
func (d *Diff) AddEntry(outpoint appmessage.Outpoint, entry *Entry) error {
	if d.ToRemove.containsWithBlueScore(outpoint, entry.blockBlueScore) {
		d.ToRemove.remove(outpoint)
	} else if _, exists := d.ToAdd[outpoint]; exists {
		return errors.Errorf("AddEntry: Cannot Add outpoint %s twice", outpoint)
	} else {
		d.ToAdd.Add(outpoint, entry)
	}
	return nil
}

// RemoveEntry removes a Entry from the Diff.
//
// If d.useMultiset is true, this function MUST be
// called with the DAG lock held.
func (d *Diff) RemoveEntry(outpoint appmessage.Outpoint, entry *Entry) error {
	if d.ToAdd.containsWithBlueScore(outpoint, entry.blockBlueScore) {
		d.ToAdd.remove(outpoint)
	} else if _, exists := d.ToRemove[outpoint]; exists {
		return errors.Errorf("removeEntry: Cannot remove outpoint %s twice", outpoint)
	} else {
		d.ToRemove.Add(outpoint, entry)
	}
	return nil
}

func (d Diff) String() string {
	return fmt.Sprintf("ToAdd: %s; ToRemove: %s", d.ToAdd, d.ToRemove)
}

// Set represents a set of unspent transaction outputs
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
type Set interface {
	fmt.Stringer
	DiffFrom(other Set) (*Diff, error)
	WithDiff(utxoDiff *Diff) (Set, error)
	AddTx(tx *appmessage.MsgTx, blockBlueScore uint64) (ok bool, err error)
	Clone() Set
	Get(outpoint appmessage.Outpoint) (*Entry, bool)
}

// FullUTXOSet represents a full list of transaction outputs and their values
type FullUTXOSet struct {
	UTXOCache        utxoCollection
	dbContext        dbaccess.Context
	estimatedSize    uint64
	maxUTXOCacheSize uint64
	outpointBuff     *bytes.Buffer
}

// NewFullUTXOSet creates a new utxoSet with full list of transaction outputs and their values
func NewFullUTXOSet() *FullUTXOSet {
	return &FullUTXOSet{
		UTXOCache: utxoCollection{},
	}
}

// NewFullUTXOSetFromContext creates a new utxoSet and map the data context with caching
func NewFullUTXOSetFromContext(context dbaccess.Context, cacheSize uint64) *FullUTXOSet {
	return &FullUTXOSet{
		dbContext:        context,
		maxUTXOCacheSize: cacheSize,
		UTXOCache:        make(utxoCollection),
	}
}

// DiffFrom returns the difference between this utxoSet and another
// DiffFrom can only work when other is a diffUTXOSet, and its Base utxoSet is this.
func (fus *FullUTXOSet) DiffFrom(other Set) (*Diff, error) {
	otherDiffSet, ok := other.(*DiffUTXOSet)
	if !ok {
		return nil, errors.New("can't DiffFrom two fullUTXOSets")
	}

	if otherDiffSet.Base != fus {
		return nil, errors.New("can DiffFrom only with diffUTXOSet where this fullUTXOSet is the Base")
	}

	return otherDiffSet.UTXODiff, nil
}

// WithDiff returns a utxoSet which is a Diff between this and another utxoSet
func (fus *FullUTXOSet) WithDiff(other *Diff) (Set, error) {
	return NewDiffUTXOSet(fus, other.Clone()), nil
}

// AddTx adds a transaction to this utxoSet and returns isAccepted=true iff it's valid in this UTXO's context.
// It returns error if something unexpected happens, such as serialization error (isAccepted=false doesn't
// necessarily means there's an error).
//
// This function MUST be called with the DAG lock held.
func (fus *FullUTXOSet) AddTx(tx *appmessage.MsgTx, blueScore uint64) (isAccepted bool, err error) {
	if !fus.containsInputs(tx) {
		return false, nil
	}

	for _, txIn := range tx.TxIn {
		fus.remove(txIn.PreviousOutpoint)
	}

	isCoinbase := tx.IsCoinBase()
	for i, txOut := range tx.TxOut {
		outpoint := *appmessage.NewOutpoint(tx.TxID(), uint32(i))
		entry := NewEntry(txOut, isCoinbase, blueScore)
		fus.Add(outpoint, entry)
	}

	return true, nil
}

func (fus *FullUTXOSet) containsInputs(tx *appmessage.MsgTx) bool {
	for _, txIn := range tx.TxIn {
		outpoint := *appmessage.NewOutpoint(&txIn.PreviousOutpoint.TxID, txIn.PreviousOutpoint.Index)
		if !fus.contains(outpoint) {
			return false
		}
	}

	return true
}

// contains returns a boolean value indicating whether a UTXO entry is in the set
func (fus *FullUTXOSet) contains(outpoint appmessage.Outpoint) bool {
	_, ok := fus.Get(outpoint)
	return ok
}

// Clone returns a Clone of this utxoSet
func (fus *FullUTXOSet) Clone() Set {
	return &FullUTXOSet{
		UTXOCache:        fus.UTXOCache.clone(),
		dbContext:        fus.dbContext,
		estimatedSize:    fus.estimatedSize,
		maxUTXOCacheSize: fus.maxUTXOCacheSize,
	}
}

// get returns the Entry associated with the given Outpoint, and a boolean indicating if such entry was found
func (fus *FullUTXOSet) get(outpoint appmessage.Outpoint) (*Entry, bool) {
	return fus.Get(outpoint)
}

// getSizeOfUTXOEntryAndOutpoint returns estimated size of Entry & Outpoint in bytes
func getSizeOfUTXOEntryAndOutpoint(entry *Entry) uint64 {
	const staticSize = uint64(unsafe.Sizeof(Entry{}) + unsafe.Sizeof(appmessage.Outpoint{}))
	return staticSize + uint64(len(entry.scriptPubKey))
}

// checkAndCleanCachedData checks the FullUTXOSet estimated size and clean it if it reaches the limit
func (fus *FullUTXOSet) checkAndCleanCachedData() {
	if fus.estimatedSize > fus.maxUTXOCacheSize {
		fus.UTXOCache = make(utxoCollection)
		fus.estimatedSize = 0
	}
}

// Add adds a new UTXO entry to this FullUTXOSet
func (fus *FullUTXOSet) Add(outpoint appmessage.Outpoint, entry *Entry) {
	fus.UTXOCache[outpoint] = entry
	fus.estimatedSize += getSizeOfUTXOEntryAndOutpoint(entry)
	fus.checkAndCleanCachedData()
}

// remove removes a UTXO entry from this collection if it exists
func (fus *FullUTXOSet) remove(outpoint appmessage.Outpoint) {
	entry, ok := fus.UTXOCache.get(outpoint)
	if ok {
		delete(fus.UTXOCache, outpoint)
		fus.estimatedSize -= getSizeOfUTXOEntryAndOutpoint(entry)
	}
}

// Get returns the Entry associated with the given Outpoint, and a boolean indicating if such entry was found
// If the Entry doesn't not exist in the memory then check in the database
func (fus *FullUTXOSet) Get(outpoint appmessage.Outpoint) (*Entry, bool) {
	utxoEntry, ok := fus.UTXOCache[outpoint]
	if ok {
		return utxoEntry, ok
	}

	if fus.outpointBuff == nil {
		fus.outpointBuff = bytes.NewBuffer(make([]byte, outpointSerializeSize))
	}

	fus.outpointBuff.Reset()
	err := SerializeOutpoint(fus.outpointBuff, &outpoint)
	if err != nil {
		return nil, false
	}

	key := fus.outpointBuff.Bytes()
	value, err := dbaccess.GetFromUTXOSet(fus.dbContext, key)

	if err != nil {
		return nil, false
	}

	entry, err := DeserializeUTXOEntry(bytes.NewReader(value))
	if err != nil {
		return nil, false
	}

	fus.Add(outpoint, entry)
	return entry, true
}

func (fus *FullUTXOSet) String() string {
	return fus.UTXOCache.String()
}

// DiffUTXOSet represents a utxoSet with a Base fullUTXOSet and a Diff
type DiffUTXOSet struct {
	Base     *FullUTXOSet
	UTXODiff *Diff
}

// NewDiffUTXOSet Creates a new utxoSet based on a Base fullUTXOSet and a Diff
func NewDiffUTXOSet(base *FullUTXOSet, diff *Diff) *DiffUTXOSet {
	return &DiffUTXOSet{
		Base:     base,
		UTXODiff: diff,
	}
}

// DiffFrom returns the difference between this utxoSet and another.
// DiffFrom can work if other is this's Base fullUTXOSet, or a diffUTXOSet with the same Base as this
func (dus *DiffUTXOSet) DiffFrom(other Set) (*Diff, error) {
	otherDiffSet, ok := other.(*DiffUTXOSet)
	if !ok {
		return nil, errors.New("can't DiffFrom diffUTXOSet with fullUTXOSet")
	}

	if otherDiffSet.Base != dus.Base {
		return nil, errors.New("can't DiffFrom with another diffUTXOSet with a different Base")
	}

	return dus.UTXODiff.diffFrom(otherDiffSet.UTXODiff)
}

// WithDiff return a new utxoSet which is a DiffFrom between this and another utxoSet
func (dus *DiffUTXOSet) WithDiff(other *Diff) (Set, error) {
	diff, err := dus.UTXODiff.WithDiff(other)
	if err != nil {
		return nil, err
	}

	return NewDiffUTXOSet(dus.Base, diff), nil
}

// AddTx adds a transaction to this utxoSet and returns true iff it's valid in this UTXO's context.
//
// If dus.Diff.useMultiset is true, this function MUST be
// called with the DAG lock held.
func (dus *DiffUTXOSet) AddTx(tx *appmessage.MsgTx, blockBlueScore uint64) (bool, error) {
	if !dus.containsInputs(tx) {
		return false, nil
	}

	err := dus.appendTx(tx, blockBlueScore)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (dus *DiffUTXOSet) appendTx(tx *appmessage.MsgTx, blockBlueScore uint64) error {
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
		outpoint := *appmessage.NewOutpoint(tx.TxID(), uint32(i))
		entry := NewEntry(txOut, isCoinbase, blockBlueScore)

		err := dus.UTXODiff.AddEntry(outpoint, entry)
		if err != nil {
			return err
		}
	}
	return nil
}

func (dus *DiffUTXOSet) containsInputs(tx *appmessage.MsgTx) bool {
	for _, txIn := range tx.TxIn {
		outpoint := *appmessage.NewOutpoint(&txIn.PreviousOutpoint.TxID, txIn.PreviousOutpoint.Index)
		isInBase := dus.Base.contains(outpoint)
		isInDiffToAdd := dus.UTXODiff.ToAdd.contains(outpoint)
		isInDiffToRemove := dus.UTXODiff.ToRemove.contains(outpoint)
		if (!isInBase && !isInDiffToAdd) || (isInDiffToRemove && !(isInBase && isInDiffToAdd)) {
			return false
		}
	}

	return true
}

// MeldToBase updates the Base fullUTXOSet with all changes in Diff
func (dus *DiffUTXOSet) MeldToBase() error {
	for outpoint := range dus.UTXODiff.ToRemove {
		if _, ok := dus.Base.Get(outpoint); ok {
			dus.Base.remove(outpoint)
		} else {
			return errors.Errorf("Couldn't remove outpoint %s because it doesn't exist in the DiffUTXOSet Base", outpoint)
		}
	}

	for outpoint, utxoEntry := range dus.UTXODiff.ToAdd {
		dus.Base.Add(outpoint, utxoEntry)
	}
	dus.UTXODiff = NewDiff()
	return nil
}

func (dus *DiffUTXOSet) String() string {
	return fmt.Sprintf("{Base: %s, To Add: %s, To Remove: %s}", dus.Base, dus.UTXODiff.ToAdd, dus.UTXODiff.ToRemove)
}

// Clone returns a Clone of this UTXO Set
func (dus *DiffUTXOSet) Clone() Set {
	return NewDiffUTXOSet(dus.Base.Clone().(*FullUTXOSet), dus.UTXODiff.Clone())
}

// CloneWithoutBase returns a *DiffUTXOSet with same
// Base as this *DiffUTXOSet and a cloned Diff.
func (dus *DiffUTXOSet) CloneWithoutBase() Set {
	return NewDiffUTXOSet(dus.Base, dus.UTXODiff.Clone())
}

// Get returns the Entry associated with provided outpoint in this Set.
// Returns false in second output if this Entry was not found
func (dus *DiffUTXOSet) Get(outpoint appmessage.Outpoint) (*Entry, bool) {
	if toRemoveEntry, ok := dus.UTXODiff.ToRemove.get(outpoint); ok {
		// An exception is made for entries with unequal blue scores
		// These are just "updates" to accepted blue score
		if toAddEntry, ok := dus.UTXODiff.ToAdd.get(outpoint); ok && toAddEntry.blockBlueScore != toRemoveEntry.blockBlueScore {
			return toAddEntry, true
		}
		return nil, false
	}
	if txOut, ok := dus.Base.get(outpoint); ok {
		return txOut, true
	}
	txOut, ok := dus.UTXODiff.ToAdd.get(outpoint)
	return txOut, ok
}
