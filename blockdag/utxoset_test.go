package blockdag

import (
	"github.com/daglabs/btcd/dagconfig/daghash"
	"github.com/daglabs/btcd/wire"
	"reflect"
	"testing"
)

// TestUTXOCollection makes sure that utxoCollection cloning and string representations work as expected.
func TestUTXOCollection(t *testing.T) {
	hash0, _ := daghash.NewHashFromStr("0000000000000000000000000000000000000000000000000000000000000000")
	hash1, _ := daghash.NewHashFromStr("1111111111111111111111111111111111111111111111111111111111111111")
	outPoint0 := *wire.NewOutPoint(hash0, 0)
	outPoint1 := *wire.NewOutPoint(hash1, 0)
	utxoEntry0 := newUTXOEntry(&wire.TxOut{PkScript: []byte{}, Value: 10})
	utxoEntry1 := newUTXOEntry(&wire.TxOut{PkScript: []byte{}, Value: 20})

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
	utxoEntry0 := newUTXOEntry(&wire.TxOut{PkScript: []byte{}, Value: 10})
	utxoEntry1 := newUTXOEntry(&wire.TxOut{PkScript: []byte{}, Value: 20})
	diff := utxoDiff{
		toAdd:    utxoCollection{outPoint0: utxoEntry0},
		toRemove: utxoCollection{outPoint1: utxoEntry1},
	}

	// Test utxoDiff creation
	newDiff := newUTXODiff()
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

// TestUTXODiffRules makes sure that all diffFrom and withDiff rules are followed.
// Each test case represents a cell in the two tables outlined in the documentation for utxoDiff.
func TestUTXODiffRules(t *testing.T) {
	hash0, _ := daghash.NewHashFromStr("0000000000000000000000000000000000000000000000000000000000000000")
	outPoint0 := *wire.NewOutPoint(hash0, 0)
	utxoEntry0 := newUTXOEntry(&wire.TxOut{PkScript: []byte{}, Value: 10})

	// For each of the following test cases, we will:
	// this.diffFrom(other) and compare it to expectedDiffFromResult
	// this.withDiff(other) and compare it to expectedWithDiffResult
	//
	// Note: an expected nil result means that we expect the respective operation to fail
	tests := []struct {
		name                   string
		this                   *utxoDiff
		other                  *utxoDiff
		expectedDiffFromResult *utxoDiff
		expectedWithDiffResult *utxoDiff
	}{
		{
			name: "one toAdd in this, one toAdd in other",
			this: &utxoDiff{
				toAdd:    utxoCollection{outPoint0: utxoEntry0},
				toRemove: utxoCollection{},
			},
			other: &utxoDiff{
				toAdd:    utxoCollection{outPoint0: utxoEntry0},
				toRemove: utxoCollection{},
			},
			expectedDiffFromResult: &utxoDiff{
				toAdd:    utxoCollection{},
				toRemove: utxoCollection{},
			},
			expectedWithDiffResult: nil,
		},
		{
			name: "one toAdd in this, one toRemove in other",
			this: &utxoDiff{
				toAdd:    utxoCollection{outPoint0: utxoEntry0},
				toRemove: utxoCollection{},
			},
			other: &utxoDiff{
				toAdd:    utxoCollection{},
				toRemove: utxoCollection{outPoint0: utxoEntry0},
			},
			expectedDiffFromResult: nil,
			expectedWithDiffResult: &utxoDiff{
				toAdd:    utxoCollection{},
				toRemove: utxoCollection{},
			},
		},
		{
			name: "one toAdd in this, empty other",
			this: &utxoDiff{
				toAdd:    utxoCollection{outPoint0: utxoEntry0},
				toRemove: utxoCollection{},
			},
			other: &utxoDiff{
				toAdd:    utxoCollection{},
				toRemove: utxoCollection{},
			},
			expectedDiffFromResult: &utxoDiff{
				toAdd:    utxoCollection{},
				toRemove: utxoCollection{outPoint0: utxoEntry0},
			},
			expectedWithDiffResult: &utxoDiff{
				toAdd:    utxoCollection{outPoint0: utxoEntry0},
				toRemove: utxoCollection{},
			},
		},
		{
			name: "one toRemove in this, one toAdd in other",
			this: &utxoDiff{
				toAdd:    utxoCollection{},
				toRemove: utxoCollection{outPoint0: utxoEntry0},
			},
			other: &utxoDiff{
				toAdd:    utxoCollection{outPoint0: utxoEntry0},
				toRemove: utxoCollection{},
			},
			expectedDiffFromResult: nil,
			expectedWithDiffResult: &utxoDiff{
				toAdd:    utxoCollection{},
				toRemove: utxoCollection{},
			},
		},
		{
			name: "one toRemove in this, one toRemove in other",
			this: &utxoDiff{
				toAdd:    utxoCollection{},
				toRemove: utxoCollection{outPoint0: utxoEntry0},
			},
			other: &utxoDiff{
				toAdd:    utxoCollection{},
				toRemove: utxoCollection{outPoint0: utxoEntry0},
			},
			expectedDiffFromResult: &utxoDiff{
				toAdd:    utxoCollection{},
				toRemove: utxoCollection{},
			},
			expectedWithDiffResult: nil,
		},
		{
			name: "one toRemove in this, empty other",
			this: &utxoDiff{
				toAdd:    utxoCollection{},
				toRemove: utxoCollection{outPoint0: utxoEntry0},
			},
			other: &utxoDiff{
				toAdd:    utxoCollection{},
				toRemove: utxoCollection{},
			},
			expectedDiffFromResult: &utxoDiff{
				toAdd:    utxoCollection{outPoint0: utxoEntry0},
				toRemove: utxoCollection{},
			},
			expectedWithDiffResult: &utxoDiff{
				toAdd:    utxoCollection{},
				toRemove: utxoCollection{outPoint0: utxoEntry0},
			},
		},
		{
			name: "empty this, one toAdd in other",
			this: &utxoDiff{
				toAdd:    utxoCollection{},
				toRemove: utxoCollection{},
			},
			other: &utxoDiff{
				toAdd:    utxoCollection{outPoint0: utxoEntry0},
				toRemove: utxoCollection{},
			},
			expectedDiffFromResult: &utxoDiff{
				toAdd:    utxoCollection{outPoint0: utxoEntry0},
				toRemove: utxoCollection{},
			},
			expectedWithDiffResult: &utxoDiff{
				toAdd:    utxoCollection{outPoint0: utxoEntry0},
				toRemove: utxoCollection{},
			},
		},
		{
			name: "empty this, one toRemove in other",
			this: &utxoDiff{
				toAdd:    utxoCollection{},
				toRemove: utxoCollection{},
			},
			other: &utxoDiff{
				toAdd:    utxoCollection{},
				toRemove: utxoCollection{outPoint0: utxoEntry0},
			},
			expectedDiffFromResult: &utxoDiff{
				toAdd:    utxoCollection{},
				toRemove: utxoCollection{outPoint0: utxoEntry0},
			},
			expectedWithDiffResult: &utxoDiff{
				toAdd:    utxoCollection{},
				toRemove: utxoCollection{outPoint0: utxoEntry0},
			},
		},
		{
			name: "empty this, empty other",
			this: &utxoDiff{
				toAdd:    utxoCollection{},
				toRemove: utxoCollection{},
			},
			other: &utxoDiff{
				toAdd:    utxoCollection{},
				toRemove: utxoCollection{},
			},
			expectedDiffFromResult: &utxoDiff{
				toAdd:    utxoCollection{},
				toRemove: utxoCollection{},
			},
			expectedWithDiffResult: &utxoDiff{
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

		// withDiff from this to other
		withDiffResult, err := test.this.withDiff(test.other)

		// Test whether withDiff returned an error
		isWithDiffOk := err == nil
		expectedIsWithDiffOk := test.expectedWithDiffResult != nil
		if isWithDiffOk != expectedIsWithDiffOk {
			t.Errorf("unexpected withDiff error in test \"%s\". "+
				"Expected: \"%t\", got: \"%t\".", test.name, expectedIsWithDiffOk, isWithDiffOk)
		}

		// Ig not error, test the withDiff result
		if isWithDiffOk && !reflect.DeepEqual(withDiffResult, test.expectedWithDiffResult) {
			t.Errorf("unexpected withDiff result in test \"%s\". "+
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
	utxoEntry0 := newUTXOEntry(txOut0)
	utxoEntry1 := newUTXOEntry(txOut1)
	diff := &utxoDiff{
		toAdd:    utxoCollection{outPoint0: utxoEntry0},
		toRemove: utxoCollection{outPoint1: utxoEntry1},
	}

	// Test fullUTXOSet creation
	emptySet := newFullUTXOSet()
	if len(emptySet.collection()) != 0 {
		t.Errorf("new set is not empty")
	}

	// Test fullUTXOSet withDiff
	withDiffResult, err := emptySet.withDiff(diff)
	if err != nil {
		t.Errorf("withDiff unexpectedly failed")
	}
	withDiffUTXOSet, ok := withDiffResult.(*diffUTXOSet)
	if !ok {
		t.Errorf("withDiff is of unexpected type")
	}
	if !reflect.DeepEqual(withDiffUTXOSet.base, emptySet) || !reflect.DeepEqual(withDiffUTXOSet.utxoDiff, diff) {
		t.Errorf("withDiff is of unexpected composition")
	}

	// Test fullUTXOSet addTx
	txIn0 := &wire.TxIn{SignatureScript: []byte{}, PreviousOutPoint: wire.OutPoint{Hash: *hash0, Index: 0}, Sequence: 0}
	transaction0 := wire.NewMsgTx(1)
	transaction0.TxIn = []*wire.TxIn{txIn0}
	transaction0.TxOut = []*wire.TxOut{txOut0}
	if ok = emptySet.addTx(transaction0); ok {
		t.Errorf("addTx unexpectedly succeeded")
	}
	emptySet = &fullUTXOSet{utxoCollection: utxoCollection{outPoint0: utxoEntry0}}
	if ok = emptySet.addTx(transaction0); !ok {
		t.Errorf("addTx unexpectedly failed")
	}

	// Test fullUTXOSet collection
	if !reflect.DeepEqual(emptySet.collection(), emptySet.utxoCollection) {
		t.Errorf("collection does not equal the set's utxoCollection")
	}

	// Test fullUTXOSet cloning
	clonedEmptySet := emptySet.clone().(*fullUTXOSet)
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
	utxoEntry0 := newUTXOEntry(txOut0)
	utxoEntry1 := newUTXOEntry(txOut1)
	diff := &utxoDiff{
		toAdd:    utxoCollection{outPoint0: utxoEntry0},
		toRemove: utxoCollection{outPoint1: utxoEntry1},
	}

	// Test diffUTXOSet creation
	emptySet := newDiffUTXOSet(newFullUTXOSet(), newUTXODiff())
	if len(emptySet.collection()) != 0 {
		t.Errorf("new set is not empty")
	}

	// Test diffUTXOSet withDiff
	withDiffResult, err := emptySet.withDiff(diff)
	if err != nil {
		t.Errorf("withDiff unexpectedly failed")
	}
	withDiffUTXOSet, ok := withDiffResult.(*diffUTXOSet)
	if !ok {
		t.Errorf("withDiff is of unexpected type")
	}
	withDiff, _ := newUTXODiff().withDiff(diff)
	if !reflect.DeepEqual(withDiffUTXOSet.base, emptySet.base) || !reflect.DeepEqual(withDiffUTXOSet.utxoDiff, withDiff) {
		t.Errorf("withDiff is of unexpected composition")
	}
	_, err = newDiffUTXOSet(newFullUTXOSet(), diff).withDiff(diff)
	if err == nil {
		t.Errorf("withDiff unexpectedly succeeded")
	}

	// Given a diffSet, each case tests that meldToBase, String, collection, and cloning work as expected
	// For each of the following test cases, we will:
	// .meldToBase() the given diffSet and compare it to expectedMeldSet
	// .String() the given diffSet and compare it to expectedString
	// .collection() the given diffSet and compare it to expectedCollection
	// .clone() the given diffSet and compare its value to itself (expected: equals) and its reference to itself (expected: not equal)
	tests := []struct {
		name               string
		diffSet            *diffUTXOSet
		expectedMeldSet    *diffUTXOSet
		expectedString     string
		expectedCollection utxoCollection
	}{
		{
			name: "empty base, empty diff",
			diffSet: &diffUTXOSet{
				base: newFullUTXOSet(),
				utxoDiff: &utxoDiff{
					toAdd:    utxoCollection{},
					toRemove: utxoCollection{},
				},
			},
			expectedMeldSet: &diffUTXOSet{
				base: newFullUTXOSet(),
				utxoDiff: &utxoDiff{
					toAdd:    utxoCollection{},
					toRemove: utxoCollection{},
				},
			},
			expectedString:     "{Base: [  ], To Add: [  ], To Remove: [  ]}",
			expectedCollection: utxoCollection{},
		},
		{
			name: "empty base, one member in diff toAdd",
			diffSet: &diffUTXOSet{
				base: newFullUTXOSet(),
				utxoDiff: &utxoDiff{
					toAdd:    utxoCollection{outPoint0: utxoEntry0},
					toRemove: utxoCollection{},
				},
			},
			expectedMeldSet: &diffUTXOSet{
				base: &fullUTXOSet{utxoCollection: utxoCollection{outPoint0: utxoEntry0}},
				utxoDiff: &utxoDiff{
					toAdd:    utxoCollection{},
					toRemove: utxoCollection{},
				},
			},
			expectedString:     "{Base: [  ], To Add: [ (0000000000000000000000000000000000000000000000000000000000000000, 0) => 10 ], To Remove: [  ]}",
			expectedCollection: utxoCollection{outPoint0: utxoEntry0},
		},
		{
			name: "empty base, one member in diff toRemove",
			diffSet: &diffUTXOSet{
				base: newFullUTXOSet(),
				utxoDiff: &utxoDiff{
					toAdd:    utxoCollection{},
					toRemove: utxoCollection{outPoint0: utxoEntry0},
				},
			},
			expectedMeldSet: &diffUTXOSet{
				base: &fullUTXOSet{utxoCollection: utxoCollection{}},
				utxoDiff: &utxoDiff{
					toAdd:    utxoCollection{},
					toRemove: utxoCollection{},
				},
			},
			expectedString:     "{Base: [  ], To Add: [  ], To Remove: [ (0000000000000000000000000000000000000000000000000000000000000000, 0) => 10 ]}",
			expectedCollection: utxoCollection{},
		},
		{
			name: "one member in base toAdd, one member in diff toAdd",
			diffSet: &diffUTXOSet{
				base: &fullUTXOSet{utxoCollection: utxoCollection{outPoint0: utxoEntry0}},
				utxoDiff: &utxoDiff{
					toAdd:    utxoCollection{outPoint1: utxoEntry1},
					toRemove: utxoCollection{},
				},
			},
			expectedMeldSet: &diffUTXOSet{
				base: &fullUTXOSet{
					utxoCollection: utxoCollection{
						outPoint0: utxoEntry0,
						outPoint1: utxoEntry1,
					},
				},
				utxoDiff: &utxoDiff{
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
			diffSet: &diffUTXOSet{
				base: &fullUTXOSet{utxoCollection: utxoCollection{outPoint0: utxoEntry0}},
				utxoDiff: &utxoDiff{
					toAdd:    utxoCollection{},
					toRemove: utxoCollection{outPoint0: utxoEntry0},
				},
			},
			expectedMeldSet: &diffUTXOSet{
				base: &fullUTXOSet{
					utxoCollection: utxoCollection{},
				},
				utxoDiff: &utxoDiff{
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
		meldSet := test.diffSet.clone().(*diffUTXOSet)
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
		clonedSet := test.diffSet.clone().(*diffUTXOSet)
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
	fullSet := newFullUTXOSet()
	diffSet := newDiffUTXOSet(fullSet, newUTXODiff())

	// For each of the following test cases, we will call utxoSet.diffFrom(diffSet) and compare
	// whether the function succeeded with expectedSuccess
	//
	// Note: since test cases are similar for both fullUTXOSet and diffUTXOSet, we test both using the same test cases
	run := func(set utxoSet) {
		tests := []struct {
			name            string
			diffSet         utxoSet
			expectedSuccess bool
		}{
			{
				name:            "diff from fullSet",
				diffSet:         newFullUTXOSet(),
				expectedSuccess: false,
			},
			{
				name:            "diff from diffSet with different base",
				diffSet:         newDiffUTXOSet(newFullUTXOSet(), newUTXODiff()),
				expectedSuccess: false,
			},
			{
				name:            "diff from diffSet with same base",
				diffSet:         newDiffUTXOSet(fullSet, newUTXODiff()),
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
	// transaction0 is coinbase. As such, it does not have any inputs
	txOut0 := &wire.TxOut{PkScript: []byte{0}, Value: 10}
	utxoEntry0 := newUTXOEntry(txOut0)
	transaction0 := wire.NewMsgTx(1)
	transaction0.TxIn = []*wire.TxIn{}
	transaction0.TxOut = []*wire.TxOut{txOut0}

	// transaction1 spends transaction0
	hash1 := transaction0.TxHash()
	outPoint1 := *wire.NewOutPoint(&hash1, 0)
	txIn1 := &wire.TxIn{SignatureScript: []byte{}, PreviousOutPoint: wire.OutPoint{Hash: hash1, Index: 0}, Sequence: 0}
	txOut1 := &wire.TxOut{PkScript: []byte{1}, Value: 20}
	utxoEntry1 := newUTXOEntry(txOut1)
	transaction1 := wire.NewMsgTx(1)
	transaction1.TxIn = []*wire.TxIn{txIn1}
	transaction1.TxOut = []*wire.TxOut{txOut1}

	// transaction2 spends transaction1
	hash2 := transaction1.TxHash()
	outPoint2 := *wire.NewOutPoint(&hash2, 0)
	txIn2 := &wire.TxIn{SignatureScript: []byte{}, PreviousOutPoint: wire.OutPoint{Hash: hash2, Index: 0}, Sequence: 0}
	txOut2 := &wire.TxOut{PkScript: []byte{2}, Value: 30}
	utxoEntry2 := newUTXOEntry(txOut2)
	transaction2 := wire.NewMsgTx(1)
	transaction2.TxIn = []*wire.TxIn{txIn2}
	transaction2.TxOut = []*wire.TxOut{txOut2}

	// outpoint3 is the outpoint for transaction2
	hash3 := transaction2.TxHash()
	outPoint3 := *wire.NewOutPoint(&hash3, 0)

	// For each of the following test cases, we will:
	// 1. startSet.addTx() all the transactions in toAdd, in order
	// 2. Compare the result set with expectedSet
	tests := []struct {
		name        string
		startSet    *diffUTXOSet
		toAdd       []*wire.MsgTx
		expectedSet *diffUTXOSet
	}{
		{
			name:     "add coinbase transaction to empty set",
			startSet: newDiffUTXOSet(newFullUTXOSet(), newUTXODiff()),
			toAdd:    []*wire.MsgTx{transaction0},
			expectedSet: &diffUTXOSet{
				base: &fullUTXOSet{utxoCollection: utxoCollection{}},
				utxoDiff: &utxoDiff{
					toAdd:    utxoCollection{outPoint1: utxoEntry0},
					toRemove: utxoCollection{},
				},
			},
		},
		{
			name:     "add regular transaction to empty set",
			startSet: newDiffUTXOSet(newFullUTXOSet(), newUTXODiff()),
			toAdd:    []*wire.MsgTx{transaction1},
			expectedSet: &diffUTXOSet{
				base: &fullUTXOSet{utxoCollection: utxoCollection{}},
				utxoDiff: &utxoDiff{
					toAdd:    utxoCollection{},
					toRemove: utxoCollection{},
				},
			},
		},
		{
			name: "add transaction to set with its input in base",
			startSet: &diffUTXOSet{
				base: &fullUTXOSet{utxoCollection: utxoCollection{outPoint1: utxoEntry0}},
				utxoDiff: &utxoDiff{
					toAdd:    utxoCollection{},
					toRemove: utxoCollection{},
				},
			},
			toAdd: []*wire.MsgTx{transaction1},
			expectedSet: &diffUTXOSet{
				base: &fullUTXOSet{utxoCollection: utxoCollection{outPoint1: utxoEntry0}},
				utxoDiff: &utxoDiff{
					toAdd:    utxoCollection{outPoint2: utxoEntry1},
					toRemove: utxoCollection{outPoint1: utxoEntry0},
				},
			},
		},
		{
			name: "add transaction to set with its input in diff toAdd",
			startSet: &diffUTXOSet{
				base: newFullUTXOSet(),
				utxoDiff: &utxoDiff{
					toAdd:    utxoCollection{outPoint1: utxoEntry0},
					toRemove: utxoCollection{},
				},
			},
			toAdd: []*wire.MsgTx{transaction1},
			expectedSet: &diffUTXOSet{
				base: newFullUTXOSet(),
				utxoDiff: &utxoDiff{
					toAdd:    utxoCollection{outPoint2: utxoEntry1},
					toRemove: utxoCollection{},
				},
			},
		},
		{
			name: "add transaction to set with its input in diff toAdd and its output in diff toRemove",
			startSet: &diffUTXOSet{
				base: newFullUTXOSet(),
				utxoDiff: &utxoDiff{
					toAdd:    utxoCollection{outPoint1: utxoEntry0},
					toRemove: utxoCollection{outPoint2: utxoEntry1},
				},
			},
			toAdd: []*wire.MsgTx{transaction1},
			expectedSet: &diffUTXOSet{
				base: newFullUTXOSet(),
				utxoDiff: &utxoDiff{
					toAdd:    utxoCollection{},
					toRemove: utxoCollection{},
				},
			},
		},
		{
			name: "add two transactions, one spending the other, to set with the first input in base",
			startSet: &diffUTXOSet{
				base: &fullUTXOSet{utxoCollection: utxoCollection{outPoint1: utxoEntry0}},
				utxoDiff: &utxoDiff{
					toAdd:    utxoCollection{},
					toRemove: utxoCollection{},
				},
			},
			toAdd: []*wire.MsgTx{transaction1, transaction2},
			expectedSet: &diffUTXOSet{
				base: &fullUTXOSet{utxoCollection: utxoCollection{outPoint1: utxoEntry0}},
				utxoDiff: &utxoDiff{
					toAdd:    utxoCollection{outPoint3: utxoEntry2},
					toRemove: utxoCollection{outPoint1: utxoEntry0},
				},
			},
		},
	}

	for _, test := range tests {
		diffSet := test.startSet.clone()

		// Apply all transactions, in order, to diffSet
		for _, transaction := range test.toAdd {
			diffSet.addTx(transaction)
		}

		// Make sure that the result diffSet equals to the expectedSet
		if !reflect.DeepEqual(diffSet, test.expectedSet) {
			t.Errorf("unexpected diffSet in test \"%s\". "+
				"Expected: \"%v\", got: \"%v\".", test.name, test.expectedSet, diffSet)
		}
	}
}
