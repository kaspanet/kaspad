package dagtraversalmanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// blueBlockWindow returns a blockWindow of the given size that contains the
// blues in the past of startindNode, the sorting is unspecified.
// If the number of blues in the past of startingNode is less then windowSize,
// the window will be padded by genesis blocks to achieve a size of windowSize.
func (dtm *dagTraversalManager) BlueWindow(startingBlock *externalapi.DomainHash, windowSize int) ([]*externalapi.DomainHash, error) {
	currentHash := startingBlock
	currentGHOSTDAGData, err := dtm.ghostdagDataStore.Get(dtm.databaseContext, currentHash)
	if err != nil {
		return nil, err
	}

	windowHeap := dtm.newSizedUpHeap(windowSize)

	for windowHeap.len() <= windowSize && currentGHOSTDAGData.SelectedParent() != nil {
		added, err := windowHeap.tryPush(currentGHOSTDAGData.SelectedParent())
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
		currentHash = currentGHOSTDAGData.SelectedParent()
		currentGHOSTDAGData, err = dtm.ghostdagDataStore.Get(dtm.databaseContext, currentHash)
		if err != nil {
			return nil, err
		}
	}

	window := make([]*externalapi.DomainHash, 0, windowSize)
	for _, b := range windowHeap.impl.slice {
		window = append(window, b.hash)
	}

	if len(window) < windowSize {
		genesis := currentHash
		for len(window) < windowSize {
			window = append(window, genesis)
		}
	}

	return window, nil
}
