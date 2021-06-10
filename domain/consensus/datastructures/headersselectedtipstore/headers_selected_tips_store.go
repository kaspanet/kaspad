package headersselectedtipstore

import (
	"github.com/golang/protobuf/proto"
	"github.com/kaspanet/kaspad/domain/consensus/database"
	"github.com/kaspanet/kaspad/domain/consensus/database/serialization"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

var keyName = []byte("headers-selected-tip")

type headerSelectedTipStore struct {
	cache *externalapi.DomainHash
	key   model.DBKey
}

// New instantiates a new HeaderSelectedTipStore
func New(prefix byte) model.HeaderSelectedTipStore {
	return &headerSelectedTipStore{
		key: database.MakeBucket([]byte{prefix}).Key(keyName),
	}
}

func (hsts *headerSelectedTipStore) Has(dbContext model.DBReader, stagingArea *model.StagingArea) (bool, error) {
	stagingShard := hsts.stagingShard(stagingArea)

	if stagingShard.newSelectedTip != nil {
		return true, nil
	}

	if hsts.cache != nil {
		return true, nil
	}

	return dbContext.Has(hsts.key)
}

func (hsts *headerSelectedTipStore) Stage(stagingArea *model.StagingArea, selectedTip *externalapi.DomainHash) {
	stagingShard := hsts.stagingShard(stagingArea)
	stagingShard.newSelectedTip = selectedTip
}

func (hsts *headerSelectedTipStore) IsStaged(stagingArea *model.StagingArea) bool {
	return hsts.stagingShard(stagingArea).isStaged()
}

func (hsts *headerSelectedTipStore) HeadersSelectedTip(dbContext model.DBReader, stagingArea *model.StagingArea) (
	*externalapi.DomainHash, error) {

	stagingShard := hsts.stagingShard(stagingArea)

	if stagingShard.newSelectedTip != nil {
		return stagingShard.newSelectedTip, nil
	}

	if hsts.cache != nil {
		return hsts.cache, nil
	}

	selectedTipBytes, err := dbContext.Get(hsts.key)
	if err != nil {
		return nil, err
	}

	selectedTip, err := hsts.deserializeHeadersSelectedTip(selectedTipBytes)
	if err != nil {
		return nil, err
	}
	hsts.cache = selectedTip
	return hsts.cache, nil
}

func (hsts *headerSelectedTipStore) serializeHeadersSelectedTip(selectedTip *externalapi.DomainHash) ([]byte, error) {
	return proto.Marshal(serialization.DomainHashToDbHash(selectedTip))
}

func (hsts *headerSelectedTipStore) deserializeHeadersSelectedTip(selectedTipBytes []byte) (*externalapi.DomainHash, error) {
	dbHash := &serialization.DbHash{}
	err := proto.Unmarshal(selectedTipBytes, dbHash)
	if err != nil {
		return nil, err
	}

	return serialization.DbHashToDomainHash(dbHash)
}
