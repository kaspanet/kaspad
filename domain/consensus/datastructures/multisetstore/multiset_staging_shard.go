package multisetstore

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

type multisetStagingShard struct {
	store    *multisetStore
	toAdd    map[externalapi.DomainHash]model.Multiset
	toDelete map[externalapi.DomainHash]struct{}
}

func (ms *multisetStore) stagingShard(stagingArea *model.StagingArea) *multisetStagingShard {
	return stagingArea.GetOrCreateShard(model.StagingShardIDMultiset, func() model.StagingShard {
		return &multisetStagingShard{
			store:    ms,
			toAdd:    make(map[externalapi.DomainHash]model.Multiset),
			toDelete: make(map[externalapi.DomainHash]struct{}),
		}
	}).(*multisetStagingShard)
}

func (mss *multisetStagingShard) Commit(dbTx model.DBTransaction) error {
	for hash, multiset := range mss.toAdd {
		multisetBytes, err := mss.store.serializeMultiset(multiset)
		if err != nil {
			return err
		}
		err = dbTx.Put(mss.store.hashAsKey(&hash), multisetBytes)
		if err != nil {
			return err
		}
		mss.store.cache.Add(&hash, multiset)
	}

	for hash := range mss.toDelete {
		err := dbTx.Delete(mss.store.hashAsKey(&hash))
		if err != nil {
			return err
		}
		mss.store.cache.Remove(&hash)
	}

	return nil
}

func (mss *multisetStagingShard) isStaged() bool {
	return len(mss.toAdd) != 0 || len(mss.toDelete) != 0
}
