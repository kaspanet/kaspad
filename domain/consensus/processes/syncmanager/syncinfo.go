package syncmanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

func (sm *syncManager) syncInfo(stagingArea *model.StagingArea) (*externalapi.SyncInfo, error) {
	headerCount := sm.getHeaderCount(stagingArea)
	blockCount := sm.getBlockCount(stagingArea)

	return &externalapi.SyncInfo{
		HeaderCount: headerCount,
		BlockCount:  blockCount,
	}, nil
}

func (sm *syncManager) getHeaderCount(stagingArea *model.StagingArea) uint64 {
	return sm.blockHeaderStore.Count(stagingArea)
}

func (sm *syncManager) getBlockCount(stagingArea *model.StagingArea) uint64 {
	return sm.blockStore.Count(stagingArea)
}
