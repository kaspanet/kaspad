package blockstatusstore

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

type blockStatusStagingShard struct {
	store *blockStatusStore
	toAdd map[externalapi.DomainHash]externalapi.BlockStatus
}

func (bss *blockStatusStore) stagingShard(stagingArea *model.StagingArea) *blockStatusStagingShard {
	return stagingArea.GetOrCreateShard(bss.shardID, func() model.StagingShard {
		return &blockStatusStagingShard{
			store: bss,
			toAdd: make(map[externalapi.DomainHash]externalapi.BlockStatus),
		}
	}).(*blockStatusStagingShard)
}

func (bsss *blockStatusStagingShard) Commit(dbTx model.DBTransaction) error {
	for hash, status := range bsss.toAdd {
		blockStatusBytes, err := bsss.store.serializeBlockStatus(status)
		if err != nil {
			return err
		}
		err = dbTx.Put(bsss.store.hashAsKey(&hash), blockStatusBytes)
		if err != nil {
			return err
		}
		bsss.store.cache.Add(&hash, status)
	}

	return nil
}

func (bsss *blockStatusStagingShard) isStaged() bool {
	return len(bsss.toAdd) != 0
}
