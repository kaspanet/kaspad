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
	"github.com/kaspanet/kaspad/util/mstime"
)

// TestBlockHeader tests the MsgBlockHeader API.
func TestBlockHeader(t *testing.T) {
	nonce := uint64(0xba4d87a69924a93d)

	hashes := []*externalapi.DomainHash{mainnetGenesisHash, simnetGenesisHash}

	merkleHash := mainnetGenesisMerkleRoot
	acceptedIDMerkleRoot := exampleAcceptedIDMerkleRoot
	bits := uint32(0x1d00ffff)
	daaScore := uint64(123)
	blueWork := big.NewInt(456)
	finalityPoint := simnetGenesisHash
	bh := NewBlockHeader(1, hashes, merkleHash, acceptedIDMerkleRoot, exampleUTXOCommitment, bits, nonce,
		daaScore, blueWork, finalityPoint)

	// Ensure we get the same data back out.
	if !reflect.DeepEqual(bh.Parents, hashes) {
		t.Errorf("NewBlockHeader: wrong prev hashes - got %v, want %v",
			spew.Sprint(bh.Parents), spew.Sprint(hashes))
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
	if bh.BlueWork != blueWork {
		t.Errorf("NewBlockHeader: wrong blueWork - got %v, want %v",
			bh.BlueWork, blueWork)
	}
	if !bh.FinalityPoint.Equal(finalityPoint) {
		t.Errorf("NewBlockHeader: wrong finalityHash - got %v, want %v",
			bh.FinalityPoint, finalityPoint)
	}
}

func TestIsGenesis(t *testing.T) {
	nonce := uint64(123123) // 0x1e0f3
	bits := uint32(0x1d00ffff)
	timestamp := mstime.UnixMilliseconds(0x495fab29000)

	baseBlockHdr := &MsgBlockHeader{
		Version:        1,
		Parents:        []*externalapi.DomainHash{mainnetGenesisHash, simnetGenesisHash},
		HashMerkleRoot: mainnetGenesisMerkleRoot,
		Timestamp:      timestamp,
		Bits:           bits,
		Nonce:          nonce,
	}
	genesisBlockHdr := &MsgBlockHeader{
		Version:        1,
		Parents:        []*externalapi.DomainHash{},
		HashMerkleRoot: mainnetGenesisMerkleRoot,
		Timestamp:      timestamp,
		Bits:           bits,
		Nonce:          nonce,
	}

	tests := []struct {
		in        *MsgBlockHeader // Block header to encode
		isGenesis bool            // Expected result for call of .IsGenesis
	}{
		{genesisBlockHdr, true},
		{baseBlockHdr, false},
	}

	t.Logf("Running %d tests", len(tests))
	for i, test := range tests {
		isGenesis := test.in.IsGenesis()
		if isGenesis != test.isGenesis {
			t.Errorf("MsgBlockHeader.IsGenesis: #%d got: %t, want: %t",
				i, isGenesis, test.isGenesis)
		}
	}
}
