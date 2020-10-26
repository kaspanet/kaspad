package pastmediantimemanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/pkg/errors"
	"sort"
)

// blueBlockWindow returns a blockWindow of the given size that contains the
// blues in the past of startindNode, sorted by GHOSTDAG order.
// If the number of blues in the past of startingNode is less then windowSize,
// the window will be padded by genesis blocks to achieve a size of windowSize.
func (pmtm *pastMedianTimeManager) blueBlockWindow(startingBlock *externalapi.DomainHash, windowSize uint64) ([]*externalapi.DomainHash, error) {
	window := make([]*externalapi.DomainHash, 0, windowSize)

	currentHash := startingBlock
	currentGHOSTDAGData, err := pmtm.ghostdagDataStore.Get(pmtm.databaseContext, currentHash)
	if err != nil {
		return nil, err
	}

	for uint64(len(window)) < windowSize && currentGHOSTDAGData.SelectedParent != nil {
		for _, blue := range currentGHOSTDAGData.MergeSetBlues {
			window = append(window, blue)
			if uint64(len(window)) == windowSize {
				break
			}
		}

		currentHash = currentGHOSTDAGData.SelectedParent
		currentGHOSTDAGData, err = pmtm.ghostdagDataStore.Get(pmtm.databaseContext, currentHash)
		if err != nil {
			return nil, err
		}
	}

	if uint64(len(window)) < windowSize {
		genesis := currentHash
		for uint64(len(window)) < windowSize {
			window = append(window, genesis)
		}
	}

	return window, nil
}

func (pmtm *pastMedianTimeManager) windowMedianTimestamp(window []*externalapi.DomainHash) (int64, error) {
	if len(window) == 0 {
		return 0, errors.New("Cannot calculate median timestamp for an empty block window")
	}

	timestamps := make([]int64, len(window))
	for i, blockHash := range window {
		block, err := pmtm.blockStore.Block(pmtm.databaseContext, blockHash)
		if err != nil {
			return 0, err
		}

		// TODO: Use headers store
		timestamps[i] = block.Header.TimeInMilliseconds
	}

	sort.Slice(timestamps, func(i, j int) bool {
		return timestamps[i] < timestamps[j]
	})

	return timestamps[len(timestamps)/2], nil
}
