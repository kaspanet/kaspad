// Copyright (c) 2014-2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package blockdag

import (
	"fmt"
	"github.com/daglabs/btcd/dagconfig"
	"github.com/daglabs/btcd/util/daghash"
	"github.com/daglabs/btcd/wire"
	"math/big"
	"reflect"
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
	params := dagconfig.SimNetParams
	params.K = 1
	dag := newTestDAG(&params)
	nonce := uint64(0)
	zeroTime := time.Unix(0, 0)
	addNode := func(parents blockSet, blockTime time.Time) *blockNode {
		bluestParent := parents.bluest()
		if blockTime == zeroTime {
			blockTime = time.Unix(bluestParent.timestamp+1, 0)
		}
		header := &wire.BlockHeader{
			ParentHashes:         parents.hashes(),
			Bits:                 dag.requiredDifficulty(bluestParent, blockTime),
			Nonce:                nonce,
			Timestamp:            blockTime,
			HashMerkleRoot:       &daghash.ZeroHash,
			AcceptedIDMerkleRoot: &daghash.ZeroHash,
			UTXOCommitment:       &daghash.ZeroHash,
		}
		node := newBlockNode(header, parents, dag.dagParams.K)
		node.updateParentsChildren()
		nonce++
		return node
	}
	tip := dag.genesis
	for i := uint64(0); i < dag.difficultyAdjustmentWindowSize; i++ {
		tip = addNode(setFromSlice(tip), zeroTime)
		if tip.bits != dag.genesis.bits {
			t.Fatalf("As long as the bluest parent's blue score is less then the difficulty adjustment window size, the difficulty should be the same as genesis'")
		}
	}
	for i := uint64(0); i < dag.difficultyAdjustmentWindowSize+1000; i++ {
		tip = addNode(setFromSlice(tip), zeroTime)
		if tip.bits != dag.genesis.bits {
			t.Fatalf("As long as the block rate remains the same, the difficulty shouldn't change")
		}
	}
	nodeInThePast := addNode(setFromSlice(tip), tip.PastMedianTime(dag))
	if nodeInThePast.bits != tip.bits {
		t.Fatalf("The difficulty should only change when nodeInThePast is in the past of a block bluest parent")
	}
	tip = nodeInThePast

	tip = addNode(setFromSlice(tip), zeroTime)
	if tip.bits != nodeInThePast.bits {
		t.Fatalf("The difficulty should only change when nodeInThePast is in the past of a block bluest parent")
	}
	tip = addNode(setFromSlice(tip), zeroTime)
	if compareBits(tip.bits, nodeInThePast.bits) >= 0 {
		t.Fatalf("tip.bits should be smaller than nodeInThePast.bits because nodeInThePast increased the block rate, so the difficulty should increase as well")
	}
	expectedBits := uint32(0x207ff395)
	if tip.bits != expectedBits {
		t.Errorf("tip.bits was expected to be %x but got %x", expectedBits, tip.bits)
	}

	// Increase block rate to increase difficulty
	for i := uint64(0); i < dag.difficultyAdjustmentWindowSize; i++ {
		tip = addNode(setFromSlice(tip), tip.PastMedianTime(dag))
		if compareBits(tip.bits, tip.parents.bluest().bits) > 0 {
			t.Fatalf("Because we're increasing the block rate, the difficulty can't decrease")
		}
	}

	// Add blocks until difficulty stabilizes
	lastBits := tip.bits
	sameBitsCount := uint64(0)
	for sameBitsCount < dag.difficultyAdjustmentWindowSize+1 {
		tip = addNode(setFromSlice(tip), zeroTime)
		if tip.bits == lastBits {
			sameBitsCount++
		} else {
			lastBits = tip.bits
			sameBitsCount = 0
		}
	}
	slowNode := addNode(setFromSlice(tip), time.Unix(tip.timestamp+2, 0))
	if slowNode.bits != tip.bits {
		t.Fatalf("The difficulty should only change when slowNode is in the past of a block bluest parent")
	}

	tip = slowNode

	tip = addNode(setFromSlice(tip), zeroTime)
	if tip.bits != slowNode.bits {
		t.Fatalf("The difficulty should only change when slowNode is in the past of a block bluest parent")
	}
	tip = addNode(setFromSlice(tip), zeroTime)
	if compareBits(tip.bits, slowNode.bits) <= 0 {
		t.Fatalf("tip.bits should be smaller than slowNode.bits because slowNode decreased the block rate, so the difficulty should decrease as well")
	}

	splitNode := addNode(setFromSlice(tip), zeroTime)
	tip = splitNode
	for i := 0; i < 100; i++ {
		tip = addNode(setFromSlice(tip), zeroTime)
	}
	blueTip := tip

	redChainTip := splitNode
	for i := 0; i < 10; i++ {
		redChainTip = addNode(setFromSlice(redChainTip), redChainTip.PastMedianTime(dag))
	}
	tipWithRedPast := addNode(setFromSlice(redChainTip, blueTip), zeroTime)
	tipWithoutRedPast := addNode(setFromSlice(blueTip), zeroTime)
	if tipWithoutRedPast.bits != tipWithRedPast.bits {
		t.Fatalf("tipWithoutRedPast.bits should be the same as tipWithRedPast.bits because red blocks shouldn't affect the difficulty")
	}
}

func compareBits(a uint32, b uint32) int {
	aTarget := util.CompactToBig(a)
	bTarget := util.CompactToBig(b)
	return aTarget.Cmp(bTarget)
}

func TestBlueBlockWindow(t *testing.T) {
	params := dagconfig.SimNetParams
	params.K = 1
	dag := newTestDAG(&params)

	windowSize := uint64(10)
	genesisNode := dag.genesis
	blockTime := genesisNode.Header().Timestamp
	blockByIDMap := make(map[string]*blockNode)
	idByBlockMap := make(map[*blockNode]string)
	blockByIDMap["A"] = genesisNode
	idByBlockMap[genesisNode] = "A"
	blockVersion := int32(0x10000000)

	blocksData := []*struct {
		parents                             []string
		id                                  string //id is a virtual entity that is used only for tests so we can define relations between blocks without knowing their hash
		expectedWindowWithoutGenesisPadding []string
		expectedWindowWithGenesisPadding    []string
		expectedOKWithoutGenesisPadding     bool
	}{
		{
			parents:                             []string{"A"},
			id:                                  "B",
			expectedWindowWithGenesisPadding:    []string{"A", "A", "A", "A", "A", "A", "A", "A", "A", "A"},
			expectedWindowWithoutGenesisPadding: nil,
			expectedOKWithoutGenesisPadding:     false,
		},
		{
			parents:                             []string{"B"},
			id:                                  "C",
			expectedWindowWithGenesisPadding:    []string{"B", "A", "A", "A", "A", "A", "A", "A", "A", "A"},
			expectedWindowWithoutGenesisPadding: nil,
			expectedOKWithoutGenesisPadding:     false,
		},
		{
			parents:                             []string{"B"},
			id:                                  "D",
			expectedWindowWithGenesisPadding:    []string{"B", "A", "A", "A", "A", "A", "A", "A", "A", "A"},
			expectedWindowWithoutGenesisPadding: nil,
			expectedOKWithoutGenesisPadding:     false,
		},
		{
			parents:                             []string{"C", "D"},
			id:                                  "E",
			expectedWindowWithGenesisPadding:    []string{"D", "C", "B", "A", "A", "A", "A", "A", "A", "A"},
			expectedWindowWithoutGenesisPadding: nil,
			expectedOKWithoutGenesisPadding:     false,
		},
		{
			parents:                             []string{"C", "D"},
			id:                                  "F",
			expectedWindowWithGenesisPadding:    []string{"D", "C", "B", "A", "A", "A", "A", "A", "A", "A"},
			expectedWindowWithoutGenesisPadding: nil,
			expectedOKWithoutGenesisPadding:     false,
		},
		{
			parents:                             []string{"A"},
			id:                                  "G",
			expectedWindowWithGenesisPadding:    []string{"A", "A", "A", "A", "A", "A", "A", "A", "A", "A"},
			expectedWindowWithoutGenesisPadding: nil,
			expectedOKWithoutGenesisPadding:     false,
		},
		{
			parents:                             []string{"G"},
			id:                                  "H",
			expectedWindowWithGenesisPadding:    []string{"G", "A", "A", "A", "A", "A", "A", "A", "A", "A"},
			expectedWindowWithoutGenesisPadding: nil,
			expectedOKWithoutGenesisPadding:     false,
		},
		{
			parents:                             []string{"H", "F"},
			id:                                  "I",
			expectedWindowWithGenesisPadding:    []string{"F", "D", "C", "B", "A", "A", "A", "A", "A", "A"},
			expectedWindowWithoutGenesisPadding: nil,
			expectedOKWithoutGenesisPadding:     false,
		},
		{
			parents:                             []string{"I"},
			id:                                  "J",
			expectedWindowWithGenesisPadding:    []string{"I", "F", "D", "C", "B", "A", "A", "A", "A", "A"},
			expectedWindowWithoutGenesisPadding: nil,
			expectedOKWithoutGenesisPadding:     false,
		},
		{
			parents:                             []string{"J"},
			id:                                  "K",
			expectedWindowWithGenesisPadding:    []string{"J", "I", "F", "D", "C", "B", "A", "A", "A", "A"},
			expectedWindowWithoutGenesisPadding: nil,
			expectedOKWithoutGenesisPadding:     false,
		},
		{
			parents:                             []string{"K"},
			id:                                  "L",
			expectedWindowWithGenesisPadding:    []string{"K", "J", "I", "F", "D", "C", "B", "A", "A", "A"},
			expectedWindowWithoutGenesisPadding: nil,
			expectedOKWithoutGenesisPadding:     false,
		},
		{
			parents:                             []string{"L"},
			id:                                  "M",
			expectedWindowWithGenesisPadding:    []string{"L", "K", "J", "I", "F", "D", "C", "B", "A", "A"},
			expectedWindowWithoutGenesisPadding: nil,
			expectedOKWithoutGenesisPadding:     false,
		},
		{
			parents:                             []string{"M"},
			id:                                  "N",
			expectedWindowWithGenesisPadding:    []string{"M", "L", "K", "J", "I", "F", "D", "C", "B", "A"},
			expectedWindowWithoutGenesisPadding: []string{"M", "L", "K", "J", "I", "F", "D", "C", "B", "A"},
			expectedOKWithoutGenesisPadding:     true,
		},
		{
			parents:                             []string{"N"},
			id:                                  "O",
			expectedWindowWithGenesisPadding:    []string{"N", "M", "L", "K", "J", "I", "F", "D", "C", "B"},
			expectedWindowWithoutGenesisPadding: []string{"N", "M", "L", "K", "J", "I", "F", "D", "C", "B"},
			expectedOKWithoutGenesisPadding:     true,
		},
	}

	for _, blockData := range blocksData {
		blockTime = blockTime.Add(time.Second)
		parents := blockSet{}
		for _, parentID := range blockData.parents {
			parent := blockByIDMap[parentID]
			parents.add(parent)
		}
		node := newTestNode(parents, blockVersion, 0, blockTime, dag.dagParams.K)
		node.hash = &daghash.Hash{} // It helps to predict hash order
		for i, char := range blockData.id {
			node.hash[i] = byte(char)
		}

		dag.index.AddNode(node)
		node.updateParentsChildren()

		blockByIDMap[blockData.id] = node
		idByBlockMap[node] = blockData.id

		window, ok := blueBlockWindow(node, windowSize, true)
		if !ok {
			t.Errorf("when padWithGenesis is set to true, ok should always be true")
		}
		if err := checkWindowIDs(window, blockData.expectedWindowWithGenesisPadding, idByBlockMap); err != nil {
			t.Errorf("Unexpected values for window with genesis padding for block %s: %s", blockData.id, err)
		}

		window, ok = blueBlockWindow(node, windowSize, false)
		if ok != blockData.expectedOKWithoutGenesisPadding {
			t.Errorf("Unexpected ok value for window without genesis padding for block %s: expected ok to be %t but got %t", blockData.id, blockData.expectedOKWithoutGenesisPadding, ok)
		}
		if ok {
			if err := checkWindowIDs(window, blockData.expectedWindowWithoutGenesisPadding, idByBlockMap); err != nil {
				t.Errorf("Unexpected values for widnow without genesis padding for block %s: %s", blockData.id, err)
			}
		}
	}
}

func checkWindowIDs(window []*blockNode, expectedIDs []string, idByBlockMap map[*blockNode]string) error {
	if len(window) != len(expectedIDs) {

	}
	ids := make([]string, len(window))
	for i, node := range window {
		ids[i] = idByBlockMap[node]
	}
	if !reflect.DeepEqual(ids, expectedIDs) {
		return fmt.Errorf("window expected to have blocks %s but got %s", expectedIDs, ids)
	}
	return nil
}
