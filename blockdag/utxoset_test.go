package blockdag

import (
	"math"
	"reflect"
	"testing"

	"github.com/daglabs/btcd/dagconfig"
	"github.com/daglabs/btcd/dagconfig/daghash"
	"github.com/daglabs/btcd/util"
	"github.com/daglabs/btcd/wire"
)

// TestUTXOCollection makes sure that utxoCollection cloning and string representations work as expected.
func TestUTXOCollection(t *testing.T) {
	hash0, _ := daghash.NewHashFromStr("0000000000000000000000000000000000000000000000000000000000000000")
	hash1, _ := daghash.NewHashFromStr("1111111111111111111111111111111111111111111111111111111111111111")
	outPoint0 := *wire.NewOutPoint(hash0, 0)
	outPoint1 := *wire.NewOutPoint(hash1, 0)
	utxoEntry0 := NewUTXOEntry(&wire.TxOut{PkScript: []byte{}, Value: 10}, true, 0)
	utxoEntry1 := NewUTXOEntry(&wire.TxOut{PkScript: []byte{}, Value: 20}, false, 1)

	// For each of the following test cases, we will:
	// .String() the given collection and compare it to expectedString
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
				outPoint0: utxoEntry1,
			},
			expectedString: "[ (0000000000000000000000000000000000000000000000000000000000000000, 0) => 20 ]",
		},
		{
			name: "two members",
			collection: utxoCollection{
				outPoint0: utxoEntry0,
				outPoint1: utxoEntry1,
			},
			expectedString: "[ (0000000000000000000000000000000000000000000000000000000000000000, 0) => 10, (1111111111111111111111111111111111111111111111111111111111111111, 0) => 20 ]",
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
			t.Errorf("collection is reference-equal to its clone in test \"%s\". ", test.name)
		}
		if !reflect.DeepEqual(test.collection, collectionClone) {
			t.Errorf("collection is not equal to its clone in test \"%s\". "+
				"Expected: \"%s\", got: \"%s\".", test.name, collectionString, collectionClone.String())
		}
	}
}

// TestUTXODiff makes sure that utxoDiff creation, cloning, and string representations work as expected.
func TestUTXODiff(t *testing.T) {
	hash0, _ := daghash.NewHashFromStr("0000000000000000000000000000000000000000000000000000000000000000")
	hash1, _ := daghash.NewHashFromStr("1111111111111111111111111111111111111111111111111111111111111111")
	outPoint0 := *wire.NewOutPoint(hash0, 0)
	outPoint1 := *wire.NewOutPoint(hash1, 0)
	utxoEntry0 := NewUTXOEntry(&wire.TxOut{PkScript: []byte{}, Value: 10}, true, 0)
	utxoEntry1 := NewUTXOEntry(&wire.TxOut{PkScript: []byte{}, Value: 20}, false, 1)
	diff := UTXODiff{
		toAdd:    utxoCollection{outPoint0: utxoEntry0},
		toRemove: utxoCollection{outPoint1: utxoEntry1},
	}

	// Test utxoDiff creation
	newDiff := NewUTXODiff()
	if len(newDiff.toAdd) != 0 || len(newDiff.toRemove) != 0 {
		t.Errorf("new diff is not empty")
	}

	// Test utxoDiff cloning
	clonedDiff := *diff.clone()
	if &clonedDiff == &diff {
		t.Errorf("cloned diff is reference-equal to the original")
	}
	if !reflect.DeepEqual(clonedDiff, diff) {
		t.Errorf("cloned diff not equal to the original"+
			"Original: \"%v\", cloned: \"%v\".", diff, clonedDiff)
	}

	// Test utxoDiff string representation
	expectedDiffString := "toAdd: [ (0000000000000000000000000000000000000000000000000000000000000000, 0) => 10 ]; toRemove: [ (1111111111111111111111111111111111111111111111111111111111111111, 0) => 20 ]"
	diffString := clonedDiff.String()
	if diffString != expectedDiffString {
		t.Errorf("unexpected diff string. "+
			"Expected: \"%s\", got: \"%s\".", expectedDiffString, diffString)
	}
}

// TestUTXODiffRules makes sure that all diffFrom and WithDiff rules are followed.
// Each test case represents a cell in the two tables outlined in the documentation for utxoDiff.
func TestUTXODiffRules(t *testing.T) {
	hash0, _ := daghash.NewHashFromStr("0000000000000000000000000000000000000000000000000000000000000000")
	outPoint0 := *wire.NewOutPoint(hash0, 0)
	utxoEntry0 := NewUTXOEntry(&wire.TxOut{PkScript: []byte{}, Value: 10}, true, 0)

	// For each of the following test cases, we will:
	// this.diffFrom(other) and compare it to expectedDiffFromResult
	// this.WithDiff(other) and compare it to expectedWithDiffResult
	//
	// Note: an expected nil result means that we expect the respective operation to fail
	tests := []struct {
		name                   string
		this                   *UTXODiff
		other                  *UTXODiff
		expectedDiffFromResult *UTXODiff
		expectedWithDiffResult *UTXODiff
	}{
		{
			name: "one toAdd in this, one toAdd in other",
			this: &UTXODiff{
				toAdd:    utxoCollection{outPoint0: utxoEntry0},
				toRemove: utxoCollection{},
			},
			other: &UTXODiff{
				toAdd:    utxoCollection{outPoint0: utxoEntry0},
				toRemove: utxoCollection{},
			},
			expectedDiffFromResult: &UTXODiff{
				toAdd:    utxoCollection{},
				toRemove: utxoCollection{},
			},
			expectedWithDiffResult: nil,
		},
		{
			name: "one toAdd in this, one toRemove in other",
			this: &UTXODiff{
				toAdd:    utxoCollection{outPoint0: utxoEntry0},
				toRemove: utxoCollection{},
			},
			other: &UTXODiff{
				toAdd:    utxoCollection{},
				toRemove: utxoCollection{outPoint0: utxoEntry0},
			},
			expectedDiffFromResult: nil,
			expectedWithDiffResult: &UTXODiff{
				toAdd:    utxoCollection{},
				toRemove: utxoCollection{},
			},
		},
		{
			name: "one toAdd in this, empty other",
			this: &UTXODiff{
				toAdd:    utxoCollection{outPoint0: utxoEntry0},
				toRemove: utxoCollection{},
			},
			other: &UTXODiff{
				toAdd:    utxoCollection{},
				toRemove: utxoCollection{},
			},
			expectedDiffFromResult: &UTXODiff{
				toAdd:    utxoCollection{},
				toRemove: utxoCollection{outPoint0: utxoEntry0},
			},
			expectedWithDiffResult: &UTXODiff{
				toAdd:    utxoCollection{outPoint0: utxoEntry0},
				toRemove: utxoCollection{},
			},
		},
		{
			name: "one toRemove in this, one toAdd in other",
			this: &UTXODiff{
				toAdd:    utxoCollection{},
				toRemove: utxoCollection{outPoint0: utxoEntry0},
			},
			other: &UTXODiff{
				toAdd:    utxoCollection{outPoint0: utxoEntry0},
				toRemove: utxoCollection{},
			},
			expectedDiffFromResult: nil,
			expectedWithDiffResult: &UTXODiff{
				toAdd:    utxoCollection{},
				toRemove: utxoCollection{},
			},
		},
		{
			name: "one toRemove in this, one toRemove in other",
			this: &UTXODiff{
				toAdd:    utxoCollection{},
				toRemove: utxoCollection{outPoint0: utxoEntry0},
			},
			other: &UTXODiff{
				toAdd:    utxoCollection{},
				toRemove: utxoCollection{outPoint0: utxoEntry0},
			},
			expectedDiffFromResult: &UTXODiff{
				toAdd:    utxoCollection{},
				toRemove: utxoCollection{},
			},
			expectedWithDiffResult: nil,
		},
		{
			name: "one toRemove in this, empty other",
			this: &UTXODiff{
				toAdd:    utxoCollection{},
				toRemove: utxoCollection{outPoint0: utxoEntry0},
			},
			other: &UTXODiff{
				toAdd:    utxoCollection{},
				toRemove: utxoCollection{},
			},
			expectedDiffFromResult: &UTXODiff{
				toAdd:    utxoCollection{outPoint0: utxoEntry0},
				toRemove: utxoCollection{},
			},
			expectedWithDiffResult: &UTXODiff{
				toAdd:    utxoCollection{},
				toRemove: utxoCollection{outPoint0: utxoEntry0},
			},
		},
		{
			name: "empty this, one toAdd in other",
			this: &UTXODiff{
				toAdd:    utxoCollection{},
				toRemove: utxoCollection{},
			},
			other: &UTXODiff{
				toAdd:    utxoCollection{outPoint0: utxoEntry0},
				toRemove: utxoCollection{},
			},
			expectedDiffFromResult: &UTXODiff{
				toAdd:    utxoCollection{outPoint0: utxoEntry0},
				toRemove: utxoCollection{},
			},
			expectedWithDiffResult: &UTXODiff{
				toAdd:    utxoCollection{outPoint0: utxoEntry0},
				toRemove: utxoCollection{},
			},
		},
		{
			name: "empty this, one toRemove in other",
			this: &UTXODiff{
				toAdd:    utxoCollection{},
				toRemove: utxoCollection{},
			},
			other: &UTXODiff{
				toAdd:    utxoCollection{},
				toRemove: utxoCollection{outPoint0: utxoEntry0},
			},
			expectedDiffFromResult: &UTXODiff{
				toAdd:    utxoCollection{},
				toRemove: utxoCollection{outPoint0: utxoEntry0},
			},
			expectedWithDiffResult: &UTXODiff{
				toAdd:    utxoCollection{},
				toRemove: utxoCollection{outPoint0: utxoEntry0},
			},
		},
		{
			name: "empty this, empty other",
			this: &UTXODiff{
				toAdd:    utxoCollection{},
				toRemove: utxoCollection{},
			},
			other: &UTXODiff{
				toAdd:    utxoCollection{},
				toRemove: utxoCollection{},
			},
			expectedDiffFromResult: &UTXODiff{
				toAdd:    utxoCollection{},
				toRemove: utxoCollection{},
			},
			expectedWithDiffResult: &UTXODiff{
				toAdd:    utxoCollection{},
				toRemove: utxoCollection{},
			},
		},
	}

	for _, test := range tests {
		// diffFrom from this to other
		diffResult, err := test.this.diffFrom(test.other)

		// Test whether diffFrom returned an error
		isDiffFromOk := err == nil
		expectedIsDiffFromOk := test.expectedDiffFromResult != nil
		if isDiffFromOk != expectedIsDiffFromOk {
			t.Errorf("unexpected diffFrom error in test \"%s\". "+
				"Expected: \"%t\", got: \"%t\".", test.name, expectedIsDiffFromOk, isDiffFromOk)
		}

		// If not error, test the diffFrom result
		if isDiffFromOk && !reflect.DeepEqual(diffResult, test.expectedDiffFromResult) {
			t.Errorf("unexpected diffFrom result in test \"%s\". "+
				"Expected: \"%v\", got: \"%v\".", test.name, test.expectedDiffFromResult, diffResult)
		}

		// WithDiff from this to other
		withDiffResult, err := test.this.WithDiff(test.other)

		// Test whether WithDiff returned an error
		isWithDiffOk := err == nil
		expectedIsWithDiffOk := test.expectedWithDiffResult != nil
		if isWithDiffOk != expectedIsWithDiffOk {
			t.Errorf("unexpected WithDiff error in test \"%s\". "+
				"Expected: \"%t\", got: \"%t\".", test.name, expectedIsWithDiffOk, isWithDiffOk)
		}

		// Ig not error, test the WithDiff result
		if isWithDiffOk && !reflect.DeepEqual(withDiffResult, test.expectedWithDiffResult) {
			t.Errorf("unexpected WithDiff result in test \"%s\". "+
				"Expected: \"%v\", got: \"%v\".", test.name, test.expectedWithDiffResult, withDiffResult)
		}
	}
}

// TestFullUTXOSet makes sure that fullUTXOSet is working as expected.
func TestFullUTXOSet(t *testing.T) {
	hash0, _ := daghash.NewHashFromStr("0000000000000000000000000000000000000000000000000000000000000000")
	hash1, _ := daghash.NewHashFromStr("1111111111111111111111111111111111111111111111111111111111111111")
	outPoint0 := *wire.NewOutPoint(hash0, 0)
	outPoint1 := *wire.NewOutPoint(hash1, 0)
	txOut0 := &wire.TxOut{PkScript: []byte{}, Value: 10}
	txOut1 := &wire.TxOut{PkScript: []byte{}, Value: 20}
	utxoEntry0 := NewUTXOEntry(txOut0, true, 0)
	utxoEntry1 := NewUTXOEntry(txOut1, false, 1)
	diff := &UTXODiff{
		toAdd:    utxoCollection{outPoint0: utxoEntry0},
		toRemove: utxoCollection{outPoint1: utxoEntry1},
	}

	// Test fullUTXOSet creation
	emptySet := NewFullUTXOSet()
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
	if !reflect.DeepEqual(withDiffUTXOSet.base, emptySet) || !reflect.DeepEqual(withDiffUTXOSet.UTXODiff, diff) {
		t.Errorf("WithDiff is of unexpected composition")
	}

	// Test fullUTXOSet addTx
	txIn0 := &wire.TxIn{SignatureScript: []byte{}, PreviousOutPoint: wire.OutPoint{TxID: *hash0, Index: 0}, Sequence: 0}
	transaction0 := wire.NewMsgTx(1)
	transaction0.TxIn = []*wire.TxIn{txIn0}
	transaction0.TxOut = []*wire.TxOut{txOut0}
	if ok = emptySet.AddTx(transaction0, 0); ok {
		t.Errorf("addTx unexpectedly succeeded")
	}
	emptySet = &FullUTXOSet{utxoCollection: utxoCollection{outPoint0: utxoEntry0}}
	if ok = emptySet.AddTx(transaction0, 0); !ok {
		t.Errorf("addTx unexpectedly failed")
	}

	// Test fullUTXOSet collection
	if !reflect.DeepEqual(emptySet.collection(), emptySet.utxoCollection) {
		t.Errorf("collection does not equal the set's utxoCollection")
	}

	// Test fullUTXOSet cloning
	clonedEmptySet := emptySet.clone().(*FullUTXOSet)
	if !reflect.DeepEqual(clonedEmptySet, emptySet) {
		t.Errorf("clone does not equal the original set")
	}
	if clonedEmptySet == emptySet {
		t.Errorf("cloned set is reference-equal to the original")
	}
}

// TestDiffUTXOSet makes sure that diffUTXOSet is working as expected.
func TestDiffUTXOSet(t *testing.T) {
	hash0, _ := daghash.NewHashFromStr("0000000000000000000000000000000000000000000000000000000000000000")
	hash1, _ := daghash.NewHashFromStr("1111111111111111111111111111111111111111111111111111111111111111")
	outPoint0 := *wire.NewOutPoint(hash0, 0)
	outPoint1 := *wire.NewOutPoint(hash1, 0)
	txOut0 := &wire.TxOut{PkScript: []byte{}, Value: 10}
	txOut1 := &wire.TxOut{PkScript: []byte{}, Value: 20}
	utxoEntry0 := NewUTXOEntry(txOut0, true, 0)
	utxoEntry1 := NewUTXOEntry(txOut1, false, 1)
	diff := &UTXODiff{
		toAdd:    utxoCollection{outPoint0: utxoEntry0},
		toRemove: utxoCollection{outPoint1: utxoEntry1},
	}

	// Test diffUTXOSet creation
	emptySet := NewDiffUTXOSet(NewFullUTXOSet(), NewUTXODiff())
	if len(emptySet.collection()) != 0 {
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
	withDiff, _ := NewUTXODiff().WithDiff(diff)
	if !reflect.DeepEqual(withDiffUTXOSet.base, emptySet.base) || !reflect.DeepEqual(withDiffUTXOSet.UTXODiff, withDiff) {
		t.Errorf("WithDiff is of unexpected composition")
	}
	_, err = NewDiffUTXOSet(NewFullUTXOSet(), diff).WithDiff(diff)
	if err == nil {
		t.Errorf("WithDiff unexpectedly succeeded")
	}

	// Given a diffSet, each case tests that meldToBase, String, collection, and cloning work as expected
	// For each of the following test cases, we will:
	// .meldToBase() the given diffSet and compare it to expectedMeldSet
	// .String() the given diffSet and compare it to expectedString
	// .collection() the given diffSet and compare it to expectedCollection
	// .clone() the given diffSet and compare its value to itself (expected: equals) and its reference to itself (expected: not equal)
	tests := []struct {
		name               string
		diffSet            *DiffUTXOSet
		expectedMeldSet    *DiffUTXOSet
		expectedString     string
		expectedCollection utxoCollection
	}{
		{
			name: "empty base, empty diff",
			diffSet: &DiffUTXOSet{
				base: NewFullUTXOSet(),
				UTXODiff: &UTXODiff{
					toAdd:    utxoCollection{},
					toRemove: utxoCollection{},
				},
			},
			expectedMeldSet: &DiffUTXOSet{
				base: NewFullUTXOSet(),
				UTXODiff: &UTXODiff{
					toAdd:    utxoCollection{},
					toRemove: utxoCollection{},
				},
			},
			expectedString:     "{Base: [  ], To Add: [  ], To Remove: [  ]}",
			expectedCollection: utxoCollection{},
		},
		{
			name: "empty base, one member in diff toAdd",
			diffSet: &DiffUTXOSet{
				base: NewFullUTXOSet(),
				UTXODiff: &UTXODiff{
					toAdd:    utxoCollection{outPoint0: utxoEntry0},
					toRemove: utxoCollection{},
				},
			},
			expectedMeldSet: &DiffUTXOSet{
				base: &FullUTXOSet{utxoCollection: utxoCollection{outPoint0: utxoEntry0}},
				UTXODiff: &UTXODiff{
					toAdd:    utxoCollection{},
					toRemove: utxoCollection{},
				},
			},
			expectedString:     "{Base: [  ], To Add: [ (0000000000000000000000000000000000000000000000000000000000000000, 0) => 10 ], To Remove: [  ]}",
			expectedCollection: utxoCollection{outPoint0: utxoEntry0},
		},
		{
			name: "empty base, one member in diff toRemove",
			diffSet: &DiffUTXOSet{
				base: NewFullUTXOSet(),
				UTXODiff: &UTXODiff{
					toAdd:    utxoCollection{},
					toRemove: utxoCollection{outPoint0: utxoEntry0},
				},
			},
			expectedMeldSet: &DiffUTXOSet{
				base: &FullUTXOSet{utxoCollection: utxoCollection{}},
				UTXODiff: &UTXODiff{
					toAdd:    utxoCollection{},
					toRemove: utxoCollection{},
				},
			},
			expectedString:     "{Base: [  ], To Add: [  ], To Remove: [ (0000000000000000000000000000000000000000000000000000000000000000, 0) => 10 ]}",
			expectedCollection: utxoCollection{},
		},
		{
			name: "one member in base toAdd, one member in diff toAdd",
			diffSet: &DiffUTXOSet{
				base: &FullUTXOSet{utxoCollection: utxoCollection{outPoint0: utxoEntry0}},
				UTXODiff: &UTXODiff{
					toAdd:    utxoCollection{outPoint1: utxoEntry1},
					toRemove: utxoCollection{},
				},
			},
			expectedMeldSet: &DiffUTXOSet{
				base: &FullUTXOSet{
					utxoCollection: utxoCollection{
						outPoint0: utxoEntry0,
						outPoint1: utxoEntry1,
					},
				},
				UTXODiff: &UTXODiff{
					toAdd:    utxoCollection{},
					toRemove: utxoCollection{},
				},
			},
			expectedString: "{Base: [ (0000000000000000000000000000000000000000000000000000000000000000, 0) => 10 ], To Add: [ (1111111111111111111111111111111111111111111111111111111111111111, 0) => 20 ], To Remove: [  ]}",
			expectedCollection: utxoCollection{
				outPoint0: utxoEntry0,
				outPoint1: utxoEntry1,
			},
		},
		{
			name: "one member in base toAdd, same one member in diff toRemove",
			diffSet: &DiffUTXOSet{
				base: &FullUTXOSet{utxoCollection: utxoCollection{outPoint0: utxoEntry0}},
				UTXODiff: &UTXODiff{
					toAdd:    utxoCollection{},
					toRemove: utxoCollection{outPoint0: utxoEntry0},
				},
			},
			expectedMeldSet: &DiffUTXOSet{
				base: &FullUTXOSet{
					utxoCollection: utxoCollection{},
				},
				UTXODiff: &UTXODiff{
					toAdd:    utxoCollection{},
					toRemove: utxoCollection{},
				},
			},
			expectedString:     "{Base: [ (0000000000000000000000000000000000000000000000000000000000000000, 0) => 10 ], To Add: [  ], To Remove: [ (0000000000000000000000000000000000000000000000000000000000000000, 0) => 10 ]}",
			expectedCollection: utxoCollection{},
		},
	}

	for _, test := range tests {
		// Test meldToBase
		meldSet := test.diffSet.clone().(*DiffUTXOSet)
		meldSet.meldToBase()
		if !reflect.DeepEqual(meldSet, test.expectedMeldSet) {
			t.Errorf("unexpected melded set in test \"%s\". "+
				"Expected: \"%v\", got: \"%v\".", test.name, test.expectedMeldSet, meldSet)
		}

		// Test string representation
		setString := test.diffSet.String()
		if setString != test.expectedString {
			t.Errorf("unexpected string in test \"%s\". "+
				"Expected: \"%s\", got: \"%s\".", test.name, test.expectedString, setString)
		}

		// Test collection
		setCollection := test.diffSet.collection()
		if !reflect.DeepEqual(setCollection, test.expectedCollection) {
			t.Errorf("unexpected set collection in test \"%s\". "+
				"Expected: \"%v\", got: \"%v\".", test.name, test.expectedCollection, setCollection)
		}

		// Test cloning
		clonedSet := test.diffSet.clone().(*DiffUTXOSet)
		if !reflect.DeepEqual(clonedSet, test.diffSet) {
			t.Errorf("unexpected set clone in test \"%s\". "+
				"Expected: \"%v\", got: \"%v\".", test.name, test.diffSet, clonedSet)
		}
		if clonedSet == test.diffSet {
			t.Errorf("cloned set is reference-equal to the original")
		}
	}
}

// TestUTXOSetDiffRules makes sure that utxoSet diffFrom rules are followed.
// The rules are:
// 1. Neither fullUTXOSet nor diffUTXOSet can diffFrom a fullUTXOSet.
// 2. fullUTXOSet cannot diffFrom a diffUTXOSet with a base other that itself.
// 3. diffUTXOSet cannot diffFrom a diffUTXOSet with a different base.
func TestUTXOSetDiffRules(t *testing.T) {
	fullSet := NewFullUTXOSet()
	diffSet := NewDiffUTXOSet(fullSet, NewUTXODiff())

	// For each of the following test cases, we will call utxoSet.diffFrom(diffSet) and compare
	// whether the function succeeded with expectedSuccess
	//
	// Note: since test cases are similar for both fullUTXOSet and diffUTXOSet, we test both using the same test cases
	run := func(set UTXOSet) {
		tests := []struct {
			name            string
			diffSet         UTXOSet
			expectedSuccess bool
		}{
			{
				name:            "diff from fullSet",
				diffSet:         NewFullUTXOSet(),
				expectedSuccess: false,
			},
			{
				name:            "diff from diffSet with different base",
				diffSet:         NewDiffUTXOSet(NewFullUTXOSet(), NewUTXODiff()),
				expectedSuccess: false,
			},
			{
				name:            "diff from diffSet with same base",
				diffSet:         NewDiffUTXOSet(fullSet, NewUTXODiff()),
				expectedSuccess: true,
			},
		}

		for _, test := range tests {
			_, err := set.diffFrom(test.diffSet)
			success := err == nil
			if success != test.expectedSuccess {
				t.Errorf("unexpected diffFrom success in test \"%s\". "+
					"Expected: \"%t\", got: \"%t\".", test.name, test.expectedSuccess, success)
			}
		}
	}

	run(fullSet) // Perform the test cases above on a fullUTXOSet
	run(diffSet) // Perform the test cases above on a diffUTXOSet
}

// TestDiffUTXOSet_addTx makes sure that diffUTXOSet addTx works as expected
func TestDiffUTXOSet_addTx(t *testing.T) {
	// transaction0 is coinbase. As such, it has exactly one input with hash zero and MaxUInt32 index
	hash0, _ := daghash.NewHashFromStr("0000000000000000000000000000000000000000000000000000000000000000")
	txIn0 := &wire.TxIn{SignatureScript: []byte{}, PreviousOutPoint: wire.OutPoint{TxID: *hash0, Index: math.MaxUint32}, Sequence: 0}
	txOut0 := &wire.TxOut{PkScript: []byte{0}, Value: 10}
	utxoEntry0 := NewUTXOEntry(txOut0, true, 0)
	transaction0 := wire.NewMsgTx(1)
	transaction0.TxIn = []*wire.TxIn{txIn0}
	transaction0.TxOut = []*wire.TxOut{txOut0}

	// transaction1 spends transaction0
	id1 := transaction0.TxID()
	outPoint1 := *wire.NewOutPoint(&id1, 0)
	txIn1 := &wire.TxIn{SignatureScript: []byte{}, PreviousOutPoint: wire.OutPoint{TxID: id1, Index: 0}, Sequence: 0}
	txOut1 := &wire.TxOut{PkScript: []byte{1}, Value: 20}
	utxoEntry1 := NewUTXOEntry(txOut1, false, 1)
	transaction1 := wire.NewMsgTx(1)
	transaction1.TxIn = []*wire.TxIn{txIn1}
	transaction1.TxOut = []*wire.TxOut{txOut1}

	// transaction2 spends transaction1
	id2 := transaction1.TxID()
	outPoint2 := *wire.NewOutPoint(&id2, 0)
	txIn2 := &wire.TxIn{SignatureScript: []byte{}, PreviousOutPoint: wire.OutPoint{TxID: id2, Index: 0}, Sequence: 0}
	txOut2 := &wire.TxOut{PkScript: []byte{2}, Value: 30}
	utxoEntry2 := NewUTXOEntry(txOut2, false, 2)
	transaction2 := wire.NewMsgTx(1)
	transaction2.TxIn = []*wire.TxIn{txIn2}
	transaction2.TxOut = []*wire.TxOut{txOut2}

	// outpoint3 is the outpoint for transaction2
	id3 := transaction2.TxID()
	outPoint3 := *wire.NewOutPoint(&id3, 0)

	// For each of the following test cases, we will:
	// 1. startSet.addTx() all the transactions in toAdd, in order, with the initial block height startHeight
	// 2. Compare the result set with expectedSet
	tests := []struct {
		name        string
		startSet    *DiffUTXOSet
		startHeight int32
		toAdd       []*wire.MsgTx
		expectedSet *DiffUTXOSet
	}{
		{
			name:        "add coinbase transaction to empty set",
			startSet:    NewDiffUTXOSet(NewFullUTXOSet(), NewUTXODiff()),
			startHeight: 0,
			toAdd:       []*wire.MsgTx{transaction0},
			expectedSet: &DiffUTXOSet{
				base: &FullUTXOSet{utxoCollection: utxoCollection{}},
				UTXODiff: &UTXODiff{
					toAdd:    utxoCollection{outPoint1: utxoEntry0},
					toRemove: utxoCollection{},
				},
			},
		},
		{
			name:        "add regular transaction to empty set",
			startSet:    NewDiffUTXOSet(NewFullUTXOSet(), NewUTXODiff()),
			startHeight: 0,
			toAdd:       []*wire.MsgTx{transaction1},
			expectedSet: &DiffUTXOSet{
				base: &FullUTXOSet{utxoCollection: utxoCollection{}},
				UTXODiff: &UTXODiff{
					toAdd:    utxoCollection{},
					toRemove: utxoCollection{},
				},
			},
		},
		{
			name: "add transaction to set with its input in base",
			startSet: &DiffUTXOSet{
				base: &FullUTXOSet{utxoCollection: utxoCollection{outPoint1: utxoEntry0}},
				UTXODiff: &UTXODiff{
					toAdd:    utxoCollection{},
					toRemove: utxoCollection{},
				},
			},
			startHeight: 1,
			toAdd:       []*wire.MsgTx{transaction1},
			expectedSet: &DiffUTXOSet{
				base: &FullUTXOSet{utxoCollection: utxoCollection{outPoint1: utxoEntry0}},
				UTXODiff: &UTXODiff{
					toAdd:    utxoCollection{outPoint2: utxoEntry1},
					toRemove: utxoCollection{outPoint1: utxoEntry0},
				},
			},
		},
		{
			name: "add transaction to set with its input in diff toAdd",
			startSet: &DiffUTXOSet{
				base: NewFullUTXOSet(),
				UTXODiff: &UTXODiff{
					toAdd:    utxoCollection{outPoint1: utxoEntry0},
					toRemove: utxoCollection{},
				},
			},
			startHeight: 1,
			toAdd:       []*wire.MsgTx{transaction1},
			expectedSet: &DiffUTXOSet{
				base: NewFullUTXOSet(),
				UTXODiff: &UTXODiff{
					toAdd:    utxoCollection{outPoint2: utxoEntry1},
					toRemove: utxoCollection{},
				},
			},
		},
		{
			name: "add transaction to set with its input in diff toAdd and its output in diff toRemove",
			startSet: &DiffUTXOSet{
				base: NewFullUTXOSet(),
				UTXODiff: &UTXODiff{
					toAdd:    utxoCollection{outPoint1: utxoEntry0},
					toRemove: utxoCollection{outPoint2: utxoEntry1},
				},
			},
			startHeight: 1,
			toAdd:       []*wire.MsgTx{transaction1},
			expectedSet: &DiffUTXOSet{
				base: NewFullUTXOSet(),
				UTXODiff: &UTXODiff{
					toAdd:    utxoCollection{},
					toRemove: utxoCollection{},
				},
			},
		},
		{
			name: "add two transactions, one spending the other, to set with the first input in base",
			startSet: &DiffUTXOSet{
				base: &FullUTXOSet{utxoCollection: utxoCollection{outPoint1: utxoEntry0}},
				UTXODiff: &UTXODiff{
					toAdd:    utxoCollection{},
					toRemove: utxoCollection{},
				},
			},
			startHeight: 1,
			toAdd:       []*wire.MsgTx{transaction1, transaction2},
			expectedSet: &DiffUTXOSet{
				base: &FullUTXOSet{utxoCollection: utxoCollection{outPoint1: utxoEntry0}},
				UTXODiff: &UTXODiff{
					toAdd:    utxoCollection{outPoint3: utxoEntry2},
					toRemove: utxoCollection{outPoint1: utxoEntry0},
				},
			},
		},
	}

	for _, test := range tests {
		diffSet := test.startSet.clone()

		// Apply all transactions to diffSet, in order, with the initial block height startHeight
		for i, transaction := range test.toAdd {
			diffSet.AddTx(transaction, test.startHeight+int32(i))
		}

		// Make sure that the result diffSet equals to the expectedSet
		if !reflect.DeepEqual(diffSet, test.expectedSet) {
			t.Errorf("unexpected diffSet in test \"%s\". "+
				"Expected: \"%v\", got: \"%v\".", test.name, test.expectedSet, diffSet)
		}
	}
}

func TestApplyUTXOChanges(t *testing.T) {
	// Create a new database and dag instance to run tests against.
	dag, teardownFunc, err := DAGSetup("TestApplyUTXOChanges", Config{
		DAGParams: &dagconfig.SimNetParams,
	})
	if err != nil {
		t.Fatalf("Failed to setup dag instance: %v", err)
	}
	defer teardownFunc()

	dag.dagParams.CoinbaseMaturity = 1
	// Create artificial checkpoint in order to prevent script validation for blocks that comes before it
	dag.checkpoints = []dagconfig.Checkpoint{
		{
			Height: 1000,
		},
	}

	cbTx, err := createCoinbaseTxForTest(1, 1, 0, dag.dagParams)
	if err != nil {
		t.Errorf("createCoinbaseTxForTest: %v", err)
	}

	chainedTx := wire.NewMsgTx(wire.TxVersion)
	chainedTx.AddTxIn(&wire.TxIn{
		PreviousOutPoint: wire.OutPoint{TxID: cbTx.TxID(), Index: 0},
		SignatureScript:  nil,
		Sequence:         wire.MaxTxInSequenceNum,
	})
	chainedTx.AddTxOut(&wire.TxOut{
		PkScript: OpTrueScript,
		Value:    uint64(1),
	})

	//Fake block header
	blockHeader := wire.NewBlockHeader(1, []daghash.Hash{dag.genesis.hash}, &daghash.Hash{}, &daghash.Hash{}, 0, 0)

	msgBlock1 := &wire.MsgBlock{
		Header:       *blockHeader,
		Transactions: []*wire.MsgTx{cbTx, chainedTx},
	}

	block1 := util.NewBlock(msgBlock1)

	var node1 blockNode
	initBlockNode(&node1, blockHeader, setFromSlice(dag.genesis), dagconfig.MainNetParams.K)

	//Checks that dag.applyUTXOChanges fails because we don't allow a transaction to spend another transaction from the same block
	_, _, err = dag.applyUTXOChanges(&node1, block1, false)
	if err == nil {
		t.Errorf("applyUTXOChanges expected an error\n")
	}

	nonChainedTx := wire.NewMsgTx(wire.TxVersion)
	nonChainedTx.AddTxIn(&wire.TxIn{
		PreviousOutPoint: wire.OutPoint{TxID: dag.dagParams.GenesisBlock.Transactions[0].TxID(), Index: 0},
		SignatureScript:  nil, //Fake SigScript, because we don't check scripts validity in this test
		Sequence:         wire.MaxTxInSequenceNum,
	})
	nonChainedTx.AddTxOut(&wire.TxOut{
		PkScript: OpTrueScript,
		Value:    uint64(1),
	})

	msgBlock2 := &wire.MsgBlock{
		Header:       *blockHeader,
		Transactions: []*wire.MsgTx{cbTx, nonChainedTx},
	}

	block2 := util.NewBlock(msgBlock2)

	var node2 blockNode
	initBlockNode(&node2, blockHeader, setFromSlice(dag.genesis), dagconfig.MainNetParams.K)

	//Checks that dag.applyUTXOChanges doesn't fail because all of its transaction are dependant on transactions from previous blocks
	_, _, err = dag.applyUTXOChanges(&node2, block2, false)
	if err != nil {
		t.Errorf("applyUTXOChanges: %v", err)
	}
}

func TestDiffFromTx(t *testing.T) {
	fus := &FullUTXOSet{
		utxoCollection: utxoCollection{},
	}
	cbTx, err := createCoinbaseTxForTest(1, 1, 0, &dagconfig.SimNetParams)
	if err != nil {
		t.Errorf("createCoinbaseTxForTest: %v", err)
	}
	fus.AddTx(cbTx, 1)
	node := &blockNode{height: 2} //Fake node
	cbOutpoint := wire.OutPoint{TxID: cbTx.TxID(), Index: 0}
	tx := wire.NewMsgTx(wire.TxVersion)
	tx.AddTxIn(&wire.TxIn{
		PreviousOutPoint: cbOutpoint,
		SignatureScript:  nil,
		Sequence:         wire.MaxTxInSequenceNum,
	})
	tx.AddTxOut(&wire.TxOut{
		PkScript: OpTrueScript,
		Value:    uint64(1),
	})
	diff, err := fus.diffFromTx(tx, node)
	if err != nil {
		t.Errorf("diffFromTx: %v", err)
	}
	if !reflect.DeepEqual(diff.toAdd, utxoCollection{
		wire.OutPoint{TxID: tx.TxID(), Index: 0}: NewUTXOEntry(tx.TxOut[0], false, 2),
	}) {
		t.Errorf("diff.toAdd doesn't have the expected values")
	}

	if !reflect.DeepEqual(diff.toRemove, utxoCollection{
		wire.OutPoint{TxID: cbTx.TxID(), Index: 0}: NewUTXOEntry(cbTx.TxOut[0], true, 1),
	}) {
		t.Errorf("diff.toRemove doesn't have the expected values")
	}

	//Test that we get an error if we don't have the outpoint inside the utxo set
	invalidTx := wire.NewMsgTx(wire.TxVersion)
	invalidTx.AddTxIn(&wire.TxIn{
		PreviousOutPoint: wire.OutPoint{TxID: daghash.Hash{}, Index: 0},
		SignatureScript:  nil,
		Sequence:         wire.MaxTxInSequenceNum,
	})
	invalidTx.AddTxOut(&wire.TxOut{
		PkScript: OpTrueScript,
		Value:    uint64(1),
	})
	_, err = fus.diffFromTx(invalidTx, node)
	if err == nil {
		t.Errorf("diffFromTx: expected an error but got <nil>")
	}

	//Test that we get an error if the outpoint is inside diffUTXOSet's toRemove
	dus := NewDiffUTXOSet(fus, &UTXODiff{
		toAdd:    utxoCollection{},
		toRemove: utxoCollection{},
	})
	dus.AddTx(tx, 2)
	_, err = dus.diffFromTx(tx, node)
	if err == nil {
		t.Errorf("diffFromTx: expected an error but got <nil>")
	}
}

// collection returns a collection of all UTXOs in this set
func (fus *FullUTXOSet) collection() utxoCollection {
	return fus.utxoCollection.clone()
}

// collection returns a collection of all UTXOs in this set
func (dus *DiffUTXOSet) collection() utxoCollection {
	clone := dus.clone().(*DiffUTXOSet)
	clone.meldToBase()

	return clone.base.collection()
}
