package coinbasemanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"math/big"
)

func (c *coinbaseManager) isBlockRewardFixed(stagingArea *model.StagingArea, blockPruningPoint *externalapi.DomainHash) (bool, error) {
	blockPruningPointIndex, found, err := c.findPruningPointIndex(stagingArea, blockPruningPoint)
	if err != nil {
		return false, err
	}

	// The given `pruningPointBlock` may only not be found under one circumstance:
	// we're currently in the process of building the next pruning point. As such,
	// we must manually set highIndex to currentIndex + 1 because the next pruning
	// point is not yet stored in the database
	highPruningPointIndex := blockPruningPointIndex
	highPruningPointHash := blockPruningPoint
	if !found {
		currentPruningPointIndex, err := c.pruningStore.CurrentPruningPointIndex(c.databaseContext, stagingArea)
		if err != nil {
			return false, err
		}
		highPruningPointIndex = currentPruningPointIndex + 1
	}

	for {
		if highPruningPointIndex <= c.fixedSubsidySwitchPruningPointInterval {
			break
		}

		lowPruningPointIndex := highPruningPointIndex - c.fixedSubsidySwitchPruningPointInterval
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
		if estimatedAverageHashRate.Cmp(c.fixedSubsidySwitchHashRateThreshold) >= 0 {
			return true, nil
		}

		highPruningPointIndex--
		highPruningPointHash, err = c.pruningStore.PruningPointByIndex(c.databaseContext, stagingArea, highPruningPointIndex)
		if err != nil {
			return false, err
		}
	}

	return false, nil
}

func (c *coinbaseManager) findPruningPointIndex(stagingArea *model.StagingArea, pruningPointHash *externalapi.DomainHash) (uint64, bool, error) {
	currentPruningPointHash, err := c.pruningStore.PruningPoint(c.databaseContext, stagingArea)
	if err != nil {
		return 0, false, err
	}
	currentPruningPointIndex, err := c.pruningStore.CurrentPruningPointIndex(c.databaseContext, stagingArea)
	if err != nil {
		return 0, false, err
	}
	for !currentPruningPointHash.Equal(pruningPointHash) && currentPruningPointIndex > 0 {
		currentPruningPointIndex--
		currentPruningPointHash, err = c.pruningStore.PruningPointByIndex(c.databaseContext, stagingArea, currentPruningPointIndex)
		if err != nil {
			return 0, false, err
		}
	}
	if currentPruningPointIndex == 0 && !currentPruningPointHash.Equal(pruningPointHash) {
		return 0, false, nil
	}
	return currentPruningPointIndex, true, nil
}
