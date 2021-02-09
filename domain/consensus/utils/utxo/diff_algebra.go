package utxo

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/pkg/errors"
)

// checkIntersection checks if there is an intersection between two utxoCollections
func checkIntersection(collection1 utxoCollection, collection2 utxoCollection) bool {
	for outpoint := range collection1 {
		if collection2.Contains(&outpoint) {
			return true
		}
	}

	return false
}

// checkIntersectionWithRule checks if there is an intersection between two utxoCollections satisfying arbitrary rule
// returns the first outpoint in the two collections' intersection satsifying the rule, and a boolean indicating whether
// such outpoint exists
func checkIntersectionWithRule(collection1 utxoCollection, collection2 utxoCollection,
	extraRule func(*externalapi.DomainOutpoint, externalapi.UTXOEntry, externalapi.UTXOEntry) bool) (
	*externalapi.DomainOutpoint, bool) {

	for outpoint, utxoEntry := range collection1 {
		if diffEntry, ok := collection2.Get(&outpoint); ok {
			if extraRule(&outpoint, utxoEntry, diffEntry) {
				return &outpoint, true
			}
		}
	}

	return nil, false
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
		if collection2.containsWithBlueScore(&outpoint, utxoEntry.BlockBlueScore()) {
			result.add(&outpoint, utxoEntry)
		} else {
			remainder.add(&outpoint, utxoEntry)
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
		if !collection2.containsWithBlueScore(&outpoint, utxoEntry.BlockBlueScore()) {
			result.add(&outpoint, utxoEntry)
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
		if !collection2.containsWithBlueScore(&outpoint, utxoEntry.BlockBlueScore()) {
			result.add(&outpoint, utxoEntry)
		} else {
			remainder.add(&outpoint, utxoEntry)
		}
	}
}

// DiffFrom returns a new mutableUTXODiff with the difference between this mutableUTXODiff and another
// Assumes that:
// Both mutableUTXODiffs are from the same base
// If a txOut exists in both mutableUTXODiffs, its underlying values would be the same
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
func diffFrom(this, other *mutableUTXODiff) (*mutableUTXODiff, error) {
	// Note that the following cases are not accounted for, as they are impossible
	// as long as the base utxoSet is the same:
	// - if utxoEntry is in this.toAdd and other.toRemove
	// - if utxoEntry is in this.toRemove and other.toAdd

	// check that NOT (entries with unequal blue scores AND utxoEntry is in this.toAdd and/or other.toRemove) -> Error
	isNotAddedOutputRemovedWithBlueScore := func(outpoint *externalapi.DomainOutpoint, utxoEntry, diffEntry externalapi.UTXOEntry) bool {
		return !(diffEntry.BlockBlueScore() != utxoEntry.BlockBlueScore() &&
			(this.toAdd.containsWithBlueScore(outpoint, diffEntry.BlockBlueScore()) ||
				other.toRemove.containsWithBlueScore(outpoint, utxoEntry.BlockBlueScore())))
	}

	if offendingOutpoint, ok :=
		checkIntersectionWithRule(this.toRemove, other.toAdd, isNotAddedOutputRemovedWithBlueScore); ok {
		return nil, errors.Errorf("diffFrom: outpoint %s both in this.toAdd and in other.toRemove", offendingOutpoint)
	}

	//check that NOT (entries with unequal blue score AND utxoEntry is in this.toRemove and/or other.toAdd) -> Error
	isNotRemovedOutputAddedWithBlueScore :=
		func(outpoint *externalapi.DomainOutpoint, utxoEntry, diffEntry externalapi.UTXOEntry) bool {

			return !(diffEntry.BlockBlueScore() != utxoEntry.BlockBlueScore() &&
				(this.toRemove.containsWithBlueScore(outpoint, diffEntry.BlockBlueScore()) ||
					other.toAdd.containsWithBlueScore(outpoint, utxoEntry.BlockBlueScore())))
		}

	if offendingOutpoint, ok :=
		checkIntersectionWithRule(this.toAdd, other.toRemove, isNotRemovedOutputAddedWithBlueScore); ok {
		return nil, errors.Errorf("diffFrom: outpoint %s both in this.toRemove and in other.toAdd", offendingOutpoint)
	}

	// if have the same entry in this.toRemove and other.toRemove
	// and existing entry is with different blue score, in this case - this is an error
	if offendingOutpoint, ok := checkIntersectionWithRule(this.toRemove, other.toRemove,
		func(outpoint *externalapi.DomainOutpoint, utxoEntry, diffEntry externalapi.UTXOEntry) bool {
			return utxoEntry.BlockBlueScore() != diffEntry.BlockBlueScore()
		}); ok {
		return nil, errors.Errorf("diffFrom: outpoint %s both in this.toRemove and other.toRemove with different "+
			"blue scores, with no corresponding entry in this.toAdd", offendingOutpoint)
	}

	result := &mutableUTXODiff{
		toAdd:    make(utxoCollection, len(this.toRemove)+len(other.toAdd)),
		toRemove: make(utxoCollection, len(this.toAdd)+len(other.toRemove)),
	}

	// All transactions in this.toAdd:
	// If they are not in other.toAdd - should be added in result.toRemove
	inBothToAdd := make(utxoCollection, len(this.toAdd))
	subtractionWithRemainderHavingBlueScoreInPlace(this.toAdd, other.toAdd, result.toRemove, inBothToAdd)
	// If they are in other.toRemove - base utxoSet is not the same
	if checkIntersection(inBothToAdd, this.toRemove) != checkIntersection(inBothToAdd, other.toRemove) {
		return nil, errors.New(
			"diffFrom: outpoint both in this.toAdd, other.toAdd, and only one of this.toRemove and other.toRemove")
	}

	// All transactions in other.toRemove:
	// If they are not in this.toRemove - should be added in result.toRemove
	subtractionHavingBlueScoreInPlace(other.toRemove, this.toRemove, result.toRemove)

	// All transactions in this.toRemove:
	// If they are not in other.toRemove - should be added in result.toAdd
	subtractionHavingBlueScoreInPlace(this.toRemove, other.toRemove, result.toAdd)

	// All transactions in other.toAdd:
	// If they are not in this.toAdd - should be added in result.toAdd
	subtractionHavingBlueScoreInPlace(other.toAdd, this.toAdd, result.toAdd)

	return result, nil
}

// WithDiffInPlace applies provided diff to this diff in-place, that would be the result if
// first d, and than diff were applied to the same base
func withDiffInPlace(this *mutableUTXODiff, other *mutableUTXODiff) error {
	if offendingOutpoint, ok := checkIntersectionWithRule(other.toRemove, this.toRemove,
		func(outpoint *externalapi.DomainOutpoint, entryToAdd, existingEntry externalapi.UTXOEntry) bool {
			return !this.toAdd.containsWithBlueScore(outpoint, entryToAdd.BlockBlueScore())
		}); ok {
		return errors.Errorf(
			"withDiffInPlace: outpoint %s both in this.toRemove and in other.toRemove", offendingOutpoint)
	}

	if offendingOutpoint, ok := checkIntersectionWithRule(other.toAdd, this.toAdd,
		func(outpoint *externalapi.DomainOutpoint, entryToAdd, existingEntry externalapi.UTXOEntry) bool {
			return !other.toRemove.containsWithBlueScore(outpoint, existingEntry.BlockBlueScore())
		}); ok {
		return errors.Errorf(
			"withDiffInPlace: outpoint %s both in this.toAdd and in other.toAdd", offendingOutpoint)
	}

	intersection := make(utxoCollection, minInt(len(other.toRemove), len(this.toAdd)))
	// If not exists neither in toAdd nor in toRemove - add to toRemove
	intersectionWithRemainderHavingBlueScoreInPlace(other.toRemove, this.toAdd, intersection, this.toRemove)
	// If already exists in toAdd with the same blueScore - remove from toAdd
	this.toAdd.removeMultiple(intersection)

	intersection = make(utxoCollection, minInt(len(other.toAdd), len(this.toRemove)))
	// If not exists neither in toAdd nor in toRemove, or exists in toRemove with different blueScore - add to toAdd
	intersectionWithRemainderHavingBlueScoreInPlace(other.toAdd, this.toRemove, intersection, this.toAdd)
	// If already exists in toRemove with the same blueScore - remove from toRemove
	this.toRemove.removeMultiple(intersection)

	return nil
}

// WithDiff applies provided diff to this diff, creating a new mutableUTXODiff, that would be the result if
// first d, and than diff were applied to some base
func withDiff(this *mutableUTXODiff, diff *mutableUTXODiff) (*mutableUTXODiff, error) {
	clone := this.clone()

	err := withDiffInPlace(clone, diff)
	if err != nil {
		return nil, err
	}

	return clone, nil
}
