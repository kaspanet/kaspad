package blockdag

import (
	"math"
	"reflect"
	"testing"

	"github.com/daglabs/btcd/btcec"
	"github.com/daglabs/btcd/dagconfig"
	"github.com/daglabs/btcd/util/daghash"
	"github.com/daglabs/btcd/wire"
)

// TestUTXOCollection makes sure that utxoCollection cloning and string representations work as expected.
func TestUTXOCollection(t *testing.T) {
	txID0, _ := daghash.NewTxIDFromStr("0000000000000000000000000000000000000000000000000000000000000000")
	txID1, _ := daghash.NewTxIDFromStr("1111111111111111111111111111111111111111111111111111111111111111")
	outpoint0 := *wire.NewOutpoint(txID0, 0)
	outpoint1 := *wire.NewOutpoint(txID1, 0)
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
				outpoint0: utxoEntry1,
			},
			expectedString: "[ (0000000000000000000000000000000000000000000000000000000000000000, 0) => 20 ]",
		},
		{
			name: "two members",
			collection: utxoCollection{
				outpoint0: utxoEntry0,
				outpoint1: utxoEntry1,
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
	txID0, _ := daghash.NewTxIDFromStr("0000000000000000000000000000000000000000000000000000000000000000")
	txID1, _ := daghash.NewTxIDFromStr("1111111111111111111111111111111111111111111111111111111111111111")
	outpoint0 := *wire.NewOutpoint(txID0, 0)
	outpoint1 := *wire.NewOutpoint(txID1, 0)
	utxoEntry0 := NewUTXOEntry(&wire.TxOut{PkScript: []byte{}, Value: 10}, true, 0)
	utxoEntry1 := NewUTXOEntry(&wire.TxOut{PkScript: []byte{}, Value: 20}, false, 1)

	// Test utxoDiff creation
	diff := NewUTXODiff()
	if len(diff.toAdd) != 0 || len(diff.toRemove) != 0 {
		t.Errorf("new diff is not empty")
	}

	err := diff.AddEntry(outpoint0, utxoEntry0)
	if err != nil {
		t.Fatalf("Error adding entry to utxo diff: %s", err)
	}

	err = diff.RemoveEntry(outpoint1, utxoEntry1)
	if err != nil {
		t.Fatalf("Error adding entry to utxo diff: %s", err)
	}

	// Test utxoDiff cloning
	clonedDiff := diff.clone()
	if clonedDiff == diff {
		t.Errorf("cloned diff is reference-equal to the original")
	}
	if !reflect.DeepEqual(clonedDiff, diff) {
		t.Errorf("cloned diff not equal to the original"+
			"Original: \"%v\", cloned: \"%v\".", diff, clonedDiff)
	}

	// Test utxoDiff string representation
	expectedDiffString := "toAdd: [ (0000000000000000000000000000000000000000000000000000000000000000, 0) => 10 ]; toRemove: [ (1111111111111111111111111111111111111111111111111111111111111111, 0) => 20 ], Multiset-Hash: 7cb61e48005b0c817211d04589d719bff87d86a6a6ce2454515f57265382ded7"
	diffString := clonedDiff.String()
	if diffString != expectedDiffString {
		t.Errorf("unexpected diff string. "+
			"Expected: \"%s\", got: \"%s\".", expectedDiffString, diffString)
	}
}

// TestUTXODiffRules makes sure that all diffFrom and WithDiff rules are followed.
// Each test case represents a cell in the two tables outlined in the documentation for utxoDiff.
func TestUTXODiffRules(t *testing.T) {
	txID0, _ := daghash.NewTxIDFromStr("0000000000000000000000000000000000000000000000000000000000000000")
	outpoint0 := *wire.NewOutpoint(txID0, 0)
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
				toAdd:    utxoCollection{outpoint0: utxoEntry0},
				toRemove: utxoCollection{},
			},
			other: &UTXODiff{
				toAdd:    utxoCollection{outpoint0: utxoEntry0},
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
				toAdd:    utxoCollection{outpoint0: utxoEntry0},
				toRemove: utxoCollection{},
			},
			other: &UTXODiff{
				toAdd:    utxoCollection{},
				toRemove: utxoCollection{outpoint0: utxoEntry0},
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
				toAdd:    utxoCollection{outpoint0: utxoEntry0},
				toRemove: utxoCollection{},
			},
			other: &UTXODiff{
				toAdd:    utxoCollection{},
				toRemove: utxoCollection{},
			},
			expectedDiffFromResult: &UTXODiff{
				toAdd:    utxoCollection{},
				toRemove: utxoCollection{outpoint0: utxoEntry0},
			},
			expectedWithDiffResult: &UTXODiff{
				toAdd:    utxoCollection{outpoint0: utxoEntry0},
				toRemove: utxoCollection{},
			},
		},
		{
			name: "one toRemove in this, one toAdd in other",
			this: &UTXODiff{
				toAdd:    utxoCollection{},
				toRemove: utxoCollection{outpoint0: utxoEntry0},
			},
			other: &UTXODiff{
				toAdd:    utxoCollection{outpoint0: utxoEntry0},
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
				toRemove: utxoCollection{outpoint0: utxoEntry0},
			},
			other: &UTXODiff{
				toAdd:    utxoCollection{},
				toRemove: utxoCollection{outpoint0: utxoEntry0},
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
				toRemove: utxoCollection{outpoint0: utxoEntry0},
			},
			other: &UTXODiff{
				toAdd:    utxoCollection{},
				toRemove: utxoCollection{},
			},
			expectedDiffFromResult: &UTXODiff{
				toAdd:    utxoCollection{outpoint0: utxoEntry0},
				toRemove: utxoCollection{},
			},
			expectedWithDiffResult: &UTXODiff{
				toAdd:    utxoCollection{},
				toRemove: utxoCollection{outpoint0: utxoEntry0},
			},
		},
		{
			name: "empty this, one toAdd in other",
			this: &UTXODiff{
				toAdd:    utxoCollection{},
				toRemove: utxoCollection{},
			},
			other: &UTXODiff{
				toAdd:    utxoCollection{outpoint0: utxoEntry0},
				toRemove: utxoCollection{},
			},
			expectedDiffFromResult: &UTXODiff{
				toAdd:    utxoCollection{outpoint0: utxoEntry0},
				toRemove: utxoCollection{},
			},
			expectedWithDiffResult: &UTXODiff{
				toAdd:    utxoCollection{outpoint0: utxoEntry0},
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
				toRemove: utxoCollection{outpoint0: utxoEntry0},
			},
			expectedDiffFromResult: &UTXODiff{
				toAdd:    utxoCollection{},
				toRemove: utxoCollection{outpoint0: utxoEntry0},
			},
			expectedWithDiffResult: &UTXODiff{
				toAdd:    utxoCollection{},
				toRemove: utxoCollection{outpoint0: utxoEntry0},
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
		this := addMultisetToDiff(t, test.this)
		other := addMultisetToDiff(t, test.other)
		expectedDiffFromResult := addMultisetToDiff(t, test.expectedDiffFromResult)
		expectedWithDiffResult := addMultisetToDiff(t, test.expectedWithDiffResult)

		// diffFrom from this to other
		diffResult, err := this.diffFrom(other)

		// Test whether diffFrom returned an error
		isDiffFromOk := err == nil
		expectedIsDiffFromOk := expectedDiffFromResult != nil
		if isDiffFromOk != expectedIsDiffFromOk {
			t.Errorf("unexpected diffFrom error in test \"%s\". "+
				"Expected: \"%t\", got: \"%t\".", test.name, expectedIsDiffFromOk, isDiffFromOk)
		}

		// If not error, test the diffFrom result
		if isDiffFromOk && !expectedDiffFromResult.equal(diffResult) {
			t.Errorf("unexpected diffFrom result in test \"%s\". "+
				"Expected: \"%v\", got: \"%v\".", test.name, expectedDiffFromResult, diffResult)
		}

		// WithDiff from this to other
		withDiffResult, err := this.WithDiff(other)

		// Test whether WithDiff returned an error
		isWithDiffOk := err == nil
		expectedIsWithDiffOk := expectedWithDiffResult != nil
		if isWithDiffOk != expectedIsWithDiffOk {
			t.Errorf("unexpected WithDiff error in test \"%s\". "+
				"Expected: \"%t\", got: \"%t\".", test.name, expectedIsWithDiffOk, isWithDiffOk)
		}

		// If not error, test the WithDiff result
		if isWithDiffOk && !withDiffResult.equal(expectedWithDiffResult) {
			t.Errorf("unexpected WithDiff result in test \"%s\". "+
				"Expected: \"%v\", got: \"%v\".", test.name, expectedWithDiffResult, withDiffResult)
		}
	}
}

func areMultisetsEqual(a *btcec.Multiset, b *btcec.Multiset) bool {
	aX, aY := a.Point()
	bX, bY := b.Point()
	return aX.Cmp(bX) == 0 && aY.Cmp(bY) == 0
}

func (d *UTXODiff) equal(other *UTXODiff) bool {
	return reflect.DeepEqual(d.toAdd, other.toAdd) &&
		reflect.DeepEqual(d.toRemove, other.toRemove) &&
		areMultisetsEqual(d.diffMultiset, other.diffMultiset)
}

func (fus *FullUTXOSet) equal(other *FullUTXOSet) bool {
	return reflect.DeepEqual(fus.utxoCollection, other.utxoCollection) &&
		areMultisetsEqual(fus.UTXOMultiset, other.UTXOMultiset)
}

func (dus *DiffUTXOSet) equal(other *DiffUTXOSet) bool {
	return dus.base.equal(other.base) && dus.UTXODiff.equal(other.UTXODiff)
}

func addMultisetToDiff(t *testing.T, diff *UTXODiff) *UTXODiff {
	if diff == nil {
		return nil
	}
	diffWithMs := NewUTXODiff()
	for outpoint, entry := range diff.toAdd {
		err := diffWithMs.AddEntry(outpoint, entry)
		if err != nil {
			t.Fatalf("Error with diffWithMs.AddEntry: %s", err)
		}
	}
	for outpoint, entry := range diff.toRemove {
		err := diffWithMs.RemoveEntry(outpoint, entry)
		if err != nil {
			t.Fatalf("Error with diffWithMs.removeEntry: %s", err)
		}
	}
	return diffWithMs
}

func addMultisetToFullUTXOSet(t *testing.T, fus *FullUTXOSet) *FullUTXOSet {
	if fus == nil {
		return nil
	}
	fusWithMs := NewFullUTXOSet()
	for outpoint, entry := range fus.utxoCollection {
		err := fusWithMs.addAndUpdateMultiset(outpoint, entry)
		if err != nil {
			t.Fatalf("Error with diffWithMs.AddEntry: %s", err)
		}
	}
	return fusWithMs
}

func addMultisetToDiffUTXOSet(t *testing.T, diffSet *DiffUTXOSet) *DiffUTXOSet {
	if diffSet == nil {
		return nil
	}
	diffWithMs := addMultisetToDiff(t, diffSet.UTXODiff)
	baseWithMs := addMultisetToFullUTXOSet(t, diffSet.base)
	return NewDiffUTXOSet(baseWithMs, diffWithMs)
}

// TestFullUTXOSet makes sure that fullUTXOSet is working as expected.
func TestFullUTXOSet(t *testing.T) {
	txID0, _ := daghash.NewTxIDFromStr("0000000000000000000000000000000000000000000000000000000000000000")
	txID1, _ := daghash.NewTxIDFromStr("1111111111111111111111111111111111111111111111111111111111111111")
	outpoint0 := *wire.NewOutpoint(txID0, 0)
	outpoint1 := *wire.NewOutpoint(txID1, 0)
	txOut0 := &wire.TxOut{PkScript: []byte{}, Value: 10}
	txOut1 := &wire.TxOut{PkScript: []byte{}, Value: 20}
	utxoEntry0 := NewUTXOEntry(txOut0, true, 0)
	utxoEntry1 := NewUTXOEntry(txOut1, false, 1)
	diff := addMultisetToDiff(t, &UTXODiff{
		toAdd:    utxoCollection{outpoint0: utxoEntry0},
		toRemove: utxoCollection{outpoint1: utxoEntry1},
	})

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
	txIn0 := &wire.TxIn{SignatureScript: []byte{}, PreviousOutpoint: wire.Outpoint{TxID: *txID0, Index: 0}, Sequence: 0}
	transaction0 := wire.NewNativeMsgTx(1, []*wire.TxIn{txIn0}, []*wire.TxOut{txOut0})
	if isAccepted, err := emptySet.AddTx(transaction0, 0); err != nil {
		t.Errorf("AddTx unexpectedly failed: %s", err)
	} else if isAccepted {
		t.Errorf("addTx unexpectedly succeeded")
	}
	emptySet = addMultisetToFullUTXOSet(t, &FullUTXOSet{utxoCollection: utxoCollection{outpoint0: utxoEntry0}})
	if isAccepted, err := emptySet.AddTx(transaction0, 0); err != nil {
		t.Errorf("addTx unexpectedly failed. Error: %s", err)
	} else if !isAccepted {
		t.Fatalf("AddTx unexpectedly didn't add tx %s", transaction0.TxID())
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
	txID0, _ := daghash.NewTxIDFromStr("0000000000000000000000000000000000000000000000000000000000000000")
	txID1, _ := daghash.NewTxIDFromStr("1111111111111111111111111111111111111111111111111111111111111111")
	outpoint0 := *wire.NewOutpoint(txID0, 0)
	outpoint1 := *wire.NewOutpoint(txID1, 0)
	txOut0 := &wire.TxOut{PkScript: []byte{}, Value: 10}
	txOut1 := &wire.TxOut{PkScript: []byte{}, Value: 20}
	utxoEntry0 := NewUTXOEntry(txOut0, true, 0)
	utxoEntry1 := NewUTXOEntry(txOut1, false, 1)
	diff := addMultisetToDiff(t, &UTXODiff{
		toAdd:    utxoCollection{outpoint0: utxoEntry0},
		toRemove: utxoCollection{outpoint1: utxoEntry1},
	})

	// Test diffUTXOSet creation
	emptySet := NewDiffUTXOSet(NewFullUTXOSet(), NewUTXODiff())
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
		name                    string
		diffSet                 *DiffUTXOSet
		expectedMeldSet         *DiffUTXOSet
		expectedString          string
		expectedCollection      utxoCollection
		expectedMeldToBaseError string
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
			expectedString:     "{Base: [  ], To Add: [  ], To Remove: [  ], Multiset-Hash:0000000000000000000000000000000000000000000000000000000000000000}",
			expectedCollection: utxoCollection{},
		},
		{
			name: "empty base, one member in diff toAdd",
			diffSet: &DiffUTXOSet{
				base: NewFullUTXOSet(),
				UTXODiff: &UTXODiff{
					toAdd:    utxoCollection{outpoint0: utxoEntry0},
					toRemove: utxoCollection{},
				},
			},
			expectedMeldSet: &DiffUTXOSet{
				base: &FullUTXOSet{utxoCollection: utxoCollection{outpoint0: utxoEntry0}},
				UTXODiff: &UTXODiff{
					toAdd:    utxoCollection{},
					toRemove: utxoCollection{},
				},
			},
			expectedString:     "{Base: [  ], To Add: [ (0000000000000000000000000000000000000000000000000000000000000000, 0) => 10 ], To Remove: [  ], Multiset-Hash:da4768bd0359c3426268d6707c1fc17a68c45ef1ea734331b07568418234487f}",
			expectedCollection: utxoCollection{outpoint0: utxoEntry0},
		},
		{
			name: "empty base, one member in diff toRemove",
			diffSet: &DiffUTXOSet{
				base: NewFullUTXOSet(),
				UTXODiff: &UTXODiff{
					toAdd:    utxoCollection{},
					toRemove: utxoCollection{outpoint0: utxoEntry0},
				},
			},
			expectedMeldSet:         nil,
			expectedString:          "{Base: [  ], To Add: [  ], To Remove: [ (0000000000000000000000000000000000000000000000000000000000000000, 0) => 10 ], Multiset-Hash:046242cb1bb1e6d3fd91d0f181e1b2d4a597ac57fa2584fc3c2eb0e0f46c9369}",
			expectedCollection:      utxoCollection{},
			expectedMeldToBaseError: "Couldn't remove outpoint 0000000000000000000000000000000000000000000000000000000000000000:0 because it doesn't exist in the DiffUTXOSet base",
		},
		{
			name: "one member in base toAdd, one member in diff toAdd",
			diffSet: &DiffUTXOSet{
				base: &FullUTXOSet{utxoCollection: utxoCollection{outpoint0: utxoEntry0}},
				UTXODiff: &UTXODiff{
					toAdd:    utxoCollection{outpoint1: utxoEntry1},
					toRemove: utxoCollection{},
				},
			},
			expectedMeldSet: &DiffUTXOSet{
				base: &FullUTXOSet{
					utxoCollection: utxoCollection{
						outpoint0: utxoEntry0,
						outpoint1: utxoEntry1,
					},
				},
				UTXODiff: &UTXODiff{
					toAdd:    utxoCollection{},
					toRemove: utxoCollection{},
				},
			},
			expectedString: "{Base: [ (0000000000000000000000000000000000000000000000000000000000000000, 0) => 10 ], To Add: [ (1111111111111111111111111111111111111111111111111111111111111111, 0) => 20 ], To Remove: [  ], Multiset-Hash:556cc61fd4d7e74d7807ca2298c5320375a6a20310a18920e54667220924baff}",
			expectedCollection: utxoCollection{
				outpoint0: utxoEntry0,
				outpoint1: utxoEntry1,
			},
		},
		{
			name: "one member in base toAdd, same one member in diff toRemove",
			diffSet: &DiffUTXOSet{
				base: &FullUTXOSet{utxoCollection: utxoCollection{outpoint0: utxoEntry0}},
				UTXODiff: &UTXODiff{
					toAdd:    utxoCollection{},
					toRemove: utxoCollection{outpoint0: utxoEntry0},
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
			expectedString:     "{Base: [ (0000000000000000000000000000000000000000000000000000000000000000, 0) => 10 ], To Add: [  ], To Remove: [ (0000000000000000000000000000000000000000000000000000000000000000, 0) => 10 ], Multiset-Hash:0000000000000000000000000000000000000000000000000000000000000000}",
			expectedCollection: utxoCollection{},
		},
	}

	for _, test := range tests {
		diffSet := addMultisetToDiffUTXOSet(t, test.diffSet)
		expectedMeldSet := addMultisetToDiffUTXOSet(t, test.expectedMeldSet)

		// Test string representation
		setString := diffSet.String()
		if setString != test.expectedString {
			t.Errorf("unexpected string in test \"%s\". "+
				"Expected: \"%s\", got: \"%s\".", test.name, test.expectedString, setString)
		}

		// Test meldToBase
		meldSet := diffSet.clone().(*DiffUTXOSet)
		err := meldSet.meldToBase()
		errString := ""
		if err != nil {
			errString = err.Error()
		}
		if test.expectedMeldToBaseError != errString {
			t.Errorf("meldToBase in test \"%s\" expected error \"%s\" but got: \"%s\"", test.name, test.expectedMeldToBaseError, errString)
		}
		if err != nil {
			continue
		}
		if !meldSet.equal(expectedMeldSet) {
			t.Errorf("unexpected melded set in test \"%s\". "+
				"Expected: \"%v\", got: \"%v\".", test.name, expectedMeldSet, meldSet)
		}

		// Test collection
		setCollection, err := diffSet.collection()
		if err != nil {
			t.Errorf("Error getting diffSet collection: %s", err)
		} else if !reflect.DeepEqual(setCollection, test.expectedCollection) {
			t.Errorf("unexpected set collection in test \"%s\". "+
				"Expected: \"%v\", got: \"%v\".", test.name, test.expectedCollection, setCollection)
		}

		// Test cloning
		clonedSet := diffSet.clone().(*DiffUTXOSet)
		if !reflect.DeepEqual(clonedSet, diffSet) {
			t.Errorf("unexpected set clone in test \"%s\". "+
				"Expected: \"%v\", got: \"%v\".", test.name, diffSet, clonedSet)
		}
		if clonedSet == diffSet {
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
	txID0, _ := daghash.NewTxIDFromStr("0000000000000000000000000000000000000000000000000000000000000000")
	txIn0 := &wire.TxIn{SignatureScript: []byte{}, PreviousOutpoint: wire.Outpoint{TxID: *txID0, Index: math.MaxUint32}, Sequence: 0}
	txOut0 := &wire.TxOut{PkScript: []byte{0}, Value: 10}
	utxoEntry0 := NewUTXOEntry(txOut0, true, 0)
	transaction0 := wire.NewNativeMsgTx(1, []*wire.TxIn{txIn0}, []*wire.TxOut{txOut0})

	// transaction1 spends transaction0
	id1 := transaction0.TxID()
	outpoint1 := *wire.NewOutpoint(id1, 0)
	txIn1 := &wire.TxIn{SignatureScript: []byte{}, PreviousOutpoint: outpoint1, Sequence: 0}
	txOut1 := &wire.TxOut{PkScript: []byte{1}, Value: 20}
	utxoEntry1 := NewUTXOEntry(txOut1, false, 1)
	transaction1 := wire.NewNativeMsgTx(1, []*wire.TxIn{txIn1}, []*wire.TxOut{txOut1})

	// transaction2 spends transaction1
	id2 := transaction1.TxID()
	outpoint2 := *wire.NewOutpoint(id2, 0)
	txIn2 := &wire.TxIn{SignatureScript: []byte{}, PreviousOutpoint: outpoint2, Sequence: 0}
	txOut2 := &wire.TxOut{PkScript: []byte{2}, Value: 30}
	utxoEntry2 := NewUTXOEntry(txOut2, false, 2)
	transaction2 := wire.NewNativeMsgTx(1, []*wire.TxIn{txIn2}, []*wire.TxOut{txOut2})

	// outpoint3 is the outpoint for transaction2
	id3 := transaction2.TxID()
	outpoint3 := *wire.NewOutpoint(id3, 0)

	// For each of the following test cases, we will:
	// 1. startSet.addTx() all the transactions in toAdd, in order, with the initial block height startHeight
	// 2. Compare the result set with expectedSet
	tests := []struct {
		name        string
		startSet    *DiffUTXOSet
		startHeight uint64
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
					toAdd:    utxoCollection{outpoint1: utxoEntry0},
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
				base: &FullUTXOSet{utxoCollection: utxoCollection{outpoint1: utxoEntry0}},
				UTXODiff: &UTXODiff{
					toAdd:    utxoCollection{},
					toRemove: utxoCollection{},
				},
			},
			startHeight: 1,
			toAdd:       []*wire.MsgTx{transaction1},
			expectedSet: &DiffUTXOSet{
				base: &FullUTXOSet{utxoCollection: utxoCollection{outpoint1: utxoEntry0}},
				UTXODiff: &UTXODiff{
					toAdd:    utxoCollection{outpoint2: utxoEntry1},
					toRemove: utxoCollection{outpoint1: utxoEntry0},
				},
			},
		},
		{
			name: "add transaction to set with its input in diff toAdd",
			startSet: &DiffUTXOSet{
				base: NewFullUTXOSet(),
				UTXODiff: &UTXODiff{
					toAdd:    utxoCollection{outpoint1: utxoEntry0},
					toRemove: utxoCollection{},
				},
			},
			startHeight: 1,
			toAdd:       []*wire.MsgTx{transaction1},
			expectedSet: &DiffUTXOSet{
				base: NewFullUTXOSet(),
				UTXODiff: &UTXODiff{
					toAdd:    utxoCollection{outpoint2: utxoEntry1},
					toRemove: utxoCollection{},
				},
			},
		},
		{
			name: "add transaction to set with its input in diff toAdd and its output in diff toRemove",
			startSet: &DiffUTXOSet{
				base: NewFullUTXOSet(),
				UTXODiff: &UTXODiff{
					toAdd:    utxoCollection{outpoint1: utxoEntry0},
					toRemove: utxoCollection{outpoint2: utxoEntry1},
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
				base: &FullUTXOSet{utxoCollection: utxoCollection{outpoint1: utxoEntry0}},
				UTXODiff: &UTXODiff{
					toAdd:    utxoCollection{},
					toRemove: utxoCollection{},
				},
			},
			startHeight: 1,
			toAdd:       []*wire.MsgTx{transaction1, transaction2},
			expectedSet: &DiffUTXOSet{
				base: &FullUTXOSet{utxoCollection: utxoCollection{outpoint1: utxoEntry0}},
				UTXODiff: &UTXODiff{
					toAdd:    utxoCollection{outpoint3: utxoEntry2},
					toRemove: utxoCollection{outpoint1: utxoEntry0},
				},
			},
		},
	}

testLoop:
	for _, test := range tests {
		startSet := addMultisetToDiffUTXOSet(t, test.startSet)
		expectedSet := addMultisetToDiffUTXOSet(t, test.expectedSet)

		diffSet := startSet.clone()

		// Apply all transactions to diffSet, in order, with the initial block height startHeight
		for i, transaction := range test.toAdd {
			_, err := diffSet.AddTx(transaction, test.startHeight+uint64(i))
			if err != nil {
				t.Errorf("Error adding tx %s in test \"%s\": %s", transaction.TxID(), test.name, err)
				continue testLoop
			}
		}

		// Make sure that the result diffSet equals to the expectedSet
		if !diffSet.(*DiffUTXOSet).equal(expectedSet) {
			t.Errorf("unexpected diffSet in test \"%s\". "+
				"Expected: \"%v\", got: \"%v\".", test.name, expectedSet, diffSet)
		}
	}
}

func TestDiffFromTx(t *testing.T) {
	fus := addMultisetToFullUTXOSet(t, &FullUTXOSet{
		utxoCollection: utxoCollection{},
	})
	cbTx, err := createCoinbaseTxForTest(1, 1, 0, &dagconfig.SimNetParams)
	if err != nil {
		t.Errorf("createCoinbaseTxForTest: %v", err)
	}
	if isAccepted, err := fus.AddTx(cbTx, 1); err != nil {
		t.Fatalf("AddTx unexpectedly failed. Error: %s", err)
	} else if !isAccepted {
		t.Fatalf("AddTx unexpectedly didn't add tx %s", cbTx.TxID())
	}
	node := &blockNode{height: 2} //Fake node
	cbOutpoint := wire.Outpoint{TxID: *cbTx.TxID(), Index: 0}
	txIns := []*wire.TxIn{&wire.TxIn{
		PreviousOutpoint: cbOutpoint,
		SignatureScript:  nil,
		Sequence:         wire.MaxTxInSequenceNum,
	}}
	txOuts := []*wire.TxOut{&wire.TxOut{
		PkScript: OpTrueScript,
		Value:    uint64(1),
	}}
	tx := wire.NewNativeMsgTx(wire.TxVersion, txIns, txOuts)
	diff, err := fus.diffFromTx(tx, node)
	if err != nil {
		t.Errorf("diffFromTx: %v", err)
	}
	if !reflect.DeepEqual(diff.toAdd, utxoCollection{
		wire.Outpoint{TxID: *tx.TxID(), Index: 0}: NewUTXOEntry(tx.TxOut[0], false, 2),
	}) {
		t.Errorf("diff.toAdd doesn't have the expected values")
	}

	if !reflect.DeepEqual(diff.toRemove, utxoCollection{
		wire.Outpoint{TxID: *cbTx.TxID(), Index: 0}: NewUTXOEntry(cbTx.TxOut[0], true, 1),
	}) {
		t.Errorf("diff.toRemove doesn't have the expected values")
	}

	//Test that we get an error if we don't have the outpoint inside the utxo set
	invalidTxIns := []*wire.TxIn{&wire.TxIn{
		PreviousOutpoint: wire.Outpoint{TxID: daghash.TxID{}, Index: 0},
		SignatureScript:  nil,
		Sequence:         wire.MaxTxInSequenceNum,
	}}
	invalidTxOuts := []*wire.TxOut{&wire.TxOut{
		PkScript: OpTrueScript,
		Value:    uint64(1),
	}}
	invalidTx := wire.NewNativeMsgTx(wire.TxVersion, invalidTxIns, invalidTxOuts)
	_, err = fus.diffFromTx(invalidTx, node)
	if err == nil {
		t.Errorf("diffFromTx: expected an error but got <nil>")
	}

	//Test that we get an error if the outpoint is inside diffUTXOSet's toRemove
	diff2 := addMultisetToDiff(t, &UTXODiff{
		toAdd:    utxoCollection{},
		toRemove: utxoCollection{},
	})
	dus := NewDiffUTXOSet(fus, diff2)
	if isAccepted, err := dus.AddTx(tx, 2); err != nil {
		t.Fatalf("AddTx unexpectedly failed. Error: %s", err)
	} else if !isAccepted {
		t.Fatalf("AddTx unexpectedly didn't add tx %s", tx.TxID())
	}
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
func (dus *DiffUTXOSet) collection() (utxoCollection, error) {
	clone := dus.clone().(*DiffUTXOSet)
	err := clone.meldToBase()
	if err != nil {
		return nil, err
	}

	return clone.base.collection(), nil
}

func TestUTXOSetAddEntry(t *testing.T) {
	txID0, _ := daghash.NewTxIDFromStr("0000000000000000000000000000000000000000000000000000000000000000")
	txID1, _ := daghash.NewTxIDFromStr("1111111111111111111111111111111111111111111111111111111111111111")
	outpoint0 := wire.NewOutpoint(txID0, 0)
	outpoint1 := wire.NewOutpoint(txID1, 0)
	utxoEntry0 := NewUTXOEntry(&wire.TxOut{PkScript: []byte{}, Value: 10}, true, 0)
	utxoEntry1 := NewUTXOEntry(&wire.TxOut{PkScript: []byte{}, Value: 20}, false, 1)

	utxoDiff := NewUTXODiff()

	tests := []struct {
		name             string
		outpointToAdd    *wire.Outpoint
		utxoEntryToAdd   *UTXOEntry
		expectedUTXODiff *UTXODiff
		expectedError    string
	}{
		{
			name:           "add an entry",
			outpointToAdd:  outpoint0,
			utxoEntryToAdd: utxoEntry0,
			expectedUTXODiff: &UTXODiff{
				toAdd:    utxoCollection{*outpoint0: utxoEntry0},
				toRemove: utxoCollection{},
			},
		},
		{
			name:           "add another entry",
			outpointToAdd:  outpoint1,
			utxoEntryToAdd: utxoEntry1,
			expectedUTXODiff: &UTXODiff{
				toAdd:    utxoCollection{*outpoint0: utxoEntry0, *outpoint1: utxoEntry1},
				toRemove: utxoCollection{},
			},
		},
		{
			name:           "add first entry again",
			outpointToAdd:  outpoint0,
			utxoEntryToAdd: utxoEntry0,
			expectedUTXODiff: &UTXODiff{
				toAdd:    utxoCollection{*outpoint0: utxoEntry0, *outpoint1: utxoEntry1},
				toRemove: utxoCollection{},
			},
			expectedError: "AddEntry: Cannot add outpoint 0000000000000000000000000000000000000000000000000000000000000000:0 twice",
		},
	}

	for _, test := range tests {
		expectedUTXODiff := addMultisetToDiff(t, test.expectedUTXODiff)
		err := utxoDiff.AddEntry(*test.outpointToAdd, test.utxoEntryToAdd)
		errString := ""
		if err != nil {
			errString = err.Error()
		}
		if errString != test.expectedError {
			t.Fatalf("utxoDiff.AddEntry: unexpected err in test \"%s\". Expected: %s but got: %s", test.name, test.expectedError, err)
		}
		if err == nil && !utxoDiff.equal(expectedUTXODiff) {
			t.Fatalf("utxoDiff.AddEntry: unexpected utxoDiff in test \"%s\". "+
				"Expected: %v, got: %v", test.name, expectedUTXODiff, utxoDiff)
		}
	}
}
