package acceptancedatastore

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

type acceptanceDataStagingShard struct {
	store    *acceptanceDataStore
	toAdd    map[externalapi.DomainHash]externalapi.AcceptanceData
	toDelete map[externalapi.DomainHash]struct{}
}

func (ads *acceptanceDataStore) stagingShard(stagingArea *model.StagingArea) *acceptanceDataStagingShard {
	return stagingArea.GetOrCreateShard("AcceptanceDataStore", func() model.StagingShard {
		return &acceptanceDataStagingShard{
			store:    ads,
			toAdd:    make(map[externalapi.DomainHash]externalapi.AcceptanceData),
			toDelete: make(map[externalapi.DomainHash]struct{}),
		}
	}).(*acceptanceDataStagingShard)
}

func (adss acceptanceDataStagingShard) Commit(dbTx model.DBTransaction) error {
	for hash, acceptanceData := range adss.toAdd {
		acceptanceDataBytes, err := adss.store.serializeAcceptanceData(acceptanceData)
		if err != nil {
			return err
		}
		err = dbTx.Put(adss.store.hashAsKey(&hash), acceptanceDataBytes)
		if err != nil {
			return err
		}
		adss.store.cache.Add(&hash, acceptanceData)
	}

	for hash := range adss.toDelete {
		err := dbTx.Delete(adss.store.hashAsKey(&hash))
		if err != nil {
			return err
		}
		adss.store.cache.Remove(&hash)
	}

	return nil
}
