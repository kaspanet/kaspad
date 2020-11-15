package consensusstatestore

import (
	"github.com/golang/protobuf/proto"
	"github.com/kaspanet/kaspad/domain/consensus/database/serialization"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/dbkeys"
)

var tipsKey = dbkeys.MakeBucket().Key([]byte("tips"))

func (c *consensusStateStore) Tips(dbContext model.DBReader) ([]*externalapi.DomainHash, error) {
	if c.stagedTips != nil {
		return c.stagedTips, nil
	}

	tipsBytes, err := dbContext.Get(tipsKey)
	if err != nil {
		return nil, err
	}

	return c.deserializeTips(tipsBytes)
}

func (c *consensusStateStore) StageTips(tipHashes []*externalapi.DomainHash) error {
	clone, err := c.cloneTips(tipHashes)
	if err != nil {
		return err
	}

	c.stagedTips = clone
	return nil
}

func (c *consensusStateStore) commitTips(dbTx model.DBTransaction) error {
	if c.stagedTips == nil {
		return nil
	}

	tipsBytes, err := c.serializeTips(c.stagedTips)
	if err != nil {
		return err
	}

	err = dbTx.Put(tipsKey, tipsBytes)
	if err != nil {
		return err
	}

	return nil
}

func (c *consensusStateStore) serializeTips(tips []*externalapi.DomainHash) ([]byte, error) {
	dbTips := serialization.TipsToDBTips(tips)
	return proto.Marshal(dbTips)
}

func (c *consensusStateStore) deserializeTips(tipsBytes []byte) ([]*externalapi.DomainHash,
	error) {

	dbTips := &serialization.DbTips{}
	err := proto.Unmarshal(tipsBytes, dbTips)
	if err != nil {
		return nil, err
	}

	return serialization.DBTipsToTips(dbTips)
}

func (c *consensusStateStore) cloneTips(tips []*externalapi.DomainHash,
) ([]*externalapi.DomainHash, error) {

	serialized, err := c.serializeTips(tips)
	if err != nil {
		return nil, err
	}

	return c.deserializeTips(serialized)
}
