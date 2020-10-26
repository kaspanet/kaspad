package utxoalgebra

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/pkg/errors"
)

// checkIntersection checks if there is an intersection between two model.UTXOCollections
func checkIntersection(collection1 model.UTXOCollection, collection2 model.UTXOCollection) bool {
	for outpoint := range collection1 {
		if collectionContains(collection2, &outpoint) {
			return true
		}
	}

	return false
}

// checkIntersectionWithRule checks if there is an intersection between two model.UTXOCollections satisfying arbitrary rule
func checkIntersectionWithRule(collection1 model.UTXOCollection, collection2 model.UTXOCollection,
	extraRule func(*externalapi.DomainOutpoint, *externalapi.UTXOEntry, *externalapi.UTXOEntry) bool) bool {

	for outpoint, utxoEntry := range collection1 {
		if diffEntry, ok := collectionGet(collection2, &outpoint); ok {
			if extraRule(&outpoint, utxoEntry, diffEntry) {
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
		if collectionContainsWithBlueScore(collection2, &outpoint, utxoEntry.BlockBlueScore) {
			collectionAdd(result, &outpoint, utxoEntry)
		} else {
			collectionAdd(remainder, &outpoint, utxoEntry)
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
		if !collectionContainsWithBlueScore(collection2, &outpoint, utxoEntry.BlockBlueScore) {
			collectionAdd(result, &outpoint, utxoEntry)
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
		if !collectionContainsWithBlueScore(collection2, &outpoint, utxoEntry.BlockBlueScore) {
			collectionAdd(result, &outpoint, utxoEntry)
		} else {
			collectionAdd(remainder, &outpoint, utxoEntry)
		}
	}
}

// DiffFrom returns a new utxoDiff with the difference between this utxoDiff and another
// Assumes that:
// Both utxoDiffs are from the same base
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
// -		Don't add anything to the result
// X		Return an error
// ToAdd	Add the UTXO into the ToAdd collection of the result
// ToRemove	Add the UTXO into the ToRemove collection of the result
//
// Examples:
// 1. This diff contains a UTXO in ToAdd, and the other diff contains it in ToRemove
//    diffFrom results in an error
// 2. This diff contains a UTXO in ToRemove, and the other diff does not contain it
//    diffFrom results in the UTXO being added to ToAdd
func DiffFrom(this, other *model.UTXODiff) (*model.UTXODiff, error) {
	// Note that the following cases are not accounted for, as they are impossible
	// as long as the base utxoSet is the same:
	// - if utxoEntry is in this.ToAdd and other.ToRemove
	// - if utxoEntry is in this.ToRemove and other.ToAdd

	// check that NOT (entries with unequal blue scores AND utxoEntry is in this.ToAdd and/or other.ToRemove) -> Error
	isNotAddedOutputRemovedWithBlueScore := func(outpoint *externalapi.DomainOutpoint, utxoEntry, diffEntry *externalapi.UTXOEntry) bool {
		return !(diffEntry.BlockBlueScore != utxoEntry.BlockBlueScore &&
			(collectionContainsWithBlueScore(this.ToAdd, outpoint, diffEntry.BlockBlueScore) ||
				collectionContainsWithBlueScore(other.ToRemove, outpoint, utxoEntry.BlockBlueScore)))
	}

	if checkIntersectionWithRule(this.ToRemove, other.ToAdd, isNotAddedOutputRemovedWithBlueScore) {
		return nil, errors.New("diffFrom: outpoint both in this.ToAdd and in other.ToRemove")
	}

	//check that NOT (entries with unequal blue score AND utxoEntry is in this.ToRemove and/or other.ToAdd) -> Error
	isNotRemovedOutputAddedWithBlueScore :=
		func(outpoint *externalapi.DomainOutpoint, utxoEntry, diffEntry *externalapi.UTXOEntry) bool {

			return !(diffEntry.BlockBlueScore != utxoEntry.BlockBlueScore &&
				(collectionContainsWithBlueScore(this.ToRemove, outpoint, diffEntry.BlockBlueScore) ||
					collectionContainsWithBlueScore(other.ToAdd, outpoint, utxoEntry.BlockBlueScore)))
		}

	if checkIntersectionWithRule(this.ToAdd, other.ToRemove, isNotRemovedOutputAddedWithBlueScore) {
		return nil, errors.New("diffFrom: outpoint both in this.ToRemove and in other.ToAdd")
	}

	// if have the same entry in this.ToRemove and other.ToRemove
	// and existing entry is with different blue score, in this case - this is an error
	if checkIntersectionWithRule(this.ToRemove, other.ToRemove,
		func(outpoint *externalapi.DomainOutpoint, utxoEntry, diffEntry *externalapi.UTXOEntry) bool {
			return utxoEntry.BlockBlueScore != diffEntry.BlockBlueScore
		}) {
		return nil, errors.New("diffFrom: outpoint both in this.ToRemove and other.ToRemove with different " +
			"blue scores, with no corresponding entry in this.ToAdd")
	}

	result := model.UTXODiff{
		ToAdd:    make(model.UTXOCollection, len(this.ToRemove)+len(other.ToAdd)),
		ToRemove: make(model.UTXOCollection, len(this.ToAdd)+len(other.ToRemove)),
	}

	// All transactions in this.ToAdd:
	// If they are not in other.ToAdd - should be added in result.ToRemove
	inBothToAdd := make(model.UTXOCollection, len(this.ToAdd))
	subtractionWithRemainderHavingBlueScoreInPlace(this.ToAdd, other.ToAdd, result.ToRemove, inBothToAdd)
	// If they are in other.ToRemove - base utxoSet is not the same
	if checkIntersection(inBothToAdd, this.ToRemove) != checkIntersection(inBothToAdd, other.ToRemove) {
		return nil, errors.New(
			"diffFrom: outpoint both in this.ToAdd, other.ToAdd, and only one of this.ToRemove and other.ToRemove")
	}

	// All transactions in other.ToRemove:
	// If they are not in this.ToRemove - should be added in result.ToRemove
	subtractionHavingBlueScoreInPlace(other.ToRemove, this.ToRemove, result.ToRemove)

	// All transactions in this.ToRemove:
	// If they are not in other.ToRemove - should be added in result.ToAdd
	subtractionHavingBlueScoreInPlace(this.ToRemove, other.ToRemove, result.ToAdd)

	// All transactions in other.ToAdd:
	// If they are not in this.ToAdd - should be added in result.ToAdd
	subtractionHavingBlueScoreInPlace(other.ToAdd, this.ToAdd, result.ToAdd)

	return &result, nil
}

// WithDiffInPlace applies provided diff to this diff in-place, that would be the result if
// first d, and than diff were applied to the same base
func WithDiffInPlace(this *model.UTXODiff, diff *model.UTXODiff) error {
	if checkIntersectionWithRule(diff.ToRemove, this.ToRemove,
		func(outpoint *externalapi.DomainOutpoint, entryToAdd, existingEntry *externalapi.UTXOEntry) bool {
			return !collectionContainsWithBlueScore(this.ToAdd, outpoint, entryToAdd.BlockBlueScore)

		}) {
		return errors.New(
			"withDiffInPlace: outpoint both in this.ToRemove and in diff.ToRemove")
	}

	if checkIntersectionWithRule(diff.ToAdd, this.ToAdd,
		func(outpoint *externalapi.DomainOutpoint, entryToAdd, existingEntry *externalapi.UTXOEntry) bool {
			return !collectionContainsWithBlueScore(diff.ToRemove, outpoint, existingEntry.BlockBlueScore)
		}) {
		return errors.New(
			"withDiffInPlace: outpoint both in this.ToAdd and in diff.ToAdd")
	}

	intersection := make(model.UTXOCollection, minInt(len(diff.ToRemove), len(this.ToAdd)))
	// If not exists neither in ToAdd nor in ToRemove - add to ToRemove
	intersectionWithRemainderHavingBlueScoreInPlace(diff.ToRemove, this.ToAdd, intersection, this.ToRemove)
	// If already exists in ToAdd with the same blueScore - remove from ToAdd
	collectionRemoveMultiple(this.ToAdd, intersection)

	intersection = make(model.UTXOCollection, minInt(len(diff.ToAdd), len(this.ToRemove)))
	// If not exists neither in ToAdd nor in ToRemove, or exists in ToRemove with different blueScore - add to ToAdd
	intersectionWithRemainderHavingBlueScoreInPlace(diff.ToAdd, this.ToRemove, intersection, this.ToAdd)
	// If already exists in ToRemove with the same blueScore - remove from ToRemove
	collectionRemoveMultiple(this.ToRemove, intersection)

	return nil
}

// WithDiff applies provided diff to this diff, creating a new utxoDiff, that would be the result if
// first d, and than diff were applied to some base
func WithDiff(this *model.UTXODiff, diff *model.UTXODiff) (*model.UTXODiff, error) {
	clone := diffClone(this)

	err := WithDiffInPlace(clone, diff)
	if err != nil {
		return nil, err
	}

	return clone, nil
}
