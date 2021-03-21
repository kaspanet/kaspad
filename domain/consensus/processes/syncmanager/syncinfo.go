package syncmanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

func (sm *syncManager) syncInfo() (*externalapi.SyncInfo, error) {
	headerCount := sm.getHeaderCount()
	blockCount := sm.getBlockCount()

	return &externalapi.SyncInfo{
		HeaderCount: headerCount,
		BlockCount:  blockCount,
	}, nil
}

func (sm *syncManager) getHeaderCount() uint64 {
	return sm.blockHeaderStore.Count(nil)
}

func (sm *syncManager) getBlockCount() uint64 {
	return sm.blockStore.Count()
}
