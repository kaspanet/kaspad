package pastmediantimemanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// pastMedianTimeManager provides a method to resolve the
// past median time of a block
type pastMedianTimeManager struct {
	timestampDeviationTolerance uint64

	databaseContext model.DBContextProxy

	ghostdagManager model.GHOSTDAGManager

	ghostdagDataStore model.GHOSTDAGDataStore
	blockStore        model.BlockStore
}

// New instantiates a new PastMedianTimeManager
func New(timestampDeviationTolerance uint64,
	databaseContext model.DBContextProxy,
	ghostdagManager model.GHOSTDAGManager,
	ghostdagDataStore model.GHOSTDAGDataStore,
	blockStore model.BlockStore) model.PastMedianTimeManager {
	return &pastMedianTimeManager{
		timestampDeviationTolerance: timestampDeviationTolerance,
		databaseContext:             databaseContext,
		ghostdagManager:             ghostdagManager,
		ghostdagDataStore:           ghostdagDataStore,
		blockStore:                  blockStore,
	}
}

// PastMedianTime returns the past median time for some block
func (pmtm *pastMedianTimeManager) PastMedianTime(blockHash *externalapi.DomainHash) (int64, error) {
	window, err := pmtm.blueBlockWindow(blockHash, 2*pmtm.timestampDeviationTolerance-1)
	if err != nil {
		return 0, err
	}

	return pmtm.windowMedianTimestamp(window)
}
