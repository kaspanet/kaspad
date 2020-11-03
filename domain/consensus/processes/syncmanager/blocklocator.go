package syncmanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/pkg/errors"
)

// createBlockLocator creates a block locator for the passed high and low hashes.
// See the BlockLocator type comments for more details.
func (sm *syncManager) createBlockLocator(lowHash, highHash *externalapi.DomainHash) (externalapi.BlockLocator, error) {
	// We use the selected parent of the high block, so that the
	// block locator won't contain it.
	highBlockGHOSTDAGData, err := sm.ghostdagDataStore.Get(sm.databaseContext, highHash)
	if err != nil {
		return nil, err
	}
	highHash = highBlockGHOSTDAGData.SelectedParent

	lowBlockGHOSTDAGData, err := sm.ghostdagDataStore.Get(sm.databaseContext, lowHash)
	if err != nil {
		return nil, err
	}
	lowBlockBlueScore := lowBlockGHOSTDAGData.BlueScore

	currentHash := highHash
	step := uint64(1)
	locator := make(externalapi.BlockLocator, 0)
	for currentHash != nil {
		locator = append(locator, currentHash)

		currentBlockGHOSTDAGData, err := sm.ghostdagDataStore.Get(sm.databaseContext, currentHash)
		if err != nil {
			return nil, err
		}
		currentBlockBlueScore := currentBlockGHOSTDAGData.BlueScore

		// Nothing more to add once the low node has been added.
		if currentBlockBlueScore <= lowBlockBlueScore {
			if currentHash != lowHash {
				return nil, errors.Errorf("highHash and lowHash are " +
					"not in the same selected parent chain.")
			}
			break
		}

		// Calculate blueScore of previous node to include ensuring the
		// final node is lowNode.
		nextBlueScore := currentBlockBlueScore - step
		if nextBlueScore < lowBlockGHOSTDAGData.BlueScore {
			nextBlueScore = lowBlockGHOSTDAGData.BlueScore
		}

		// Walk down currentHash's selected parent chain to the appropriate ancestor
		currentHash, err = sm.dagTraversalManager.HighestChainBlockBelowBlueScore(currentHash, nextBlueScore)
		if err != nil {
			return nil, err
		}

		// Double the distance between included hashes
		step *= 2
	}

	return locator, nil
}

func (sm *syncManager) findNextBlockLocatorBoundaries(blockLocator externalapi.BlockLocator) (lowHash, highHash *externalapi.DomainHash, err error) {
	panic("implement me")
}
