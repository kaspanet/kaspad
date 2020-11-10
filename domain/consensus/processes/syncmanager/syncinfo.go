package syncmanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/util/mstime"
)

// areHeaderTipsSyncedMaxTimeDifference is the number of blocks from
// the header virtual selected parent (estimated by timestamps) for
// kaspad to be considered not synced
const areHeaderTipsSyncedMaxTimeDifference = 300

func (sm *syncManager) syncInfo() (*externalapi.SyncInfo, error) {
	syncState, err := sm.resolveSyncState()
	if err != nil {
		return nil, err
	}

	var ibdRootUTXOBlockHash *externalapi.DomainHash
	if syncState == externalapi.SyncStateMissingUTXOSet {
		ibdRootUTXOBlockHash, err = sm.consensusStateManager.HeaderTipsPruningPoint()
		if err != nil {
			return nil, err
		}
	}

	return &externalapi.SyncInfo{
		State:                syncState,
		IBDRootUTXOBlockHash: ibdRootUTXOBlockHash,
	}, nil
}

func (sm *syncManager) resolveSyncState() (externalapi.SyncState, error) {
	hasTips, err := sm.headerTipsStore.HasTips(sm.databaseContext)
	if err != nil {
		return 0, err
	}

	if !hasTips {
		return externalapi.SyncStateMissingGenesis, nil
	}

	headerVirtualSelectedParentHash, err := sm.headerVirtualSelectedParentHash()
	if err != nil {
		return 0, err
	}

	isSynced, err := sm.areHeaderTipsSynced(headerVirtualSelectedParentHash)
	if err != nil {
		return 0, err
	}
	if !isSynced {
		return externalapi.SyncStateHeadersFirst, nil
	}

	headerVirtualSelectedParentBlockStatus, err := sm.blockStatusStore.Get(sm.databaseContext, headerVirtualSelectedParentHash)
	if err != nil {
		return 0, err
	}
	if headerVirtualSelectedParentBlockStatus != externalapi.StatusValid {
		return externalapi.SyncStateMissingUTXOSet, nil
	}

	virtualSelectedParentHash, err := sm.virtualSelectedParentHash()
	if err != nil {
		return 0, err
	}
	if *virtualSelectedParentHash != *headerVirtualSelectedParentHash {
		return externalapi.SyncStateMissingBlockBodies, nil
	}

	return externalapi.SyncStateRelay, nil
}

func (sm *syncManager) virtualSelectedParentHash() (*externalapi.DomainHash, error) {
	virtualGHOSTDAGData, err := sm.ghostdagDataStore.Get(sm.databaseContext, model.VirtualBlockHash)
	if err != nil {
		return nil, err
	}
	return virtualGHOSTDAGData.SelectedParent, nil
}

func (sm *syncManager) headerVirtualSelectedParentHash() (*externalapi.DomainHash, error) {
	headerTips, err := sm.headerTipsStore.Tips(sm.databaseContext)
	if err != nil {
		return nil, err
	}
	return sm.ghostdagManager.ChooseSelectedParent(headerTips...)
}

func (sm *syncManager) areHeaderTipsSynced(headerVirtualSelectedParentHash *externalapi.DomainHash) (bool, error) {
	virtualSelectedParentHeader, err := sm.blockHeaderStore.BlockHeader(sm.databaseContext, headerVirtualSelectedParentHash)
	if err != nil {
		return false, err
	}
	virtualSelectedParentTimeInMilliseconds := virtualSelectedParentHeader.TimeInMilliseconds

	nowInMilliseconds := mstime.Now().UnixMilliseconds()
	timeDifference := nowInMilliseconds - virtualSelectedParentTimeInMilliseconds

	maxTimeDifference := areHeaderTipsSyncedMaxTimeDifference * sm.targetTimePerBlock

	return timeDifference <= maxTimeDifference, nil
}
