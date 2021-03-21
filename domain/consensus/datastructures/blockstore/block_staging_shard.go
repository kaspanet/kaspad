package blockstore

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

type blockStagingShard struct {
	blockStore     *blockStore
	blocksToAdd    map[externalapi.DomainHash]*externalapi.DomainBlock
	blocksToDelete map[externalapi.DomainHash]struct{}
}

func (bs *blockStore) stagingShard(stagingArea model.StagingArea) *blockStagingShard {
	return stagingArea.GetOrCreateShard("BlockStore", func() model.StagingShard {
		return &blockStagingShard{
			blockStore:     bs,
			blocksToAdd:    make(map[externalapi.DomainHash]*externalapi.DomainBlock),
			blocksToDelete: make(map[externalapi.DomainHash]struct{}),
		}
	}).(*blockStagingShard)
}

func (bss blockStagingShard) Commit(dbTx model.DBTransaction) error {
	for hash, block := range bss.blocksToAdd {
		blockBytes, err := bss.blockStore.serializeBlock(block)
		if err != nil {
			return err
		}
		err = dbTx.Put(bss.blockStore.hashAsKey(&hash), blockBytes)
		if err != nil {
			return err
		}
		bss.blockStore.cache.Add(&hash, block)
	}

	for hash := range bss.blocksToDelete {
		err := dbTx.Delete(bss.blockStore.hashAsKey(&hash))
		if err != nil {
			return err
		}
		bss.blockStore.cache.Remove(&hash)
	}

	err := bss.commitCount(dbTx)
	if err != nil {
		return err
	}

	return nil
}

func (bss *blockStagingShard) commitCount(dbTx model.DBTransaction) error {
	count := bss.blockStore.count(bss)
	countBytes, err := bss.blockStore.serializeBlockCount(count)
	if err != nil {
		return err
	}
	err = dbTx.Put(countKey, countBytes)
	if err != nil {
		return err
	}
	bss.blockStore.countCached = count
	return nil
}
