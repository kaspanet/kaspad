package ghostdagdatastore

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

type ghostdagDataStagingShard struct {
	store *ghostdagDataStore
	toAdd map[externalapi.DomainHash]*externalapi.BlockGHOSTDAGData
}

func (gds *ghostdagDataStore) stagingShard(stagingShardID model.StagingShardID, stagingArea *model.StagingArea) *ghostdagDataStagingShard {
	return stagingArea.GetOrCreateShard(stagingShardID, func() model.StagingShard {
		return &ghostdagDataStagingShard{
			store: gds,
			toAdd: make(map[externalapi.DomainHash]*externalapi.BlockGHOSTDAGData),
		}
	}).(*ghostdagDataStagingShard)
}

func (gdss *ghostdagDataStagingShard) Commit(dbTx model.DBTransaction) error {
	for hash, blockGHOSTDAGData := range gdss.toAdd {
		blockGhostdagDataBytes, err := gdss.store.serializeBlockGHOSTDAGData(blockGHOSTDAGData)
		if err != nil {
			return err
		}
		err = dbTx.Put(gdss.store.hashAsKey(&hash), blockGhostdagDataBytes)
		if err != nil {
			return err
		}
		gdss.store.cache.Add(&hash, blockGHOSTDAGData)
	}

	return nil
}

func (gdss *ghostdagDataStagingShard) isStaged() bool {
	return len(gdss.toAdd) != 0
}
