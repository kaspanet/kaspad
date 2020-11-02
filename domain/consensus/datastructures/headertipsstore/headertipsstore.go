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
}

func (h headerTipsStore) Discard() {
	h.staging = nil
}

func (h headerTipsStore) Commit(dbTx model.DBTransaction) error {
	tipsBytes, err := h.serializeTips(h.staging)
	if err != nil {
		return err
	}

	return dbTx.Put(headerTipsKey, tipsBytes)
}

func (h headerTipsStore) Stage(tips []*externalapi.DomainHash) {
	h.staging = tips
}

func (h headerTipsStore) IsStaged() bool {
	return h.staging != nil
}

func (h headerTipsStore) Tips(dbContext model.DBReader) ([]*externalapi.DomainHash, error) {
	if h.staging != nil {
		return h.staging, nil
	}

	tipsBytes, err := dbContext.Get(headerTipsKey)
	if err != nil {
		return nil, err
	}

	return h.deserializeTips(tipsBytes)
}

func (h headerTipsStore) serializeTips(tips []*externalapi.DomainHash) ([]byte, error) {
	dbTips := serialization.HeaderTipsToDBHeaderTips(tips)
	return proto.Marshal(dbTips)
}

func (h headerTipsStore) deserializeTips(tipsBytes []byte) ([]*externalapi.DomainHash, error) {
	dbTips := &serialization.DbHeaderTips{}
	err := proto.Unmarshal(tipsBytes, dbTips)
	if err != nil {
		return nil, err
	}

	return serialization.DBHeaderTipsTOHeaderTips(dbTips)
}

// New instantiates a new HeaderTipsStore
func New() model.HeaderTipsStore {
	return &headerTipsStore{}
}
