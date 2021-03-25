package daablocksstore

import (
	"github.com/kaspanet/kaspad/domain/consensus/database/binaryserialization"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

type daaBlocksStagingShard struct {
	store                  *daaBlocksStore
	daaScoreToAdd          map[externalapi.DomainHash]uint64
	daaAddedBlocksToAdd    map[externalapi.DomainHash][]*externalapi.DomainHash
	daaScoreToDelete       map[externalapi.DomainHash]struct{}
	daaAddedBlocksToDelete map[externalapi.DomainHash]struct{}
}

func (daas *daaBlocksStore) stagingShard(stagingArea *model.StagingArea) *daaBlocksStagingShard {
	return stagingArea.GetOrCreateShard(model.StagingShardIDDAABlocks, func() model.StagingShard {
		return &daaBlocksStagingShard{
			store:                  daas,
			daaScoreToAdd:          make(map[externalapi.DomainHash]uint64),
			daaAddedBlocksToAdd:    make(map[externalapi.DomainHash][]*externalapi.DomainHash),
			daaScoreToDelete:       make(map[externalapi.DomainHash]struct{}),
			daaAddedBlocksToDelete: make(map[externalapi.DomainHash]struct{}),
		}
	}).(*daaBlocksStagingShard)
}

func (daass daaBlocksStagingShard) Commit(dbTx model.DBTransaction) error {
	for hash, daaScore := range daass.daaScoreToAdd {
		daaScoreBytes := binaryserialization.SerializeUint64(daaScore)
		err := dbTx.Put(daass.store.daaScoreHashAsKey(&hash), daaScoreBytes)
		if err != nil {
			return err
		}
		daass.store.daaScoreLRUCache.Add(&hash, daaScore)
	}

	for hash, addedBlocks := range daass.daaAddedBlocksToAdd {
		addedBlocksBytes := binaryserialization.SerializeHashes(addedBlocks)
		err := dbTx.Put(daass.store.daaAddedBlocksHashAsKey(&hash), addedBlocksBytes)
		if err != nil {
			return err
		}
		daass.store.daaAddedBlocksLRUCache.Add(&hash, addedBlocks)
	}

	for hash := range daass.daaScoreToDelete {
		err := dbTx.Delete(daass.store.daaScoreHashAsKey(&hash))
		if err != nil {
			return err
		}
		daass.store.daaScoreLRUCache.Remove(&hash)
	}

	for hash := range daass.daaAddedBlocksToDelete {
		err := dbTx.Delete(daass.store.daaAddedBlocksHashAsKey(&hash))
		if err != nil {
			return err
		}
		daass.store.daaAddedBlocksLRUCache.Remove(&hash)
	}

	return nil
}

func (daass daaBlocksStagingShard) isStaged() bool {
	return len(daass.daaScoreToAdd) != 0 ||
		len(daass.daaAddedBlocksToAdd) != 0 ||
		len(daass.daaScoreToDelete) != 0 ||
		len(daass.daaAddedBlocksToDelete) != 0
}
