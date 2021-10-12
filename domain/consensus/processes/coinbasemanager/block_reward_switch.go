package coinbasemanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"math/big"
)

func (c *coinbaseManager) isBlockRewardFixed(stagingArea *model.StagingArea) (bool, error) {
	if c.hasBlockRewardSwitchedToFixed {
		return true, nil
	}

	currentPruningPointIndex, err := c.pruningStore.CurrentPruningPointIndex(c.databaseContext, stagingArea)
	if err != nil {
		return false, err
	}

	for highPruningPointIndex := currentPruningPointIndex; highPruningPointIndex > c.fixedSubsidySwitchPruningPointInterval; highPruningPointIndex-- {
		lowPruningPointIndex := highPruningPointIndex - c.fixedSubsidySwitchPruningPointInterval

		highPruningPointHash, err := c.pruningStore.PruningPointByIndex(c.databaseContext, stagingArea, highPruningPointIndex)
		if err != nil {
			return false, err
		}
		lowPruningPointHash, err := c.pruningStore.PruningPointByIndex(c.databaseContext, stagingArea, lowPruningPointIndex)
		if err != nil {
			return false, err
		}

		highPruningPointGHOSTDAGData, err := c.ghostdagDataStore.Get(c.databaseContext, stagingArea, highPruningPointHash, false)
		if err != nil {
			return false, err
		}
		lowPruningPointGHOSTDAGData, err := c.ghostdagDataStore.Get(c.databaseContext, stagingArea, lowPruningPointHash, true)
		if err != nil {
			return false, err
		}

		blueWorkDifference := new(big.Int).Sub(highPruningPointGHOSTDAGData.BlueWork(), lowPruningPointGHOSTDAGData.BlueWork())
		blueScoreDifference := new(big.Int).SetUint64(highPruningPointGHOSTDAGData.BlueScore() - lowPruningPointGHOSTDAGData.BlueScore())
		estimatedAverageHashRate := new(big.Int).Div(blueWorkDifference, blueScoreDifference)
		if estimatedAverageHashRate.Cmp(c.fixedSubsidySwitchHashRateDifference) >= 0 {
			c.hasBlockRewardSwitchedToFixed = true
			return true, nil
		}
	}

	return false, nil
}
