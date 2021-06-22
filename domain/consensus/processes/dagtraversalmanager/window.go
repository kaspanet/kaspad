package dagtraversalmanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
	"github.com/kaspanet/kaspad/infrastructure/db/database"
	"github.com/pkg/errors"
)

// BlockWindow returns a blockWindow of the given size that contains the
// blocks in the past of highHash, the sorting is unspecified.
// If the number of blocks in the past of startingNode is less then windowSize,
func (dtm *dagTraversalManager) BlockWindow(stagingArea *model.StagingArea,
	highHash *externalapi.DomainHash,
	windowSize int,
	isBlockWithPrefilledData bool) ([]*externalapi.DomainHash, error) {

	window, err := dtm.blockWindow(stagingArea, highHash, windowSize)
	if isBlockWithPrefilledData && database.IsNotFoundError(err) {
		return nil, errors.Wrapf(ruleerrors.ErrBlockWindowMissingBlocks, "some blocks are missing from the block window")
	}

	if err != nil {
		return nil, err
	}

	return window, nil
}

func (dtm *dagTraversalManager) blockWindow(stagingArea *model.StagingArea, highHash *externalapi.DomainHash, windowSize int) ([]*externalapi.DomainHash, error) {
	if highHash.Equal(dtm.genesisHash) {
		return nil, nil
	}

	currentGHOSTDAGData, err := dtm.ghostdagDataStore.Get(dtm.databaseContext, stagingArea, highHash)
	if err != nil {
		return nil, err
	}

	windowHeap := dtm.newSizedUpHeap(stagingArea, windowSize)

	for windowHeap.len() <= windowSize &&
		currentGHOSTDAGData.SelectedParent() != nil &&
		!currentGHOSTDAGData.SelectedParent().Equal(dtm.genesisHash) {

		selectedParentGHOSTDAGData, err := dtm.ghostdagDataStore.Get(
			dtm.databaseContext, stagingArea, currentGHOSTDAGData.SelectedParent())
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
			// TODO: What if merge set blues is not found on isBlockWithPrefilledData because it's not part of the final DAA window?
			// The easiest way to solve it is to probably send the full merge set of each chain block in the DAA window.
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
		currentGHOSTDAGData = selectedParentGHOSTDAGData
	}

	window := make([]*externalapi.DomainHash, 0, len(windowHeap.impl.slice))
	for _, b := range windowHeap.impl.slice {
		window = append(window, b.hash)
	}
	return window, nil
}
