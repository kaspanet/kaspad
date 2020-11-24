package utxoalgebra

import (
	"reflect"
	"testing"

	"github.com/kaspanet/kaspad/domain/consensus/utils/transactionid"

	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// TestUTXOCollection makes sure that model.UTXOCollection cloning and string representations work as expected.
func TestUTXOCollection(t *testing.T) {
	txID0, _ := transactionid.FromString("0000000000000000000000000000000000000000000000000000000000000000")
	txID1, _ := transactionid.FromString("1111111111111111111111111111111111111111111111111111111111111111")
	outpoint0 := externalapi.NewDomainOutpoint(txID0, 0)
	outpoint1 := externalapi.NewDomainOutpoint(txID1, 0)
	utxoEntry0 := externalapi.NewUTXOEntry(10, []byte{}, true, 0)
	utxoEntry1 := externalapi.NewUTXOEntry(20, []byte{}, false, 1)

	// For each of the following test cases, we will:
	// .String() the given collection and compare it to expectedStringWithMultiset
	// .clone() the given collection and compare its value to itself (expected: equals) and its reference to itself (expected: not equal)
	tests := []struct {
		name           string
		collection     model.UTXOCollection
		expectedString string
	}{
		{
			name:           "empty collection",
			collection:     model.UTXOCollection{},
			expectedString: "[  ]",
		},
		{
			name: "one member",
			collection: model.UTXOCollection{
				*outpoint0: utxoEntry1,
			},
			expectedString: "[ (0000000000000000000000000000000000000000000000000000000000000000, 0) => 20, blueScore: 1 ]",
		},
		{
			name: "two members",
			collection: model.UTXOCollection{
				*outpoint0: utxoEntry0,
				*outpoint1: utxoEntry1,
			},
			expectedString: "[ (0000000000000000000000000000000000000000000000000000000000000000, 0) => 10, blueScore: 0, (1111111111111111111111111111111111111111111111111111111111111111, 0) => 20, blueScore: 1 ]",
		},
	}

	for _, test := range tests {
		// Test model.UTXOCollection string representation
		collectionString := test.collection.String()
		if collectionString != test.expectedString {
			t.Errorf("unexpected string in test \"%s\". "+
				"Expected: \"%s\", got: \"%s\".", test.name, test.expectedString, collectionString)
		}

		// Test model.UTXOCollection cloning
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

// Testmodel.UTXODiff makes sure that utxoDiff creation, cloning, and string representations work as expected.
func TestUTXODiff(t *testing.T) {
	txID0, _ := transactionid.FromString("0000000000000000000000000000000000000000000000000000000000000000")
	txID1, _ := transactionid.FromString("1111111111111111111111111111111111111111111111111111111111111111")
	outpoint0 := externalapi.NewDomainOutpoint(txID0, 0)
	outpoint1 := externalapi.NewDomainOutpoint(txID1, 0)
	utxoEntry0 := externalapi.NewUTXOEntry(10, []byte{}, true, 0)
	utxoEntry1 := externalapi.NewUTXOEntry(20, []byte{}, false, 1)

	diff := model.NewUTXODiff()

	if len(diff.ToAdd) != 0 || len(diff.ToRemove) != 0 {
		t.Errorf("new diff is not empty")
	}

	err := diffAddEntry(diff, outpoint0, utxoEntry0)
	if err != nil {
		t.Fatalf("error adding entry to utxo diff: %s", err)
	}

	err = diffRemoveEntry(diff, outpoint1, utxoEntry1)
	if err != nil {
		t.Fatalf("error adding entry to utxo diff: %s", err)
	}

	// Test utxoDiff cloning
	clonedDiff := diff.Clone()
	if clonedDiff == diff {
		t.Errorf("cloned diff is reference-equal to the original")
	}
	if !reflect.DeepEqual(clonedDiff, diff) {
		t.Errorf("cloned diff not equal to the original"+
			"Original: \"%v\", cloned: \"%v\".", diff, clonedDiff)
	}

	// Test utxoDiff string representation
	expectedDiffString := "ToAdd: [ (0000000000000000000000000000000000000000000000000000000000000000, 0) => 10, blueScore: 0 ]; ToRemove: [ (1111111111111111111111111111111111111111111111111111111111111111, 0) => 20, blueScore: 1 ]"
	diffString := clonedDiff.String()
	if diffString != expectedDiffString {
		t.Errorf("unexpected diff string. "+
			"Expected: \"%s\", got: \"%s\".", expectedDiffString, diffString)
	}
}

// Testmodel.UTXODiffRules makes sure that all diffFrom and WithDiff rules are followed.
// Each test case represents a cell in the two tables outlined in the documentation for utxoDiff.
func TestUTXODiffRules(t *testing.T) {
	txID0, _ := transactionid.FromString("0000000000000000000000000000000000000000000000000000000000000000")
	outpoint0 := externalapi.NewDomainOutpoint(txID0, 0)
	utxoEntry1 := externalapi.NewUTXOEntry(10, []byte{}, true, 0)
	utxoEntry2 := externalapi.NewUTXOEntry(20, []byte{}, true, 1)

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
		this                   *model.UTXODiff
		other                  *model.UTXODiff
		expectedDiffFromResult *model.UTXODiff
		expectedWithDiffResult *model.UTXODiff
	}{
		{
			name: "first ToAdd in this, first ToAdd in other",
			this: &model.UTXODiff{
				ToAdd:    model.UTXOCollection{*outpoint0: utxoEntry1},
				ToRemove: model.UTXOCollection{},
			},
			other: &model.UTXODiff{
				ToAdd:    model.UTXOCollection{*outpoint0: utxoEntry1},
				ToRemove: model.UTXOCollection{},
			},
			expectedDiffFromResult: &model.UTXODiff{
				ToAdd:    model.UTXOCollection{},
				ToRemove: model.UTXOCollection{},
			},
			expectedWithDiffResult: nil,
		},
		{
			name: "first in ToAdd in this, second in ToAdd in other",
			this: &model.UTXODiff{
				ToAdd:    model.UTXOCollection{*outpoint0: utxoEntry1},
				ToRemove: model.UTXOCollection{},
			},
			other: &model.UTXODiff{
				ToAdd:    model.UTXOCollection{*outpoint0: utxoEntry2},
				ToRemove: model.UTXOCollection{},
			},
			expectedDiffFromResult: &model.UTXODiff{
				ToAdd:    model.UTXOCollection{*outpoint0: utxoEntry2},
				ToRemove: model.UTXOCollection{*outpoint0: utxoEntry1},
			},
			expectedWithDiffResult: nil,
		},
		{
			name: "first in ToAdd in this, second in ToRemove in other",
			this: &model.UTXODiff{
				ToAdd:    model.UTXOCollection{*outpoint0: utxoEntry1},
				ToRemove: model.UTXOCollection{},
			},
			other: &model.UTXODiff{
				ToAdd:    model.UTXOCollection{},
				ToRemove: model.UTXOCollection{*outpoint0: utxoEntry1},
			},
			expectedDiffFromResult: nil,
			expectedWithDiffResult: &model.UTXODiff{
				ToAdd:    model.UTXOCollection{},
				ToRemove: model.UTXOCollection{},
			},
		},
		{
			name: "first in ToAdd in this and other, second in ToRemove in other",
			this: &model.UTXODiff{
				ToAdd:    model.UTXOCollection{*outpoint0: utxoEntry1},
				ToRemove: model.UTXOCollection{},
			},
			other: &model.UTXODiff{
				ToAdd:    model.UTXOCollection{*outpoint0: utxoEntry1},
				ToRemove: model.UTXOCollection{*outpoint0: utxoEntry2},
			},
			expectedDiffFromResult: nil,
			expectedWithDiffResult: nil,
		},
		{
			name: "first in ToAdd in this and ToRemove in other, second in ToAdd in other",
			this: &model.UTXODiff{
				ToAdd:    model.UTXOCollection{*outpoint0: utxoEntry1},
				ToRemove: model.UTXOCollection{},
			},
			other: &model.UTXODiff{
				ToAdd:    model.UTXOCollection{*outpoint0: utxoEntry2},
				ToRemove: model.UTXOCollection{*outpoint0: utxoEntry1},
			},
			expectedDiffFromResult: nil,
			expectedWithDiffResult: &model.UTXODiff{
				ToAdd:    model.UTXOCollection{*outpoint0: utxoEntry2},
				ToRemove: model.UTXOCollection{},
			},
		},
		{
			name: "first in ToAdd in this, empty other",
			this: &model.UTXODiff{
				ToAdd:    model.UTXOCollection{*outpoint0: utxoEntry1},
				ToRemove: model.UTXOCollection{},
			},
			other: &model.UTXODiff{
				ToAdd:    model.UTXOCollection{},
				ToRemove: model.UTXOCollection{},
			},
			expectedDiffFromResult: &model.UTXODiff{
				ToAdd:    model.UTXOCollection{},
				ToRemove: model.UTXOCollection{*outpoint0: utxoEntry1},
			},
			expectedWithDiffResult: &model.UTXODiff{
				ToAdd:    model.UTXOCollection{*outpoint0: utxoEntry1},
				ToRemove: model.UTXOCollection{},
			},
		},
		{
			name: "first in ToRemove in this and in ToAdd in other",
			this: &model.UTXODiff{
				ToAdd:    model.UTXOCollection{},
				ToRemove: model.UTXOCollection{*outpoint0: utxoEntry1},
			},
			other: &model.UTXODiff{
				ToAdd:    model.UTXOCollection{*outpoint0: utxoEntry1},
				ToRemove: model.UTXOCollection{},
			},
			expectedDiffFromResult: nil,
			expectedWithDiffResult: &model.UTXODiff{
				ToAdd:    model.UTXOCollection{},
				ToRemove: model.UTXOCollection{},
			},
		},
		{
			name: "first in ToRemove in this, second in ToAdd in other",
			this: &model.UTXODiff{
				ToAdd:    model.UTXOCollection{},
				ToRemove: model.UTXOCollection{*outpoint0: utxoEntry1},
			},
			other: &model.UTXODiff{
				ToAdd:    model.UTXOCollection{*outpoint0: utxoEntry2},
				ToRemove: model.UTXOCollection{},
			},
			expectedDiffFromResult: nil,
			expectedWithDiffResult: &model.UTXODiff{
				ToAdd:    model.UTXOCollection{*outpoint0: utxoEntry2},
				ToRemove: model.UTXOCollection{*outpoint0: utxoEntry1},
			},
		},
		{
			name: "first in ToRemove in this and other",
			this: &model.UTXODiff{
				ToAdd:    model.UTXOCollection{},
				ToRemove: model.UTXOCollection{*outpoint0: utxoEntry1},
			},
			other: &model.UTXODiff{
				ToAdd:    model.UTXOCollection{},
				ToRemove: model.UTXOCollection{*outpoint0: utxoEntry1},
			},
			expectedDiffFromResult: &model.UTXODiff{
				ToAdd:    model.UTXOCollection{},
				ToRemove: model.UTXOCollection{},
			},
			expectedWithDiffResult: nil,
		},
		{
			name: "first in ToRemove in this, second in ToRemove in other",
			this: &model.UTXODiff{
				ToAdd:    model.UTXOCollection{},
				ToRemove: model.UTXOCollection{*outpoint0: utxoEntry1},
			},
			other: &model.UTXODiff{
				ToAdd:    model.UTXOCollection{},
				ToRemove: model.UTXOCollection{*outpoint0: utxoEntry2},
			},
			expectedDiffFromResult: nil,
			expectedWithDiffResult: nil,
		},
		{
			name: "first in ToRemove in this and ToAdd in other, second in ToRemove in other",
			this: &model.UTXODiff{
				ToAdd:    model.UTXOCollection{},
				ToRemove: model.UTXOCollection{*outpoint0: utxoEntry1},
			},
			other: &model.UTXODiff{
				ToAdd:    model.UTXOCollection{*outpoint0: utxoEntry1},
				ToRemove: model.UTXOCollection{*outpoint0: utxoEntry2},
			},
			expectedDiffFromResult: nil,
			expectedWithDiffResult: nil,
		},
		{
			name: "first in ToRemove in this and other, second in ToAdd in other",
			this: &model.UTXODiff{
				ToAdd:    model.UTXOCollection{},
				ToRemove: model.UTXOCollection{*outpoint0: utxoEntry1},
			},
			other: &model.UTXODiff{
				ToAdd:    model.UTXOCollection{*outpoint0: utxoEntry2},
				ToRemove: model.UTXOCollection{*outpoint0: utxoEntry1},
			},
			expectedDiffFromResult: &model.UTXODiff{
				ToAdd:    model.UTXOCollection{*outpoint0: utxoEntry2},
				ToRemove: model.UTXOCollection{},
			},
			expectedWithDiffResult: nil,
		},
		{
			name: "first in ToRemove in this, empty other",
			this: &model.UTXODiff{
				ToAdd:    model.UTXOCollection{},
				ToRemove: model.UTXOCollection{*outpoint0: utxoEntry1},
			},
			other: &model.UTXODiff{
				ToAdd:    model.UTXOCollection{},
				ToRemove: model.UTXOCollection{},
			},
			expectedDiffFromResult: &model.UTXODiff{
				ToAdd:    model.UTXOCollection{*outpoint0: utxoEntry1},
				ToRemove: model.UTXOCollection{},
			},
			expectedWithDiffResult: &model.UTXODiff{
				ToAdd:    model.UTXOCollection{},
				ToRemove: model.UTXOCollection{*outpoint0: utxoEntry1},
			},
		},
		{
			name: "first in ToAdd in this and other, second in ToRemove in this",
			this: &model.UTXODiff{
				ToAdd:    model.UTXOCollection{*outpoint0: utxoEntry1},
				ToRemove: model.UTXOCollection{*outpoint0: utxoEntry2},
			},
			other: &model.UTXODiff{
				ToAdd:    model.UTXOCollection{*outpoint0: utxoEntry1},
				ToRemove: model.UTXOCollection{},
			},
			expectedDiffFromResult: nil,
			expectedWithDiffResult: nil,
		},
		{
			name: "first in ToAdd in this, second in ToRemove in this and ToAdd in other",
			this: &model.UTXODiff{
				ToAdd:    model.UTXOCollection{*outpoint0: utxoEntry1},
				ToRemove: model.UTXOCollection{*outpoint0: utxoEntry2},
			},
			other: &model.UTXODiff{
				ToAdd:    model.UTXOCollection{*outpoint0: utxoEntry2},
				ToRemove: model.UTXOCollection{},
			},
			expectedDiffFromResult: nil,
			expectedWithDiffResult: nil,
		},
		{
			name: "first in ToAdd in this and ToRemove in other, second in ToRemove in this",
			this: &model.UTXODiff{
				ToAdd:    model.UTXOCollection{*outpoint0: utxoEntry1},
				ToRemove: model.UTXOCollection{*outpoint0: utxoEntry2},
			},
			other: &model.UTXODiff{
				ToAdd:    model.UTXOCollection{},
				ToRemove: model.UTXOCollection{*outpoint0: utxoEntry1},
			},
			expectedDiffFromResult: nil,
			expectedWithDiffResult: &model.UTXODiff{
				ToAdd:    model.UTXOCollection{},
				ToRemove: model.UTXOCollection{*outpoint0: utxoEntry2},
			},
		},
		{
			name: "first in ToAdd in this, second in ToRemove in this and in other",
			this: &model.UTXODiff{
				ToAdd:    model.UTXOCollection{*outpoint0: utxoEntry1},
				ToRemove: model.UTXOCollection{*outpoint0: utxoEntry2},
			},
			other: &model.UTXODiff{
				ToAdd:    model.UTXOCollection{},
				ToRemove: model.UTXOCollection{*outpoint0: utxoEntry2},
			},
			expectedDiffFromResult: &model.UTXODiff{
				ToAdd:    model.UTXOCollection{},
				ToRemove: model.UTXOCollection{*outpoint0: utxoEntry1},
			},
			expectedWithDiffResult: nil,
		},
		{
			name: "first in ToAdd and second in ToRemove in both this and other",
			this: &model.UTXODiff{
				ToAdd:    model.UTXOCollection{*outpoint0: utxoEntry1},
				ToRemove: model.UTXOCollection{*outpoint0: utxoEntry2},
			},
			other: &model.UTXODiff{
				ToAdd:    model.UTXOCollection{*outpoint0: utxoEntry1},
				ToRemove: model.UTXOCollection{*outpoint0: utxoEntry2},
			},
			expectedDiffFromResult: &model.UTXODiff{
				ToAdd:    model.UTXOCollection{},
				ToRemove: model.UTXOCollection{},
			},
			expectedWithDiffResult: nil,
		},
		{
			name: "first in ToAdd in this and ToRemove in other, second in ToRemove in this and ToAdd in other",
			this: &model.UTXODiff{
				ToAdd:    model.UTXOCollection{*outpoint0: utxoEntry1},
				ToRemove: model.UTXOCollection{*outpoint0: utxoEntry2},
			},
			other: &model.UTXODiff{
				ToAdd:    model.UTXOCollection{*outpoint0: utxoEntry2},
				ToRemove: model.UTXOCollection{*outpoint0: utxoEntry1},
			},
			expectedDiffFromResult: nil,
			expectedWithDiffResult: &model.UTXODiff{
				ToAdd:    model.UTXOCollection{},
				ToRemove: model.UTXOCollection{},
			},
		},
		{
			name: "first in ToAdd and second in ToRemove in this, empty other",
			this: &model.UTXODiff{
				ToAdd:    model.UTXOCollection{*outpoint0: utxoEntry1},
				ToRemove: model.UTXOCollection{*outpoint0: utxoEntry2},
			},
			other: &model.UTXODiff{
				ToAdd:    model.UTXOCollection{},
				ToRemove: model.UTXOCollection{},
			},
			expectedDiffFromResult: &model.UTXODiff{
				ToAdd:    model.UTXOCollection{*outpoint0: utxoEntry2},
				ToRemove: model.UTXOCollection{*outpoint0: utxoEntry1},
			},
			expectedWithDiffResult: &model.UTXODiff{
				ToAdd:    model.UTXOCollection{*outpoint0: utxoEntry1},
				ToRemove: model.UTXOCollection{*outpoint0: utxoEntry2},
			},
		},
		{
			name: "empty this, first in ToAdd in other",
			this: &model.UTXODiff{
				ToAdd:    model.UTXOCollection{},
				ToRemove: model.UTXOCollection{},
			},
			other: &model.UTXODiff{
				ToAdd:    model.UTXOCollection{*outpoint0: utxoEntry1},
				ToRemove: model.UTXOCollection{},
			},
			expectedDiffFromResult: &model.UTXODiff{
				ToAdd:    model.UTXOCollection{*outpoint0: utxoEntry1},
				ToRemove: model.UTXOCollection{},
			},
			expectedWithDiffResult: &model.UTXODiff{
				ToAdd:    model.UTXOCollection{*outpoint0: utxoEntry1},
				ToRemove: model.UTXOCollection{},
			},
		},
		{
			name: "empty this, first in ToRemove in other",
			this: &model.UTXODiff{
				ToAdd:    model.UTXOCollection{},
				ToRemove: model.UTXOCollection{},
			},
			other: &model.UTXODiff{
				ToAdd:    model.UTXOCollection{},
				ToRemove: model.UTXOCollection{*outpoint0: utxoEntry1},
			},
			expectedDiffFromResult: &model.UTXODiff{
				ToAdd:    model.UTXOCollection{},
				ToRemove: model.UTXOCollection{*outpoint0: utxoEntry1},
			},
			expectedWithDiffResult: &model.UTXODiff{
				ToAdd:    model.UTXOCollection{},
				ToRemove: model.UTXOCollection{*outpoint0: utxoEntry1},
			},
		},
		{
			name: "empty this, first in ToAdd and second in ToRemove in other",
			this: &model.UTXODiff{
				ToAdd:    model.UTXOCollection{},
				ToRemove: model.UTXOCollection{},
			},
			other: &model.UTXODiff{
				ToAdd:    model.UTXOCollection{*outpoint0: utxoEntry1},
				ToRemove: model.UTXOCollection{*outpoint0: utxoEntry2},
			},
			expectedDiffFromResult: &model.UTXODiff{
				ToAdd:    model.UTXOCollection{*outpoint0: utxoEntry1},
				ToRemove: model.UTXOCollection{*outpoint0: utxoEntry2},
			},
			expectedWithDiffResult: &model.UTXODiff{
				ToAdd:    model.UTXOCollection{*outpoint0: utxoEntry1},
				ToRemove: model.UTXOCollection{*outpoint0: utxoEntry2},
			},
		},
		{
			name: "empty this, empty other",
			this: &model.UTXODiff{
				ToAdd:    model.UTXOCollection{},
				ToRemove: model.UTXOCollection{},
			},
			other: &model.UTXODiff{
				ToAdd:    model.UTXOCollection{},
				ToRemove: model.UTXOCollection{},
			},
			expectedDiffFromResult: &model.UTXODiff{
				ToAdd:    model.UTXOCollection{},
				ToRemove: model.UTXOCollection{},
			},
			expectedWithDiffResult: &model.UTXODiff{
				ToAdd:    model.UTXOCollection{},
				ToRemove: model.UTXOCollection{},
			},
		},
	}

	for _, test := range tests {
		// diffFrom from test.this to test.other
		diffResult, err := DiffFrom(test.this, test.other)

		// Test whether diffFrom returned an error
		isDiffFromOk := err == nil
		expectedIsDiffFromOk := test.expectedDiffFromResult != nil
		if isDiffFromOk != expectedIsDiffFromOk {
			t.Errorf("unexpected diffFrom error in test \"%s\". "+
				"Expected: \"%t\", got: \"%t\".", test.name, expectedIsDiffFromOk, isDiffFromOk)
		}

		// If not error, test the diffFrom result
		if isDiffFromOk && !diffEqual(test.expectedDiffFromResult, diffResult) {
			t.Errorf("unexpected diffFrom result in test \"%s\". "+
				"Expected: \"%v\", got: \"%v\".", test.name, test.expectedDiffFromResult, diffResult)
		}

		// Make sure that WithDiff after diffFrom results in the original test.other
		if isDiffFromOk {
			otherResult, err := WithDiff(test.this, diffResult)
			if err != nil {
				t.Errorf("WithDiff unexpectedly failed in test \"%s\": %s", test.name, err)
			}
			if !diffEqual(test.other, otherResult) {
				t.Errorf("unexpected WithDiff result in test \"%s\". "+
					"Expected: \"%v\", got: \"%v\".", test.name, test.other, otherResult)
			}
		}

		// WithDiff from test.this to test.other
		withDiffResult, err := WithDiff(test.this, test.other)

		// Test whether WithDiff returned an error
		isWithDiffOk := err == nil
		expectedIsWithDiffOk := test.expectedWithDiffResult != nil
		if isWithDiffOk != expectedIsWithDiffOk {
			t.Errorf("unexpected WithDiff error in test \"%s\". "+
				"Expected: \"%t\", got: \"%t\".", test.name, expectedIsWithDiffOk, isWithDiffOk)
		}

		// If not error, test the WithDiff result
		if isWithDiffOk && !diffEqual(withDiffResult, test.expectedWithDiffResult) {
			t.Errorf("unexpected WithDiff result in test \"%s\". "+
				"Expected: \"%v\", got: \"%v\".", test.name, test.expectedWithDiffResult, withDiffResult)
		}

		// Repeat WithDiff check test.this time using withDiffInPlace
		thisClone := test.this.Clone()
		err = WithDiffInPlace(thisClone, test.other)

		// Test whether withDiffInPlace returned an error
		isWithDiffInPlaceOk := err == nil
		expectedIsWithDiffInPlaceOk := test.expectedWithDiffResult != nil
		if isWithDiffInPlaceOk != expectedIsWithDiffInPlaceOk {
			t.Errorf("unexpected withDiffInPlace error in test \"%s\". "+
				"Expected: \"%t\", got: \"%t\".", test.name, expectedIsWithDiffInPlaceOk, isWithDiffInPlaceOk)
		}

		// If not error, test the withDiffInPlace result
		if isWithDiffInPlaceOk && !diffEqual(thisClone, test.expectedWithDiffResult) {
			t.Errorf("unexpected withDiffInPlace result in test \"%s\". "+
				"Expected: \"%v\", got: \"%v\".", test.name, test.expectedWithDiffResult, thisClone)
		}

		// Make sure that diffFrom after WithDiff results in the original test.other
		if isWithDiffOk {
			otherResult, err := DiffFrom(test.this, withDiffResult)
			if err != nil {
				t.Errorf("diffFrom unexpectedly failed in test \"%s\": %s", test.name, err)
			}
			if !diffEqual(test.other, otherResult) {
				t.Errorf("unexpected diffFrom result in test \"%s\". "+
					"Expected: \"%v\", got: \"%v\".", test.name, test.other, otherResult)
			}
		}
	}
}
