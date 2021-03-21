package pastmediantimemanager

import (
	"sort"

	"github.com/kaspanet/kaspad/domain/consensus/utils/sorters"

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

	genesisHash *externalapi.DomainHash
}

// New instantiates a new PastMedianTimeManager
func New(timestampDeviationTolerance int,
	databaseContext model.DBReader,
	dagTraversalManager model.DAGTraversalManager,
	blockHeaderStore model.BlockHeaderStore,
	ghostdagDataStore model.GHOSTDAGDataStore,
	genesisHash *externalapi.DomainHash) model.PastMedianTimeManager {

	return &pastMedianTimeManager{
		timestampDeviationTolerance: timestampDeviationTolerance,
		databaseContext:             databaseContext,

		dagTraversalManager: dagTraversalManager,

		blockHeaderStore:  blockHeaderStore,
		ghostdagDataStore: ghostdagDataStore,
		genesisHash:       genesisHash,
	}
}

// PastMedianTime returns the past median time for some block
func (pmtm *pastMedianTimeManager) PastMedianTime(blockHash *externalapi.DomainHash) (int64, error) {
	window, err := pmtm.dagTraversalManager.BlockWindow(blockHash, 2*pmtm.timestampDeviationTolerance-1)
	if err != nil {
		return 0, err
	}
	if len(window) == 0 {
		header, err := pmtm.blockHeaderStore.BlockHeader(pmtm.databaseContext, nil, pmtm.genesisHash)
		if err != nil {
			return 0, err
		}
		return header.TimeInMilliseconds(), nil
	}

	return pmtm.windowMedianTimestamp(window)
}

func (pmtm *pastMedianTimeManager) windowMedianTimestamp(window []*externalapi.DomainHash) (int64, error) {
	if len(window) == 0 {
		return 0, errors.New("Cannot calculate median timestamp for an empty block window")
	}

	timestamps := make([]int64, len(window))
	for i, blockHash := range window {
		blockHeader, err := pmtm.blockHeaderStore.BlockHeader(pmtm.databaseContext, nil, blockHash)
		if err != nil {
			return 0, err
		}
		timestamps[i] = blockHeader.TimeInMilliseconds()
	}

	sort.Sort(sorters.Int64Slice(timestamps))

	return timestamps[len(timestamps)/2], nil
}
