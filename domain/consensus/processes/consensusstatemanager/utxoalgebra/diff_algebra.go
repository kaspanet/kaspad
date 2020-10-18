package utxoalgebra

import (
	"fmt"

	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/pkg/errors"
)

// checkIntersection checks if there is an intersection between two model.UTXOCollections
func checkIntersection(collection1 model.UTXOCollection, collection2 model.UTXOCollection) bool {
	for outpoint := range collection1 {
		if collectionContains(collection2, outpoint) {
			return true
		}
	}

	return false
}

// checkIntersectionWithRule checks if there is an intersection between two model.UTXOCollections satisfying arbitrary rule
func checkIntersectionWithRule(collection1 model.UTXOCollection, collection2 model.UTXOCollection, extraRule func(model.DomainOutpoint, *model.UTXOEntry, *model.UTXOEntry) bool) bool {
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

// intersectionWithRemainderHavingBlueScore calculates an intersection between two model.UTXOCollections
// having same blue score, returns the result and the remainder from collection1
func intersectionWithRemainderHavingBlueScore(collection1, collection2 model.UTXOCollection) (result, remainder model.UTXOCollection) {
	result = make(model.UTXOCollection, minInt(len(collection1), len(collection2)))
	remainder = make(model.UTXOCollection, len(collection1))
	intersectionWithRemainderHavingBlueScoreInPlace(collection1, collection2, result, remainder)
	return
}

// intersectionWithRemainderHavingBlueScoreInPlace calculates an intersection between two model.UTXOCollections
// having same blue score, puts it into result and into remainder from collection1
func intersectionWithRemainderHavingBlueScoreInPlace(collection1, collection2, result, remainder model.UTXOCollection) {
	for outpoint, utxoEntry := range collection1 {
		if collection2.containsWithBlueScore(outpoint, utxoEntry.blockBlueScore) {
			result.add(outpoint, utxoEntry)
		} else {
			remainder.add(outpoint, utxoEntry)
		}
	}
}

// subtractionHavingBlueScore calculates a subtraction between collection1 and collection2
// having same blue score, returns the result
func subtractionHavingBlueScore(collection1, collection2 model.UTXOCollection) (result model.UTXOCollection) {
	result = make(model.UTXOCollection, len(collection1))

	subtractionHavingBlueScoreInPlace(collection1, collection2, result)
	return
}

// subtractionHavingBlueScoreInPlace calculates a subtraction between collection1 and collection2
// having same blue score, puts it into result
func subtractionHavingBlueScoreInPlace(collection1, collection2, result model.UTXOCollection) {
	for outpoint, utxoEntry := range collection1 {
		if !collection2.containsWithBlueScore(outpoint, utxoEntry.blockBlueScore) {
			result.add(outpoint, utxoEntry)
		}
	}
}

// subtractionWithRemainderHavingBlueScore calculates a subtraction between collection1 and collection2
// having same blue score, returns the result and the remainder from collection1
func subtractionWithRemainderHavingBlueScore(collection1, collection2 model.UTXOCollection) (result, remainder model.UTXOCollection) {
	result = make(model.UTXOCollection, len(collection1))
	remainder = make(model.UTXOCollection, len(collection1))

	subtractionWithRemainderHavingBlueScoreInPlace(collection1, collection2, result, remainder)
	return
}

// subtractionWithRemainderHavingBlueScoreInPlace calculates a subtraction between collection1 and collection2
// having same blue score, puts it into result and into remainder from collection1
func subtractionWithRemainderHavingBlueScoreInPlace(collection1, collection2, result, remainder model.UTXOCollection) {
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
func (d *model.UTXODiff) diffFrom(other *model.UTXODiff) (*model.UTXODiff, error) {
	// Note that the following cases are not accounted for, as they are impossible
	// as long as the base utxoSet is the same:
	// - if utxoEntry is in d.toAdd and other.toRemove
	// - if utxoEntry is in d.toRemove and other.toAdd

	// check that NOT (entries with unequal blue scores AND utxoEntry is in d.toAdd and/or other.toRemove) -> Error
	isNotAddedOutputRemovedWithBlueScore := func(outpoint model.DomainOutpoint, utxoEntry, diffEntry *model.UTXOEntry) bool {
		return !(diffEntry.blockBlueScore != utxoEntry.blockBlueScore &&
			(d.toAdd.containsWithBlueScore(outpoint, diffEntry.blockBlueScore) ||
				other.toRemove.containsWithBlueScore(outpoint, utxoEntry.blockBlueScore)))
	}

	if checkIntersectionWithRule(d.toRemove, other.toAdd, isNotAddedOutputRemovedWithBlueScore) {
		return nil, errors.New("diffFrom: outpoint both in d.toAdd and in other.toRemove")
	}

	//check that NOT (entries with unequal blue score AND utxoEntry is in d.toRemove and/or other.toAdd) -> Error
	isNotRemovedOutputAddedWithBlueScore := func(outpoint model.DomainOutpoint, utxoEntry, diffEntry *model.UTXOEntry) bool {
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
		func(outpoint model.DomainOutpoint, utxoEntry, diffEntry *model.UTXOEntry) bool {
			return utxoEntry.blockBlueScore != diffEntry.blockBlueScore
		}) {
		return nil, errors.New("diffFrom: outpoint both in d.toRemove and other.toRemove with different " +
			"blue scores, with no corresponding entry in d.toAdd")
	}

	result := model.UTXODiff{
		toAdd:    make(model.UTXOCollection, len(d.toRemove)+len(other.toAdd)),
		toRemove: make(model.UTXOCollection, len(d.toAdd)+len(other.toRemove)),
	}

	// All transactions in d.toAdd:
	// If they are not in other.toAdd - should be added in result.toRemove
	inBothToAdd := make(model.UTXOCollection, len(d.toAdd))
	subtractionWithRemainderHavingBlueScoreInPlace(d.toAdd, other.toAdd, result.toRemove, inBothToAdd)
	// If they are in other.toRemove - base utxoSet is not the same
	if checkIntersection(inBothToAdd, d.toRemove) != checkIntersection(inBothToAdd, other.toRemove) {
		return nil, errors.New(
			"diffFrom: outpoint both in d.toAdd, other.toAdd, and only one of d.toRemove and other.toRemove")
	}

	// All transactions in other.toRemove:
	// If they are not in d.toRemove - should be added in result.toRemove
	subtractionHavingBlueScoreInPlace(other.toRemove, d.toRemove, result.toRemove)

	// All transactions in d.toRemove:
	// If they are not in other.toRemove - should be added in result.toAdd
	subtractionHavingBlueScoreInPlace(d.toRemove, other.toRemove, result.toAdd)

	// All transactions in other.toAdd:
	// If they are not in d.toAdd - should be added in result.toAdd
	subtractionHavingBlueScoreInPlace(other.toAdd, d.toAdd, result.toAdd)

	return &result, nil
}

// withDiffInPlace applies provided diff to this diff in-place, that would be the result if
// first d, and than diff were applied to the same base
func (d *model.UTXODiff) withDiffInPlace(diff *model.UTXODiff) error {
	if checkIntersectionWithRule(diff.toRemove, d.toRemove,
		func(outpoint model.DomainOutpoint, entryToAdd, existingEntry *model.UTXOEntry) bool {
			return !d.toAdd.containsWithBlueScore(outpoint, entryToAdd.blockBlueScore)

		}) {
		return errors.New(
			"withDiffInPlace: outpoint both in d.toRemove and in diff.toRemove")
	}

	if checkIntersectionWithRule(diff.toAdd, d.toAdd,
		func(outpoint model.DomainOutpoint, entryToAdd, existingEntry *model.UTXOEntry) bool {
			return !diff.toRemove.containsWithBlueScore(outpoint, existingEntry.blockBlueScore)
		}) {
		return errors.New(
			"withDiffInPlace: outpoint both in d.toAdd and in diff.toAdd")
	}

	intersection := make(model.UTXOCollection, minInt(len(diff.toRemove), len(d.toAdd)))
	// If not exists neither in toAdd nor in toRemove - add to toRemove
	intersectionWithRemainderHavingBlueScoreInPlace(diff.toRemove, d.toAdd, intersection, d.toRemove)
	// If already exists in toAdd with the same blueScore - remove from toAdd
	d.toAdd.removeMultiple(intersection)

	intersection = make(model.UTXOCollection, minInt(len(diff.toAdd), len(d.toRemove)))
	// If not exists neither in toAdd nor in toRemove, or exists in toRemove with different blueScore - add to toAdd
	intersectionWithRemainderHavingBlueScoreInPlace(diff.toAdd, d.toRemove, intersection, d.toAdd)
	// If already exists in toRemove with the same blueScore - remove from toRemove
	d.toRemove.removeMultiple(intersection)

	return nil
}

// WithDiff applies provided diff to this diff, creating a new utxoDiff, that would be the result if
// first d, and than diff were applied to some base
func (d *model.UTXODiff) WithDiff(diff *model.UTXODiff) (*model.UTXODiff, error) {
	clone := d.clone()

	err := clone.withDiffInPlace(diff)
	if err != nil {
		return nil, err
	}

	return clone, nil
}

// clone returns a clone of this utxoDiff
func (d *model.UTXODiff) clone() *model.UTXODiff {
	clone := &model.UTXODiff{
		toAdd:    d.toAdd.clone(),
		toRemove: d.toRemove.clone(),
	}
	return clone
}

// AddEntry adds a model.UTXOEntry to the diff
//
// If d.useMultiset is true, this function MUST be
// called with the DAG lock held.
func (d *model.UTXODiff) AddEntry(outpoint model.DomainOutpoint, entry *model.UTXOEntry) error {
	if d.toRemove.containsWithBlueScore(outpoint, entry.blockBlueScore) {
		d.toRemove.remove(outpoint)
	} else if _, exists := d.toAdd[outpoint]; exists {
		return errors.Errorf("AddEntry: Cannot add outpoint %s twice", outpoint)
	} else {
		d.toAdd.add(outpoint, entry)
	}
	return nil
}

// RemoveEntry removes a model.UTXOEntry from the diff.
//
// If d.useMultiset is true, this function MUST be
// called with the DAG lock held.
func (d *model.UTXODiff) RemoveEntry(outpoint model.DomainOutpoint, entry *model.UTXOEntry) error {
	if d.toAdd.containsWithBlueScore(outpoint, entry.blockBlueScore) {
		d.toAdd.remove(outpoint)
	} else if _, exists := d.toRemove[outpoint]; exists {
		return errors.Errorf("removeEntry: Cannot remove outpoint %s twice", outpoint)
	} else {
		d.toRemove.add(outpoint, entry)
	}
	return nil
}

func (d model.UTXODiff) String() string {
	return fmt.Sprintf("toAdd: %s; toRemove: %s", d.toAdd, d.toRemove)
}
