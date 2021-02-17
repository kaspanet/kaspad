package utxo

import (
	"reflect"
	"testing"

	"github.com/kaspanet/kaspad/domain/consensus/utils/transactionid"

	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

func (mud *mutableUTXODiff) equal(other *mutableUTXODiff) bool {
	if mud == nil || other == nil {
		return mud == other
	}

	return reflect.DeepEqual(mud.toAdd, other.toAdd) &&
		reflect.DeepEqual(mud.toRemove, other.toRemove)
}

// TestUTXOCollection makes sure that utxoCollection cloning and string representations work as expected.
func TestUTXOCollection(t *testing.T) {
	txID0, _ := transactionid.FromString("0000000000000000000000000000000000000000000000000000000000000000")
	txID1, _ := transactionid.FromString("1111111111111111111111111111111111111111111111111111111111111111")
	outpoint0 := externalapi.NewDomainOutpoint(txID0, 0)
	outpoint1 := externalapi.NewDomainOutpoint(txID1, 0)
	utxoEntry0 := NewUTXOEntry(10, &externalapi.ScriptPublicKey{[]byte{}, 0}, true, 0)
	utxoEntry1 := NewUTXOEntry(20, &externalapi.ScriptPublicKey{[]byte{}, 0}, false, 1)

	// For each of the following test cases, we will:
	// .String() the given collection and compare it to expectedStringWithMultiset
	// .clone() the given collection and compare its value to itself (expected: equals) and its reference to itself (expected: not equal)
	tests := []struct {
		name           string
		collection     utxoCollection
		expectedString string
	}{
		{
			name:           "empty collection",
			collection:     utxoCollection{},
			expectedString: "[  ]",
		},
		{
			name: "one member",
			collection: utxoCollection{
				*outpoint0: utxoEntry1,
			},
			expectedString: "[ (0000000000000000000000000000000000000000000000000000000000000000, 0) => 20, blueScore: 1 ]",
		},
		{
			name: "two members",
			collection: utxoCollection{
				*outpoint0: utxoEntry0,
				*outpoint1: utxoEntry1,
			},
			expectedString: "[ (0000000000000000000000000000000000000000000000000000000000000000, 0) => 10, blueScore: 0, (1111111111111111111111111111111111111111111111111111111111111111, 0) => 20, blueScore: 1 ]",
		},
	}

	for _, test := range tests {
		// Test utxoCollection string representation
		collectionString := test.collection.String()
		if collectionString != test.expectedString {
			t.Errorf("unexpected string in test \"%s\". "+
				"Expected: \"%s\", got: \"%s\".", test.name, test.expectedString, collectionString)
		}

		// Test utxoCollection cloning
		collectionClone := test.collection.Clone()
		if reflect.ValueOf(collectionClone).Pointer() == reflect.ValueOf(test.collection).Pointer() {
			t.Errorf("collection is reference-equal to its clone in test \"%s\". ", test.name)
		}
		if !reflect.DeepEqual(test.collection, collectionClone) {
			t.Errorf("collection is not equal to its clone in test \"%s\". "+
				"Expected: \"%s\", got: \"%s\".", test.name, collectionString, collectionClone.String())
		}
	}
}

// TestutxoDiff makes sure that mutableUTXODiff creation, cloning, and string representations work as expected.
func TestUTXODiff(t *testing.T) {
	txID0, _ := transactionid.FromString("0000000000000000000000000000000000000000000000000000000000000000")
	txID1, _ := transactionid.FromString("1111111111111111111111111111111111111111111111111111111111111111")
	outpoint0 := externalapi.NewDomainOutpoint(txID0, 0)
	outpoint1 := externalapi.NewDomainOutpoint(txID1, 0)
	utxoEntry0 := NewUTXOEntry(10, &externalapi.ScriptPublicKey{[]byte{}, 0}, true, 0)
	utxoEntry1 := NewUTXOEntry(20, &externalapi.ScriptPublicKey{[]byte{}, 0}, false, 1)

	diff := newMutableUTXODiff()

	if len(diff.toAdd) != 0 || len(diff.toRemove) != 0 {
		t.Errorf("new diff is not empty")
	}

	err := diff.addEntry(outpoint0, utxoEntry0)
	if err != nil {
		t.Fatalf("error adding entry to utxo diff: %s", err)
	}

	err = diff.removeEntry(outpoint1, utxoEntry1)
	if err != nil {
		t.Fatalf("error adding entry to utxo diff: %s", err)
	}

	// Test mutableUTXODiff cloning
	clonedDiff := diff.clone()
	if clonedDiff == diff {
		t.Errorf("cloned diff is reference-equal to the original")
	}
	if !reflect.DeepEqual(clonedDiff, diff) {
		t.Errorf("cloned diff not equal to the original"+
			"Original: \"%v\", cloned: \"%v\".", diff, clonedDiff)
	}

	// Test mutableUTXODiff string representation
	expectedDiffString := "toAdd: [ (0000000000000000000000000000000000000000000000000000000000000000, 0) => 10, blueScore: 0 ]; toRemove: [ (1111111111111111111111111111111111111111111111111111111111111111, 0) => 20, blueScore: 1 ]"
	diffString := clonedDiff.String()
	if diffString != expectedDiffString {
		t.Errorf("unexpected diff string. "+
			"Expected: \"%s\", got: \"%s\".", expectedDiffString, diffString)
	}
}

// TestutxoDiffRules makes sure that all diffFrom and WithDiff rules are followed.
// Each test case represents a cell in the two tables outlined in the documentation for mutableUTXODiff.
func TestUTXODiffRules(t *testing.T) {
	txID0, _ := transactionid.FromString("0000000000000000000000000000000000000000000000000000000000000000")
	outpoint0 := externalapi.NewDomainOutpoint(txID0, 0)
	utxoEntry1 := NewUTXOEntry(10, &externalapi.ScriptPublicKey{[]byte{}, 0}, true, 0)
	utxoEntry2 := NewUTXOEntry(20, &externalapi.ScriptPublicKey{[]byte{}, 0}, true, 1)

	// For each of the following test cases, we will:
	// this.diffFrom(other) and compare it to expectedDiffFromResult
	// this.WithDiff(other) and compare it to expectedWithDiffResult
	// this.withDiffInPlace(other) and compare it to expectedWithDiffResult
	//
	// Note: an expected nil result means that we expect the respective operation to fail
	// See the following spreadsheet for a summary of all test-cases:
	// https://docs.google.com/spreadsheets/d/1E8G3mp5y1-yifouwLLXRLueSRfXdDRwRKFieYE07buY/edit?usp=sharing
	tests := []struct {
		name                   string
		this                   *mutableUTXODiff
		other                  *mutableUTXODiff
		expectedDiffFromResult *mutableUTXODiff
		expectedWithDiffResult *mutableUTXODiff
	}{
		{
			name: "first toAdd in this, first toAdd in other",
			this: &mutableUTXODiff{
				toAdd:    utxoCollection{*outpoint0: utxoEntry1},
				toRemove: utxoCollection{},
			},
			other: &mutableUTXODiff{
				toAdd:    utxoCollection{*outpoint0: utxoEntry1},
				toRemove: utxoCollection{},
			},
			expectedDiffFromResult: &mutableUTXODiff{
				toAdd:    utxoCollection{},
				toRemove: utxoCollection{},
			},
			expectedWithDiffResult: nil,
		},
		{
			name: "first in toAdd in this, second in toAdd in other",
			this: &mutableUTXODiff{
				toAdd:    utxoCollection{*outpoint0: utxoEntry1},
				toRemove: utxoCollection{},
			},
			other: &mutableUTXODiff{
				toAdd:    utxoCollection{*outpoint0: utxoEntry2},
				toRemove: utxoCollection{},
			},
			expectedDiffFromResult: &mutableUTXODiff{
				toAdd:    utxoCollection{*outpoint0: utxoEntry2},
				toRemove: utxoCollection{*outpoint0: utxoEntry1},
			},
			expectedWithDiffResult: nil,
		},
		{
			name: "first in toAdd in this, second in toRemove in other",
			this: &mutableUTXODiff{
				toAdd:    utxoCollection{*outpoint0: utxoEntry1},
				toRemove: utxoCollection{},
			},
			other: &mutableUTXODiff{
				toAdd:    utxoCollection{},
				toRemove: utxoCollection{*outpoint0: utxoEntry1},
			},
			expectedDiffFromResult: nil,
			expectedWithDiffResult: &mutableUTXODiff{
				toAdd:    utxoCollection{},
				toRemove: utxoCollection{},
			},
		},
		{
			name: "first in toAdd in this and other, second in toRemove in other",
			this: &mutableUTXODiff{
				toAdd:    utxoCollection{*outpoint0: utxoEntry1},
				toRemove: utxoCollection{},
			},
			other: &mutableUTXODiff{
				toAdd:    utxoCollection{*outpoint0: utxoEntry1},
				toRemove: utxoCollection{*outpoint0: utxoEntry2},
			},
			expectedDiffFromResult: nil,
			expectedWithDiffResult: nil,
		},
		{
			name: "first in toAdd in this and toRemove in other, second in toAdd in other",
			this: &mutableUTXODiff{
				toAdd:    utxoCollection{*outpoint0: utxoEntry1},
				toRemove: utxoCollection{},
			},
			other: &mutableUTXODiff{
				toAdd:    utxoCollection{*outpoint0: utxoEntry2},
				toRemove: utxoCollection{*outpoint0: utxoEntry1},
			},
			expectedDiffFromResult: nil,
			expectedWithDiffResult: &mutableUTXODiff{
				toAdd:    utxoCollection{*outpoint0: utxoEntry2},
				toRemove: utxoCollection{},
			},
		},
		{
			name: "first in toAdd in this, empty other",
			this: &mutableUTXODiff{
				toAdd:    utxoCollection{*outpoint0: utxoEntry1},
				toRemove: utxoCollection{},
			},
			other: &mutableUTXODiff{
				toAdd:    utxoCollection{},
				toRemove: utxoCollection{},
			},
			expectedDiffFromResult: &mutableUTXODiff{
				toAdd:    utxoCollection{},
				toRemove: utxoCollection{*outpoint0: utxoEntry1},
			},
			expectedWithDiffResult: &mutableUTXODiff{
				toAdd:    utxoCollection{*outpoint0: utxoEntry1},
				toRemove: utxoCollection{},
			},
		},
		{
			name: "first in toRemove in this and in toAdd in other",
			this: &mutableUTXODiff{
				toAdd:    utxoCollection{},
				toRemove: utxoCollection{*outpoint0: utxoEntry1},
			},
			other: &mutableUTXODiff{
				toAdd:    utxoCollection{*outpoint0: utxoEntry1},
				toRemove: utxoCollection{},
			},
			expectedDiffFromResult: nil,
			expectedWithDiffResult: &mutableUTXODiff{
				toAdd:    utxoCollection{},
				toRemove: utxoCollection{},
			},
		},
		{
			name: "first in toRemove in this, second in toAdd in other",
			this: &mutableUTXODiff{
				toAdd:    utxoCollection{},
				toRemove: utxoCollection{*outpoint0: utxoEntry1},
			},
			other: &mutableUTXODiff{
				toAdd:    utxoCollection{*outpoint0: utxoEntry2},
				toRemove: utxoCollection{},
			},
			expectedDiffFromResult: nil,
			expectedWithDiffResult: &mutableUTXODiff{
				toAdd:    utxoCollection{*outpoint0: utxoEntry2},
				toRemove: utxoCollection{*outpoint0: utxoEntry1},
			},
		},
		{
			name: "first in toRemove in this and other",
			this: &mutableUTXODiff{
				toAdd:    utxoCollection{},
				toRemove: utxoCollection{*outpoint0: utxoEntry1},
			},
			other: &mutableUTXODiff{
				toAdd:    utxoCollection{},
				toRemove: utxoCollection{*outpoint0: utxoEntry1},
			},
			expectedDiffFromResult: &mutableUTXODiff{
				toAdd:    utxoCollection{},
				toRemove: utxoCollection{},
			},
			expectedWithDiffResult: nil,
		},
		{
			name: "first in toRemove in this, second in toRemove in other",
			this: &mutableUTXODiff{
				toAdd:    utxoCollection{},
				toRemove: utxoCollection{*outpoint0: utxoEntry1},
			},
			other: &mutableUTXODiff{
				toAdd:    utxoCollection{},
				toRemove: utxoCollection{*outpoint0: utxoEntry2},
			},
			expectedDiffFromResult: nil,
			expectedWithDiffResult: nil,
		},
		{
			name: "first in toRemove in this and toAdd in other, second in toRemove in other",
			this: &mutableUTXODiff{
				toAdd:    utxoCollection{},
				toRemove: utxoCollection{*outpoint0: utxoEntry1},
			},
			other: &mutableUTXODiff{
				toAdd:    utxoCollection{*outpoint0: utxoEntry1},
				toRemove: utxoCollection{*outpoint0: utxoEntry2},
			},
			expectedDiffFromResult: nil,
			expectedWithDiffResult: nil,
		},
		{
			name: "first in toRemove in this and other, second in toAdd in other",
			this: &mutableUTXODiff{
				toAdd:    utxoCollection{},
				toRemove: utxoCollection{*outpoint0: utxoEntry1},
			},
			other: &mutableUTXODiff{
				toAdd:    utxoCollection{*outpoint0: utxoEntry2},
				toRemove: utxoCollection{*outpoint0: utxoEntry1},
			},
			expectedDiffFromResult: &mutableUTXODiff{
				toAdd:    utxoCollection{*outpoint0: utxoEntry2},
				toRemove: utxoCollection{},
			},
			expectedWithDiffResult: nil,
		},
		{
			name: "first in toRemove in this, empty other",
			this: &mutableUTXODiff{
				toAdd:    utxoCollection{},
				toRemove: utxoCollection{*outpoint0: utxoEntry1},
			},
			other: &mutableUTXODiff{
				toAdd:    utxoCollection{},
				toRemove: utxoCollection{},
			},
			expectedDiffFromResult: &mutableUTXODiff{
				toAdd:    utxoCollection{*outpoint0: utxoEntry1},
				toRemove: utxoCollection{},
			},
			expectedWithDiffResult: &mutableUTXODiff{
				toAdd:    utxoCollection{},
				toRemove: utxoCollection{*outpoint0: utxoEntry1},
			},
		},
		{
			name: "first in toAdd in this and other, second in toRemove in this",
			this: &mutableUTXODiff{
				toAdd:    utxoCollection{*outpoint0: utxoEntry1},
				toRemove: utxoCollection{*outpoint0: utxoEntry2},
			},
			other: &mutableUTXODiff{
				toAdd:    utxoCollection{*outpoint0: utxoEntry1},
				toRemove: utxoCollection{},
			},
			expectedDiffFromResult: nil,
			expectedWithDiffResult: nil,
		},
		{
			name: "first in toAdd in this, second in toRemove in this and toAdd in other",
			this: &mutableUTXODiff{
				toAdd:    utxoCollection{*outpoint0: utxoEntry1},
				toRemove: utxoCollection{*outpoint0: utxoEntry2},
			},
			other: &mutableUTXODiff{
				toAdd:    utxoCollection{*outpoint0: utxoEntry2},
				toRemove: utxoCollection{},
			},
			expectedDiffFromResult: nil,
			expectedWithDiffResult: nil,
		},
		{
			name: "first in toAdd in this and toRemove in other, second in toRemove in this",
			this: &mutableUTXODiff{
				toAdd:    utxoCollection{*outpoint0: utxoEntry1},
				toRemove: utxoCollection{*outpoint0: utxoEntry2},
			},
			other: &mutableUTXODiff{
				toAdd:    utxoCollection{},
				toRemove: utxoCollection{*outpoint0: utxoEntry1},
			},
			expectedDiffFromResult: nil,
			expectedWithDiffResult: &mutableUTXODiff{
				toAdd:    utxoCollection{},
				toRemove: utxoCollection{*outpoint0: utxoEntry2},
			},
		},
		{
			name: "first in toAdd in this, second in toRemove in this and in other",
			this: &mutableUTXODiff{
				toAdd:    utxoCollection{*outpoint0: utxoEntry1},
				toRemove: utxoCollection{*outpoint0: utxoEntry2},
			},
			other: &mutableUTXODiff{
				toAdd:    utxoCollection{},
				toRemove: utxoCollection{*outpoint0: utxoEntry2},
			},
			expectedDiffFromResult: &mutableUTXODiff{
				toAdd:    utxoCollection{},
				toRemove: utxoCollection{*outpoint0: utxoEntry1},
			},
			expectedWithDiffResult: nil,
		},
		{
			name: "first in toAdd and second in toRemove in both this and other",
			this: &mutableUTXODiff{
				toAdd:    utxoCollection{*outpoint0: utxoEntry1},
				toRemove: utxoCollection{*outpoint0: utxoEntry2},
			},
			other: &mutableUTXODiff{
				toAdd:    utxoCollection{*outpoint0: utxoEntry1},
				toRemove: utxoCollection{*outpoint0: utxoEntry2},
			},
			expectedDiffFromResult: &mutableUTXODiff{
				toAdd:    utxoCollection{},
				toRemove: utxoCollection{},
			},
			expectedWithDiffResult: nil,
		},
		{
			name: "first in toAdd in this and toRemove in other, second in toRemove in this and toAdd in other",
			this: &mutableUTXODiff{
				toAdd:    utxoCollection{*outpoint0: utxoEntry1},
				toRemove: utxoCollection{*outpoint0: utxoEntry2},
			},
			other: &mutableUTXODiff{
				toAdd:    utxoCollection{*outpoint0: utxoEntry2},
				toRemove: utxoCollection{*outpoint0: utxoEntry1},
			},
			expectedDiffFromResult: nil,
			expectedWithDiffResult: &mutableUTXODiff{
				toAdd:    utxoCollection{},
				toRemove: utxoCollection{},
			},
		},
		{
			name: "first in toAdd and second in toRemove in this, empty other",
			this: &mutableUTXODiff{
				toAdd:    utxoCollection{*outpoint0: utxoEntry1},
				toRemove: utxoCollection{*outpoint0: utxoEntry2},
			},
			other: &mutableUTXODiff{
				toAdd:    utxoCollection{},
				toRemove: utxoCollection{},
			},
			expectedDiffFromResult: &mutableUTXODiff{
				toAdd:    utxoCollection{*outpoint0: utxoEntry2},
				toRemove: utxoCollection{*outpoint0: utxoEntry1},
			},
			expectedWithDiffResult: &mutableUTXODiff{
				toAdd:    utxoCollection{*outpoint0: utxoEntry1},
				toRemove: utxoCollection{*outpoint0: utxoEntry2},
			},
		},
		{
			name: "empty this, first in toAdd in other",
			this: &mutableUTXODiff{
				toAdd:    utxoCollection{},
				toRemove: utxoCollection{},
			},
			other: &mutableUTXODiff{
				toAdd:    utxoCollection{*outpoint0: utxoEntry1},
				toRemove: utxoCollection{},
			},
			expectedDiffFromResult: &mutableUTXODiff{
				toAdd:    utxoCollection{*outpoint0: utxoEntry1},
				toRemove: utxoCollection{},
			},
			expectedWithDiffResult: &mutableUTXODiff{
				toAdd:    utxoCollection{*outpoint0: utxoEntry1},
				toRemove: utxoCollection{},
			},
		},
		{
			name: "empty this, first in toRemove in other",
			this: &mutableUTXODiff{
				toAdd:    utxoCollection{},
				toRemove: utxoCollection{},
			},
			other: &mutableUTXODiff{
				toAdd:    utxoCollection{},
				toRemove: utxoCollection{*outpoint0: utxoEntry1},
			},
			expectedDiffFromResult: &mutableUTXODiff{
				toAdd:    utxoCollection{},
				toRemove: utxoCollection{*outpoint0: utxoEntry1},
			},
			expectedWithDiffResult: &mutableUTXODiff{
				toAdd:    utxoCollection{},
				toRemove: utxoCollection{*outpoint0: utxoEntry1},
			},
		},
		{
			name: "empty this, first in toAdd and second in toRemove in other",
			this: &mutableUTXODiff{
				toAdd:    utxoCollection{},
				toRemove: utxoCollection{},
			},
			other: &mutableUTXODiff{
				toAdd:    utxoCollection{*outpoint0: utxoEntry1},
				toRemove: utxoCollection{*outpoint0: utxoEntry2},
			},
			expectedDiffFromResult: &mutableUTXODiff{
				toAdd:    utxoCollection{*outpoint0: utxoEntry1},
				toRemove: utxoCollection{*outpoint0: utxoEntry2},
			},
			expectedWithDiffResult: &mutableUTXODiff{
				toAdd:    utxoCollection{*outpoint0: utxoEntry1},
				toRemove: utxoCollection{*outpoint0: utxoEntry2},
			},
		},
		{
			name: "empty this, empty other",
			this: &mutableUTXODiff{
				toAdd:    utxoCollection{},
				toRemove: utxoCollection{},
			},
			other: &mutableUTXODiff{
				toAdd:    utxoCollection{},
				toRemove: utxoCollection{},
			},
			expectedDiffFromResult: &mutableUTXODiff{
				toAdd:    utxoCollection{},
				toRemove: utxoCollection{},
			},
			expectedWithDiffResult: &mutableUTXODiff{
				toAdd:    utxoCollection{},
				toRemove: utxoCollection{},
			},
		},
	}

	for _, test := range tests {
		// diffFrom from test.this to test.other
		diffResult, err := diffFrom(test.this, test.other)

		// Test whether diffFrom returned an error
		isDiffFromOk := err == nil
		expectedIsDiffFromOk := test.expectedDiffFromResult != nil
		if isDiffFromOk != expectedIsDiffFromOk {
			t.Errorf("unexpected diffFrom error in test \"%s\". "+
				"Expected: \"%t\", got: \"%t\".", test.name, expectedIsDiffFromOk, isDiffFromOk)
		}

		// If not error, test the diffFrom result
		if isDiffFromOk && !test.expectedDiffFromResult.equal(diffResult) {
			t.Errorf("unexpected diffFrom result in test \"%s\". "+
				"Expected: \"%v\", got: \"%v\".", test.name, test.expectedDiffFromResult, diffResult)
		}

		// Make sure that WithDiff after diffFrom results in the original test.other
		if isDiffFromOk {
			otherResult, err := withDiff(test.this, diffResult)
			if err != nil {
				t.Errorf("WithDiff unexpectedly failed in test \"%s\": %s", test.name, err)
			}
			if !test.other.equal(otherResult) {
				t.Errorf("unexpected WithDiff result in test \"%s\". "+
					"Expected: \"%v\", got: \"%v\".", test.name, test.other, otherResult)
			}
		}

		// WithDiff from test.this to test.other
		withDiffResult, err := withDiff(test.this, test.other)

		// Test whether WithDiff returned an error
		isWithDiffOk := err == nil
		expectedIsWithDiffOk := test.expectedWithDiffResult != nil
		if isWithDiffOk != expectedIsWithDiffOk {
			t.Errorf("unexpected WithDiff error in test \"%s\". "+
				"Expected: \"%t\", got: \"%t\".", test.name, expectedIsWithDiffOk, isWithDiffOk)
		}

		// If not error, test the WithDiff result
		if isWithDiffOk && !withDiffResult.equal(test.expectedWithDiffResult) {
			t.Errorf("unexpected WithDiff result in test \"%s\". "+
				"Expected: \"%v\", got: \"%v\".", test.name, test.expectedWithDiffResult, withDiffResult)
		}

		// Repeat WithDiff check test.this time using withDiffInPlace
		thisClone := test.this.clone()
		err = withDiffInPlace(thisClone, test.other)

		// Test whether withDiffInPlace returned an error
		isWithDiffInPlaceOk := err == nil
		expectedIsWithDiffInPlaceOk := test.expectedWithDiffResult != nil
		if isWithDiffInPlaceOk != expectedIsWithDiffInPlaceOk {
			t.Errorf("unexpected withDiffInPlace error in test \"%s\". "+
				"Expected: \"%t\", got: \"%t\".", test.name, expectedIsWithDiffInPlaceOk, isWithDiffInPlaceOk)
		}

		// If not error, test the withDiffInPlace result
		if isWithDiffInPlaceOk && !thisClone.equal(test.expectedWithDiffResult) {
			t.Errorf("unexpected withDiffInPlace result in test \"%s\". "+
				"Expected: \"%v\", got: \"%v\".", test.name, test.expectedWithDiffResult, thisClone)
		}

		// Make sure that diffFrom after WithDiff results in the original test.other
		if isWithDiffOk {
			otherResult, err := diffFrom(test.this, withDiffResult)
			if err != nil {
				t.Errorf("diffFrom unexpectedly failed in test \"%s\": %s", test.name, err)
			}
			if !test.other.equal(otherResult) {
				t.Errorf("unexpected diffFrom result in test \"%s\". "+
					"Expected: \"%v\", got: \"%v\".", test.name, test.other, otherResult)
			}
		}
	}
}
