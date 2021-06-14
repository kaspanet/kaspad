package blockheaderstore

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

type blockHeaderStagingShard struct {
	store    *blockHeaderStore
	toAdd    map[externalapi.DomainHash]externalapi.BlockHeader
	toDelete map[externalapi.DomainHash]struct{}
}

func (bhs *blockHeaderStore) stagingShard(stagingArea *model.StagingArea) *blockHeaderStagingShard {
	return stagingArea.GetOrCreateShard(model.StagingShardIDBlockHeader, func() model.StagingShard {
		return &blockHeaderStagingShard{
			store:    bhs,
			toAdd:    make(map[externalapi.DomainHash]externalapi.BlockHeader),
			toDelete: make(map[externalapi.DomainHash]struct{}),
		}
	}).(*blockHeaderStagingShard)
}

func (bhss *blockHeaderStagingShard) Commit(dbTx model.DBTransaction) error {
	for hash, header := range bhss.toAdd {
		headerBytes, err := bhss.store.serializeHeader(header)
		if err != nil {
			return err
		}
		err = dbTx.Put(bhss.store.hashAsKey(&hash), headerBytes)
		if err != nil {
			return err
		}
		bhss.store.cache.Add(&hash, header)
	}

	for hash := range bhss.toDelete {
		err := dbTx.Delete(bhss.store.hashAsKey(&hash))
		if err != nil {
			return err
		}
		bhss.store.cache.Remove(&hash)
	}

	err := bhss.commitCount(dbTx)
	if err != nil {
		return err
	}

	return nil
}

func (bhss *blockHeaderStagingShard) commitCount(dbTx model.DBTransaction) error {
	count := bhss.store.count(bhss)
	countBytes, err := bhss.store.serializeHeaderCount(count)
	if err != nil {
		return err
	}
	err = dbTx.Put(bhss.store.countKey, countBytes)
	if err != nil {
		return err
	}
	bhss.store.countCached = count
	return nil
}

func (bhss *blockHeaderStagingShard) isStaged() bool {
	return len(bhss.toAdd) != 0 || len(bhss.toDelete) != 0
}
