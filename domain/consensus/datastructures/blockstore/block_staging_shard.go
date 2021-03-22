package blockstore

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

type blockStagingShard struct {
	store    *blockStore
	toAdd    map[externalapi.DomainHash]*externalapi.DomainBlock
	toDelete map[externalapi.DomainHash]struct{}
}

func (bs *blockStore) stagingShard(stagingArea *model.StagingArea) *blockStagingShard {
	return stagingArea.GetOrCreateShard("BlockStore", func() model.StagingShard {
		return &blockStagingShard{
			store:    bs,
			toAdd:    make(map[externalapi.DomainHash]*externalapi.DomainBlock),
			toDelete: make(map[externalapi.DomainHash]struct{}),
		}
	}).(*blockStagingShard)
}

func (bss blockStagingShard) Commit(dbTx model.DBTransaction) error {
	for hash, block := range bss.toAdd {
		blockBytes, err := bss.store.serializeBlock(block)
		if err != nil {
			return err
		}
		err = dbTx.Put(bss.store.hashAsKey(&hash), blockBytes)
		if err != nil {
			return err
		}
		bss.store.cache.Add(&hash, block)
	}

	for hash := range bss.toDelete {
		err := dbTx.Delete(bss.store.hashAsKey(&hash))
		if err != nil {
			return err
		}
		bss.store.cache.Remove(&hash)
	}

	err := bss.commitCount(dbTx)
	if err != nil {
		return err
	}

	return nil
}

func (bss *blockStagingShard) commitCount(dbTx model.DBTransaction) error {
	count := bss.store.count(bss)
	countBytes, err := bss.store.serializeBlockCount(count)
	if err != nil {
		return err
	}
	err = dbTx.Put(countKey, countBytes)
	if err != nil {
		return err
	}
	bss.store.countCached = count
	return nil
}
