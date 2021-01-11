package syncmanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/pkg/errors"
)

// createBlockLocator creates a block locator for the passed high and low hashes.
// See the BlockLocator type comments for more details.
func (sm *syncManager) createBlockLocator(lowHash, highHash *externalapi.DomainHash, limit uint32) (externalapi.BlockLocator, error) {
	// We use the selected parent of the high block, so that the
	// block locator won't contain it.
	highBlockGHOSTDAGData, err := sm.ghostdagDataStore.Get(sm.databaseContext, highHash)
	if err != nil {
		return nil, err
	}
	highHash = highBlockGHOSTDAGData.SelectedParent()

	lowBlockGHOSTDAGData, err := sm.ghostdagDataStore.Get(sm.databaseContext, lowHash)
	if err != nil {
		return nil, err
	}
	lowBlockBlueScore := lowBlockGHOSTDAGData.BlueScore()

	currentHash := highHash
	step := uint64(1)
	locator := make(externalapi.BlockLocator, 0)
	// The loop will break if we reached the limit or if we got to lowHash.
	for {
		locator = append(locator, currentHash)

		// Stop if we've reached the limit (if it's set)
		if limit > 0 && uint32(len(locator)) == limit {
			break
		}

		currentBlockGHOSTDAGData, err := sm.ghostdagDataStore.Get(sm.databaseContext, currentHash)
		if err != nil {
			return nil, err
		}
		currentBlockBlueScore := currentBlockGHOSTDAGData.BlueScore()

		// Nothing more to add once the low node has been added.
		if currentBlockBlueScore <= lowBlockBlueScore {
			isCurrentHashInSelectedParentChainOfLowHash, err := sm.dagTopologyManager.IsInSelectedParentChainOf(currentHash, lowHash)
			if err != nil {
				return nil, err
			}
			if !isCurrentHashInSelectedParentChainOfLowHash {
				return nil, errors.Errorf("highHash and lowHash are " +
					"not in the same selected parent chain.")
			}
			break
		}

		// Calculate blueScore of previous node to include ensuring the
		// final node is lowNode.
		nextBlueScore := currentBlockBlueScore - step
		if currentBlockBlueScore < step || nextBlueScore < lowBlockGHOSTDAGData.BlueScore() {
			nextBlueScore = lowBlockGHOSTDAGData.BlueScore()
		}

		// Walk down currentHash's selected parent chain to the appropriate ancestor
		currentHash, err = sm.dagTraversalManager.LowestChainBlockAboveOrEqualToBlueScore(currentHash, nextBlueScore)
		if err != nil {
			return nil, err
		}

		// Double the distance between included hashes
		step *= 2
	}

	return locator, nil
}

// findNextBlockLocatorBoundaries finds the lowest unknown block locator
// hash and the highest known block locator hash. This is used to create the
// next block locator to find the highest shared known chain block with a
// remote kaspad.
func (sm *syncManager) findNextBlockLocatorBoundaries(blockLocator externalapi.BlockLocator) (
	lowHash, highHash *externalapi.DomainHash, err error) {

	// Find the most recent locator block hash in the DAG. In case none of
	// the hashes in the locator are in the DAG, fall back to the genesis block.
	highestKnownHash := sm.genesisBlockHash
	lowestUnknownHash := blockLocator[len(blockLocator)-1]
	for _, hash := range blockLocator {
		exists, err := sm.blockStatusStore.Exists(sm.databaseContext, hash)
		if err != nil {
			return nil, nil, err
		}
		if !exists {
			lowestUnknownHash = hash
		} else {
			highestKnownHash = hash
			break
		}
	}
	return highestKnownHash, lowestUnknownHash, nil
}

func (sm *syncManager) createHeadersSelectedChainBlockLocator(lowHash,
	highHash *externalapi.DomainHash) (externalapi.BlockLocator, error) {

	if highHash.Equal(sm.genesisBlockHash) && lowHash.Equal(sm.genesisBlockHash) {
		return externalapi.BlockLocator{sm.genesisBlockHash}, nil
	}

	lowHashIndex, exists, err := sm.headersSelectedChainStore.GetIndexByHash(sm.databaseContext, lowHash)
	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, errors.Errorf("couldn't find index for low hash %s", lowHash)
	}

	highHashIndex, exists, err := sm.headersSelectedChainStore.GetIndexByHash(sm.databaseContext, highHash)
	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, errors.Errorf("couldn't find index for high hash %s", highHash)
	}

	if highHashIndex < lowHashIndex {
		return nil, errors.Errorf("cannot build block locator while highHash is lower than lowHash")
	}

	locator := externalapi.BlockLocator{}
	currentIndex := highHashIndex
	step := uint64(1)
	for currentIndex > lowHashIndex {
		blockHash, exists, err := sm.headersSelectedChainStore.GetHashByIndex(sm.databaseContext, currentIndex)
		if err != nil {
			return nil, err
		}

		if !exists {
			return nil, errors.Errorf("couldn't find hash for index %d", currentIndex)
		}

		locator = append(locator, blockHash)
		if currentIndex < step {
			break
		}

		currentIndex -= step
		step *= 2
	}

	locator = append(locator, lowHash)
	return locator, nil
}
