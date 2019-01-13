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
	merkles := BuildHashMerkleTreeStore(block.Transactions())
	calculatedMerkleRoot := merkles[len(merkles)-1]
	wantMerkle := &Block100000.Header.HashMerkleRoot
	if !wantMerkle.IsEqual(calculatedMerkleRoot) {
		t.Errorf("BuildHashMerkleTreeStore: merkle root mismatch - "+
			"got %v, want %v", calculatedMerkleRoot, wantMerkle)
	}
}
