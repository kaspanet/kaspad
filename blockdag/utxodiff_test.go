package blockdag

import (
	"testing"
	"reflect"
	"github.com/daglabs/btcd/dagconfig/daghash"
	"github.com/daglabs/btcd/wire"
)

func TestUTXODiff(t *testing.T) {
	newDiff := newUTXODiff()
	if newDiff.toAdd.len() != 0 || newDiff.toRemove.len() != 0 {
		t.Errorf("new diff is not empty")
	}

	hash0, _ := daghash.NewHashFromStr("0000000000000000000000000000000000000000000000000000000000000000")
	hash1, _ := daghash.NewHashFromStr("1111111111111111111111111111111111111111111111111111111111111111")
	txOut0 := &wire.TxOut{PkScript: []byte{}, Value: 10}
	txOut1 := &wire.TxOut{PkScript: []byte{}, Value: 20}
	diff := utxoDiff{
		toAdd:    utxoCollection{*hash0: map[uint32]*wire.TxOut{0: txOut0}},
		toRemove: utxoCollection{*hash1: map[uint32]*wire.TxOut{0: txOut1}},
	}

	clonedDiff := *diff.clone()
	if !reflect.DeepEqual(clonedDiff, diff) {
		t.Errorf("cloned diff not equal to the original"+
			"Original: \"%v\", cloned: \"%v\".", diff, clonedDiff)
	}

	expectedDiffString := "toAdd: [ (0000000000000000000000000000000000000000000000000000000000000000, 0) => 10 ]; toRemove: [ (1111111111111111111111111111111111111111111111111111111111111111, 0) => 20 ]"
	diffString := clonedDiff.String()
	if diffString != expectedDiffString {
		t.Errorf("unexpected diff string. "+
			"Expected: \"%s\", got: \"%s\".", expectedDiffString, diffString)
	}
}

func TestUTXODiffRules(t *testing.T) {
	hash0, _ := daghash.NewHashFromStr("0000000000000000000000000000000000000000000000000000000000000000")
	txOut0 := &wire.TxOut{PkScript: []byte{}, Value: 10}

	tests := []struct {
		name                   string
		this                   *utxoDiff
		other                  *utxoDiff
		expectedDiffResult     *utxoDiff
		expectedWithDiffResult *utxoDiff
	}{
		{
			name: "one toAdd in this, one toAdd in other",
			this: &utxoDiff{
				toAdd:    utxoCollection{*hash0: map[uint32]*wire.TxOut{0: txOut0}},
				toRemove: utxoCollection{},
			},
			other: &utxoDiff{
				toAdd:    utxoCollection{*hash0: map[uint32]*wire.TxOut{0: txOut0}},
				toRemove: utxoCollection{},
			},
			expectedDiffResult: &utxoDiff{
				toAdd:    utxoCollection{},
				toRemove: utxoCollection{},
			},
			expectedWithDiffResult: nil,
		},
		{
			name: "one toAdd in this, one toRemove in other",
			this: &utxoDiff{
				toAdd:    utxoCollection{*hash0: map[uint32]*wire.TxOut{0: txOut0}},
				toRemove: utxoCollection{},
			},
			other: &utxoDiff{
				toAdd:    utxoCollection{},
				toRemove: utxoCollection{*hash0: map[uint32]*wire.TxOut{0: txOut0}},
			},
			expectedDiffResult: nil,
			expectedWithDiffResult: &utxoDiff{
				toAdd:    utxoCollection{},
				toRemove: utxoCollection{},
			},
		},
		{
			name: "one toAdd in this, empty other",
			this: &utxoDiff{
				toAdd:    utxoCollection{*hash0: map[uint32]*wire.TxOut{0: txOut0}},
				toRemove: utxoCollection{},
			},
			other: &utxoDiff{
				toAdd:    utxoCollection{},
				toRemove: utxoCollection{},
			},
			expectedDiffResult: &utxoDiff{
				toAdd:    utxoCollection{},
				toRemove: utxoCollection{*hash0: map[uint32]*wire.TxOut{0: txOut0}},
			},
			expectedWithDiffResult: &utxoDiff{
				toAdd:    utxoCollection{*hash0: map[uint32]*wire.TxOut{0: txOut0}},
				toRemove: utxoCollection{},
			},
		},
		{
			name: "one toRemove in this, one toAdd in other",
			this: &utxoDiff{
				toAdd:    utxoCollection{},
				toRemove: utxoCollection{*hash0: map[uint32]*wire.TxOut{0: txOut0}},
			},
			other: &utxoDiff{
				toAdd:    utxoCollection{*hash0: map[uint32]*wire.TxOut{0: txOut0}},
				toRemove: utxoCollection{},
			},
			expectedDiffResult: nil,
			expectedWithDiffResult: &utxoDiff{
				toAdd:    utxoCollection{},
				toRemove: utxoCollection{},
			},
		},
		{
			name: "one toRemove in this, one toRemove in other",
			this: &utxoDiff{
				toAdd:    utxoCollection{},
				toRemove: utxoCollection{*hash0: map[uint32]*wire.TxOut{0: txOut0}},
			},
			other: &utxoDiff{
				toAdd:    utxoCollection{},
				toRemove: utxoCollection{*hash0: map[uint32]*wire.TxOut{0: txOut0}},
			},
			expectedDiffResult: &utxoDiff{
				toAdd:    utxoCollection{},
				toRemove: utxoCollection{},
			},
			expectedWithDiffResult: nil,
		},
		{
			name: "one toRemove in this, empty other",
			this: &utxoDiff{
				toAdd:    utxoCollection{},
				toRemove: utxoCollection{*hash0: map[uint32]*wire.TxOut{0: txOut0}},
			},
			other: &utxoDiff{
				toAdd:    utxoCollection{},
				toRemove: utxoCollection{},
			},
			expectedDiffResult: &utxoDiff{
				toAdd:    utxoCollection{*hash0: map[uint32]*wire.TxOut{0: txOut0}},
				toRemove: utxoCollection{},
			},
			expectedWithDiffResult: &utxoDiff{
				toAdd:    utxoCollection{},
				toRemove: utxoCollection{*hash0: map[uint32]*wire.TxOut{0: txOut0}},
			},
		},
		{
			name: "empty this, one toAdd in other",
			this: &utxoDiff{
				toAdd:    utxoCollection{},
				toRemove: utxoCollection{},
			},
			other: &utxoDiff{
				toAdd:    utxoCollection{*hash0: map[uint32]*wire.TxOut{0: txOut0}},
				toRemove: utxoCollection{},
			},
			expectedDiffResult: &utxoDiff{
				toAdd:    utxoCollection{*hash0: map[uint32]*wire.TxOut{0: txOut0}},
				toRemove: utxoCollection{},
			},
			expectedWithDiffResult: &utxoDiff{
				toAdd:    utxoCollection{*hash0: map[uint32]*wire.TxOut{0: txOut0}},
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
				toRemove: utxoCollection{*hash0: map[uint32]*wire.TxOut{0: txOut0}},
			},
			expectedDiffResult: &utxoDiff{
				toAdd:    utxoCollection{},
				toRemove: utxoCollection{*hash0: map[uint32]*wire.TxOut{0: txOut0}},
			},
			expectedWithDiffResult: &utxoDiff{
				toAdd:    utxoCollection{},
				toRemove: utxoCollection{*hash0: map[uint32]*wire.TxOut{0: txOut0}},
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
			expectedDiffResult: &utxoDiff{
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
		diffResult, err := test.this.diff(test.other)
		isDiffOk := err == nil
		if isDiffOk && !reflect.DeepEqual(diffResult, test.expectedDiffResult) {
			t.Errorf("unexpected diff result in test \"%s\". "+
				"Expected: \"%v\", got: \"%v\".", test.name, test.expectedDiffResult, diffResult)
		}
		expectedIsDiffOk := test.expectedDiffResult != nil
		if isDiffOk != expectedIsDiffOk {
			t.Errorf("unexpected diff error in test \"%s\". "+
				"Expected: \"%t\", got: \"%t\".", test.name, expectedIsDiffOk, isDiffOk)
		}

		withDiffResult, err := test.this.withDiff(test.other)
		isWithDiffOk := err == nil
		if isWithDiffOk && !reflect.DeepEqual(withDiffResult, test.expectedWithDiffResult) {
			t.Errorf("unexpected withDiff result in test \"%s\". "+
				"Expected: \"%v\", got: \"%v\".", test.name, test.expectedWithDiffResult, withDiffResult)
		}
		expectedIsWithDiffOk := test.expectedWithDiffResult != nil
		if isWithDiffOk != expectedIsWithDiffOk {
			t.Errorf("unexpected withDiff error in test \"%s\". "+
				"Expected: \"%t\", got: \"%t\".", test.name, expectedIsWithDiffOk, isWithDiffOk)
		}
	}
}
