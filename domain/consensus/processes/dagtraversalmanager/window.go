package dagtraversalmanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/infrastructure/db/database"
)

// BlockWindow returns a blockWindow of the given size that contains the
// blocks in the past of highHash, the sorting is unspecified.
// If the number of blocks in the past of startingNode is less then windowSize,
func (dtm *dagTraversalManager) BlockWindow(stagingArea *model.StagingArea, highHash *externalapi.DomainHash, windowSize int) ([]*externalapi.DomainHash, error) {

	if highHash.Equal(dtm.genesisHash) {
		return nil, nil
	}

	current := highHash
	currentGHOSTDAGData, err := dtm.ghostdagDataStore.Get(dtm.databaseContext, stagingArea, highHash, false)
	if err != nil {
		return nil, err
	}

	windowHeap := dtm.newSizedUpHeap(stagingArea, windowSize)

	for {
		if currentGHOSTDAGData.SelectedParent().Equal(dtm.genesisHash) {
			break
		}

		if currentGHOSTDAGData.SelectedParent().Equal(model.VirtualGenesisBlockHash) {
			for i := uint64(0); ; i++ {
				daaBlock, err := dtm.daaWindowStore.DAAWindowBlock(dtm.databaseContext, stagingArea, current, i)
				if database.IsNotFoundError(err) {
					break
				}
				if err != nil {
					return nil, err
				}

				added, err := windowHeap.tryPushWithGHOSTDAGData(daaBlock.Hash, daaBlock.GHOSTDAGData)
				if err != nil {
					return nil, err
				}

				// Because the DAA window is sorted by blue work, if this block is not added the next one
				// won't be added as well.
				if !added {
					break
				}
			}
			break
		}

		selectedParentGHOSTDAGData, err := dtm.ghostdagDataStore.Get(
			dtm.databaseContext, stagingArea, currentGHOSTDAGData.SelectedParent(), false)
		if err != nil {
			return nil, err
		}
		added, err := windowHeap.tryPushWithGHOSTDAGData(currentGHOSTDAGData.SelectedParent(), selectedParentGHOSTDAGData)
		if err != nil {
			return nil, err
		}

		// If the window is full and the selected parent is less than the minimum then we break
		// because this means that there cannot be any more blocks in the past with higher blueWork
		if !added {
			break
		}

		// Now we go over the merge set.
		// Remove the SP from the blue merge set because we already added it.
		mergeSetBlues := currentGHOSTDAGData.MergeSetBlues()[1:]
		// Go over the merge set in reverse because it's ordered in reverse by blueWork.
		for i := len(mergeSetBlues) - 1; i >= 0; i-- {
			added, err := windowHeap.tryPush(mergeSetBlues[i])
			if err != nil {
				return nil, err
			}
			// If it's smaller than minimum then we won't be able to add the rest because they're even smaller.
			if !added {
				break
			}
		}

		mergeSetReds := currentGHOSTDAGData.MergeSetReds()
		for i := len(mergeSetReds) - 1; i >= 0; i-- {
			added, err := windowHeap.tryPush(mergeSetReds[i])
			if err != nil {
				return nil, err
			}
			// If it's smaller than minimum then we won't be able to add the rest because they're even smaller.
			if !added {
				break
			}
		}

		current = currentGHOSTDAGData.SelectedParent()
		currentGHOSTDAGData = selectedParentGHOSTDAGData
	}

	window := make([]*externalapi.DomainHash, 0, len(windowHeap.impl.slice))
	for _, b := range windowHeap.impl.slice {
		window = append(window, b.hash)
	}
	return window, nil
}
