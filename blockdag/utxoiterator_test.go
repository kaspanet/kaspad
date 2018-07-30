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
		diffSet        diffUTXOSet
		expectedTxOut  []*wire.TxOut
		expectedHashes []daghash.Hash
		expectedLength int
	}{
		{
			name:       "empty collection should not iterate",
			collection: utxoCollection{},
			diffSet: diffUTXOSet{
				base: &fullUTXOSet{},
				utxoDiff: &utxoDiff{
					toAdd:    utxoCollection{},
					toRemove: utxoCollection{},
				},
			},
			expectedTxOut:  []*wire.TxOut{},
			expectedHashes: []daghash.Hash{},
			expectedLength: 0,
		},
		{
			name:       "collection with one nil member should not iterate",
			collection: utxoCollection{*hash0: nil},
			diffSet: diffUTXOSet{
				base: &fullUTXOSet{utxoCollection: utxoCollection{*hash0: nil}},
				utxoDiff: &utxoDiff{
					toAdd:    utxoCollection{},
					toRemove: utxoCollection{},
				},
			},
			expectedTxOut:  []*wire.TxOut{},
			expectedHashes: []daghash.Hash{},
			expectedLength: 0,
		},
		{
			name:       "collection with one empty member should not iterate",
			collection: utxoCollection{*hash0: map[uint32]*wire.TxOut{}},
			diffSet: diffUTXOSet{
				base: &fullUTXOSet{utxoCollection: utxoCollection{*hash0: map[uint32]*wire.TxOut{}}},
				utxoDiff: &utxoDiff{
					toAdd:    utxoCollection{},
					toRemove: utxoCollection{},
				},
			},
			expectedTxOut:  []*wire.TxOut{},
			expectedHashes: []daghash.Hash{},
			expectedLength: 0,
		},
		{
			name:       "collection with one txOut member should iterate once",
			collection: utxoCollection{*hash0: map[uint32]*wire.TxOut{0: txOut0}},
			diffSet: diffUTXOSet{
				base: &fullUTXOSet{utxoCollection: utxoCollection{*hash1: map[uint32]*wire.TxOut{0: txOut1}}},
				utxoDiff: &utxoDiff{
					toAdd:    utxoCollection{*hash0: map[uint32]*wire.TxOut{0: txOut0}},
					toRemove: utxoCollection{*hash1: map[uint32]*wire.TxOut{0: txOut1}},
				},
			},
			expectedTxOut:  []*wire.TxOut{txOut0},
			expectedHashes: []daghash.Hash{*hash0},
			expectedLength: 1,
		},
		{
			name:       "collection with two txOut members with the same previousHash should iterate twice",
			collection: utxoCollection{*hash0: map[uint32]*wire.TxOut{0: txOut0, 1: txOut1}},
			diffSet: diffUTXOSet{
				base: &fullUTXOSet{utxoCollection: utxoCollection{*hash0: map[uint32]*wire.TxOut{0: txOut0}}},
				utxoDiff: &utxoDiff{
					toAdd:    utxoCollection{*hash0: map[uint32]*wire.TxOut{0: txOut1}},
					toRemove: utxoCollection{},
				},
			},
			expectedTxOut:  []*wire.TxOut{txOut0, txOut1},
			expectedHashes: []daghash.Hash{*hash0},
			expectedLength: 2,
		},
		{
			name:       "collection with two txOut members with different previousHashes should iterate twice",
			collection: utxoCollection{*hash0: map[uint32]*wire.TxOut{0: txOut0}, *hash1: map[uint32]*wire.TxOut{0: txOut1}},
			diffSet: diffUTXOSet{
				base: &fullUTXOSet{utxoCollection: utxoCollection{*hash0: map[uint32]*wire.TxOut{0: txOut0}}},
				utxoDiff: &utxoDiff{
					toAdd:    utxoCollection{*hash1: map[uint32]*wire.TxOut{0: txOut1}},
					toRemove: utxoCollection{},
				},
			},
			expectedTxOut:  []*wire.TxOut{txOut0, txOut1},
			expectedHashes: []daghash.Hash{*hash0, *hash1},
			expectedLength: 2,
		},
	}

	for _, test := range tests {
		run := func(iterator utxoIterator) {
			iteratedTimes := 0
			expectedTxOutSet := make(map[*wire.TxOut]bool)
			for _, txOut := range test.expectedTxOut {
				expectedTxOutSet[txOut] = false
			}
			expectedHashSet := make(map[daghash.Hash]bool)
			for _, hash := range test.expectedHashes {
				expectedHashSet[hash] = false
			}

			for utxo := range iterator {
				expectedTxOutSet[utxo.txOut] = true
				expectedHashSet[utxo.previousHash] = true
				iteratedTimes++
			}

			for txOut, wasVisited := range expectedTxOutSet {
				if !wasVisited {
					t.Errorf("missing txOut [%v] in test \"%s\".", txOut, test.name)
				}
			}
			for hash, wasVisited := range expectedHashSet {
				if !wasVisited {
					t.Errorf("missing previousHash [%v] in test \"%s\".", hash, test.name)
				}
			}
			if iteratedTimes != test.expectedLength {
				t.Errorf("unexpected length in test \"%s\". "+
					"Expected: %d, got: %d.", test.name, test.expectedLength, iteratedTimes)
			}
		}

		run(test.collection.iterate())
		run(test.diffSet.iterate())
	}
}
