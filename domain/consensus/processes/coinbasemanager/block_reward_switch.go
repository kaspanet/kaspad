package coinbasemanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"math/big"
)

func (c *coinbaseManager) isBlockRewardFixed(stagingArea *model.StagingArea, blockPruningPoint *externalapi.DomainHash) (bool, error) {
	blockPruningPointIndex, err := c.findPruningPointIndex(stagingArea, blockPruningPoint)
	if err != nil {
		return false, err
	}

	for highPruningPointIndex := blockPruningPointIndex; highPruningPointIndex > c.fixedSubsidySwitchPruningPointInterval; highPruningPointIndex-- {
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
			return true, nil
		}
	}

	return false, nil
}

func (c *coinbaseManager) findPruningPointIndex(stagingArea *model.StagingArea, pruningPointHash *externalapi.DomainHash) (uint64, error) {
	currentPruningPointHash, err := c.pruningStore.PruningPoint(c.databaseContext, stagingArea)
	if err != nil {
		return 0, err
	}
	currentPruningPointIndex, err := c.pruningStore.CurrentPruningPointIndex(c.databaseContext, stagingArea)
	if err != nil {
		return 0, err
	}
	for !currentPruningPointHash.Equal(pruningPointHash) && currentPruningPointIndex > 0 {
		currentPruningPointIndex--
		currentPruningPointHash, err = c.pruningStore.PruningPointByIndex(c.databaseContext, stagingArea, currentPruningPointIndex)
		if err != nil {
			return 0, err
		}
	}
	return currentPruningPointIndex, nil
}
