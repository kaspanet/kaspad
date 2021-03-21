package reachabilitydatastore

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

type reachabilityDataStagingShard struct {
	store                   *reachabilityDataStore
	reachabilityData        map[externalapi.DomainHash]model.ReachabilityData
	reachabilityReindexRoot *externalapi.DomainHash
}

func (rds *reachabilityDataStore) stagingShard(stagingArea *model.StagingArea) *reachabilityDataStagingShard {
	return stagingArea.GetOrCreateShard("BlockStore", func() model.StagingShard {
		return &reachabilityDataStagingShard{
			store:                   rds,
			reachabilityData:        make(map[externalapi.DomainHash]model.ReachabilityData),
			reachabilityReindexRoot: nil,
		}
	}).(*reachabilityDataStagingShard)
}

func (rdss reachabilityDataStagingShard) Commit(dbTx model.DBTransaction) error {
	if rdss.reachabilityReindexRoot != nil {
		reachabilityReindexRootBytes, err := rdss.store.serializeReachabilityReindexRoot(rdss.reachabilityReindexRoot)
		if err != nil {
			return err
		}
		err = dbTx.Put(reachabilityReindexRootKey, reachabilityReindexRootBytes)
		if err != nil {
			return err
		}
		rdss.store.reachabilityReindexRootCache = rdss.reachabilityReindexRoot
	}
	for hash, reachabilityData := range rdss.reachabilityData {
		reachabilityDataBytes, err := rdss.store.serializeReachabilityData(reachabilityData)
		if err != nil {
			return err
		}
		err = dbTx.Put(rdss.store.reachabilityDataBlockHashAsKey(&hash), reachabilityDataBytes)
		if err != nil {
			return err
		}
		rdss.store.reachabilityDataCache.Add(&hash, reachabilityData)
	}

	return nil
}
