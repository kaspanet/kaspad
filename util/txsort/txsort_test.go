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
			unsortedHash: "55721464fe5511e70792da14d7c4f20f6e81d5e7197919e536d0598796daaef3",
			sortedHash:   "573054025c067ab92e3e8b66cf25d16dbb8ab4ff9e9ef9ece79d2aee83f06785",
		},
		{
			name:         "second test case from BIP 69 - already sorted",
			hexFile:      "bip69-2.hex",
			isSorted:     true,
			unsortedHash: "b9ce1181a9f94375000c9a49fc097a9abe9eb85a0f2e6792e1ec0ac24f72172b",
			sortedHash:   "b9ce1181a9f94375000c9a49fc097a9abe9eb85a0f2e6792e1ec0ac24f72172b",
		},
		{
			name:         "block 100001 tx[1] - sorts outputs only, based on amount",
			hexFile:      "bip69-3.hex",
			isSorted:     false,
			unsortedHash: "a1fd1029514fb555ecaa6126849d66a19c4239b2a77bdace6f6ef7db3cc23f30",
			sortedHash:   "4234b089ff83ec954b76e8b56449c095718124ebedd4cf8642d49f02ae55ade2",
		},
		{
			name:         "block 100001 tx[2] - sorts both inputs and outputs",
			hexFile:      "bip69-4.hex",
			isSorted:     false,
			unsortedHash: "d2eb3b56e3be83886dc5ee5789c332ba77f0f3f53abe98d306c1b2e9c25045a2",
			sortedHash:   "280567fe8d4cda60aa1d1b6ca87dd5944d761e8dee72a704af71365696988cda",
		},
		{
			name:         "block 100998 tx[6] - sorts outputs only, based on output script",
			hexFile:      "bip69-5.hex",
			isSorted:     false,
			unsortedHash: "5c83c1b03b7a1c80a44686323ea17d5bc60a976efb5a3de116cb99cac6ec2557",
			sortedHash:   "23c0abbbd17ca34980c9330e10b2c9230d72cec5aba38306ea676bbfa789a125",
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
