package utxo

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/kaspanet/kaspad/infrastructure/config"
	"github.com/kaspanet/kaspad/infrastructure/db/dbaccess"
	"github.com/kaspanet/kaspad/util/subnetworkid"

	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/util/daghash"
)

func prepareDatabaseForTest(t *testing.T, testName string) (*dbaccess.DatabaseContext, func()) {
	var err error
	tmpDir, err := ioutil.TempDir("", "utxoset_test")
	if err != nil {
		t.Fatalf("error creating temp dir: %s", err)
		return nil, nil
	}

	dbPath := filepath.Join(tmpDir, testName)
	_ = os.RemoveAll(dbPath)
	databaseContext, err := dbaccess.New(dbPath)
	if err != nil {
		t.Fatalf("error creating db: %s", err)
		return nil, nil
	}

	// Setup a teardown function for cleaning up. This function is
	// returned to the caller to be invoked when it is done testing.
	teardown := func() {
		databaseContext.Close()
		os.RemoveAll(dbPath)
	}

	return databaseContext, teardown

}

// TestUTXOCollection makes sure that utxoCollection cloning and string representations work as expected.
func TestUTXOCollection(t *testing.T) {
	txID0, _ := daghash.NewTxIDFromStr("0000000000000000000000000000000000000000000000000000000000000000")
	txID1, _ := daghash.NewTxIDFromStr("1111111111111111111111111111111111111111111111111111111111111111")
	outpoint0 := *appmessage.NewOutpoint(txID0, 0)
	outpoint1 := *appmessage.NewOutpoint(txID1, 0)
	utxoEntry0 := NewEntry(&appmessage.TxOut{ScriptPubKey: []byte{}, Value: 10}, true, 0)
	utxoEntry1 := NewEntry(&appmessage.TxOut{ScriptPubKey: []byte{}, Value: 20}, false, 1)

	// For each of the following test cases, we will:
	// .String() the given collection and compare it to expectedStringWithMultiset
	// .Clone() the given collection and compare its value to itself (expected: equals) and its reference to itself (expected: not equal)
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
				outpoint0: utxoEntry1,
			},
			expectedString: "[ (0000000000000000000000000000000000000000000000000000000000000000, 0) => 20, blueScore: 1 ]",
		},
		{
			name: "two members",
			collection: utxoCollection{
				outpoint0: utxoEntry0,
				outpoint1: utxoEntry1,
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
		collectionClone := test.collection.clone()
		if reflect.ValueOf(collectionClone).Pointer() == reflect.ValueOf(test.collection).Pointer() {
			t.Errorf("collection is reference-equal to its Clone in test \"%s\". ", test.name)
		}
		if !reflect.DeepEqual(test.collection, collectionClone) {
			t.Errorf("collection is not equal to its Clone in test \"%s\". "+
				"Expected: \"%s\", got: \"%s\".", test.name, collectionString, collectionClone.String())
		}
	}
}

// TestUTXODiff makes sure that utxoDiff creation, cloning, and string representations work as expected.
func TestUTXODiff(t *testing.T) {
	txID0, _ := daghash.NewTxIDFromStr("0000000000000000000000000000000000000000000000000000000000000000")
	txID1, _ := daghash.NewTxIDFromStr("1111111111111111111111111111111111111111111111111111111111111111")
	outpoint0 := *appmessage.NewOutpoint(txID0, 0)
	outpoint1 := *appmessage.NewOutpoint(txID1, 0)
	utxoEntry0 := NewEntry(&appmessage.TxOut{ScriptPubKey: []byte{}, Value: 10}, true, 0)
	utxoEntry1 := NewEntry(&appmessage.TxOut{ScriptPubKey: []byte{}, Value: 20}, false, 1)

	// Test utxoDiff creation

	diff := NewDiff()

	if len(diff.ToAdd) != 0 || len(diff.ToRemove) != 0 {
		t.Errorf("new Diff is not empty")
	}

	err := diff.AddEntry(outpoint0, utxoEntry0)
	if err != nil {
		t.Fatalf("error adding entry to utxo Diff: %s", err)
	}

	err = diff.RemoveEntry(outpoint1, utxoEntry1)
	if err != nil {
		t.Fatalf("error adding entry to utxo Diff: %s", err)
	}

	// Test utxoDiff cloning
	clonedDiff := diff.Clone()
	if clonedDiff == diff {
		t.Errorf("cloned Diff is reference-equal to the original")
	}
	if !reflect.DeepEqual(clonedDiff, diff) {
		t.Errorf("cloned Diff not equal to the original"+
			"Original: \"%v\", cloned: \"%v\".", diff, clonedDiff)
	}

	// Test utxoDiff string representation
	expectedDiffString := "ToAdd: [ (0000000000000000000000000000000000000000000000000000000000000000, 0) => 10, blueScore: 0 ]; ToRemove: [ (1111111111111111111111111111111111111111111111111111111111111111, 0) => 20, blueScore: 1 ]"
	diffString := clonedDiff.String()
	if diffString != expectedDiffString {
		t.Errorf("unexpected Diff string. "+
			"Expected: \"%s\", got: \"%s\".", expectedDiffString, diffString)
	}
}

// TestUTXODiffRules makes sure that all DiffFrom and WithDiff rules are followed.
// Each test case represents a cell in the two tables outlined in the documentation for utxoDiff.
func TestUTXODiffRules(t *testing.T) {
	txID0, _ := daghash.NewTxIDFromStr("0000000000000000000000000000000000000000000000000000000000000000")
	outpoint0 := *appmessage.NewOutpoint(txID0, 0)
	utxoEntry1 := NewEntry(&appmessage.TxOut{ScriptPubKey: []byte{}, Value: 10}, true, 10)
	utxoEntry2 := NewEntry(&appmessage.TxOut{ScriptPubKey: []byte{}, Value: 10}, true, 20)

	// For each of the following test cases, we will:
	// this.DiffFrom(other) and compare it to expectedDiffFromResult
	// this.WithDiff(other) and compare it to expectedWithDiffResult
	// this.WithDiffInPlace(other) and compare it to expectedWithDiffResult
	//
	// Note: an expected nil result means that we expect the respective operation to fail
	// See the following spreadsheet for a summary of all test-cases:
	// https://docs.google.com/spreadsheets/d/1E8G3mp5y1-yifouwLLXRLueSRfXdDRwRKFieYE07buY/edit?usp=sharing
	tests := []struct {
		name                   string
		this                   *Diff
		other                  *Diff
		expectedDiffFromResult *Diff
		expectedWithDiffResult *Diff
	}{
		{
			name: "first ToAdd in this, first ToAdd in other",
			this: &Diff{
				ToAdd:    utxoCollection{outpoint0: utxoEntry1},
				ToRemove: utxoCollection{},
			},
			other: &Diff{
				ToAdd:    utxoCollection{outpoint0: utxoEntry1},
				ToRemove: utxoCollection{},
			},
			expectedDiffFromResult: &Diff{
				ToAdd:    utxoCollection{},
				ToRemove: utxoCollection{},
			},
			expectedWithDiffResult: nil,
		},
		{
			name: "first in ToAdd in this, second in ToAdd in other",
			this: &Diff{
				ToAdd:    utxoCollection{outpoint0: utxoEntry1},
				ToRemove: utxoCollection{},
			},
			other: &Diff{
				ToAdd:    utxoCollection{outpoint0: utxoEntry2},
				ToRemove: utxoCollection{},
			},
			expectedDiffFromResult: &Diff{
				ToAdd:    utxoCollection{outpoint0: utxoEntry2},
				ToRemove: utxoCollection{outpoint0: utxoEntry1},
			},
			expectedWithDiffResult: nil,
		},
		{
			name: "first in ToAdd in this, second in ToRemove in other",
			this: &Diff{
				ToAdd:    utxoCollection{outpoint0: utxoEntry1},
				ToRemove: utxoCollection{},
			},
			other: &Diff{
				ToAdd:    utxoCollection{},
				ToRemove: utxoCollection{outpoint0: utxoEntry1},
			},
			expectedDiffFromResult: nil,
			expectedWithDiffResult: &Diff{
				ToAdd:    utxoCollection{},
				ToRemove: utxoCollection{},
			},
		},
		{
			name: "first in ToAdd in this and other, second in ToRemove in other",
			this: &Diff{
				ToAdd:    utxoCollection{outpoint0: utxoEntry1},
				ToRemove: utxoCollection{},
			},
			other: &Diff{
				ToAdd:    utxoCollection{outpoint0: utxoEntry1},
				ToRemove: utxoCollection{outpoint0: utxoEntry2},
			},
			expectedDiffFromResult: nil,
			expectedWithDiffResult: nil,
		},
		{
			name: "first in ToAdd in this and ToRemove in other, second in ToAdd in other",
			this: &Diff{
				ToAdd:    utxoCollection{outpoint0: utxoEntry1},
				ToRemove: utxoCollection{},
			},
			other: &Diff{
				ToAdd:    utxoCollection{outpoint0: utxoEntry2},
				ToRemove: utxoCollection{outpoint0: utxoEntry1},
			},
			expectedDiffFromResult: nil,
			expectedWithDiffResult: &Diff{
				ToAdd:    utxoCollection{outpoint0: utxoEntry2},
				ToRemove: utxoCollection{},
			},
		},
		{
			name: "first in ToAdd in this, empty other",
			this: &Diff{
				ToAdd:    utxoCollection{outpoint0: utxoEntry1},
				ToRemove: utxoCollection{},
			},
			other: &Diff{
				ToAdd:    utxoCollection{},
				ToRemove: utxoCollection{},
			},
			expectedDiffFromResult: &Diff{
				ToAdd:    utxoCollection{},
				ToRemove: utxoCollection{outpoint0: utxoEntry1},
			},
			expectedWithDiffResult: &Diff{
				ToAdd:    utxoCollection{outpoint0: utxoEntry1},
				ToRemove: utxoCollection{},
			},
		},
		{
			name: "first in ToRemove in this and in ToAdd in other",
			this: &Diff{
				ToAdd:    utxoCollection{},
				ToRemove: utxoCollection{outpoint0: utxoEntry1},
			},
			other: &Diff{
				ToAdd:    utxoCollection{outpoint0: utxoEntry1},
				ToRemove: utxoCollection{},
			},
			expectedDiffFromResult: nil,
			expectedWithDiffResult: &Diff{
				ToAdd:    utxoCollection{},
				ToRemove: utxoCollection{},
			},
		},
		{
			name: "first in ToRemove in this, second in ToAdd in other",
			this: &Diff{
				ToAdd:    utxoCollection{},
				ToRemove: utxoCollection{outpoint0: utxoEntry1},
			},
			other: &Diff{
				ToAdd:    utxoCollection{outpoint0: utxoEntry2},
				ToRemove: utxoCollection{},
			},
			expectedDiffFromResult: nil,
			expectedWithDiffResult: &Diff{
				ToAdd:    utxoCollection{outpoint0: utxoEntry2},
				ToRemove: utxoCollection{outpoint0: utxoEntry1},
			},
		},
		{
			name: "first in ToRemove in this and other",
			this: &Diff{
				ToAdd:    utxoCollection{},
				ToRemove: utxoCollection{outpoint0: utxoEntry1},
			},
			other: &Diff{
				ToAdd:    utxoCollection{},
				ToRemove: utxoCollection{outpoint0: utxoEntry1},
			},
			expectedDiffFromResult: &Diff{
				ToAdd:    utxoCollection{},
				ToRemove: utxoCollection{},
			},
			expectedWithDiffResult: nil,
		},
		{
			name: "first in ToRemove in this, second in ToRemove in other",
			this: &Diff{
				ToAdd:    utxoCollection{},
				ToRemove: utxoCollection{outpoint0: utxoEntry1},
			},
			other: &Diff{
				ToAdd:    utxoCollection{},
				ToRemove: utxoCollection{outpoint0: utxoEntry2},
			},
			expectedDiffFromResult: nil,
			expectedWithDiffResult: nil,
		},
		{
			name: "first in ToRemove in this and ToAdd in other, second in ToRemove in other",
			this: &Diff{
				ToAdd:    utxoCollection{},
				ToRemove: utxoCollection{outpoint0: utxoEntry1},
			},
			other: &Diff{
				ToAdd:    utxoCollection{outpoint0: utxoEntry1},
				ToRemove: utxoCollection{outpoint0: utxoEntry2},
			},
			expectedDiffFromResult: nil,
			expectedWithDiffResult: nil,
		},
		{
			name: "first in ToRemove in this and other, second in ToAdd in other",
			this: &Diff{
				ToAdd:    utxoCollection{},
				ToRemove: utxoCollection{outpoint0: utxoEntry1},
			},
			other: &Diff{
				ToAdd:    utxoCollection{outpoint0: utxoEntry2},
				ToRemove: utxoCollection{outpoint0: utxoEntry1},
			},
			expectedDiffFromResult: &Diff{
				ToAdd:    utxoCollection{outpoint0: utxoEntry2},
				ToRemove: utxoCollection{},
			},
			expectedWithDiffResult: nil,
		},
		{
			name: "first in ToRemove in this, empty other",
			this: &Diff{
				ToAdd:    utxoCollection{},
				ToRemove: utxoCollection{outpoint0: utxoEntry1},
			},
			other: &Diff{
				ToAdd:    utxoCollection{},
				ToRemove: utxoCollection{},
			},
			expectedDiffFromResult: &Diff{
				ToAdd:    utxoCollection{outpoint0: utxoEntry1},
				ToRemove: utxoCollection{},
			},
			expectedWithDiffResult: &Diff{
				ToAdd:    utxoCollection{},
				ToRemove: utxoCollection{outpoint0: utxoEntry1},
			},
		},
		{
			name: "first in ToAdd in this and other, second in ToRemove in this",
			this: &Diff{
				ToAdd:    utxoCollection{outpoint0: utxoEntry1},
				ToRemove: utxoCollection{outpoint0: utxoEntry2},
			},
			other: &Diff{
				ToAdd:    utxoCollection{outpoint0: utxoEntry1},
				ToRemove: utxoCollection{},
			},
			expectedDiffFromResult: nil,
			expectedWithDiffResult: nil,
		},
		{
			name: "first in ToAdd in this, second in ToRemove in this and ToAdd in other",
			this: &Diff{
				ToAdd:    utxoCollection{outpoint0: utxoEntry1},
				ToRemove: utxoCollection{outpoint0: utxoEntry2},
			},
			other: &Diff{
				ToAdd:    utxoCollection{outpoint0: utxoEntry2},
				ToRemove: utxoCollection{},
			},
			expectedDiffFromResult: nil,
			expectedWithDiffResult: nil,
		},
		{
			name: "first in ToAdd in this and ToRemove in other, second in ToRemove in this",
			this: &Diff{
				ToAdd:    utxoCollection{outpoint0: utxoEntry1},
				ToRemove: utxoCollection{outpoint0: utxoEntry2},
			},
			other: &Diff{
				ToAdd:    utxoCollection{},
				ToRemove: utxoCollection{outpoint0: utxoEntry1},
			},
			expectedDiffFromResult: nil,
			expectedWithDiffResult: &Diff{
				ToAdd:    utxoCollection{},
				ToRemove: utxoCollection{outpoint0: utxoEntry2},
			},
		},
		{
			name: "first in ToAdd in this, second in ToRemove in this and in other",
			this: &Diff{
				ToAdd:    utxoCollection{outpoint0: utxoEntry1},
				ToRemove: utxoCollection{outpoint0: utxoEntry2},
			},
			other: &Diff{
				ToAdd:    utxoCollection{},
				ToRemove: utxoCollection{outpoint0: utxoEntry2},
			},
			expectedDiffFromResult: &Diff{
				ToAdd:    utxoCollection{},
				ToRemove: utxoCollection{outpoint0: utxoEntry1},
			},
			expectedWithDiffResult: nil,
		},
		{
			name: "first in ToAdd and second in ToRemove in both this and other",
			this: &Diff{
				ToAdd:    utxoCollection{outpoint0: utxoEntry1},
				ToRemove: utxoCollection{outpoint0: utxoEntry2},
			},
			other: &Diff{
				ToAdd:    utxoCollection{outpoint0: utxoEntry1},
				ToRemove: utxoCollection{outpoint0: utxoEntry2},
			},
			expectedDiffFromResult: &Diff{
				ToAdd:    utxoCollection{},
				ToRemove: utxoCollection{},
			},
			expectedWithDiffResult: nil,
		},
		{
			name: "first in ToAdd in this and ToRemove in other, second in ToRemove in this and ToAdd in other",
			this: &Diff{
				ToAdd:    utxoCollection{outpoint0: utxoEntry1},
				ToRemove: utxoCollection{outpoint0: utxoEntry2},
			},
			other: &Diff{
				ToAdd:    utxoCollection{outpoint0: utxoEntry2},
				ToRemove: utxoCollection{outpoint0: utxoEntry1},
			},
			expectedDiffFromResult: nil,
			expectedWithDiffResult: &Diff{
				ToAdd:    utxoCollection{},
				ToRemove: utxoCollection{},
			},
		},
		{
			name: "first in ToAdd and second in ToRemove in this, empty other",
			this: &Diff{
				ToAdd:    utxoCollection{outpoint0: utxoEntry1},
				ToRemove: utxoCollection{outpoint0: utxoEntry2},
			},
			other: &Diff{
				ToAdd:    utxoCollection{},
				ToRemove: utxoCollection{},
			},
			expectedDiffFromResult: &Diff{
				ToAdd:    utxoCollection{outpoint0: utxoEntry2},
				ToRemove: utxoCollection{outpoint0: utxoEntry1},
			},
			expectedWithDiffResult: &Diff{
				ToAdd:    utxoCollection{outpoint0: utxoEntry1},
				ToRemove: utxoCollection{outpoint0: utxoEntry2},
			},
		},
		{
			name: "empty this, first in ToAdd in other",
			this: &Diff{
				ToAdd:    utxoCollection{},
				ToRemove: utxoCollection{},
			},
			other: &Diff{
				ToAdd:    utxoCollection{outpoint0: utxoEntry1},
				ToRemove: utxoCollection{},
			},
			expectedDiffFromResult: &Diff{
				ToAdd:    utxoCollection{outpoint0: utxoEntry1},
				ToRemove: utxoCollection{},
			},
			expectedWithDiffResult: &Diff{
				ToAdd:    utxoCollection{outpoint0: utxoEntry1},
				ToRemove: utxoCollection{},
			},
		},
		{
			name: "empty this, first in ToRemove in other",
			this: &Diff{
				ToAdd:    utxoCollection{},
				ToRemove: utxoCollection{},
			},
			other: &Diff{
				ToAdd:    utxoCollection{},
				ToRemove: utxoCollection{outpoint0: utxoEntry1},
			},
			expectedDiffFromResult: &Diff{
				ToAdd:    utxoCollection{},
				ToRemove: utxoCollection{outpoint0: utxoEntry1},
			},
			expectedWithDiffResult: &Diff{
				ToAdd:    utxoCollection{},
				ToRemove: utxoCollection{outpoint0: utxoEntry1},
			},
		},
		{
			name: "empty this, first in ToAdd and second in ToRemove in other",
			this: &Diff{
				ToAdd:    utxoCollection{},
				ToRemove: utxoCollection{},
			},
			other: &Diff{
				ToAdd:    utxoCollection{outpoint0: utxoEntry1},
				ToRemove: utxoCollection{outpoint0: utxoEntry2},
			},
			expectedDiffFromResult: &Diff{
				ToAdd:    utxoCollection{outpoint0: utxoEntry1},
				ToRemove: utxoCollection{outpoint0: utxoEntry2},
			},
			expectedWithDiffResult: &Diff{
				ToAdd:    utxoCollection{outpoint0: utxoEntry1},
				ToRemove: utxoCollection{outpoint0: utxoEntry2},
			},
		},
		{
			name: "empty this, empty other",
			this: &Diff{
				ToAdd:    utxoCollection{},
				ToRemove: utxoCollection{},
			},
			other: &Diff{
				ToAdd:    utxoCollection{},
				ToRemove: utxoCollection{},
			},
			expectedDiffFromResult: &Diff{
				ToAdd:    utxoCollection{},
				ToRemove: utxoCollection{},
			},
			expectedWithDiffResult: &Diff{
				ToAdd:    utxoCollection{},
				ToRemove: utxoCollection{},
			},
		},
	}

	for _, test := range tests {
		// DiffFrom from test.this to test.other
		diffResult, err := test.this.diffFrom(test.other)

		// Test whether DiffFrom returned an error
		isDiffFromOk := err == nil
		expectedIsDiffFromOk := test.expectedDiffFromResult != nil
		if isDiffFromOk != expectedIsDiffFromOk {
			t.Errorf("unexpected DiffFrom error in test \"%s\". "+
				"Expected: \"%t\", got: \"%t\".", test.name, expectedIsDiffFromOk, isDiffFromOk)
		}

		// If not error, test the DiffFrom result
		if isDiffFromOk && !test.expectedDiffFromResult.equal(diffResult) {
			t.Errorf("unexpected DiffFrom result in test \"%s\". "+
				"Expected: \"%v\", got: \"%v\".", test.name, test.expectedDiffFromResult, diffResult)
		}

		// Make sure that WithDiff after DiffFrom results in the original test.other
		if isDiffFromOk {
			otherResult, err := test.this.WithDiff(diffResult)
			if err != nil {
				t.Errorf("WithDiff unexpectedly failed in test \"%s\": %s", test.name, err)
			}
			if !test.other.equal(otherResult) {
				t.Errorf("unexpected WithDiff result in test \"%s\". "+
					"Expected: \"%v\", got: \"%v\".", test.name, test.other, otherResult)
			}
		}

		// WithDiff from test.this to test.other
		withDiffResult, err := test.this.WithDiff(test.other)

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

		// Repeat WithDiff check test.this time using WithDiffInPlace
		thisClone := test.this.Clone()
		err = thisClone.WithDiffInPlace(test.other)

		// Test whether WithDiffInPlace returned an error
		isWithDiffInPlaceOk := err == nil
		expectedIsWithDiffInPlaceOk := test.expectedWithDiffResult != nil
		if isWithDiffInPlaceOk != expectedIsWithDiffInPlaceOk {
			t.Errorf("unexpected WithDiffInPlace error in test \"%s\". "+
				"Expected: \"%t\", got: \"%t\".", test.name, expectedIsWithDiffInPlaceOk, isWithDiffInPlaceOk)
		}

		// If not error, test the WithDiffInPlace result
		if isWithDiffInPlaceOk && !thisClone.equal(test.expectedWithDiffResult) {
			t.Errorf("unexpected WithDiffInPlace result in test \"%s\". "+
				"Expected: \"%v\", got: \"%v\".", test.name, test.expectedWithDiffResult, thisClone)
		}

		// Make sure that DiffFrom after WithDiff results in the original test.other
		if isWithDiffOk {
			otherResult, err := test.this.diffFrom(withDiffResult)
			if err != nil {
				t.Errorf("DiffFrom unexpectedly failed in test \"%s\": %s", test.name, err)
			}
			if !test.other.equal(otherResult) {
				t.Errorf("unexpected DiffFrom result in test \"%s\". "+
					"Expected: \"%v\", got: \"%v\".", test.name, test.other, otherResult)
			}
		}
	}
}

func (d *Diff) equal(other *Diff) bool {
	if d == nil || other == nil {
		return d == other
	}

	return reflect.DeepEqual(d.ToAdd, other.ToAdd) &&
		reflect.DeepEqual(d.ToRemove, other.ToRemove)
}

func (fus *FullUTXOSet) equal(other *FullUTXOSet) bool {
	return reflect.DeepEqual(fus.UTXOCache, other.UTXOCache)
}

func (dus *DiffUTXOSet) equal(other *DiffUTXOSet) bool {
	return dus.Base.equal(other.Base) && dus.UTXODiff.equal(other.UTXODiff)
}

// TestFullUTXOSet makes sure that fullUTXOSet is working as expected.
func TestFullUTXOSet(t *testing.T) {
	txID0, _ := daghash.NewTxIDFromStr("0000000000000000000000000000000000000000000000000000000000000000")
	txID1, _ := daghash.NewTxIDFromStr("1111111111111111111111111111111111111111111111111111111111111111")
	outpoint0 := *appmessage.NewOutpoint(txID0, 0)
	outpoint1 := *appmessage.NewOutpoint(txID1, 0)
	txOut0 := &appmessage.TxOut{ScriptPubKey: []byte{}, Value: 10}
	txOut1 := &appmessage.TxOut{ScriptPubKey: []byte{}, Value: 20}
	utxoEntry0 := NewEntry(txOut0, true, 0)
	utxoEntry1 := NewEntry(txOut1, false, 1)
	diff := &Diff{
		ToAdd:    utxoCollection{outpoint0: utxoEntry0},
		ToRemove: utxoCollection{outpoint1: utxoEntry1},
	}

	// Test fullUTXOSet creation
	fullUTXOCacheSize := config.DefaultConfig().MaxUTXOCacheSize
	db, teardown := prepareDatabaseForTest(t, "TestDiffUTXOSet")
	defer teardown()
	emptySet := NewFullUTXOSetFromContext(db, fullUTXOCacheSize)
	if len(emptySet.collection()) != 0 {
		t.Errorf("new set is not empty")
	}

	// Test fullUTXOSet WithDiff
	withDiffResult, err := emptySet.WithDiff(diff)
	if err != nil {
		t.Errorf("WithDiff unexpectedly failed")
	}
	withDiffUTXOSet, ok := withDiffResult.(*DiffUTXOSet)
	if !ok {
		t.Errorf("WithDiff is of unexpected type")
	}
	if !reflect.DeepEqual(withDiffUTXOSet.Base, emptySet) || !reflect.DeepEqual(withDiffUTXOSet.UTXODiff, diff) {
		t.Errorf("WithDiff is of unexpected composition")
	}

	// Test fullUTXOSet addTx
	txIn0 := &appmessage.TxIn{SignatureScript: []byte{}, PreviousOutpoint: appmessage.Outpoint{TxID: *txID0, Index: 0}, Sequence: 0}
	transaction0 := appmessage.NewNativeMsgTx(1, []*appmessage.TxIn{txIn0}, []*appmessage.TxOut{txOut0})
	if isAccepted, err := emptySet.AddTx(transaction0, 0); err != nil {
		t.Errorf("AddTx unexpectedly failed: %s", err)
	} else if isAccepted {
		t.Errorf("addTx unexpectedly succeeded")
	}
	emptySet = NewFullUTXOSetFromContext(db, fullUTXOCacheSize)
	emptySet.Add(outpoint0, utxoEntry0)
	if isAccepted, err := emptySet.AddTx(transaction0, 0); err != nil {
		t.Errorf("addTx unexpectedly failed. Error: %s", err)
	} else if !isAccepted {
		t.Fatalf("AddTx unexpectedly didn't Add tx %s", transaction0.TxID())
	}

	// Test fullUTXOSet collection
	if !reflect.DeepEqual(emptySet.collection(), emptySet.UTXOCache) {
		t.Errorf("collection does not equal the set's utxoCollection")
	}

	// Test fullUTXOSet cloning
	clonedEmptySet := emptySet.Clone().(*FullUTXOSet)
	if !reflect.DeepEqual(clonedEmptySet, emptySet) {
		t.Errorf("Clone does not equal the original set")
	}
	if clonedEmptySet == emptySet {
		t.Errorf("cloned set is reference-equal to the original")
	}
}

// TestDiffUTXOSet makes sure that diffUTXOSet is working as expected.
func TestDiffUTXOSet(t *testing.T) {
	txID0, _ := daghash.NewTxIDFromStr("0000000000000000000000000000000000000000000000000000000000000000")
	txID1, _ := daghash.NewTxIDFromStr("1111111111111111111111111111111111111111111111111111111111111111")
	outpoint0 := *appmessage.NewOutpoint(txID0, 0)
	outpoint1 := *appmessage.NewOutpoint(txID1, 0)
	txOut0 := &appmessage.TxOut{ScriptPubKey: []byte{}, Value: 10}
	txOut1 := &appmessage.TxOut{ScriptPubKey: []byte{}, Value: 20}
	utxoEntry0 := NewEntry(txOut0, true, 0)
	utxoEntry1 := NewEntry(txOut1, false, 1)
	diff := &Diff{
		ToAdd:    utxoCollection{outpoint0: utxoEntry0},
		ToRemove: utxoCollection{outpoint1: utxoEntry1},
	}
	fullUTXOCacheSize := config.DefaultConfig().MaxUTXOCacheSize
	db, teardown := prepareDatabaseForTest(t, "TestDiffUTXOSet")
	defer teardown()

	// Test diffUTXOSet creation
	emptySet := NewDiffUTXOSet(NewFullUTXOSetFromContext(db, fullUTXOCacheSize), NewDiff())
	if collection, err := emptySet.collection(); err != nil {
		t.Errorf("Error getting emptySet collection: %s", err)
	} else if len(collection) != 0 {
		t.Errorf("new set is not empty")
	}

	// Test diffUTXOSet WithDiff
	withDiffResult, err := emptySet.WithDiff(diff)
	if err != nil {
		t.Errorf("WithDiff unexpectedly failed")
	}
	withDiffUTXOSet, ok := withDiffResult.(*DiffUTXOSet)
	if !ok {
		t.Errorf("WithDiff is of unexpected type")
	}
	withDiff, _ := NewDiff().WithDiff(diff)
	if !reflect.DeepEqual(withDiffUTXOSet.Base, emptySet.Base) || !reflect.DeepEqual(withDiffUTXOSet.UTXODiff, withDiff) {
		t.Errorf("WithDiff is of unexpected composition")
	}
	_, err = NewDiffUTXOSet(NewFullUTXOSetFromContext(db, fullUTXOCacheSize), diff).WithDiff(diff)
	if err == nil {
		t.Errorf("WithDiff unexpectedly succeeded")
	}

	// Given a diffSet, each case tests that MeldToBase, String, collection, and cloning work as expected
	// For each of the following test cases, we will:
	// .MeldToBase() the given diffSet and compare it to expectedMeldSet
	// .String() the given diffSet and compare it to expectedString
	// .collection() the given diffSet and compare it to expectedCollection
	// .Clone() the given diffSet and compare its value to itself (expected: equals) and its reference to itself (expected: not equal)
	tests := []struct {
		name                    string
		diffSet                 *DiffUTXOSet
		expectedMeldSet         *DiffUTXOSet
		expectedString          string
		expectedCollection      utxoCollection
		expectedMeldToBaseError string
	}{
		{
			name: "empty Base, empty Diff",
			diffSet: &DiffUTXOSet{
				Base: NewFullUTXOSetFromContext(db, fullUTXOCacheSize),
				UTXODiff: &Diff{
					ToAdd:    utxoCollection{},
					ToRemove: utxoCollection{},
				},
			},
			expectedMeldSet: &DiffUTXOSet{
				Base: NewFullUTXOSetFromContext(db, fullUTXOCacheSize),
				UTXODiff: &Diff{
					ToAdd:    utxoCollection{},
					ToRemove: utxoCollection{},
				},
			},
			expectedString:     "{Base: [  ], To Add: [  ], To Remove: [  ]}",
			expectedCollection: utxoCollection{},
		},
		{
			name: "empty Base, one member in Diff ToAdd",
			diffSet: &DiffUTXOSet{
				Base: NewFullUTXOSetFromContext(db, fullUTXOCacheSize),
				UTXODiff: &Diff{
					ToAdd:    utxoCollection{outpoint0: utxoEntry0},
					ToRemove: utxoCollection{},
				},
			},
			expectedMeldSet: &DiffUTXOSet{
				Base: func() *FullUTXOSet {
					futxo := NewFullUTXOSetFromContext(db, fullUTXOCacheSize)
					futxo.Add(outpoint0, utxoEntry0)
					return futxo
				}(),
				UTXODiff: &Diff{
					ToAdd:    utxoCollection{},
					ToRemove: utxoCollection{},
				},
			},
			expectedString:     "{Base: [  ], To Add: [ (0000000000000000000000000000000000000000000000000000000000000000, 0) => 10, blueScore: 0 ], To Remove: [  ]}",
			expectedCollection: utxoCollection{outpoint0: utxoEntry0},
		},
		{
			name: "empty Base, one member in Diff ToRemove",
			diffSet: &DiffUTXOSet{
				Base: NewFullUTXOSetFromContext(db, fullUTXOCacheSize),
				UTXODiff: &Diff{
					ToAdd:    utxoCollection{},
					ToRemove: utxoCollection{outpoint0: utxoEntry0},
				},
			},
			expectedMeldSet:         nil,
			expectedString:          "{Base: [  ], To Add: [  ], To Remove: [ (0000000000000000000000000000000000000000000000000000000000000000, 0) => 10, blueScore: 0 ]}",
			expectedCollection:      utxoCollection{},
			expectedMeldToBaseError: "Couldn't remove outpoint 0000000000000000000000000000000000000000000000000000000000000000:0 because it doesn't exist in the DiffUTXOSet Base",
		},
		{
			name: "one member in Base ToAdd, one member in Diff ToAdd",
			diffSet: &DiffUTXOSet{
				Base: func() *FullUTXOSet {
					futxo := NewFullUTXOSetFromContext(db, fullUTXOCacheSize)
					futxo.Add(outpoint0, utxoEntry0)
					return futxo
				}(),
				UTXODiff: &Diff{
					ToAdd:    utxoCollection{outpoint1: utxoEntry1},
					ToRemove: utxoCollection{},
				},
			},
			expectedMeldSet: &DiffUTXOSet{
				Base: func() *FullUTXOSet {
					futxo := NewFullUTXOSetFromContext(db, fullUTXOCacheSize)
					futxo.Add(outpoint0, utxoEntry0)
					futxo.Add(outpoint1, utxoEntry1)
					return futxo
				}(),
				UTXODiff: &Diff{
					ToAdd:    utxoCollection{},
					ToRemove: utxoCollection{},
				},
			},
			expectedString: "{Base: [ (0000000000000000000000000000000000000000000000000000000000000000, 0) => 10, blueScore: 0 ], To Add: [ (1111111111111111111111111111111111111111111111111111111111111111, 0) => 20, blueScore: 1 ], To Remove: [  ]}",
			expectedCollection: utxoCollection{
				outpoint0: utxoEntry0,
				outpoint1: utxoEntry1,
			},
		},
		{
			name: "one member in Base ToAdd, same one member in Diff ToRemove",
			diffSet: &DiffUTXOSet{
				Base: func() *FullUTXOSet {
					futxo := NewFullUTXOSetFromContext(db, fullUTXOCacheSize)
					futxo.Add(outpoint0, utxoEntry0)
					return futxo
				}(),
				UTXODiff: &Diff{
					ToAdd:    utxoCollection{},
					ToRemove: utxoCollection{outpoint0: utxoEntry0},
				},
			},
			expectedMeldSet: &DiffUTXOSet{
				Base: NewFullUTXOSetFromContext(db, fullUTXOCacheSize),
				UTXODiff: &Diff{
					ToAdd:    utxoCollection{},
					ToRemove: utxoCollection{},
				},
			},
			expectedString:     "{Base: [ (0000000000000000000000000000000000000000000000000000000000000000, 0) => 10, blueScore: 0 ], To Add: [  ], To Remove: [ (0000000000000000000000000000000000000000000000000000000000000000, 0) => 10, blueScore: 0 ]}",
			expectedCollection: utxoCollection{},
		},
	}

	for _, test := range tests {
		// Test string representation
		setString := test.diffSet.String()
		if setString != test.expectedString {
			t.Errorf("unexpected string in test \"%s\". "+
				"Expected: \"%s\", got: \"%s\".", test.name, test.expectedString, setString)
		}

		// Test MeldToBase
		meldSet := test.diffSet.Clone().(*DiffUTXOSet)
		err := meldSet.MeldToBase()
		errString := ""
		if err != nil {
			errString = err.Error()
		}
		if test.expectedMeldToBaseError != errString {
			t.Errorf("MeldToBase in test \"%s\" expected error \"%s\" but got: \"%s\"", test.name, test.expectedMeldToBaseError, errString)
		}
		if err != nil {
			continue
		}
		if !meldSet.equal(test.expectedMeldSet) {
			t.Errorf("unexpected melded set in test \"%s\". "+
				"Expected: \"%v\", got: \"%v\".", test.name, test.expectedMeldSet, meldSet)
		}

		// Test collection
		setCollection, err := test.diffSet.collection()
		if err != nil {
			t.Errorf("Error getting test.diffSet collection: %s", err)
		} else if !reflect.DeepEqual(setCollection, test.expectedCollection) {
			t.Errorf("unexpected set collection in test \"%s\". "+
				"Expected: \"%v\", got: \"%v\".", test.name, test.expectedCollection, setCollection)
		}

		// Test cloning
		clonedSet := test.diffSet.Clone().(*DiffUTXOSet)
		if !reflect.DeepEqual(clonedSet, test.diffSet) {
			t.Errorf("unexpected set Clone in test \"%s\". "+
				"Expected: \"%v\", got: \"%v\".", test.name, test.diffSet, clonedSet)
		}
		if clonedSet == test.diffSet {
			t.Errorf("cloned set is reference-equal to the original")
		}
	}
}

// TestUTXOSetDiffRules makes sure that utxoSet DiffFrom rules are followed.
// The rules are:
// 1. Neither fullUTXOSet nor diffUTXOSet can DiffFrom a fullUTXOSet.
// 2. fullUTXOSet cannot DiffFrom a diffUTXOSet with a Base other that itself.
// 3. diffUTXOSet cannot DiffFrom a diffUTXOSet with a different Base.
func TestUTXOSetDiffRules(t *testing.T) {
	fullSet := NewFullUTXOSet()
	diffSet := NewDiffUTXOSet(fullSet, NewDiff())

	// For each of the following test cases, we will call utxoSet.DiffFrom(diffSet) and compare
	// whether the function succeeded with expectedSuccess
	//
	// Note: since test cases are similar for both fullUTXOSet and diffUTXOSet, we test both using the same test cases
	run := func(set Set) {
		tests := []struct {
			name            string
			diffSet         Set
			expectedSuccess bool
		}{
			{
				name:            "Diff from fullSet",
				diffSet:         NewFullUTXOSet(),
				expectedSuccess: false,
			},
			{
				name:            "Diff from diffSet with different Base",
				diffSet:         NewDiffUTXOSet(NewFullUTXOSet(), NewDiff()),
				expectedSuccess: false,
			},
			{
				name:            "Diff from diffSet with same Base",
				diffSet:         NewDiffUTXOSet(fullSet, NewDiff()),
				expectedSuccess: true,
			},
		}

		for _, test := range tests {
			_, err := set.DiffFrom(test.diffSet)
			success := err == nil
			if success != test.expectedSuccess {
				t.Errorf("unexpected DiffFrom success in test \"%s\". "+
					"Expected: \"%t\", got: \"%t\".", test.name, test.expectedSuccess, success)
			}
		}
	}

	run(fullSet) // Perform the test cases above on a fullUTXOSet
	run(diffSet) // Perform the test cases above on a diffUTXOSet
}

// TestDiffUTXOSet_addTx makes sure that diffUTXOSet addTx works as expected
func TestDiffUTXOSet_addTx(t *testing.T) {
	txOut0 := &appmessage.TxOut{ScriptPubKey: []byte{0}, Value: 10}
	utxoEntry0 := NewEntry(txOut0, true, 0)
	coinbaseTX := appmessage.NewSubnetworkMsgTx(1, []*appmessage.TxIn{}, []*appmessage.TxOut{txOut0}, subnetworkid.SubnetworkIDCoinbase, 0, nil)
	fullUTXOCacheSize := config.DefaultConfig().MaxUTXOCacheSize
	db, teardown := prepareDatabaseForTest(t, "TestDiffUTXOSet")
	defer teardown()

	// transaction1 spends coinbaseTX
	id1 := coinbaseTX.TxID()
	outpoint1 := *appmessage.NewOutpoint(id1, 0)
	txIn1 := &appmessage.TxIn{SignatureScript: []byte{}, PreviousOutpoint: outpoint1, Sequence: 0}
	txOut1 := &appmessage.TxOut{ScriptPubKey: []byte{1}, Value: 20}
	utxoEntry1 := NewEntry(txOut1, false, 1)
	transaction1 := appmessage.NewNativeMsgTx(1, []*appmessage.TxIn{txIn1}, []*appmessage.TxOut{txOut1})

	// transaction2 spends transaction1
	id2 := transaction1.TxID()
	outpoint2 := *appmessage.NewOutpoint(id2, 0)
	txIn2 := &appmessage.TxIn{SignatureScript: []byte{}, PreviousOutpoint: outpoint2, Sequence: 0}
	txOut2 := &appmessage.TxOut{ScriptPubKey: []byte{2}, Value: 30}
	utxoEntry2 := NewEntry(txOut2, false, 2)
	transaction2 := appmessage.NewNativeMsgTx(1, []*appmessage.TxIn{txIn2}, []*appmessage.TxOut{txOut2})

	// outpoint3 is the outpoint for transaction2
	id3 := transaction2.TxID()
	outpoint3 := *appmessage.NewOutpoint(id3, 0)

	// For each of the following test cases, we will:
	// 1. startSet.addTx() all the transactions in ToAdd, in order, with the initial block height startHeight
	// 2. Compare the result set with expectedSet
	tests := []struct {
		name        string
		startSet    *DiffUTXOSet
		startHeight uint64
		toAdd       []*appmessage.MsgTx
		expectedSet *DiffUTXOSet
	}{
		{
			name:        "Add coinbase transaction to empty set",
			startSet:    NewDiffUTXOSet(NewFullUTXOSetFromContext(db, fullUTXOCacheSize), NewDiff()),
			startHeight: 0,
			toAdd:       []*appmessage.MsgTx{coinbaseTX},
			expectedSet: &DiffUTXOSet{
				Base: NewFullUTXOSetFromContext(db, fullUTXOCacheSize),
				UTXODiff: &Diff{
					ToAdd:    utxoCollection{outpoint1: utxoEntry0},
					ToRemove: utxoCollection{},
				},
			},
		},
		{
			name:        "Add regular transaction to empty set",
			startSet:    NewDiffUTXOSet(NewFullUTXOSetFromContext(db, fullUTXOCacheSize), NewDiff()),
			startHeight: 0,
			toAdd:       []*appmessage.MsgTx{transaction1},
			expectedSet: &DiffUTXOSet{
				Base: NewFullUTXOSetFromContext(db, fullUTXOCacheSize),
				UTXODiff: &Diff{
					ToAdd:    utxoCollection{},
					ToRemove: utxoCollection{},
				},
			},
		},
		{
			name: "Add transaction to set with its input in Base",
			startSet: &DiffUTXOSet{
				Base: func() *FullUTXOSet {
					futxo := NewFullUTXOSetFromContext(db, fullUTXOCacheSize)
					futxo.Add(outpoint1, utxoEntry0)
					return futxo
				}(),
				UTXODiff: &Diff{
					ToAdd:    utxoCollection{},
					ToRemove: utxoCollection{},
				},
			},
			startHeight: 1,
			toAdd:       []*appmessage.MsgTx{transaction1},
			expectedSet: &DiffUTXOSet{
				Base: func() *FullUTXOSet {
					futxo := NewFullUTXOSetFromContext(db, fullUTXOCacheSize)
					futxo.Add(outpoint1, utxoEntry0)
					return futxo
				}(),
				UTXODiff: &Diff{
					ToAdd:    utxoCollection{outpoint2: utxoEntry1},
					ToRemove: utxoCollection{outpoint1: utxoEntry0},
				},
			},
		},
		{
			name: "Add transaction to set with its input in Diff ToAdd",
			startSet: &DiffUTXOSet{
				Base: NewFullUTXOSetFromContext(db, fullUTXOCacheSize),
				UTXODiff: &Diff{
					ToAdd:    utxoCollection{outpoint1: utxoEntry0},
					ToRemove: utxoCollection{},
				},
			},
			startHeight: 1,
			toAdd:       []*appmessage.MsgTx{transaction1},
			expectedSet: &DiffUTXOSet{
				Base: NewFullUTXOSetFromContext(db, fullUTXOCacheSize),
				UTXODiff: &Diff{
					ToAdd:    utxoCollection{outpoint2: utxoEntry1},
					ToRemove: utxoCollection{},
				},
			},
		},
		{
			name: "Add transaction to set with its input in Diff ToAdd and its output in Diff ToRemove",
			startSet: &DiffUTXOSet{
				Base: NewFullUTXOSetFromContext(db, fullUTXOCacheSize),
				UTXODiff: &Diff{
					ToAdd:    utxoCollection{outpoint1: utxoEntry0},
					ToRemove: utxoCollection{outpoint2: utxoEntry1},
				},
			},
			startHeight: 1,
			toAdd:       []*appmessage.MsgTx{transaction1},
			expectedSet: &DiffUTXOSet{
				Base: NewFullUTXOSetFromContext(db, fullUTXOCacheSize),
				UTXODiff: &Diff{
					ToAdd:    utxoCollection{},
					ToRemove: utxoCollection{},
				},
			},
		},
		{
			name: "Add two transactions, one spending the other, to set with the first input in Base",
			startSet: &DiffUTXOSet{
				Base: func() *FullUTXOSet {
					futxo := NewFullUTXOSetFromContext(db, fullUTXOCacheSize)
					futxo.Add(outpoint1, utxoEntry0)
					return futxo
				}(),
				UTXODiff: &Diff{
					ToAdd:    utxoCollection{},
					ToRemove: utxoCollection{},
				},
			},
			startHeight: 1,
			toAdd:       []*appmessage.MsgTx{transaction1, transaction2},
			expectedSet: &DiffUTXOSet{
				Base: func() *FullUTXOSet {
					futxo := NewFullUTXOSetFromContext(db, fullUTXOCacheSize)
					futxo.Add(outpoint1, utxoEntry0)
					return futxo
				}(),
				UTXODiff: &Diff{
					ToAdd:    utxoCollection{outpoint3: utxoEntry2},
					ToRemove: utxoCollection{outpoint1: utxoEntry0},
				},
			},
		},
	}

testLoop:
	for _, test := range tests {
		diffSet := test.startSet.Clone()

		// Apply all transactions to diffSet, in order, with the initial block height startHeight
		for i, transaction := range test.toAdd {
			height := test.startHeight + uint64(i)
			_, err := diffSet.AddTx(transaction, height)
			if err != nil {
				t.Errorf("Error adding tx %s in test \"%s\": %s", transaction.TxID(), test.name, err)
				continue testLoop
			}
		}

		// Make sure that the result diffSet equals to test.expectedSet
		if !diffSet.(*DiffUTXOSet).equal(test.expectedSet) {
			t.Errorf("unexpected diffSet in test \"%s\". "+
				"Expected: \"%v\", got: \"%v\".", test.name, test.expectedSet, diffSet)
		}
	}
}

// collection returns a collection of all UTXOs in this set
func (fus *FullUTXOSet) collection() utxoCollection {
	return fus.UTXOCache.clone()
}

// collection returns a collection of all UTXOs in this set
func (dus *DiffUTXOSet) collection() (utxoCollection, error) {
	clone := dus.Clone().(*DiffUTXOSet)
	err := clone.MeldToBase()
	if err != nil {
		return nil, err
	}

	return clone.Base.collection(), nil
}

func TestUTXOSetAddEntry(t *testing.T) {
	txID0, _ := daghash.NewTxIDFromStr("0000000000000000000000000000000000000000000000000000000000000000")
	txID1, _ := daghash.NewTxIDFromStr("1111111111111111111111111111111111111111111111111111111111111111")
	outpoint0 := appmessage.NewOutpoint(txID0, 0)
	outpoint1 := appmessage.NewOutpoint(txID1, 0)
	utxoEntry0 := NewEntry(&appmessage.TxOut{ScriptPubKey: []byte{}, Value: 10}, true, 0)
	utxoEntry1 := NewEntry(&appmessage.TxOut{ScriptPubKey: []byte{}, Value: 20}, false, 1)

	utxoDiff := NewDiff()

	tests := []struct {
		name             string
		outpointToAdd    *appmessage.Outpoint
		utxoEntryToAdd   *Entry
		expectedUTXODiff *Diff
		expectedError    string
	}{
		{
			name:           "Add an entry",
			outpointToAdd:  outpoint0,
			utxoEntryToAdd: utxoEntry0,
			expectedUTXODiff: &Diff{
				ToAdd:    utxoCollection{*outpoint0: utxoEntry0},
				ToRemove: utxoCollection{},
			},
		},
		{
			name:           "Add another entry",
			outpointToAdd:  outpoint1,
			utxoEntryToAdd: utxoEntry1,
			expectedUTXODiff: &Diff{
				ToAdd:    utxoCollection{*outpoint0: utxoEntry0, *outpoint1: utxoEntry1},
				ToRemove: utxoCollection{},
			},
		},
		{
			name:           "Add first entry again",
			outpointToAdd:  outpoint0,
			utxoEntryToAdd: utxoEntry0,
			expectedUTXODiff: &Diff{
				ToAdd:    utxoCollection{*outpoint0: utxoEntry0, *outpoint1: utxoEntry1},
				ToRemove: utxoCollection{},
			},
			expectedError: "AddEntry: Cannot Add outpoint 0000000000000000000000000000000000000000000000000000000000000000:0 twice",
		},
	}

	for _, test := range tests {
		err := utxoDiff.AddEntry(*test.outpointToAdd, test.utxoEntryToAdd)
		errString := ""
		if err != nil {
			errString = err.Error()
		}
		if errString != test.expectedError {
			t.Fatalf("utxoDiff.AddEntry: unexpected err in test \"%s\". Expected: %s but got: %s", test.name, test.expectedError, err)
		}
		if err == nil && !utxoDiff.equal(test.expectedUTXODiff) {
			t.Fatalf("utxoDiff.AddEntry: unexpected utxoDiff in test \"%s\". "+
				"Expected: %v, got: %v", test.name, test.expectedUTXODiff, utxoDiff)
		}
	}
}
