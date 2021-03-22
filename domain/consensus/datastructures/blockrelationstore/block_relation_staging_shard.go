package blockrelationstore

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

type blockRelationStagingShard struct {
	store *blockRelationStore
	toAdd map[externalapi.DomainHash]*model.BlockRelations
}

func (brs *blockRelationStore) stagingShard(stagingArea *model.StagingArea) *blockRelationStagingShard {
	return stagingArea.GetOrCreateShard("BlockRelationsStore", func() model.StagingShard {
		return &blockRelationStagingShard{
			store: brs,
			toAdd: make(map[externalapi.DomainHash]*model.BlockRelations),
		}
	}).(*blockRelationStagingShard)
}

func (brss blockRelationStagingShard) Commit(dbTx model.DBTransaction) error {
	for hash, blockRelations := range brss.toAdd {
		blockRelationBytes, err := brss.store.serializeBlockRelations(blockRelations)
		if err != nil {
			return err
		}
		err = dbTx.Put(brss.store.hashAsKey(&hash), blockRelationBytes)
		if err != nil {
			return err
		}
		brss.store.cache.Add(&hash, blockRelations)
	}

	return nil
}
