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

		highPruningPointHeader, err := c.blockHeaderStore.BlockHeader(c.databaseContext, stagingArea, highPruningPointHash)
		if err != nil {
			return false, err
		}
		lowPruningPointHeader, err := c.blockHeaderStore.BlockHeader(c.databaseContext, stagingArea, lowPruningPointHash)
		if err != nil {
			return false, err
		}

		blueWorkDifference := new(big.Int).Sub(highPruningPointHeader.BlueWork(), lowPruningPointHeader.BlueWork())
		blueScoreDifference := new(big.Int).SetUint64(highPruningPointHeader.BlueScore() - lowPruningPointHeader.BlueScore())
		estimatedAverageHashRate := new(big.Int).Div(blueWorkDifference, blueScoreDifference)
		if estimatedAverageHashRate.Cmp(c.fixedSubsidySwitchHashRateDifference) >= 0 {
			c.hasBlockRewardSwitchedToFixed = true
			return true, nil
		}
	}

	return false, nil
}
