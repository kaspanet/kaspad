package blockrelationstore

import (
	"github.com/golang/protobuf/proto"
	"github.com/kaspanet/kaspad/domain/consensus/database/serialization"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/dbkeys"
)

var bucket = dbkeys.MakeBucket([]byte("block-relations"))

// blockRelationStore represents a store of BlockRelations
type blockRelationStore struct {
	staging map[externalapi.DomainHash]*model.BlockRelations
}

// New instantiates a new BlockRelationStore
func New() model.BlockRelationStore {
	return &blockRelationStore{
		staging: make(map[externalapi.DomainHash]*model.BlockRelations),
	}
}

func (brs *blockRelationStore) StageBlockRelation(blockHash *externalapi.DomainHash, blockRelations *model.BlockRelations) {
	brs.staging[*blockHash] = blockRelations.Clone()
}

func (brs *blockRelationStore) IsStaged() bool {
	return len(brs.staging) != 0
}

func (brs *blockRelationStore) Discard() {
	brs.staging = make(map[externalapi.DomainHash]*model.BlockRelations)
}

func (brs *blockRelationStore) Commit(dbTx model.DBTransaction) error {
	for hash, blockRelations := range brs.staging {
		blockRelationBytes, err := brs.serializeBlockRelations(blockRelations)
		if err != nil {
			return err
		}
		err = dbTx.Put(brs.hashAsKey(&hash), blockRelationBytes)
		if err != nil {
			return err
		}
	}

	brs.Discard()
	return nil
}

func (brs *blockRelationStore) BlockRelation(dbContext model.DBReader, blockHash *externalapi.DomainHash) (*model.BlockRelations, error) {
	if blockRelations, ok := brs.staging[*blockHash]; ok {
		return blockRelations, nil
	}

	blockRelationsBytes, err := dbContext.Get(brs.hashAsKey(blockHash))
	if err != nil {
		return nil, err
	}

	return brs.deserializeBlockRelations(blockRelationsBytes)
}

func (brs *blockRelationStore) Has(dbContext model.DBReader, blockHash *externalapi.DomainHash) (bool, error) {
	if _, ok := brs.staging[*blockHash]; ok {
		return true, nil
	}

	return dbContext.Has(brs.hashAsKey(blockHash))
}

func (brs *blockRelationStore) hashAsKey(hash *externalapi.DomainHash) model.DBKey {
	return bucket.Key(hash[:])
}

func (brs *blockRelationStore) serializeBlockRelations(blockRelations *model.BlockRelations) ([]byte, error) {
	dbBlockRelations := serialization.DomainBlockRelationsToDbBlockRelations(blockRelations)
	return proto.Marshal(dbBlockRelations)
}

func (brs *blockRelationStore) deserializeBlockRelations(blockRelationsBytes []byte) (*model.BlockRelations, error) {
	dbBlockRelations := &serialization.DbBlockRelations{}
	err := proto.Unmarshal(blockRelationsBytes, dbBlockRelations)
	if err != nil {
		return nil, err
	}
	return serialization.DbBlockRelationsToDomainBlockRelations(dbBlockRelations)
}
