package pastmediantimemanager

import (
	"sort"

	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/pkg/errors"
)

// pastMedianTimeManager provides a method to resolve the
// past median time of a block
type pastMedianTimeManager struct {
	timestampDeviationTolerance int

	databaseContext model.DBReader

	dagTraversalManager model.DAGTraversalManager

	blockHeaderStore  model.BlockHeaderStore
	ghostdagDataStore model.GHOSTDAGDataStore
}

// New instantiates a new PastMedianTimeManager
func New(timestampDeviationTolerance int,
	databaseContext model.DBReader,
	dagTraversalManager model.DAGTraversalManager,
	blockHeaderStore model.BlockHeaderStore,
	ghostdagDataStore model.GHOSTDAGDataStore) model.PastMedianTimeManager {

	return &pastMedianTimeManager{
		timestampDeviationTolerance: timestampDeviationTolerance,
		databaseContext:             databaseContext,

		dagTraversalManager: dagTraversalManager,

		blockHeaderStore:  blockHeaderStore,
		ghostdagDataStore: ghostdagDataStore,
	}
}

// PastMedianTime returns the past median time for some block
func (pmtm *pastMedianTimeManager) PastMedianTime(blockHash *externalapi.DomainHash) (int64, error) {
	window, err := pmtm.dagTraversalManager.BlockWindow(blockHash, 2*pmtm.timestampDeviationTolerance-1)
	if err != nil {
		return 0, err
	}

	return pmtm.windowMedianTimestamp(window)
}

func (pmtm *pastMedianTimeManager) windowMedianTimestamp(window []*externalapi.DomainHash) (int64, error) {
	if len(window) == 0 {
		return 0, errors.New("Cannot calculate median timestamp for an empty block window")
	}

	timestamps := make([]int64, len(window))
	for i, blockHash := range window {
		blockHeader, err := pmtm.blockHeaderStore.BlockHeader(pmtm.databaseContext, blockHash)
		if err != nil {
			return 0, err
		}
		timestamps[i] = blockHeader.TimeInMilliseconds()
	}

	sort.Slice(timestamps, func(i, j int) bool {
		return timestamps[i] < timestamps[j]
	})

	return timestamps[len(timestamps)/2], nil
}
