package finalitystore

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

type subsidyStagingShard struct {
	store *subsidyStore
	toAdd map[externalapi.DomainHash]uint64
}

func (fs *subsidyStore) stagingShard(stagingArea *model.StagingArea) *subsidyStagingShard {
	return stagingArea.GetOrCreateShard(fs.shardID, func() model.StagingShard {
		return &subsidyStagingShard{
			store: fs,
			toAdd: make(map[externalapi.DomainHash]uint64),
		}
	}).(*subsidyStagingShard)
}

func (fss *subsidyStagingShard) Commit(dbTx model.DBTransaction) error {
	for hash, subsidy := range fss.toAdd {
		err := dbTx.Put(fss.store.hashAsKey(&hash), fss.store.serializeSubsidy(subsidy))
		if err != nil {
			return err
		}
		fss.store.cache.Add(&hash, subsidy)
	}

	return nil
}

func (fss *subsidyStagingShard) isStaged() bool {
	return len(fss.toAdd) == 0
}
