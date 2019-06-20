// Copyright (c) 2014-2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package blockdag

import (
	"github.com/daglabs/btcd/dagconfig"
	"github.com/daglabs/btcd/util/daghash"
	"github.com/daglabs/btcd/wire"
	"math/big"
	"testing"
	"time"

	"github.com/daglabs/btcd/util"
)

// TestBigToCompact ensures BigToCompact converts big integers to the expected
// compact representation.
func TestBigToCompact(t *testing.T) {
	tests := []struct {
		in  int64
		out uint32
	}{
		{0, 0},
		{-1, 25231360},
	}

	for x, test := range tests {
		n := big.NewInt(test.in)
		r := util.BigToCompact(n)
		if r != test.out {
			t.Errorf("TestBigToCompact test #%d failed: got %d want %d\n",
				x, r, test.out)
			return
		}
	}
}

// TestCompactToBig ensures CompactToBig converts numbers using the compact
// representation to the expected big intergers.
func TestCompactToBig(t *testing.T) {
	tests := []struct {
		in  uint32
		out int64
	}{
		{10000000, 0},
	}

	for x, test := range tests {
		n := util.CompactToBig(test.in)
		want := big.NewInt(test.out)
		if n.Cmp(want) != 0 {
			t.Errorf("TestCompactToBig test #%d failed: got %d want %d\n",
				x, n.Int64(), want.Int64())
			return
		}
	}
}

// TestCalcWork ensures CalcWork calculates the expected work value from values
// in compact representation.
func TestCalcWork(t *testing.T) {
	tests := []struct {
		in  uint32
		out int64
	}{
		{10000000, 0},
	}

	for x, test := range tests {
		bits := uint32(test.in)

		r := util.CalcWork(bits)
		if r.Int64() != test.out {
			t.Errorf("TestCalcWork test #%d failed: got %v want %d\n",
				x, r.Int64(), test.out)
			return
		}
	}
}

func TestDifficulty(t *testing.T) {
	dag := newTestDAG(&dagconfig.SimNetParams)
	nodes := make([]*blockNode, 0)
	nonce := uint64(0)
	zeroTime := time.Unix(0, 0)
	addNode := func(parents blockSet, blockTime time.Time) *blockNode {
		bluestParent := parents.bluest()
		if blockTime == zeroTime {
			blockTime = time.Unix(bluestParent.timestamp+1, 0)
		}
		header := &wire.BlockHeader{
			ParentHashes:         parents.hashes(),
			Bits:                 dag.calcNextRequiredDifficulty(bluestParent, blockTime),
			Nonce:                nonce,
			Timestamp:            blockTime,
			HashMerkleRoot:       &daghash.ZeroHash,
			AcceptedIDMerkleRoot: &daghash.ZeroHash,
			UTXOCommitment:       &daghash.ZeroHash,
		}
		node := newBlockNode(header, parents, dag.dagParams.K)
		node.updateParentsChildren()
		nodes = append(nodes, node)
		nonce++
		return node
	}
	firstNode := addNode(setFromSlice(dag.genesis), zeroTime)
	if firstNode.bits != dag.genesis.bits {
		t.Fatalf("First block should have the same difficulty as genesis")
	}
	tip := firstNode
	for i := 0; i < 2639; i++ {
		tip = addNode(setFromSlice(tip), zeroTime)
		if tip.bits != dag.genesis.bits {
			t.Fatalf("%d: As long as the block rate remains the same, the difficulty shouldn't change", i)
		}
	}
	nodeInThePast := addNode(setFromSlice(tip), zeroTime)
	return
	if nodeInThePast.bits == tip.bits{
		t.Fatalf("As long as the block rate remains the same, the difficulty shouldn't change")
	}
	tip = nodeInThePast
	for i := 0; i < 3000; i++ {
		prevTip := tip
		tip = addNode(setFromSlice(prevTip), zeroTime)
		if tip.bits != prevTip.bits {
			t.Fatalf("%d: As long as the block rate remains the same, the difficulty shouldn't change", i)
		}
	}
	tip = addNode(setFromSlice(tip), zeroTime)
	if tip.bits != dag.genesis.bits {
		t.Fatalf("First block should have the same difficulty as genesis")
	}
}
