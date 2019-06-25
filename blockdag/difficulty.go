// Copyright (c) 2013-2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package blockdag

import (
	"math"
	"math/big"
	"time"

	"github.com/daglabs/btcd/util"
)

func blueBlockWindow(startingNode *blockNode, windowSize uint64, padWithGenesis bool) (window []*blockNode, ok bool) {
	window = make([]*blockNode, 0, windowSize)
	currentNode := startingNode
	for uint64(len(window)) < windowSize && currentNode.selectedParent != nil {
		if currentNode.selectedParent != nil {
			for _, blue := range currentNode.blues {
				window = append(window, blue)
				if uint64(len(window)) < windowSize {
					break
				}
			}
			currentNode = currentNode.selectedParent
		} else {
			if !padWithGenesis {
				return nil, false
			}
			window = append(window, currentNode)
		}
	}

	if uint64(len(window)) < windowSize {
		if !padWithGenesis {
			return nil, false
		}
		genesis := currentNode
		for uint64(len(window)) < windowSize {
			window = append(window, genesis)
		}
	}

	return window, true
}

func blockWindowMinMaxTimestamps(window []*blockNode) (min, max int64) {
	min = math.MaxInt64
	max = 0
	for _, node := range window {
		if node.timestamp < min {
			min = node.timestamp
		}
		if node.timestamp > max {
			max = node.timestamp
		}
	}
	return
}

func averageBlockWindowTarget(window []*blockNode) *big.Int {
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

	// Fetch window of dag.difficultyAdjustmentWindowSize + 1 so we can have dag.difficultyAdjustmentWindowSize block intervals
	timestampsWindow, ok := blueBlockWindow(bluestParent, dag.difficultyAdjustmentWindowSize+1, false)
	if !ok {
		return dag.dagParams.PowLimitBits
	}
	windowMinTimestamp, windowMaxTimeStamp := blockWindowMinMaxTimestamps(timestampsWindow)

	// Remove the last block from the window so to calculate the average target of dag.difficultyAdjustmentWindowSize blocks
	targetsWindow := timestampsWindow[:dag.difficultyAdjustmentWindowSize]

	// Calculate new target difficulty as:
	// averageWindowTarget * (windowMinTimestamp / (targetTimePerBlock * windowSize))
	// The result uses integer division which means it will be slightly
	// rounded down.
	newTarget := averageBlockWindowTarget(targetsWindow)
	newTarget.
		Mul(newTarget, big.NewInt(windowMaxTimeStamp-windowMinTimestamp)).
		Div(newTarget, big.NewInt(dag.targetTimePerBlock)).
		Div(newTarget, big.NewInt(int64(dag.difficultyAdjustmentWindowSize)))
	if newTarget.Cmp(dag.dagParams.PowLimit) > 0 {
		return dag.dagParams.PowLimitBits
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
