package syncmanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

func (sm *syncManager) syncInfo() (*externalapi.SyncInfo, error) {
	isAwaitingUTXOSet, ibdRootUTXOBlockHash, err := sm.isAwaitingUTXOSet()
	if err != nil {
		return nil, err
	}

	headerCount := sm.getHeaderCount()
	blockCount := sm.getBlockCount()

	return &externalapi.SyncInfo{
		IsAwaitingUTXOSet:    isAwaitingUTXOSet,
		IBDRootUTXOBlockHash: ibdRootUTXOBlockHash,
		HeaderCount:          headerCount,
		BlockCount:           blockCount,
	}, nil
}

func (sm *syncManager) isAwaitingUTXOSet() (isAwaitingUTXOSet bool, ibdRootUTXOBlockHash *externalapi.DomainHash,
	err error) {

	pruningPointByHeaders, err := sm.pruningManager.CalculatePruningPointByHeaderSelectedTip()
	if err != nil {
		return false, nil, err
	}

	pruningPoint, err := sm.pruningStore.PruningPoint(sm.databaseContext)
	if err != nil {
		return false, nil, err
	}

	// If the pruning point by headers is different from the current point
	// it means we need to request the new pruning point UTXO set.
	if !pruningPoint.Equal(pruningPointByHeaders) {
		return true, pruningPointByHeaders, nil
	}

	return false, nil, nil
}

func (sm *syncManager) getHeaderCount() uint64 {
	return sm.blockHeaderStore.Count()
}

func (sm *syncManager) getBlockCount() uint64 {
	return sm.blockStore.Count()
}
