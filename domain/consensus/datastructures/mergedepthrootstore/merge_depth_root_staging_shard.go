package mergedepthrootstore

import (
	"github.com/c4ei/yunseokyeol/domain/consensus/model"
	"github.com/c4ei/yunseokyeol/domain/consensus/model/externalapi"
)

type mergeDepthRootStagingShard struct {
	store *mergeDepthRootStore
	toAdd map[externalapi.DomainHash]*externalapi.DomainHash
}

func (mdrs *mergeDepthRootStore) stagingShard(stagingArea *model.StagingArea) *mergeDepthRootStagingShard {
	return stagingArea.GetOrCreateShard(mdrs.shardID, func() model.StagingShard {
		return &mergeDepthRootStagingShard{
			store: mdrs,
			toAdd: make(map[externalapi.DomainHash]*externalapi.DomainHash),
		}
	}).(*mergeDepthRootStagingShard)
}

func (mdrss *mergeDepthRootStagingShard) Commit(dbTx model.DBTransaction) error {
	for hash, mergeDepthRoot := range mdrss.toAdd {
		err := dbTx.Put(mdrss.store.hashAsKey(&hash), mergeDepthRoot.ByteSlice())
		if err != nil {
			return err
		}
		mdrss.store.cache.Add(&hash, mergeDepthRoot)
	}

	return nil
}

func (mdrss *mergeDepthRootStagingShard) isStaged() bool {
	return len(mdrss.toAdd) == 0
}
