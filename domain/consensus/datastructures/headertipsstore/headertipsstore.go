package headertipsstore

import (
	"github.com/golang/protobuf/proto"
	"github.com/kaspanet/kaspad/domain/consensus/database/serialization"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/dbkeys"
)

var headerTipsKey = dbkeys.MakeBucket().Key([]byte("header-tips"))

type headerTipsStore struct {
	staging []*externalapi.DomainHash
	cache   []*externalapi.DomainHash
}

// New instantiates a new HeaderTipsStore
func New() model.HeaderTipsStore {
	return &headerTipsStore{}
}

func (hts *headerTipsStore) HasTips(dbContext model.DBReader) (bool, error) {
	if len(hts.staging) > 0 {
		return true, nil
	}

	if len(hts.cache) > 0 {
		return true, nil
	}

	return dbContext.Has(headerTipsKey)
}

func (hts *headerTipsStore) Discard() {
	hts.staging = nil
}

func (hts *headerTipsStore) Commit(dbTx model.DBTransaction) error {
	if hts.staging == nil {
		return nil
	}

	tipsBytes, err := hts.serializeTips(hts.staging)
	if err != nil {
		return err
	}
	err = dbTx.Put(headerTipsKey, tipsBytes)
	if err != nil {
		return err
	}
	hts.cache = hts.staging

	hts.Discard()
	return nil
}

func (hts *headerTipsStore) Stage(tips []*externalapi.DomainHash) {
	hts.staging = externalapi.CloneHashes(tips)
}

func (hts *headerTipsStore) IsStaged() bool {
	return hts.staging != nil
}

func (hts *headerTipsStore) Tips(dbContext model.DBReader) ([]*externalapi.DomainHash, error) {
	if hts.staging != nil {
		return externalapi.CloneHashes(hts.staging)
	}

	if hts.cache != nil {
		return externalapi.CloneHashes(hts.cache)
	}

	tipsBytes, err := dbContext.Get(headerTipsKey)
	if err != nil {
		return nil, err
	}

	tips, err := hts.deserializeTips(tipsBytes)
	if err != nil {
		return nil, err
	}
	hts.cache = tips
	return externalapi.CloneHashes(tips)
}

func (hts *headerTipsStore) serializeTips(tips []*externalapi.DomainHash) ([]byte, error) {
	dbTips := serialization.HeaderTipsToDBHeaderTips(tips)
	return proto.Marshal(dbTips)
}

func (hts *headerTipsStore) deserializeTips(tipsBytes []byte) ([]*externalapi.DomainHash, error) {
	dbTips := &serialization.DbHeaderTips{}
	err := proto.Unmarshal(tipsBytes, dbTips)
	if err != nil {
		return nil, err
	}

	return serialization.DBHeaderTipsToHeaderTips(dbTips)
}
