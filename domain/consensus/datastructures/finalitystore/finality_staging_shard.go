package finalitystore

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

type finalityStagingShard struct {
	store *finalityStore
	toAdd map[externalapi.DomainHash]*externalapi.DomainHash
}

func (fs *finalityStore) stagingShard(stagingArea *model.StagingArea) *finalityStagingShard {
	return stagingArea.GetOrCreateShard(fs.shardID, func() model.StagingShard {
		return &finalityStagingShard{
			store: fs,
			toAdd: make(map[externalapi.DomainHash]*externalapi.DomainHash),
		}
	}).(*finalityStagingShard)
}

func (fss *finalityStagingShard) Commit(dbTx model.DBTransaction) error {
	for hash, finalityPointHash := range fss.toAdd {
		err := dbTx.Put(fss.store.hashAsKey(&hash), finalityPointHash.ByteSlice())
		if err != nil {
			return err
		}
		fss.store.cache.Add(&hash, finalityPointHash)
	}

	return nil
}

func (fss *finalityStagingShard) isStaged() bool {
	return len(fss.toAdd) == 0
}
