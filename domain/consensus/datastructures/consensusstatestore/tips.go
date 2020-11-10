package consensusstatestore

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/dbkeys"
	"github.com/kaspanet/kaspad/domain/consensus/utils/hashes"
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

	return hashes.DeserializeHashSlice(tipsBytes)
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
	tipsBytes := hashes.SerializeHashSlice(c.stagedTips)

	err := dbTx.Put(tipsKey, tipsBytes)
	if err != nil {
		return err
	}

	return nil
}

func (c *consensusStateStore) serializeTips(tips []*externalapi.DomainHash) ([]byte, error) {
	panic("unimplemented")
}

func (c *consensusStateStore) deserializeTips(tipsBytes []byte) ([]*externalapi.DomainHash,
	error) {

	panic("unimplemented")
}

func (c *consensusStateStore) cloneTips(tips []*externalapi.DomainHash,
) ([]*externalapi.DomainHash, error) {

	serialized, err := c.serializeTips(tips)
	if err != nil {
		return nil, err
	}

	return c.deserializeTips(serialized)
}
