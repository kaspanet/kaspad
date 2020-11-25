package consensusstatestore

import (
	"github.com/golang/protobuf/proto"
	"github.com/kaspanet/kaspad/domain/consensus/database/serialization"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/dbkeys"
)

var tipsKey = dbkeys.MakeBucket().Key([]byte("tips"))

func (css *consensusStateStore) Tips(dbContext model.DBReader) ([]*externalapi.DomainHash, error) {
	if css.tipsStaging != nil {
		return css.cloneTips(css.tipsStaging)
	}

	if css.tipsCache != nil {
		return css.cloneTips(css.tipsCache)
	}

	tipsBytes, err := dbContext.Get(tipsKey)
	if err != nil {
		return nil, err
	}

	tips, err := css.deserializeTips(tipsBytes)
	if err != nil {
		return nil, err
	}
	css.tipsCache = tips
	return css.cloneTips(tips)
}

func (css *consensusStateStore) StageTips(tipHashes []*externalapi.DomainHash) error {
	clone, err := css.cloneTips(tipHashes)
	if err != nil {
		return err
	}

	css.tipsStaging = clone
	return nil
}

func (css *consensusStateStore) commitTips(dbTx model.DBTransaction) error {
	if css.tipsStaging == nil {
		return nil
	}

	tipsBytes, err := css.serializeTips(css.tipsStaging)
	if err != nil {
		return err
	}
	err = dbTx.Put(tipsKey, tipsBytes)
	if err != nil {
		return err
	}
	css.tipsCache = css.tipsStaging

	// Note: we don't discard the staging here since that's
	// being done at the end of Commit()
	return nil
}

func (css *consensusStateStore) serializeTips(tips []*externalapi.DomainHash) ([]byte, error) {
	dbTips := serialization.TipsToDBTips(tips)
	return proto.Marshal(dbTips)
}

func (css *consensusStateStore) deserializeTips(tipsBytes []byte) ([]*externalapi.DomainHash,
	error) {

	dbTips := &serialization.DbTips{}
	err := proto.Unmarshal(tipsBytes, dbTips)
	if err != nil {
		return nil, err
	}

	return serialization.DBTipsToTips(dbTips)
}

func (css *consensusStateStore) cloneTips(tips []*externalapi.DomainHash,
) ([]*externalapi.DomainHash, error) {

	serialized, err := css.serializeTips(tips)
	if err != nil {
		return nil, err
	}

	return css.deserializeTips(serialized)
}
