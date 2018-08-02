package blockdag

import (
	"testing"
	"github.com/daglabs/btcd/dagconfig/daghash"
	"github.com/daglabs/btcd/wire"
	"reflect"
)

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

	emptySet := newFullUTXOSet()
	if len(emptySet.collection()) != 0 {
		t.Errorf("new set is not empty")
	}

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

	if !reflect.DeepEqual(emptySet.collection(), emptySet.utxoCollection) {
		t.Errorf("collection does not equal the set's utxoCollection")
	}

	if !reflect.DeepEqual(emptySet.clone(), emptySet) {
		t.Errorf("clone does not equal the original set")
	}
}

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

	emptySet := newDiffUTXOSet(newFullUTXOSet(), newUTXODiff())
	if len(emptySet.collection()) != 0 {
		t.Errorf("new set is not empty")
	}

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

	tests := []struct {
		name               string
		startSet           *diffUTXOSet
		expectedMeldSet    *diffUTXOSet
		expectedString     string
		expectedCollection utxoCollection
	}{
		{
			name: "empty base, empty diff",
			startSet: &diffUTXOSet{
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
			startSet: &diffUTXOSet{
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
			startSet: &diffUTXOSet{
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
			startSet: &diffUTXOSet{
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
			startSet: &diffUTXOSet{
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
		meldSet := test.startSet.clone().(*diffUTXOSet)
		meldSet.meldToBase()
		if !reflect.DeepEqual(meldSet, test.expectedMeldSet) {
			t.Errorf("unexpected melded set in test \"%s\". "+
				"Expected: \"%v\", got: \"%v\".", test.name, test.expectedMeldSet, meldSet)
		}

		setString := test.startSet.String()
		if setString != test.expectedString {
			t.Errorf("unexpected string in test \"%s\". "+
				"Expected: \"%s\", got: \"%s\".", test.name, test.expectedString, setString)
		}

		setCollection := test.startSet.collection()
		if !reflect.DeepEqual(setCollection, test.expectedCollection) {
			t.Errorf("unexpected set collection in test \"%s\". "+
				"Expected: \"%v\", got: \"%v\".", test.name, test.expectedCollection, setCollection)
		}

		setClone := test.startSet.clone()
		if !reflect.DeepEqual(setClone, test.startSet) {
			t.Errorf("unexpected set clone in test \"%s\". "+
				"Expected: \"%v\", got: \"%v\".", test.name, test.startSet, setClone)
		}
	}
}

func TestUTXOSetDiffRules(t *testing.T) {
	fullSet := newFullUTXOSet()
	diffSet := newDiffUTXOSet(fullSet, newUTXODiff())
	run := func(set utxoSet) {
		tests := []struct {
			name                string
			diffSet             utxoSet
			expectedDiffSuccess bool
		}{
			{
				name:                "diff against fullSet",
				diffSet:             newFullUTXOSet(),
				expectedDiffSuccess: false,
			},
			{
				name:                "diff against diffSet with different base",
				diffSet:             newDiffUTXOSet(newFullUTXOSet(), newUTXODiff()),
				expectedDiffSuccess: false,
			},
			{
				name:                "diff against diffSet with same base",
				diffSet:             newDiffUTXOSet(fullSet, newUTXODiff()),
				expectedDiffSuccess: true,
			},
		}

		for _, test := range tests {
			_, err := set.diffFrom(test.diffSet)
			diffSuccess := err == nil
			if diffSuccess != test.expectedDiffSuccess {
				t.Errorf("unexpected diff success in test \"%s\". "+
					"Expected: \"%t\", got: \"%t\".", test.name, test.expectedDiffSuccess, diffSuccess)
			}
		}
	}

	run(fullSet)
	run(diffSet)
}

func TestDiffUTXOSet_addTx(t *testing.T) {
	hash0, _ := daghash.NewHashFromStr("0000000000000000000000000000000000000000000000000000000000000000")
	outPoint0 := *wire.NewOutPoint(hash0, 0)
	txIn0 := &wire.TxIn{SignatureScript: []byte{}, PreviousOutPoint: wire.OutPoint{Hash: *hash0, Index: 0}, Sequence: 0}
	txOut0 := &wire.TxOut{PkScript: []byte{}, Value: 10}
	utxoEntry0 := newUTXOEntry(txOut0)
	transaction0 := wire.NewMsgTx(1)
	transaction0.TxIn = []*wire.TxIn{txIn0}
	transaction0.TxOut = []*wire.TxOut{txOut0}

	hash1 := transaction0.TxHash()
	outPoint1 := *wire.NewOutPoint(&hash1, 0)
	txIn1 := &wire.TxIn{SignatureScript: []byte{}, PreviousOutPoint: wire.OutPoint{Hash: hash1, Index: 0}, Sequence: 0}
	txOut1 := &wire.TxOut{PkScript: []byte{}, Value: 20}
	utxoEntry1 := newUTXOEntry(txOut1)
	transaction1 := wire.NewMsgTx(1)
	transaction1.TxIn = []*wire.TxIn{txIn1}
	transaction1.TxOut = []*wire.TxOut{txOut1}

	hash2 := transaction1.TxHash()
	outPoint2 := *wire.NewOutPoint(&hash2, 0)

	tests := []struct {
		name        string
		startSet    *diffUTXOSet
		toAdd       []*wire.MsgTx
		expectedSet *diffUTXOSet
	}{
		{
			name:        "add transaction to empty set",
			startSet:    newDiffUTXOSet(newFullUTXOSet(), newUTXODiff()),
			toAdd:       []*wire.MsgTx{transaction0},
			expectedSet: nil,
		},
		{
			name: "add transaction to set with the tx in base",
			startSet: &diffUTXOSet{
				base: &fullUTXOSet{utxoCollection: utxoCollection{outPoint0: utxoEntry0}},
				utxoDiff: &utxoDiff{
					toAdd:    utxoCollection{},
					toRemove: utxoCollection{},
				},
			},
			toAdd: []*wire.MsgTx{transaction0},
			expectedSet: &diffUTXOSet{
				base: &fullUTXOSet{utxoCollection: utxoCollection{outPoint0: utxoEntry0}},
				utxoDiff: &utxoDiff{
					toAdd:    utxoCollection{outPoint1: utxoEntry0},
					toRemove: utxoCollection{outPoint0: utxoEntry0},
				},
			},
		},
		{
			name: "add transaction to set with the tx in diff toAdd",
			startSet: &diffUTXOSet{
				base: newFullUTXOSet(),
				utxoDiff: &utxoDiff{
					toAdd:    utxoCollection{outPoint0: utxoEntry0},
					toRemove: utxoCollection{},
				},
			},
			toAdd: []*wire.MsgTx{transaction0},
			expectedSet: &diffUTXOSet{
				base: newFullUTXOSet(),
				utxoDiff: &utxoDiff{
					toAdd:    utxoCollection{outPoint1: utxoEntry0},
					toRemove: utxoCollection{},
				},
			},
		},
		{
			name: "add transaction to set with the tx in diff toAdd and its hash in diff toRemove",
			startSet: &diffUTXOSet{
				base: newFullUTXOSet(),
				utxoDiff: &utxoDiff{
					toAdd:    utxoCollection{outPoint0: utxoEntry0},
					toRemove: utxoCollection{outPoint1: utxoEntry0},
				},
			},
			toAdd: []*wire.MsgTx{transaction0},
			expectedSet: &diffUTXOSet{
				base: newFullUTXOSet(),
				utxoDiff: &utxoDiff{
					toAdd:    utxoCollection{},
					toRemove: utxoCollection{},
				},
			},
		},
		{
			name: "add two transactions, one spending the other, to set with the first tx in base",
			startSet: &diffUTXOSet{
				base: &fullUTXOSet{utxoCollection: utxoCollection{outPoint0: utxoEntry0}},
				utxoDiff: &utxoDiff{
					toAdd:    utxoCollection{},
					toRemove: utxoCollection{},
				},
			},
			toAdd: []*wire.MsgTx{transaction0, transaction1},
			expectedSet: &diffUTXOSet{
				base: &fullUTXOSet{utxoCollection: utxoCollection{outPoint0: utxoEntry0}},
				utxoDiff: &utxoDiff{
					toAdd:    utxoCollection{outPoint2: utxoEntry1},
					toRemove: utxoCollection{outPoint0: utxoEntry0},
				},
			},
		},
	}

	for _, test := range tests {
		diffSet := test.startSet.clone()
		isSuccessful := true
		for _, transaction := range test.toAdd {
			isSuccessful = isSuccessful && diffSet.addTx(transaction)
		}

		shouldBeSuccessful := test.expectedSet != nil
		if shouldBeSuccessful != isSuccessful {
			t.Errorf("unexpected addTx success in test \"%s\". "+
				"Expected: \"%t\", got: \"%t\".", test.name, shouldBeSuccessful, isSuccessful)
		}
		if shouldBeSuccessful && !reflect.DeepEqual(diffSet, test.expectedSet) {
			t.Errorf("unexpected diffSet in test \"%s\". "+
				"Expected: \"%v\", got: \"%v\".", test.name, test.expectedSet, diffSet)
		}
	}
}

func TestIterate(t *testing.T) {
	hash0, _ := daghash.NewHashFromStr("0000000000000000000000000000000000000000000000000000000000000000")
	hash1, _ := daghash.NewHashFromStr("1111111111111111111111111111111111111111111111111111111111111111")
	outPoint0 := *wire.NewOutPoint(hash0, 0)
	outPoint1 := *wire.NewOutPoint(hash1, 0)
	utxoEntry0 := newUTXOEntry(&wire.TxOut{PkScript: []byte{}, Value: 10})
	utxoEntry1 := newUTXOEntry(&wire.TxOut{PkScript: []byte{}, Value: 20})

	tests := []struct {
		name            string
		set             utxoSet
		expectedOutputs []utxoIteratorOutput
	}{
		{
			name:            "empty fullSet should not iterate",
			set:             &fullUTXOSet{},
			expectedOutputs: []utxoIteratorOutput{},
		},
		{
			name: "empty diffSet should not iterate",
			set: &diffUTXOSet{
				base: &fullUTXOSet{},
				utxoDiff: &utxoDiff{
					toAdd:    utxoCollection{},
					toRemove: utxoCollection{},
				},
			},
			expectedOutputs: []utxoIteratorOutput{},
		},
		{
			name:            "fullSet with one nil member should iterate once",
			set:             &fullUTXOSet{utxoCollection{outPoint0: nil}},
			expectedOutputs: []utxoIteratorOutput{{outPoint0, nil}},
		},
		{
			name: "diffSet with one nil member should iterate once",
			set: &diffUTXOSet{
				base: &fullUTXOSet{utxoCollection: utxoCollection{outPoint0: nil}},
				utxoDiff: &utxoDiff{
					toAdd:    utxoCollection{},
					toRemove: utxoCollection{},
				},
			},
			expectedOutputs: []utxoIteratorOutput{{outPoint0, nil}},
		},
		{
			name:            "fullSet with one entry should iterate once",
			set:             &fullUTXOSet{utxoCollection{outPoint0: utxoEntry0}},
			expectedOutputs: []utxoIteratorOutput{{outPoint0, utxoEntry0}},
		},
		{
			name: "diffSet with one entry should iterate once",
			set: &diffUTXOSet{
				base: &fullUTXOSet{utxoCollection: utxoCollection{outPoint1: utxoEntry1}},
				utxoDiff: &utxoDiff{
					toAdd:    utxoCollection{outPoint0: utxoEntry0},
					toRemove: utxoCollection{outPoint1: utxoEntry1},
				},
			},
			expectedOutputs: []utxoIteratorOutput{{outPoint0, utxoEntry0}},
		},
		{
			name:            "fullSet with two entries should iterate twice",
			set:             &fullUTXOSet{utxoCollection{outPoint0: utxoEntry0, outPoint1: utxoEntry1}},
			expectedOutputs: []utxoIteratorOutput{{outPoint0, utxoEntry0}, {outPoint1, utxoEntry1}},
		},
		{
			name: "diffSet with two txOut members with different hashes should iterate twice",
			set: &diffUTXOSet{
				base: &fullUTXOSet{utxoCollection: utxoCollection{outPoint0: utxoEntry0}},
				utxoDiff: &utxoDiff{
					toAdd:    utxoCollection{outPoint1: utxoEntry1},
					toRemove: utxoCollection{},
				},
			},
			expectedOutputs: []utxoIteratorOutput{{outPoint0, utxoEntry0}, {outPoint1, utxoEntry1}},
		},
	}

	for _, test := range tests {
		iteratedTimes := 0
		expectedOutputSet := make(map[utxoIteratorOutput]bool)
		for _, output := range test.expectedOutputs {
			expectedOutputSet[output] = false
		}

		for output := range test.set.iterate() {
			expectedOutputSet[output] = true
			iteratedTimes++
		}

		for output, wasVisited := range expectedOutputSet {
			if !wasVisited {
				t.Errorf("missing output [%v] in test \"%s\".", output, test.name)
			}
		}
		expectedLength := len(test.expectedOutputs)
		if iteratedTimes != expectedLength {
			t.Errorf("unexpected length in test \"%s\". "+
				"Expected: %d, got: %d.", test.name, expectedLength, iteratedTimes)
		}
	}
}
