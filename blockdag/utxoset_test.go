package blockdag

import (
	"testing"
	"github.com/daglabs/btcd/dagconfig/daghash"
	"github.com/daglabs/btcd/wire"
	"reflect"
)

func TestFullUTXOSet(t *testing.T) {
	fullSet := newFullUTXOSet()
	if fullSet.len() != 0 {
		t.Errorf("new set is not empty")
	}

	tests := []struct {
		name                string
		set                 utxoSet
		expectedDiffSuccess bool
	}{
		{
			name:                "fullDiff",
			set:                 newFullUTXOSet(),
			expectedDiffSuccess: false,
		},
		{
			name:                "fullDiff with different base",
			set:                 newDiffUTXOSet(newFullUTXOSet(), newUTXODiff()),
			expectedDiffSuccess: false,
		},
		{
			name:                "fullDiff with same base",
			set:                 newDiffUTXOSet(fullSet, newUTXODiff()),
			expectedDiffSuccess: true,
		},
	}

	for _, test := range tests {
		_, err := fullSet.diff(test.set)
		diffSuccess := err == nil
		if diffSuccess != test.expectedDiffSuccess {
			t.Errorf("unexpected diff success in test \"%s\". "+
				"Expected: \"%t\", got: \"%t\".", test.name, test.expectedDiffSuccess, diffSuccess)
		}
	}

	hash0, _ := daghash.NewHashFromStr("0000000000000000000000000000000000000000000000000000000000000000")
	hash1, _ := daghash.NewHashFromStr("1111111111111111111111111111111111111111111111111111111111111111")
	txOut0 := &wire.TxOut{PkScript: []byte{}, Value: 10}
	txOut1 := &wire.TxOut{PkScript: []byte{}, Value: 20}
	diff := &utxoDiff{
		toAdd:    utxoCollection{*hash0: map[uint32]*wire.TxOut{0: txOut0}},
		toRemove: utxoCollection{*hash1: map[uint32]*wire.TxOut{0: txOut1}},
	}
	withDiffResult, err := fullSet.withDiff(diff)
	if err != nil {
		t.Errorf("withDiff unexpectedly failed")
	}
	withDiffUTXOSet, ok := withDiffResult.(*diffUTXOSet)
	if !ok {
		t.Errorf("withDiff is of unexpected type")
	}
	if !reflect.DeepEqual(withDiffUTXOSet.base, fullSet) || !reflect.DeepEqual(withDiffUTXOSet.utxoDiff, diff) {
		t.Errorf("withDiff is of unexpected composition")
	}

	txIn0 := &wire.TxIn{SignatureScript: []byte{}, PreviousOutPoint: wire.OutPoint{Hash: *hash0, Index: 0}, Sequence: 0}
	transaction0 := wire.NewMsgTx(1)
	transaction0.TxIn = []*wire.TxIn{txIn0}
	transaction0.TxOut = []*wire.TxOut{txOut0}
	if ok = fullSet.addTx(transaction0); ok {
		t.Errorf("addTx unexpectedly succeeded")
	}
	fullSet = &fullUTXOSet{utxoCollection: utxoCollection{*hash0: map[uint32]*wire.TxOut{0: txOut0}}}
	if ok = fullSet.addTx(transaction0); !ok {
		t.Errorf("addTx unexpectedly failed")
	}

	if !reflect.DeepEqual(fullSet.collection(), fullSet.utxoCollection) {
		t.Errorf("collection does not equal the set's utxoCollection")
	}

	if !reflect.DeepEqual(fullSet.clone(), fullSet) {
		t.Errorf("clone does not equal the original set")
	}
}
