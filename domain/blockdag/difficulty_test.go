// Copyright (c) 2014-2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package blockdag

import (
	"github.com/kaspanet/kaspad/domain/blocknode"
	"math/big"
	"testing"
	"time"

	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/kaspanet/kaspad/util/mstime"

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
	params := dagconfig.MainnetParams
	params.K = 1
	params.DifficultyAdjustmentWindowSize = 264
	dag, teardownFunc, err := DAGSetup("TestDifficulty", true, Config{
		DAGParams: &params,
	})
	if err != nil {
		t.Fatalf("Failed to setup DAG instance: %v", err)
	}
	defer teardownFunc()

	zeroTime := mstime.Time{}
	addNode := func(parents blocknode.Set, blockTime mstime.Time) *blocknode.Node {
		bluestParent := parents.Bluest()
		if blockTime.IsZero() {
			blockTime = bluestParent.Time()
			blockTime = blockTime.Add(params.TargetTimePerBlock)
		}
		block, err := PrepareBlockForTest(dag, parents.Hashes(), nil)
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
		node, ok := dag.Index.LookupNode(block.BlockHash())
		if !ok {
			t.Fatalf("block %s does not exist in the DAG", block.BlockHash())
		}
		return node
	}
	tip := dag.genesis
	for i := uint64(0); i < dag.difficultyAdjustmentWindowSize; i++ {
		tip = addNode(blocknode.SetFromSlice(tip), zeroTime)
		if tip.Bits != dag.genesis.Bits {
			t.Fatalf("As long as the bluest parent's blue score is less then the difficulty adjustment " +
				"window size, the difficulty should be the same as genesis'")
		}
	}
	for i := uint64(0); i < dag.difficultyAdjustmentWindowSize+100; i++ {
		tip = addNode(blocknode.SetFromSlice(tip), zeroTime)
		if tip.Bits != dag.genesis.Bits {
			t.Fatalf("As long as the block rate remains the same, the difficulty shouldn't change")
		}
	}
	nodeInThePast := addNode(blocknode.SetFromSlice(tip), dag.PastMedianTime(tip))
	if nodeInThePast.Bits != tip.Bits {
		t.Fatalf("The difficulty should only change when nodeInThePast is in the past of a block bluest parent")
	}
	tip = nodeInThePast

	tip = addNode(blocknode.SetFromSlice(tip), zeroTime)
	if tip.Bits != nodeInThePast.Bits {
		t.Fatalf("The difficulty should only change when nodeInThePast is in the past of a block bluest parent")
	}
	tip = addNode(blocknode.SetFromSlice(tip), zeroTime)
	if compareBits(tip.Bits, nodeInThePast.Bits) >= 0 {
		t.Fatalf("tip.bits should be smaller than nodeInThePast.bits because nodeInThePast increased the " +
			"block rate, so the difficulty should increase as well")
	}
	expectedBits := uint32(0x207f83df)
	if tip.Bits != expectedBits {
		t.Errorf("tip.bits was expected to be %x but got %x", expectedBits, tip.Bits)
	}

	// Increase block rate to increase difficulty
	for i := uint64(0); i < dag.difficultyAdjustmentWindowSize; i++ {
		tip = addNode(blocknode.SetFromSlice(tip), dag.PastMedianTime(tip))
		if compareBits(tip.Bits, tip.Parents.Bluest().Bits) > 0 {
			t.Fatalf("Because we're increasing the block rate, the difficulty can't decrease")
		}
	}

	// Add blocks until difficulty stabilizes
	lastBits := tip.Bits
	sameBitsCount := uint64(0)
	for sameBitsCount < dag.difficultyAdjustmentWindowSize+1 {
		tip = addNode(blocknode.SetFromSlice(tip), zeroTime)
		if tip.Bits == lastBits {
			sameBitsCount++
		} else {
			lastBits = tip.Bits
			sameBitsCount = 0
		}
	}
	slowBlockTime := tip.Time()
	slowBlockTime = slowBlockTime.Add(params.TargetTimePerBlock + time.Second)
	slowNode := addNode(blocknode.SetFromSlice(tip), slowBlockTime)
	if slowNode.Bits != tip.Bits {
		t.Fatalf("The difficulty should only change when slowNode is in the past of a block bluest parent")
	}

	tip = slowNode

	tip = addNode(blocknode.SetFromSlice(tip), zeroTime)
	if tip.Bits != slowNode.Bits {
		t.Fatalf("The difficulty should only change when slowNode is in the past of a block bluest parent")
	}
	tip = addNode(blocknode.SetFromSlice(tip), zeroTime)
	if compareBits(tip.Bits, slowNode.Bits) <= 0 {
		t.Fatalf("tip.bits should be smaller than slowNode.bits because slowNode decreased the block" +
			" rate, so the difficulty should decrease as well")
	}

	splitNode := addNode(blocknode.SetFromSlice(tip), zeroTime)
	tip = splitNode
	for i := 0; i < 100; i++ {
		tip = addNode(blocknode.SetFromSlice(tip), zeroTime)
	}
	blueTip := tip

	redChainTip := splitNode
	for i := 0; i < 10; i++ {
		redChainTip = addNode(blocknode.SetFromSlice(redChainTip), dag.PastMedianTime(redChainTip))
	}
	tipWithRedPast := addNode(blocknode.SetFromSlice(redChainTip, blueTip), zeroTime)
	tipWithoutRedPast := addNode(blocknode.SetFromSlice(blueTip), zeroTime)
	if tipWithoutRedPast.Bits != tipWithRedPast.Bits {
		t.Fatalf("tipWithoutRedPast.bits should be the same as tipWithRedPast.bits because red blocks" +
			" shouldn't affect the difficulty")
	}
}

func compareBits(a uint32, b uint32) int {
	aTarget := util.CompactToBig(a)
	bTarget := util.CompactToBig(b)
	return aTarget.Cmp(bTarget)
}
