// Copyright (c) 2013-2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package blockdag

import (
	"math/big"
	"time"

	"github.com/daglabs/btcd/util"
)

// calcEasiestDifficulty calculates the easiest possible difficulty that a block
// can have given starting difficulty bits and a duration.  It is mainly used to
// verify that claimed proof of work by a block is sane as compared to a
// known good checkpoint.
func (dag *BlockDAG) calcEasiestDifficulty(bits uint32, duration time.Duration) uint32 {
	// Convert types used in the calculations below.
	durationVal := int64(duration / time.Second)
	adjustmentFactor := big.NewInt(dag.dagParams.RetargetAdjustmentFactor)

	// The test network rules allow minimum difficulty blocks after more
	// than twice the desired amount of time needed to generate a block has
	// elapsed.
	if dag.dagParams.ReduceMinDifficulty {
		reductionTime := int64(dag.dagParams.MinDiffReductionTime /
			time.Second)
		if durationVal > reductionTime {
			return dag.dagParams.PowLimitBits
		}
	}

	// Since easier difficulty equates to higher numbers, the easiest
	// difficulty for a given duration is the largest value possible given
	// the number of retargets for the duration and starting difficulty
	// multiplied by the max adjustment factor.
	newTarget := util.CompactToBig(bits)
	for durationVal > 0 && newTarget.Cmp(dag.dagParams.PowLimit) < 0 {
		newTarget.Mul(newTarget, adjustmentFactor)
		durationVal -= dag.maxRetargetTimespan
	}

	// Limit new value to the proof of work limit.
	if newTarget.Cmp(dag.dagParams.PowLimit) > 0 {
		newTarget.Set(dag.dagParams.PowLimit)
	}

	return util.BigToCompact(newTarget)
}

// findPrevTestNetDifficulty returns the difficulty of the previous block which
// did not have the special testnet minimum difficulty rule applied.
//
// This function MUST be called with the chain state lock held (for writes).
func (dag *BlockDAG) findPrevTestNetDifficulty(startNode *blockNode) uint32 {
	// Search backwards through the chain for the last block without
	// the special rule applied.
	iterNode := startNode
	for iterNode != nil && iterNode.height%dag.blocksPerRetarget != 0 &&
		iterNode.bits == dag.dagParams.PowLimitBits {

		iterNode = iterNode.selectedParent
	}

	// Return the found difficulty or the minimum difficulty if no
	// appropriate block was found.
	lastBits := dag.dagParams.PowLimitBits
	if iterNode != nil {
		lastBits = iterNode.bits
	}
	return lastBits
}

// calcNextRequiredDifficulty calculates the required difficulty for the block
// after the passed previous block node based on the difficulty retarget rules.
// This function differs from the exported CalcNextRequiredDifficulty in that
// the exported version uses the current best chain as the previous block node
// while this function accepts any block node.
func (dag *BlockDAG) calcNextRequiredDifficulty(bluestParent *blockNode, newBlockTime time.Time) (uint32, error) {
	// Genesis block.
	if bluestParent == nil {
		return dag.dagParams.PowLimitBits, nil
	}

	// Return the previous block's difficulty requirements if this block
	// is not at a difficulty retarget interval.
	if (bluestParent.height+1)%dag.blocksPerRetarget != 0 {
		// For networks that support it, allow special reduction of the
		// required difficulty once too much time has elapsed without
		// mining a block.
		if dag.dagParams.ReduceMinDifficulty {
			// Return minimum difficulty when more than the desired
			// amount of time has elapsed without mining a block.
			reductionTime := int64(dag.dagParams.MinDiffReductionTime /
				time.Second)
			allowMinTime := bluestParent.timestamp + reductionTime
			if newBlockTime.Unix() > allowMinTime {
				return dag.dagParams.PowLimitBits, nil
			}

			// The block was mined within the desired timeframe, so
			// return the difficulty for the last block which did
			// not have the special minimum difficulty rule applied.
			return dag.findPrevTestNetDifficulty(bluestParent), nil
		}

		// For the main network (or any unrecognized networks), simply
		// return the previous block's difficulty requirements.
		return bluestParent.bits, nil
	}

	// Get the block node at the previous retarget (targetTimespan days
	// worth of blocks).
	firstNode := bluestParent.RelativeAncestor(dag.blocksPerRetarget - 1)
	if firstNode == nil {
		return 0, AssertError("unable to obtain previous retarget block")
	}

	// Limit the amount of adjustment that can occur to the previous
	// difficulty.
	actualTimespan := bluestParent.timestamp - firstNode.timestamp
	adjustedTimespan := actualTimespan
	if actualTimespan < dag.minRetargetTimespan {
		adjustedTimespan = dag.minRetargetTimespan
	} else if actualTimespan > dag.maxRetargetTimespan {
		adjustedTimespan = dag.maxRetargetTimespan
	}

	// Calculate new target difficulty as:
	//  currentDifficulty * (adjustedTimespan / targetTimespan)
	// The result uses integer division which means it will be slightly
	// rounded down.  Bitcoind also uses integer division to calculate this
	// result.
	oldTarget := util.CompactToBig(bluestParent.bits)
	newTarget := new(big.Int).Mul(oldTarget, big.NewInt(adjustedTimespan))
	targetTimeSpan := int64(dag.dagParams.TargetTimespan / time.Second)
	newTarget.Div(newTarget, big.NewInt(targetTimeSpan))

	// Limit new value to the proof of work limit.
	if newTarget.Cmp(dag.dagParams.PowLimit) > 0 {
		newTarget.Set(dag.dagParams.PowLimit)
	}

	// Log new target difficulty and return it.  The new target logging is
	// intentionally converting the bits back to a number instead of using
	// newTarget since conversion to the compact representation loses
	// precision.
	newTargetBits := util.BigToCompact(newTarget)
	log.Debugf("Difficulty retarget at block height %d", bluestParent.height+1)
	log.Debugf("Old target %08x (%064x)", bluestParent.bits, oldTarget)
	log.Debugf("New target %08x (%064x)", newTargetBits, util.CompactToBig(newTargetBits))
	log.Debugf("Actual timespan %s, adjusted timespan %s, target timespan %s",
		time.Duration(actualTimespan)*time.Second,
		time.Duration(adjustedTimespan)*time.Second,
		dag.dagParams.TargetTimespan)

	return newTargetBits, nil
}

// CalcNextRequiredDifficulty calculates the required difficulty for the block
// after the end of the current best chain based on the difficulty retarget
// rules.
//
// This function is safe for concurrent access.
func (dag *BlockDAG) CalcNextRequiredDifficulty(timestamp time.Time) (uint32, error) {
	dag.dagLock.RLock()
	difficulty, err := dag.calcNextRequiredDifficulty(dag.selectedTip(), timestamp)
	dag.dagLock.RUnlock()
	return difficulty, err
}
