// Copyright (c) 2014-2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package blockdag

import (
	"github.com/kaspanet/kaspad/dagconfig"
	"math/big"
	"testing"
	"time"

	"github.com/kaspanet/kaspad/util"
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
	params := dagconfig.SimnetParams
	params.K = 1
	params.DifficultyAdjustmentWindowSize = 264
	dag, teardownFunc, err := DAGSetup("TestDifficulty", true, Config{
		DAGParams: &params,
	})
	if err != nil {
		t.Fatalf("Failed to setup DAG instance: %v", err)
	}
	defer teardownFunc()

	zeroTime := time.Unix(0, 0)
	addNode := func(parents blockSet, blockTime time.Time) *blockNode {
		bluestParent := parents.bluest()
		if blockTime == zeroTime {
			blockTime = time.Unix(bluestParent.timestamp, 0)
			blockTime = blockTime.Add(params.TargetTimePerBlock)
		}
		block, err := PrepareBlockForTest(dag, parents.hashes(), nil)
		if err != nil {
			t.Fatalf("unexpected error in PrepareBlockForTest: %s", err)
		}
		block.Header.Timestamp = blockTime
		block.Header.Bits = dag.requiredDifficulty(bluestParent, blockTime)

		isOrphan, isDelayed, err := dag.ProcessBlock(util.NewBlock(block), BFNoPoWCheck)
		if err != nil {
			t.Fatalf("unexpected error in ProcessBlock: %s", err)
		}
		if isDelayed {
			t.Fatalf("block is too far in the future")
		}
		if isOrphan {
			t.Fatalf("block was unexpectedly orphan")
		}
		return dag.index.LookupNode(block.BlockHash())
	}
	tip := dag.genesis
	for i := uint64(0); i < dag.difficultyAdjustmentWindowSize; i++ {
		tip = addNode(blockSetFromSlice(tip), zeroTime)
		if tip.bits != dag.genesis.bits {
			t.Fatalf("As long as the bluest parent's blue score is less then the difficulty adjustment " +
				"window size, the difficulty should be the same as genesis'")
		}
	}
	for i := uint64(0); i < dag.difficultyAdjustmentWindowSize+100; i++ {
		tip = addNode(blockSetFromSlice(tip), zeroTime)
		if tip.bits != dag.genesis.bits {
			t.Fatalf("As long as the block rate remains the same, the difficulty shouldn't change")
		}
	}
	nodeInThePast := addNode(blockSetFromSlice(tip), tip.PastMedianTime(dag))
	if nodeInThePast.bits != tip.bits {
		t.Fatalf("The difficulty should only change when nodeInThePast is in the past of a block bluest parent")
	}
	tip = nodeInThePast

	tip = addNode(blockSetFromSlice(tip), zeroTime)
	if tip.bits != nodeInThePast.bits {
		t.Fatalf("The difficulty should only change when nodeInThePast is in the past of a block bluest parent")
	}
	tip = addNode(blockSetFromSlice(tip), zeroTime)
	if compareBits(tip.bits, nodeInThePast.bits) >= 0 {
		t.Fatalf("tip.bits should be smaller than nodeInThePast.bits because nodeInThePast increased the " +
			"block rate, so the difficulty should increase as well")
	}
	expectedBits := uint32(0x207f83df)
	if tip.bits != expectedBits {
		t.Errorf("tip.bits was expected to be %x but got %x", expectedBits, tip.bits)
	}

	// Increase block rate to increase difficulty
	for i := uint64(0); i < dag.difficultyAdjustmentWindowSize; i++ {
		tip = addNode(blockSetFromSlice(tip), tip.PastMedianTime(dag))
		if compareBits(tip.bits, tip.parents.bluest().bits) > 0 {
			t.Fatalf("Because we're increasing the block rate, the difficulty can't decrease")
		}
	}

	// Add blocks until difficulty stabilizes
	lastBits := tip.bits
	sameBitsCount := uint64(0)
	for sameBitsCount < dag.difficultyAdjustmentWindowSize+1 {
		tip = addNode(blockSetFromSlice(tip), zeroTime)
		if tip.bits == lastBits {
			sameBitsCount++
		} else {
			lastBits = tip.bits
			sameBitsCount = 0
		}
	}
	slowBlockTime := time.Unix(tip.timestamp, 0)
	slowBlockTime = slowBlockTime.Add(params.TargetTimePerBlock + time.Second)
	slowNode := addNode(blockSetFromSlice(tip), slowBlockTime)
	if slowNode.bits != tip.bits {
		t.Fatalf("The difficulty should only change when slowNode is in the past of a block bluest parent")
	}

	tip = slowNode

	tip = addNode(blockSetFromSlice(tip), zeroTime)
	if tip.bits != slowNode.bits {
		t.Fatalf("The difficulty should only change when slowNode is in the past of a block bluest parent")
	}
	tip = addNode(blockSetFromSlice(tip), zeroTime)
	if compareBits(tip.bits, slowNode.bits) <= 0 {
		t.Fatalf("tip.bits should be smaller than slowNode.bits because slowNode decreased the block" +
			" rate, so the difficulty should decrease as well")
	}

	splitNode := addNode(blockSetFromSlice(tip), zeroTime)
	tip = splitNode
	for i := 0; i < 100; i++ {
		tip = addNode(blockSetFromSlice(tip), zeroTime)
	}
	blueTip := tip

	redChainTip := splitNode
	for i := 0; i < 10; i++ {
		redChainTip = addNode(blockSetFromSlice(redChainTip), redChainTip.PastMedianTime(dag))
	}
	tipWithRedPast := addNode(blockSetFromSlice(redChainTip, blueTip), zeroTime)
	tipWithoutRedPast := addNode(blockSetFromSlice(blueTip), zeroTime)
	if tipWithoutRedPast.bits != tipWithRedPast.bits {
		t.Fatalf("tipWithoutRedPast.bits should be the same as tipWithRedPast.bits because red blocks" +
			" shouldn't affect the difficulty")
	}
}

func compareBits(a uint32, b uint32) int {
	aTarget := util.CompactToBig(a)
	bTarget := util.CompactToBig(b)
	return aTarget.Cmp(bTarget)
}
