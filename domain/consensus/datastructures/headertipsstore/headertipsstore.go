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

func (h *headerTipsStore) HasTips(dbContext model.DBReader) (bool, error) {
	if h.staging != nil {
		return len(h.staging) > 0, nil
	}

	return dbContext.Has(headerTipsKey)
}

func (h *headerTipsStore) Discard() {
	h.staging = nil
}

func (h *headerTipsStore) Commit(dbTx model.DBTransaction) error {
	if h.staging == nil {
		return nil
	}

	tipsBytes, err := h.serializeTips(h.staging)
	if err != nil {
		return err
	}

	err = dbTx.Put(headerTipsKey, tipsBytes)
	if err != nil {
		return err
	}

	h.Discard()
	return nil
}

func (h *headerTipsStore) Stage(tips []*externalapi.DomainHash) error {
	clone, err := h.clone(tips)
	if err != nil {
		return err
	}

	h.staging = clone
	return nil
}

func (h *headerTipsStore) IsStaged() bool {
	return h.staging != nil
}

func (h *headerTipsStore) Tips(dbContext model.DBReader) ([]*externalapi.DomainHash, error) {
	if h.staging != nil {
		return h.staging, nil
	}

	tipsBytes, err := dbContext.Get(headerTipsKey)
	if err != nil {
		return nil, err
	}

	return h.deserializeTips(tipsBytes)
}

func (h *headerTipsStore) serializeTips(tips []*externalapi.DomainHash) ([]byte, error) {
	dbTips := serialization.HeaderTipsToDBHeaderTips(tips)
	return proto.Marshal(dbTips)
}

func (h *headerTipsStore) deserializeTips(tipsBytes []byte) ([]*externalapi.DomainHash, error) {
	dbTips := &serialization.DbHeaderTips{}
	err := proto.Unmarshal(tipsBytes, dbTips)
	if err != nil {
		return nil, err
	}

	return serialization.DBHeaderTipsToHeaderTips(dbTips)
}

func (c *headerTipsStore) clone(tips []*externalapi.DomainHash,
) ([]*externalapi.DomainHash, error) {

	serialized, err := c.serializeTips(tips)
	if err != nil {
		return nil, err
	}

	return c.deserializeTips(serialized)
}

// New instantiates a new HeaderTipsStore
func New() model.HeaderTipsStore {
	return &headerTipsStore{}
}
