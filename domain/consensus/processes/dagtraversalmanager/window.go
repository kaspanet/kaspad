package dagtraversalmanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
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

func (dtm *dagTraversalManager) blockWindowHeap(stagingArea *model.StagingArea,
	highHash *externalapi.DomainHash, windowSize int) (*sizedUpBlockHeap, error) {
	windowHeapSlice, err := dtm.windowHeapSliceStore.Get(stagingArea, highHash, windowSize)
	sliceNotCached := database.IsNotFoundError(err)
	if !sliceNotCached && err != nil {
		return nil, err
	}
	if !sliceNotCached {
		return dtm.newSizedUpHeapFromSlice(stagingArea, windowHeapSlice), nil
	}

	heap, err := dtm.calculateBlockWindowHeap(stagingArea, highHash, windowSize)
	if err != nil {
		return nil, err
	}

	if !highHash.Equal(model.VirtualBlockHash) {
		dtm.windowHeapSliceStore.Stage(stagingArea, highHash, windowSize, heap.impl.slice)
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

	// If the block has a trusted DAA window attached, we just take it as is and don't use cache of selected parent to
	// build the window. This is because tryPushMergeSet might not be able to find all the GHOSTDAG data that is
	// associated with the block merge set.
	_, err = dtm.daaWindowStore.DAAWindowBlock(dtm.databaseContext, stagingArea, current, 0)
	isNonTrustedBlock := database.IsNotFoundError(err)
	if !isNonTrustedBlock && err != nil {
		return nil, err
	}

	if isNonTrustedBlock && currentGHOSTDAGData.SelectedParent() != nil {
		windowHeapSlice, err := dtm.windowHeapSliceStore.Get(stagingArea, currentGHOSTDAGData.SelectedParent(), windowSize)
		selectedParentNotCached := database.IsNotFoundError(err)
		if !selectedParentNotCached && err != nil {
			return nil, err
		}
		if !selectedParentNotCached {
			windowHeap := dtm.newSizedUpHeapFromSlice(stagingArea, windowHeapSlice)
			if !currentGHOSTDAGData.SelectedParent().Equal(dtm.genesisHash) {
				selectedParentGHOSTDAGData, err := dtm.ghostdagDataStore.Get(
					dtm.databaseContext, stagingArea, currentGHOSTDAGData.SelectedParent(), false)
				if err != nil {
					return nil, err
				}

				_, err = dtm.tryPushMergeSet(windowHeap, currentGHOSTDAGData, selectedParentGHOSTDAGData)
				if err != nil {
					return nil, err
				}
			}

			return windowHeap, nil
		}
	}

	// Walk down the chain until you finish or find a trusted block and then take complete the rest
	// of the window with the trusted window.
	for {
		if currentGHOSTDAGData.SelectedParent().Equal(dtm.genesisHash) {
			break
		}

		_, err := dtm.daaWindowStore.DAAWindowBlock(dtm.databaseContext, stagingArea, current, 0)
		currentIsNonTrustedBlock := database.IsNotFoundError(err)
		if !currentIsNonTrustedBlock && err != nil {
			return nil, err
		}

		if !currentIsNonTrustedBlock {
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
