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
			unsortedHash: "6b178242e9b6c1d448e2fa7d52fb7ec8dcea289c7852f1184880aabaef4da412",
			sortedHash:   "ab07dbc63d1376cfb4f122722748a9f1487adb66fbbb682c593da602723f22c3",
		},
		{
			name:         "second test case from BIP 69 - already sorted",
			hexFile:      "bip69-2.hex",
			isSorted:     true,
			unsortedHash: "ef51200d830bf2bc77b50e3b92f878d82f6f8d29d7e0d031782f24e2af023601",
			sortedHash:   "ef51200d830bf2bc77b50e3b92f878d82f6f8d29d7e0d031782f24e2af023601",
		},
		{
			name:         "block 100001 tx[1] - sorts outputs only, based on amount",
			hexFile:      "bip69-3.hex",
			isSorted:     false,
			unsortedHash: "36e32ff592786843a4b80dc1029129530e93c885a1cbeb9ef2bed309b336b92e",
			sortedHash:   "f971eb5dd57cc6848c0b903440c175a2fe0d93004099b1582e201de8d60ef6dc",
		},
		{
			name:         "block 100001 tx[2] - sorts both inputs and outputs",
			hexFile:      "bip69-4.hex",
			isSorted:     false,
			unsortedHash: "355b28a1b5e05c937755182169cbb6cfd58c9c50122ce430efb83c650c9aadef",
			sortedHash:   "4f0f023e7fc32a893fa156ca3001d1be4f1ff86f6d7b69d36f76fa0f427e765a",
		},
		{
			name:         "block 100998 tx[6] - sorts outputs only, based on output script",
			hexFile:      "bip69-5.hex",
			isSorted:     false,
			unsortedHash: "f1d10b059664f8711d7a0960ce7763a89a89c697a74b8b5db4dbf850c0973fd6",
			sortedHash:   "76405dc7da88f622481b5e9d6894a45a45f7c376d1f95a390b8a494b7767ff71",
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
