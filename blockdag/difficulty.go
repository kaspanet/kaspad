// Copyright (c) 2013-2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package blockdag

import (
	"math/big"
	"sort"
	"time"

	"github.com/daglabs/btcd/util"
)

func blueBlockWindow(startingNode *blockNode, windowSize uint64, fillBlankSpaceWithGenesis bool) []*blockNode {
	window := make([]*blockNode, 0, windowSize)
	currentNode := startingNode
	for uint64(len(window)) < windowSize {
		if currentNode.selectedParent != nil {
			for _, blue := range currentNode.blues {
				window = append(window, blue)
				if uint64(len(window)) < windowSize {
					break
				}
			}
			currentNode = currentNode.selectedParent
		} else {
			if !fillBlankSpaceWithGenesis {
				break
			}
			window = append(window, currentNode)
		}
	}
	return window
}

func calcBlockWindowMinMaxAndMedianTimestamps(window []*blockNode) (min, max, median int64) {
	timestamps := make([]int64, len(window))
	for i, node := range window {
		timestamps[i] = node.timestamp
	}
	sort.Sort(timeSorter(timestamps))
	min = timestamps[0]
	max = timestamps[len(timestamps)-1]
	median = timestamps[len(timestamps)/2]
	return
}

func calcAverageBlockWindowTarget(window []*blockNode) *big.Int {
	averageTarget := big.NewInt(0)
	for _, node := range window {
		target := util.CompactToBig(node.bits)
		averageTarget.Add(averageTarget, target)
	}
	return averageTarget.Div(averageTarget, big.NewInt(int64(len(window))))
}

// calcNextRequiredDifficulty calculates the required difficulty for the block
// after the passed previous block node based on the difficulty retarget rules.
// This function differs from the exported CalcNextRequiredDifficulty in that
// the exported version uses the current best chain as the previous block node
// while this function accepts any block node.
func (dag *BlockDAG) calcNextRequiredDifficulty(bluestParent *blockNode, newBlockTime time.Time) uint32 {
	// Genesis block.
	if bluestParent == nil {
		return dag.dagParams.PowLimitBits
	}

	window := blueBlockWindow(bluestParent, dag.difficultyAdjustmentWindowSize, false)
	windowMinTimestamp, _, _ := calcBlockWindowMinMaxAndMedianTimestamps(window)
	adjustmentFactor := windowMinTimestamp / int64(dag.targetTimePerBlock) / int64(len(window))
	newTarget := calcAverageBlockWindowTarget(window)
	newTarget.Mul(newTarget, big.NewInt(adjustmentFactor))
	if newTarget.Cmp(dag.dagParams.PowLimit) > 0 {
		newTarget.Set(dag.dagParams.PowLimit)
	}
	newTargetBits := util.BigToCompact(newTarget)
	return newTargetBits
}

// CalcNextRequiredDifficulty calculates the required difficulty for the block
// after the end of the current best chain based on the difficulty retarget
// rules.
//
// This function is safe for concurrent access.
func (dag *BlockDAG) CalcNextRequiredDifficulty(timestamp time.Time) uint32 {
	difficulty := dag.calcNextRequiredDifficulty(dag.selectedTip(), timestamp)
	return difficulty
}
