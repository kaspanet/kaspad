// Copyright (c) 2013-2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package blockdag

import (
	"testing"

	"github.com/daglabs/btcd/util"
)

// TestMerkle tests the BuildHashMerkleTreeStore API.
func TestMerkle(t *testing.T) {
	block := util.NewBlock(&Block100000)

	hashMerkles := BuildHashMerkleTreeStore(block.Transactions())
	calculatedHashMerkleRoot := hashMerkles[len(hashMerkles)-1]
	wantHashMerkleRoot := &Block100000.Header.HashMerkleRoot
	if !wantHashMerkleRoot.IsEqual(calculatedHashMerkleRoot) {
		t.Errorf("BuildHashMerkleTreeStore: hash merkle root mismatch - "+
			"got %v, want %v", calculatedHashMerkleRoot, wantHashMerkleRoot)
	}

	idMerkles := BuildIDMerkleTreeStore(block.Transactions())
	calculatedIDMerkleRoot := idMerkles[len(idMerkles)-1]
	wantIDMerkleRoot := &Block100000.Header.IDMerkleRoot
	if !wantIDMerkleRoot.IsEqual(calculatedIDMerkleRoot) {
		t.Errorf("BuildIDMerkleTreeStore: ID merkle root mismatch - "+
			"got %v, want %v", calculatedIDMerkleRoot, wantIDMerkleRoot)
	}
}
