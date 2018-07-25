package blockdag

import (
	"testing"
	"github.com/daglabs/btcd/dagconfig/daghash"
	"github.com/daglabs/btcd/wire"
)

func TestIterate(t *testing.T) {
	hash0, _ := daghash.NewHashFromStr("0000000000000000000000000000000000000000000000000000000000000000")
	hash1, _ := daghash.NewHashFromStr("1111111111111111111111111111111111111111111111111111111111111111")
	txOut0 := &wire.TxOut{PkScript: []byte{}, Value: 10}
	txOut1 := &wire.TxOut{PkScript: []byte{}, Value: 20}
	tests := []struct {
		name           string
		collection     utxoCollection
		expectedTxOut  []*wire.TxOut
		expectedLength int
	}{
		{
			name:           "empty collection should not iterate",
			collection:     utxoCollection{},
			expectedTxOut:  []*wire.TxOut{},
			expectedLength: 0,
		},
		{
			name:           "collection with one nil member should not iterate",
			collection:     utxoCollection{*hash0: nil},
			expectedTxOut:  []*wire.TxOut{},
			expectedLength: 0,
		},
		{
			name:           "collection with one empty member should not iterate",
			collection:     utxoCollection{*hash0: map[int]*wire.TxOut{}},
			expectedTxOut:  []*wire.TxOut{},
			expectedLength: 0,
		},
		{
			name:           "collection with one txOut member should iterate once",
			collection:     utxoCollection{*hash0: map[int]*wire.TxOut{0: txOut0}},
			expectedTxOut:  []*wire.TxOut{txOut0},
			expectedLength: 1,
		},
		{
			name:           "collection with two txOut members with the same previousHash should iterate twice",
			collection:     utxoCollection{*hash0: map[int]*wire.TxOut{0: txOut0, 1: txOut1}},
			expectedTxOut:  []*wire.TxOut{txOut0, txOut1},
			expectedLength: 2,
		},
		{
			name:           "collection with two txOut members with different previousHashes should iterate twice",
			collection:     utxoCollection{*hash0: map[int]*wire.TxOut{0: txOut0}, *hash1: map[int]*wire.TxOut{0: txOut1}},
			expectedTxOut:  []*wire.TxOut{txOut0, txOut1},
			expectedLength: 2,
		},
	}

	for _, test := range tests {
		iteratedTimes := 0
		expectedTxOutSet := make(map[*wire.TxOut]int)
		for _, txOut := range test.expectedTxOut {
			expectedTxOutSet[txOut]++
		}

		for utxo := range test.collection.Iterate() {
			expectedTxOutSet[utxo.txOut]--
			iteratedTimes++
		}

		for txOut, visitValue := range expectedTxOutSet {
			if visitValue > 0 {
				t.Errorf("too few txOut in test \"%s\". "+
					"Deficit: %d, txOut: %v", test.name, visitValue, txOut)
			} else if visitValue < 0 {
				t.Errorf("too much txOut in test \"%s\". "+
					"Surplus: %d, txOut: %v", test.name, -visitValue, txOut)
			}
		}
		if iteratedTimes != test.expectedLength {
			t.Errorf("unexpected length in test \"%s\". "+
				"Expected: %d, got: %d.", test.name, test.expectedLength, iteratedTimes)
		}
	}
}
