package blockdag

import (
	"testing"
	"reflect"
	"github.com/daglabs/btcd/dagconfig/daghash"
	"github.com/daglabs/btcd/wire"
)

func TestUTXODiff(t *testing.T) {
	newDiff := newUTXODiff()
	if len(newDiff.toAdd) != 0 || len(newDiff.toRemove) != 0 {
		t.Errorf("new diff is not empty")
	}

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
	outPoint0 := *wire.NewOutPoint(hash0, 0)
	utxoEntry0 := newUTXOEntry(&wire.TxOut{PkScript: []byte{}, Value: 10})

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
				toAdd:    utxoCollection{outPoint0: utxoEntry0},
				toRemove: utxoCollection{},
			},
			other: &utxoDiff{
				toAdd:    utxoCollection{outPoint0: utxoEntry0},
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
				toAdd:    utxoCollection{outPoint0: utxoEntry0},
				toRemove: utxoCollection{},
			},
			other: &utxoDiff{
				toAdd:    utxoCollection{},
				toRemove: utxoCollection{outPoint0: utxoEntry0},
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
				toAdd:    utxoCollection{outPoint0: utxoEntry0},
				toRemove: utxoCollection{},
			},
			other: &utxoDiff{
				toAdd:    utxoCollection{},
				toRemove: utxoCollection{},
			},
			expectedDiffResult: &utxoDiff{
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
				toRemove: utxoCollection{outPoint0: utxoEntry0},
			},
			other: &utxoDiff{
				toAdd:    utxoCollection{},
				toRemove: utxoCollection{outPoint0: utxoEntry0},
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
				toRemove: utxoCollection{outPoint0: utxoEntry0},
			},
			other: &utxoDiff{
				toAdd:    utxoCollection{},
				toRemove: utxoCollection{},
			},
			expectedDiffResult: &utxoDiff{
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
			expectedDiffResult: &utxoDiff{
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
			expectedDiffResult: &utxoDiff{
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
		diffResult, err := test.this.diffFrom(test.other)
		isDiffOk := err == nil
		if isDiffOk && !reflect.DeepEqual(diffResult, test.expectedDiffResult) {
			t.Errorf("unexpected diffFrom result in test \"%s\". "+
				"Expected: \"%v\", got: \"%v\".", test.name, test.expectedDiffResult, diffResult)
		}
		expectedIsDiffOk := test.expectedDiffResult != nil
		if isDiffOk != expectedIsDiffOk {
			t.Errorf("unexpected diffFrom error in test \"%s\". "+
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
