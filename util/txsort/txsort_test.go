// Copyright (c) 2015-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package txsort_test

import (
	"bytes"
	"encoding/hex"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/daglabs/btcd/util/txsort"
	"github.com/daglabs/btcd/wire"
)

// TestSort ensures the transaction sorting works according to the BIP.
func TestSort(t *testing.T) {
	tests := []struct {
		name         string
		hexFile      string
		isSorted     bool
		unsortedHash string
		sortedHash   string
	}{
		{
			name:         "first test case from BIP 69 - sorts inputs only, based on hash",
			hexFile:      "bip69-1.hex",
			isSorted:     false,
			unsortedHash: "652c30fa452b4706db2f4630a99e2e4d5205bc4724b758e02e7f3e5b772f2454",
			sortedHash:   "cb0d27e995be2d7670c172dd2c2b155cdeb5d6c60e8b13b3e398667a451279e8",
		},
		{
			name:         "second test case from BIP 69 - already sorted",
			hexFile:      "bip69-2.hex",
			isSorted:     true,
			unsortedHash: "1b4c7d54847f0a07a018816894cd6ee172802e4166ad0a705ad7c84ae531a2b7",
			sortedHash:   "1b4c7d54847f0a07a018816894cd6ee172802e4166ad0a705ad7c84ae531a2b7",
		},
		{
			name:         "block 100001 tx[1] - sorts outputs only, based on amount",
			hexFile:      "bip69-3.hex",
			isSorted:     false,
			unsortedHash: "fe297f705cf3a08dd398a85187060b45688008c658afa364384931d8cb091e34",
			sortedHash:   "487caaf688db44e54ed2a0dc0d5b665ad13c399259abd1e8abce1f7e762a52d2",
		},
		{
			name:         "block 100001 tx[2] - sorts both inputs and outputs",
			hexFile:      "bip69-4.hex",
			isSorted:     false,
			unsortedHash: "1132a25b7a18ede0e03dd8472313357512c788faf31da4225579436432b2606f",
			sortedHash:   "f63e7e1a71c39970e9c92a1c3fdfd50269819431249d5c560213eb270e48371f",
		},
		{
			name:         "block 100998 tx[6] - sorts outputs only, based on output script",
			hexFile:      "bip69-5.hex",
			isSorted:     false,
			unsortedHash: "712433780097a8966d5824fb3ea2a87fc0de080d5131d3257dc393f81ba40ce3",
			sortedHash:   "6bdc745db166942bba7a31ae16ae447897d2223a7d16c057db9ee9f285fea1ef",
		},
	}

	for _, test := range tests {
		// Load and deserialize the test transaction.
		filePath := filepath.Join("testdata", test.hexFile)
		txHexBytes, err := ioutil.ReadFile(filePath)
		if err != nil {
			t.Errorf("ReadFile (%s): failed to read test file: %v",
				test.name, err)
			continue
		}
		txBytes, err := hex.DecodeString(string(txHexBytes))
		if err != nil {
			t.Errorf("DecodeString (%s): failed to decode tx: %v",
				test.name, err)
			continue
		}
		var tx wire.MsgTx
		err = tx.Deserialize(bytes.NewReader(txBytes))
		if err != nil {
			t.Errorf("Deserialize (%s): unexpected error %v",
				test.name, err)
			continue
		}

		// Ensure the sort order of the original transaction matches the
		// expected value.
		if got := txsort.IsSorted(&tx); got != test.isSorted {
			t.Errorf("IsSorted (%s): sort does not match "+
				"expected - got %v, want %v", test.name, got,
				test.isSorted)
			continue
		}

		// Sort the transaction and ensure the resulting hash is the
		// expected value.
		sortedTx := txsort.Sort(&tx)
		if got := sortedTx.TxHash().String(); got != test.sortedHash {
			t.Errorf("Sort (%s): sorted hash does not match "+
				"expected - got %v, want %v", test.name, got,
				test.sortedHash)
			continue
		}

		// Ensure the original transaction is not modified.
		if got := tx.TxHash().String(); got != test.unsortedHash {
			t.Errorf("Sort (%s): unsorted hash does not match "+
				"expected - got %v, want %v", test.name, got,
				test.unsortedHash)
			continue
		}

		// Now sort the transaction using the mutable version and ensure
		// the resulting hash is the expected value.
		txsort.InPlaceSort(&tx)
		if got := tx.TxHash().String(); got != test.sortedHash {
			t.Errorf("SortMutate (%s): sorted hash does not match "+
				"expected - got %v, want %v", test.name, got,
				test.sortedHash)
			continue
		}
	}
}
