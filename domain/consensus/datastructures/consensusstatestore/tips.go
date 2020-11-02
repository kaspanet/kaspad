package consensusstatestore

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/dbkeys"
	"github.com/kaspanet/kaspad/domain/consensus/utils/hashes"
)

var tipsBucket = dbkeys.MakeBucket([]byte("tips"))
var tipsKey = tipsBucket.Key([]byte("tips"))

func (c consensusStateStore) Tips(dbContext model.DBReader) ([]*externalapi.DomainHash, error) {
	if c.stagedTips != nil {
		return c.stagedTips, nil
	}

	tipsBytes, err := dbContext.Get(tipsKey)
	if err != nil {
		return nil, err
	}

	return hashes.DeserializeHashSlice(tipsBytes)
}

func (c consensusStateStore) StageTips(tipHashes []*externalapi.DomainHash) error {
	c.stagedTips = tipHashes

	return nil
}

func (c consensusStateStore) commitTips(dbTx model.DBTransaction) error {
	tipsBytes := hashes.SerializeHashSlice(c.stagedTips)

	err := dbTx.Put(tipsKey, tipsBytes)
	if err != nil {
		return err
	}

	c.stagedTips = nil
	return nil
}
