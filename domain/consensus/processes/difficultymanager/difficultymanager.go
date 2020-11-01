package difficultymanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/util/bigintpool"
	"math/big"
	"time"
)

// DifficultyManager provides a method to resolve the
// difficulty value of a block
type difficultyManager struct {
	databaseContext                model.DBReader
	ghostdagManager                model.GHOSTDAGManager
	ghostdagStore                  model.GHOSTDAGDataStore
	headerStore                    model.BlockHeaderStore
	dagTopologyManager             model.DAGTopologyManager
	powMax                         *big.Int
	difficultyAdjustmentWindowSize uint64
	targetTimePerBlock             time.Duration
	powMaxBits                     uint32
}

// New instantiates a new DifficultyManager
func New(
	databaseContext model.DBReader,
	ghostdagManager model.GHOSTDAGManager,
	ghostdagStore model.GHOSTDAGDataStore,
	headerStore model.BlockHeaderStore,
	dagTopologyManager model.DAGTopologyManager,
	powMax *big.Int,
	difficultyAdjustmentWindowSize uint64,
	targetTimePerBlock time.Duration,
) model.DifficultyManager {
	return &difficultyManager{
		databaseContext:                databaseContext,
		ghostdagManager:                ghostdagManager,
		ghostdagStore:                  ghostdagStore,
		headerStore:                    headerStore,
		dagTopologyManager:             dagTopologyManager,
		powMax:                         powMax,
		difficultyAdjustmentWindowSize: difficultyAdjustmentWindowSize,
		targetTimePerBlock:             targetTimePerBlock,
		powMaxBits:                     util.BigToCompact(powMax),
	}
}

// RequiredDifficulty returns the difficulty required for some block
func (dm *difficultyManager) RequiredDifficulty(blockHash *externalapi.DomainHash) (uint32, error) {

	parents, err := dm.dagTopologyManager.Parents(blockHash)
	if err != nil {
		return 0, err
	}
	// Genesis block
	if len(parents) == 0 {
		return dm.powMaxBits, nil
	}

	// find bluestParent
	bluestParent := parents[0]
	bluestGhostDAG, err := dm.ghostdagStore.Get(dm.databaseContext, bluestParent)
	if err != nil {
		return 0, err
	}
	for i := 1; i < len(parents); i++ {
		parentGhostDAG, err := dm.ghostdagStore.Get(dm.databaseContext, parents[i])
		if err != nil {
			return 0, err
		}
		newBluest, err := dm.ghostdagManager.ChooseSelectedParent(bluestParent, parents[i])
		if err != nil {
			return 0, err
		}
		if bluestParent != newBluest {
			bluestParent = newBluest
			bluestGhostDAG = parentGhostDAG
		}
	}

	// Not enough blocks for building a difficulty window.
	if bluestGhostDAG.BlueScore < dm.difficultyAdjustmentWindowSize+1 {
		return dm.powMaxBits, nil
	}

	// Fetch window of dag.difficultyAdjustmentWindowSize + 1 so we can have dag.difficultyAdjustmentWindowSize block intervals
	timestampsWindow, err := dm.blueBlockWindow(bluestParent, dm.difficultyAdjustmentWindowSize+1)
	if err != nil {
		return 0, err
	}
	windowMinTimestamp, windowMaxTimeStamp := timestampsWindow.minMaxTimestamps()

	// Remove the last block from the window so to calculate the average target of dag.difficultyAdjustmentWindowSize blocks
	targetsWindow := timestampsWindow[:dm.difficultyAdjustmentWindowSize]

	// Calculate new target difficulty as:
	// averageWindowTarget * (windowMinTimestamp / (targetTimePerBlock * windowSize))
	// The result uses integer division which means it will be slightly
	// rounded down.
	newTarget := bigintpool.Acquire(0)
	defer bigintpool.Release(newTarget)
	windowTimeStampDifference := bigintpool.Acquire(windowMaxTimeStamp - windowMinTimestamp)
	defer bigintpool.Release(windowTimeStampDifference)
	targetTimePerBlock := bigintpool.Acquire(dm.targetTimePerBlock.Milliseconds())
	defer bigintpool.Release(targetTimePerBlock)
	difficultyAdjustmentWindowSize := bigintpool.Acquire(int64(dm.difficultyAdjustmentWindowSize))
	defer bigintpool.Release(difficultyAdjustmentWindowSize)

	targetsWindow.averageTarget(newTarget)
	newTarget.
		Mul(newTarget, windowTimeStampDifference).
		Div(newTarget, targetTimePerBlock).
		Div(newTarget, difficultyAdjustmentWindowSize)
	if newTarget.Cmp(dm.powMax) > 0 {
		return dm.powMaxBits, nil
	}
	newTargetBits := util.BigToCompact(newTarget)
	return newTargetBits, nil
}
