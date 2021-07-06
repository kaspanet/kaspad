package consensusstatestore

import (
	"github.com/golang/protobuf/proto"
	"github.com/kaspanet/kaspad/domain/consensus/database/serialization"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

var tipsKeyName = []byte("tips")

func (css *consensusStateStore) Tips(stagingArea *model.StagingArea, dbContext model.DBReader) ([]*externalapi.DomainHash, error) {
	stagingShard := css.stagingShard(stagingArea)

	if stagingShard.tipsStaging != nil {
		return externalapi.CloneHashes(stagingShard.tipsStaging), nil
	}

	if css.tipsCache != nil {
		return externalapi.CloneHashes(css.tipsCache), nil
	}

	tipsBytes, err := dbContext.Get(css.tipsKey)
	if err != nil {
		return nil, err
	}

	tips, err := css.deserializeTips(tipsBytes)
	if err != nil {
		return nil, err
	}
	css.tipsCache = tips
	return externalapi.CloneHashes(tips), nil
}

func (css *consensusStateStore) StageTips(stagingArea *model.StagingArea, tipHashes []*externalapi.DomainHash) {
	stagingShard := css.stagingShard(stagingArea)

	stagingShard.tipsStaging = externalapi.CloneHashes(tipHashes)
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

func (csss *consensusStateStagingShard) commitTips(dbTx model.DBTransaction) error {
	if csss.tipsStaging == nil {
		return nil
	}

	tipsBytes, err := csss.store.serializeTips(csss.tipsStaging)
	if err != nil {
		return err
	}
	err = dbTx.Put(csss.store.tipsKey, tipsBytes)
	if err != nil {
		return err
	}
	csss.store.tipsCache = csss.tipsStaging

	return nil
}
