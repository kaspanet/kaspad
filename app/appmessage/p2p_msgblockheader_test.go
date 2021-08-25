// Copyright (c) 2013-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package appmessage

import (
	"math/big"
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// TestBlockHeader tests the MsgBlockHeader API.
func TestBlockHeader(t *testing.T) {
	nonce := uint64(0xba4d87a69924a93d)

	parents := []externalapi.BlockLevelParents{[]*externalapi.DomainHash{mainnetGenesisHash, simnetGenesisHash}}

	merkleHash := mainnetGenesisMerkleRoot
	acceptedIDMerkleRoot := exampleAcceptedIDMerkleRoot
	bits := uint32(0x1d00ffff)
	daaScore := uint64(123)
	blueScore := uint64(456)
	blueWork := big.NewInt(789)
	pruningPoint := simnetGenesisHash
	bh := NewBlockHeader(1, parents, merkleHash, acceptedIDMerkleRoot, exampleUTXOCommitment, bits, nonce,
		daaScore, blueScore, blueWork, pruningPoint)

	// Ensure we get the same data back out.
	if !reflect.DeepEqual(bh.Parents, parents) {
		t.Errorf("NewBlockHeader: wrong parents - got %v, want %v",
			spew.Sprint(bh.Parents), spew.Sprint(parents))
	}
	if bh.HashMerkleRoot != merkleHash {
		t.Errorf("NewBlockHeader: wrong merkle root - got %v, want %v",
			spew.Sprint(bh.HashMerkleRoot), spew.Sprint(merkleHash))
	}
	if bh.Bits != bits {
		t.Errorf("NewBlockHeader: wrong bits - got %v, want %v",
			bh.Bits, bits)
	}
	if bh.Nonce != nonce {
		t.Errorf("NewBlockHeader: wrong nonce - got %v, want %v",
			bh.Nonce, nonce)
	}
	if bh.DAAScore != daaScore {
		t.Errorf("NewBlockHeader: wrong daaScore - got %v, want %v",
			bh.DAAScore, daaScore)
	}
	if bh.BlueScore != blueScore {
		t.Errorf("NewBlockHeader: wrong blueScore - got %v, want %v",
			bh.BlueScore, blueScore)
	}
	if bh.BlueWork != blueWork {
		t.Errorf("NewBlockHeader: wrong blueWork - got %v, want %v",
			bh.BlueWork, blueWork)
	}
	if !bh.PruningPoint.Equal(pruningPoint) {
		t.Errorf("NewBlockHeader: wrong pruningPoint - got %v, want %v",
			bh.PruningPoint, pruningPoint)
	}
}
