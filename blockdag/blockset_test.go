package blockdag

import (
	"testing"

	"github.com/daglabs/btcd/dagconfig/daghash"
)

func TestHashes(t *testing.T) {
	bs := setFromSlice(
		&blockNode{
			hash: daghash.Hash{3},
		},
		&blockNode{
			hash: daghash.Hash{1},
		},
		&blockNode{
			hash: daghash.Hash{0},
		},
		&blockNode{
			hash: daghash.Hash{2},
		},
	)

	expected := []daghash.Hash{
		{0},
		{1},
		{2},
		{3},
	}

	if !daghash.AreEqual(bs.hashes(), expected) {
		t.Errorf("TestHashes: hashes are not ordered as expected")
	}
}
