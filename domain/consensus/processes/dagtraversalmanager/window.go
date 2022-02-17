package dagtraversalmanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/lrucache"
	"github.com/kaspanet/kaspad/infrastructure/db/database"
)

func (dtm *dagTraversalManager) DAABlockWindow(stagingArea *model.StagingArea, highHash *externalapi.DomainHash) ([]*externalapi.DomainHash, error) {
	return dtm.BlockWindow(stagingArea, highHash, dtm.difficultyAdjustmentWindowSize)
}

// BlockWindow returns a blockWindow of the given size that contains the
// blocks in the past of highHash, the sorting is unspecified.
// If the number of blocks in the past of startingNode is less then windowSize,
func (dtm *dagTraversalManager) BlockWindow(stagingArea *model.StagingArea, highHash *externalapi.DomainHash,
	windowSize int) ([]*externalapi.DomainHash, error) {

	windowHeap, err := dtm.blockWindowHeap(stagingArea, highHash, windowSize)
	if err != nil {
		return nil, err
	}

	window := make([]*externalapi.DomainHash, 0, len(windowHeap.impl.slice))
	for _, b := range windowHeap.impl.slice {
		window = append(window, b.Hash)
	}
	return window, nil
}

func (dtm *dagTraversalManager) BlockWindowWithGHOSTDAGData(stagingArea *model.StagingArea,
	highHash *externalapi.DomainHash, windowSize int) ([]*externalapi.BlockGHOSTDAGDataHashPair, error) {

	windowHeap, err := dtm.blockWindowHeap(stagingArea, highHash, windowSize)
	if err != nil {
		return nil, err
	}

	return windowHeap.impl.slice, nil
}

func (dtm *dagTraversalManager) blockWindowHeap(stagingArea *model.StagingArea,
	highHash *externalapi.DomainHash, windowSize int) (*sizedUpBlockHeap, error) {
	if _, ok := dtm.blockWindowCacheByWindowSize[windowSize]; !ok {
		dtm.blockWindowCacheByWindowSize[windowSize] = lrucache.New(1000, true)
	} else if heap, ok := dtm.blockWindowCacheByWindowSize[windowSize].Get(highHash); ok {
		return heap.(*sizedUpBlockHeap), nil
	}

	heap, err := dtm.calculateBlockWindowHeap(stagingArea, highHash, windowSize)
	if err != nil {
		return nil, err
	}

	if !highHash.Equal(model.VirtualBlockHash) {
		dtm.blockWindowCacheByWindowSize[windowSize].Add(highHash, heap)
	}
	return heap, nil
}

func (dtm *dagTraversalManager) calculateBlockWindowHeap(stagingArea *model.StagingArea,
	highHash *externalapi.DomainHash, windowSize int) (*sizedUpBlockHeap, error) {

	windowHeap := dtm.newSizedUpHeap(stagingArea, windowSize)
	if highHash.Equal(dtm.genesisHash) {
		return windowHeap, nil
	}
	if windowSize == 0 {
		return windowHeap, nil
	}

	current := highHash
	currentGHOSTDAGData, err := dtm.ghostdagDataStore.Get(dtm.databaseContext, stagingArea, highHash, false)
	if err != nil {
		return nil, err
	}

	_, err = dtm.daaWindowStore.DAAWindowBlock(dtm.databaseContext, stagingArea, current, 0)
	isNotFoundError := database.IsNotFoundError(err)
	if !isNotFoundError && err != nil {
		return nil, err
	}

	if isNotFoundError && currentGHOSTDAGData.SelectedParent() != nil {
		if heap, ok := dtm.blockWindowCacheByWindowSize[windowSize].Get(currentGHOSTDAGData.SelectedParent()); ok {
			selectedParentWindowHeap := heap.(*sizedUpBlockHeap)
			windowHeap.impl.slice = make([]*externalapi.BlockGHOSTDAGDataHashPair, len(selectedParentWindowHeap.impl.slice))
			copy(windowHeap.impl.slice, selectedParentWindowHeap.impl.slice)
			selectedParentGHOSTDAGData, err := dtm.ghostdagDataStore.Get(
				dtm.databaseContext, stagingArea, currentGHOSTDAGData.SelectedParent(), false)
			if err != nil {
				return nil, err
			}

			if !currentGHOSTDAGData.SelectedParent().Equal(dtm.genesisHash) {
				_, err = dtm.tryPushMergeSet(windowHeap, currentGHOSTDAGData, selectedParentGHOSTDAGData)
				if err != nil {
					return nil, err
				}
			}

			return windowHeap, nil
		}
	}

	for {
		if currentGHOSTDAGData.SelectedParent().Equal(dtm.genesisHash) {
			break
		}

		_, err := dtm.daaWindowStore.DAAWindowBlock(dtm.databaseContext, stagingArea, current, 0)
		isNotFoundError := database.IsNotFoundError(err)
		if !isNotFoundError && err != nil {
			return nil, err
		}

		if !isNotFoundError {
			for i := uint64(0); ; i++ {
				daaBlock, err := dtm.daaWindowStore.DAAWindowBlock(dtm.databaseContext, stagingArea, current, i)
				if database.IsNotFoundError(err) {
					break
				}
				if err != nil {
					return nil, err
				}

				_, err = windowHeap.tryPushWithGHOSTDAGData(daaBlock.Hash, daaBlock.GHOSTDAGData)
				if err != nil {
					return nil, err
				}

				// Right now we go over all of the window of `current` and filter blocks on the fly.
				// We can optimize it if we make sure that daaWindowStore stores sorted windows, and
				// then return from this function once one block was not added to the heap.
			}
			break
		}

		selectedParentGHOSTDAGData, err := dtm.ghostdagDataStore.Get(
			dtm.databaseContext, stagingArea, currentGHOSTDAGData.SelectedParent(), false)
		if err != nil {
			return nil, err
		}

		done, err := dtm.tryPushMergeSet(windowHeap, currentGHOSTDAGData, selectedParentGHOSTDAGData)
		if err != nil {
			return nil, err
		}
		if done {
			break
		}

		current = currentGHOSTDAGData.SelectedParent()
		currentGHOSTDAGData = selectedParentGHOSTDAGData
	}

	return windowHeap, nil
}

func (dtm *dagTraversalManager) tryPushMergeSet(windowHeap *sizedUpBlockHeap, currentGHOSTDAGData, selectedParentGHOSTDAGData *externalapi.BlockGHOSTDAGData) (bool, error) {
	added, err := windowHeap.tryPushWithGHOSTDAGData(currentGHOSTDAGData.SelectedParent(), selectedParentGHOSTDAGData)
	if err != nil {
		return false, err
	}

	// If the window is full and the selected parent is less than the minimum then we break
	// because this means that there cannot be any more blocks in the past with higher blueWork
	if !added {
		return true, nil
	}

	// Now we go over the merge set.
	// Remove the SP from the blue merge set because we already added it.
	mergeSetBlues := currentGHOSTDAGData.MergeSetBlues()[1:]
	// Go over the merge set in reverse because it's ordered in reverse by blueWork.
	for i := len(mergeSetBlues) - 1; i >= 0; i-- {
		added, err := windowHeap.tryPush(mergeSetBlues[i])
		if err != nil {
			return false, err
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
			return false, err
		}
		// If it's smaller than minimum then we won't be able to add the rest because they're even smaller.
		if !added {
			break
		}
	}

	return false, nil
}
