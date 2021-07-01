package daawindowstore

import (
	"github.com/golang/protobuf/proto"
	"github.com/kaspanet/kaspad/domain/consensus/database/serialization"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

type dbKey struct {
	blockHash externalapi.DomainHash
	index     uint64
}

func newDBKey(blockHash *externalapi.DomainHash, index uint64) dbKey {
	return dbKey{
		blockHash: *blockHash,
		index:     index,
	}
}

type daaWindowStagingShard struct {
	store *daaWindowStore
	toAdd map[dbKey]*externalapi.BlockGHOSTDAGDataHashPair
}

func (daaws *daaWindowStore) stagingShard(stagingArea *model.StagingArea) *daaWindowStagingShard {
	return stagingArea.GetOrCreateShard(model.StagingShardIDDAAWindow, func() model.StagingShard {
		return &daaWindowStagingShard{
			store: daaws,
			toAdd: make(map[dbKey]*externalapi.BlockGHOSTDAGDataHashPair),
		}
	}).(*daaWindowStagingShard)
}

func (daawss *daaWindowStagingShard) Commit(dbTx model.DBTransaction) error {
	for key, pair := range daawss.toAdd {
		pairBytes, err := serializePair(pair)
		if err != nil {
			return err
		}

		err = dbTx.Put(daawss.store.key(key), pairBytes)
		if err != nil {
			return err
		}
		daawss.store.cache.Add(&key.blockHash, key.index, pair)
	}

	return nil
}

func serializePair(pair *externalapi.BlockGHOSTDAGDataHashPair) ([]byte, error) {
	return proto.Marshal(serialization.BlockGHOSTDAGDataHashPairToDbBlockGhostdagDataHashPair(pair))
}

func (daawss *daaWindowStagingShard) isStaged() bool {
	return len(daawss.toAdd) == 0
}
