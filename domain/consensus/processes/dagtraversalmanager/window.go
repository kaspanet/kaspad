package dagtraversalmanager

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// blueBlockWindow returns a blockWindow of the given size that contains the
// blues in the past of startindNode, sorted by GHOSTDAG order.
// If the number of blues in the past of startingNode is less then windowSize,
// the window will be padded by genesis blocks to achieve a size of windowSize.
func (dtm *dagTraversalManager) BlueWindow(startingBlock *externalapi.DomainHash, windowSize uint64) ([]*externalapi.DomainHash, error) {
	window := make([]*externalapi.DomainHash, 0, windowSize)

	currentHash := startingBlock
	currentGHOSTDAGData, err := dtm.ghostdagDataStore.Get(dtm.databaseContext, currentHash)
	if err != nil {
		return nil, err
	}

	for uint64(len(window)) < windowSize && currentGHOSTDAGData.SelectedParent() != nil {
		for _, blue := range currentGHOSTDAGData.MergeSetBlues() {
			window = append(window, blue)
			if uint64(len(window)) == windowSize {
				break
			}
		}

		currentHash = currentGHOSTDAGData.SelectedParent()
		currentGHOSTDAGData, err = dtm.ghostdagDataStore.Get(dtm.databaseContext, currentHash)
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
