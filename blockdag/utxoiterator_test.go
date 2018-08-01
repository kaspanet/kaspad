package blockdag

import (
	"testing"
	"github.com/daglabs/btcd/dagconfig/daghash"
	"github.com/daglabs/btcd/wire"
)

func TestIterate(t *testing.T) {
	hash0, _ := daghash.NewHashFromStr("0000000000000000000000000000000000000000000000000000000000000000")
	hash1, _ := daghash.NewHashFromStr("1111111111111111111111111111111111111111111111111111111111111111")
	outPoint0 := *wire.NewOutPoint(hash0, 0)
	outPoint1 := *wire.NewOutPoint(hash1, 0)
	utxoEntry0 := newUTXOEntry(&wire.TxOut{PkScript: []byte{}, Value: 10})
	utxoEntry1 := newUTXOEntry(&wire.TxOut{PkScript: []byte{}, Value: 20})

	tests := []struct {
		name            string
		iterable        utxoIterable
		expectedOutputs []utxoIteratorOutput
	}{
		{
			name:            "empty collection should not iterate",
			iterable:        utxoCollection{},
			expectedOutputs: []utxoIteratorOutput{},
		},
		{
			name: "empty diffSet should not iterate",
			iterable: &diffUTXOSet{
				base: &fullUTXOSet{},
				utxoDiff: &utxoDiff{
					toAdd:    utxoCollection{},
					toRemove: utxoCollection{},
				},
			},
			expectedOutputs: []utxoIteratorOutput{},
		},
		{
			name:            "collection with one nil member should iterate once",
			iterable:        utxoCollection{outPoint0: nil},
			expectedOutputs: []utxoIteratorOutput{{outPoint0, nil}},
		},
		{
			name: "diffSet with one nil member should iterate once",
			iterable: &diffUTXOSet{
				base: &fullUTXOSet{utxoCollection: utxoCollection{outPoint0: nil}},
				utxoDiff: &utxoDiff{
					toAdd:    utxoCollection{},
					toRemove: utxoCollection{},
				},
			},
			expectedOutputs: []utxoIteratorOutput{{outPoint0, nil}},
		},
		{
			name:            "collection with one entry should iterate once",
			iterable:        utxoCollection{outPoint0: utxoEntry0},
			expectedOutputs: []utxoIteratorOutput{{outPoint0, utxoEntry0}},
		},
		{
			name: "diffSet with one entry should iterate once",
			iterable: &diffUTXOSet{
				base: &fullUTXOSet{utxoCollection: utxoCollection{outPoint1: utxoEntry1}},
				utxoDiff: &utxoDiff{
					toAdd:    utxoCollection{outPoint0: utxoEntry0},
					toRemove: utxoCollection{outPoint1: utxoEntry1},
				},
			},
			expectedOutputs: []utxoIteratorOutput{{outPoint0, utxoEntry0}},
		},
		{
			name:            "collection with two entries should iterate twice",
			iterable:        utxoCollection{outPoint0: utxoEntry0, outPoint1: utxoEntry1},
			expectedOutputs: []utxoIteratorOutput{{outPoint0, utxoEntry0}, {outPoint1, utxoEntry1}},
		},
		{
			name: "diffSet with two txOut members with different hashes should iterate twice",
			iterable: &diffUTXOSet{
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

		for output := range test.iterable.iterate() {
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
