package acceptancedatastore

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

type acceptanceDataStagingShard struct {
	acceptanceDataStore *acceptanceDataStore
	toAdd               map[externalapi.DomainHash]externalapi.AcceptanceData
	toDelete            map[externalapi.DomainHash]struct{}
}

func (ads *acceptanceDataStore) stagingShard(stagingArea *model.StagingArea) *acceptanceDataStagingShard {
	return stagingArea.GetOrCreateShard("BlockStore", func() model.StagingShard {
		return &acceptanceDataStagingShard{
			acceptanceDataStore: ads,
			toAdd:               make(map[externalapi.DomainHash]externalapi.AcceptanceData),
			toDelete:            make(map[externalapi.DomainHash]struct{}),
		}
	}).(*acceptanceDataStagingShard)
}

func (adss acceptanceDataStagingShard) Commit(dbTx model.DBTransaction) error {
	for hash, acceptanceData := range adss.toAdd {
		acceptanceDataBytes, err := adss.acceptanceDataStore.serializeAcceptanceData(acceptanceData)
		if err != nil {
			return err
		}
		err = dbTx.Put(adss.acceptanceDataStore.hashAsKey(&hash), acceptanceDataBytes)
		if err != nil {
			return err
		}
		adss.acceptanceDataStore.cache.Add(&hash, acceptanceData)
	}

	for hash := range adss.toDelete {
		err := dbTx.Delete(adss.acceptanceDataStore.hashAsKey(&hash))
		if err != nil {
			return err
		}
		adss.acceptanceDataStore.cache.Remove(&hash)
	}

	return nil
}
