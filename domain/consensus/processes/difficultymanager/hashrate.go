package difficultymanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/infrastructure/logger"
	"github.com/pkg/errors"
	"math/big"
)

func (dm *difficultyManager) EstimateNetworkHashesPerSecond(windowSize int) (uint64, error) {
	onEnd := logger.LogAndMeasureExecutionTime(log, "EstimateNetworkHashesPerSecond")
	defer onEnd()

	stagingArea := model.NewStagingArea()
	return dm.estimateNetworkHashesPerSecond(stagingArea, windowSize)
}

func (dm *difficultyManager) estimateNetworkHashesPerSecond(stagingArea *model.StagingArea, windowSize int) (uint64, error) {
	if windowSize < 2 {
		return 0, errors.Errorf("windowSize must be equal to or greater than 2")
	}

	blockWindow, windowHashes, err := dm.blockWindow(stagingArea, model.VirtualBlockHash, windowSize)
	if err != nil {
		return 0, err
	}

	minWindowTimestamp, maxWindowTimestamp, _, _ := blockWindow.minMaxTimestamps()
	if minWindowTimestamp == maxWindowTimestamp {
		return 0, errors.Errorf("min window timestamp is equal to the max window timestamp")
	}

	firstHash := windowHashes[0]
	firstBlockGHOSTDAGData, err := dm.ghostdagStore.Get(dm.databaseContext, stagingArea, firstHash)
	if err != nil {
		return 0, err
	}
	firstBlockBlueWork := firstBlockGHOSTDAGData.BlueWork()
	minWindowBlueWork := firstBlockBlueWork
	maxWindowBlueWork := firstBlockBlueWork
	for _, hash := range windowHashes[1:] {
		blockGHOSTDAGData, err := dm.ghostdagStore.Get(dm.databaseContext, stagingArea, hash)
		if err != nil {
			return 0, err
		}
		blockBlueWork := blockGHOSTDAGData.BlueWork()
		if blockBlueWork.Cmp(minWindowBlueWork) < 0 {
			minWindowBlueWork = blockBlueWork
		}
		if blockBlueWork.Cmp(maxWindowBlueWork) > 0 {
			maxWindowBlueWork = blockBlueWork
		}
	}

	nominator := new(big.Int).Sub(maxWindowBlueWork, minWindowBlueWork)
	denominator := big.NewInt((maxWindowTimestamp - minWindowTimestamp) / 1000) // Divided by 1000 to convert milliseconds to seconds
	networkHashesPerSecondBigInt := new(big.Int).Div(nominator, denominator)
	return networkHashesPerSecondBigInt.Uint64(), nil
}
