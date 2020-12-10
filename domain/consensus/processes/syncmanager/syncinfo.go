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

	pruningPoint, err := sm.pruningStore.PruningPoint(sm.databaseContext)
	if err != nil {
		return false, nil, err
	}

	pruningPointStatus, err := sm.blockStatusStore.Get(sm.databaseContext, pruningPoint)
	if err != nil {
		return false, nil, err
	}

	isAwaitingUTXOSet = pruningPointStatus != externalapi.StatusValid
	if isAwaitingUTXOSet {
		ibdRootUTXOBlockHash = pruningPoint
	}

	return isAwaitingUTXOSet, ibdRootUTXOBlockHash, nil
}

func (sm *syncManager) getHeaderCount() uint64 {
	return sm.blockHeaderStore.Count()
}

func (sm *syncManager) getBlockCount() uint64 {
	return sm.blockStore.Count()
}
