// Copyright (c) 2013-2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package blockdag

import (
	"math/big"
	"time"

	"github.com/daglabs/btcd/util"
)

// requiredDifficulty calculates the required difficulty for a
// block given its bluest parent.
func (dag *BlockDAG) requiredDifficulty(bluestParent *blockNode, newBlockTime time.Time) uint32 {
	// Genesis block.
	if bluestParent == nil || bluestParent.blueScore < dag.difficultyAdjustmentWindowSize+1 {
		return dag.powMaxBits
	}

	// Fetch window of dag.difficultyAdjustmentWindowSize + 1 so we can have dag.difficultyAdjustmentWindowSize block intervals
	timestampsWindow := blueBlockWindow(bluestParent, dag.difficultyAdjustmentWindowSize+1)
	windowMinTimestamp, windowMaxTimeStamp := timestampsWindow.minMaxTimestamps()

	// Remove the last block from the window so to calculate the average target of dag.difficultyAdjustmentWindowSize blocks
	targetsWindow := timestampsWindow[:dag.difficultyAdjustmentWindowSize]

	// Calculate new target difficulty as:
	// averageWindowTarget * (windowMinTimestamp / (targetTimePerBlock * windowSize))
	// The result uses integer division which means it will be slightly
	// rounded down.
	newTarget := targetsWindow.averageTarget()
	newTarget.
		Mul(newTarget, big.NewInt(windowMaxTimeStamp-windowMinTimestamp)).
		Div(newTarget, big.NewInt(dag.targetTimePerBlock)).
		Div(newTarget, big.NewInt(int64(dag.difficultyAdjustmentWindowSize)))
	if newTarget.Cmp(dag.dagParams.PowMax) > 0 {
		return dag.powMaxBits
	}
	newTargetBits := util.BigToCompact(newTarget)
	return newTargetBits
}

// NextRequiredDifficulty calculates the required difficulty for a block that will
// be built on top of the current tips.
//
// This function is safe for concurrent access.
func (dag *BlockDAG) NextRequiredDifficulty(timestamp time.Time) uint32 {
	difficulty := dag.requiredDifficulty(dag.virtual.parents.bluest(), timestamp)
	return difficulty
}
