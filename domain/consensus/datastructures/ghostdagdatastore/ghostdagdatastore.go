package ghostdagdatastore

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/dbkeys"
)

var bucket = dbkeys.MakeBucket([]byte("block-ghostdag-data"))

// ghostdagDataStore represents a store of BlockGHOSTDAGData
type ghostdagDataStore struct {
	staging map[externalapi.DomainHash]*model.BlockGHOSTDAGData
}

// New instantiates a new GHOSTDAGDataStore
func New() model.GHOSTDAGDataStore {
	return &ghostdagDataStore{
		staging: make(map[externalapi.DomainHash]*model.BlockGHOSTDAGData),
	}
}

// Stage stages the given blockGHOSTDAGData for the given blockHash
func (gds *ghostdagDataStore) Stage(blockHash *externalapi.DomainHash, blockGHOSTDAGData *model.BlockGHOSTDAGData) {
	gds.staging[*blockHash] = blockGHOSTDAGData
}

func (gds *ghostdagDataStore) IsStaged() bool {
	return len(gds.staging) != 0
}

func (gds *ghostdagDataStore) Discard() {
	gds.staging = make(map[externalapi.DomainHash]*model.BlockGHOSTDAGData)
}

func (gds *ghostdagDataStore) Commit(dbTx model.DBTransaction) error {
	for hash, blockGHOSTDAGData := range gds.staging {
		err := dbTx.Put(gds.hashAsKey(&hash), gds.serializeBlockGHOSTDAGData(blockGHOSTDAGData))
		if err != nil {
			return err
		}
	}

	gds.Discard()
	return nil
}

// Get gets the blockGHOSTDAGData associated with the given blockHash
func (gds *ghostdagDataStore) Get(dbContext model.DBReader, blockHash *externalapi.DomainHash) (*model.BlockGHOSTDAGData, error) {
	if blockGHOSTDAGData, ok := gds.staging[*blockHash]; ok {
		return blockGHOSTDAGData, nil
	}

	blockGHOSTDAGDataBytes, err := dbContext.Get(gds.hashAsKey(blockHash))
	if err != nil {
		return nil, err
	}

	return gds.deserializeBlockGHOSTDAGData(blockGHOSTDAGDataBytes)
}

func (gds *ghostdagDataStore) hashAsKey(hash *externalapi.DomainHash) model.DBKey {
	return bucket.Key(hash[:])
}

func (gds *ghostdagDataStore) serializeBlockGHOSTDAGData(blockGHOSTDAGData *model.BlockGHOSTDAGData) []byte {
	panic("implement me")
}

func (gds *ghostdagDataStore) deserializeBlockGHOSTDAGData(blockGHOSTDAGDataBytes []byte) (*model.BlockGHOSTDAGData, error) {
	panic("implement me")
}
