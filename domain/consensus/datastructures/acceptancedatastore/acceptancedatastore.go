package acceptancedatastore

import (
	"github.com/kaspanet/kaspad/domain/consensus/database/serialization"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/dbkeys"
	"google.golang.org/protobuf/proto"
)

var bucket = dbkeys.MakeBucket([]byte("acceptance-data"))

// acceptanceDataStore represents a store of AcceptanceData
type acceptanceDataStore struct {
	staging map[externalapi.DomainHash]model.AcceptanceData
}

// New instantiates a new AcceptanceDataStore
func New() model.AcceptanceDataStore {
	return &acceptanceDataStore{
		staging: make(map[externalapi.DomainHash]model.AcceptanceData),
	}
}

// Stage stages the given acceptanceData for the given blockHash
func (ads *acceptanceDataStore) Stage(blockHash *externalapi.DomainHash, acceptanceData model.AcceptanceData) {
	ads.staging[*blockHash] = acceptanceData
}

func (ads *acceptanceDataStore) IsStaged() bool {
	return len(ads.staging) != 0
}

func (ads *acceptanceDataStore) Discard() {
	ads.staging = make(map[externalapi.DomainHash]model.AcceptanceData)
}

func (ads *acceptanceDataStore) Commit(dbTx model.DBTransaction) error {
	for hash, acceptanceData := range ads.staging {
		acceptanceDataBytes, err := ads.serializeAcceptanceData(acceptanceData)
		if err != nil {
			return err
		}
		err = dbTx.Put(ads.hashAsKey(&hash), acceptanceDataBytes)
		if err != nil {
			return err
		}
	}

	ads.Discard()
	return nil
}

// Get gets the acceptanceData associated with the given blockHash
func (ads *acceptanceDataStore) Get(dbContext model.DBReader, blockHash *externalapi.DomainHash) (model.AcceptanceData, error) {
	if acceptanceData, ok := ads.staging[*blockHash]; ok {
		return acceptanceData, nil
	}

	acceptanceDataBytes, err := dbContext.Get(ads.hashAsKey(blockHash))
	if err != nil {
		return nil, err
	}

	return ads.deserializeAcceptanceData(acceptanceDataBytes)
}

// Delete deletes the acceptanceData associated with the given blockHash
func (ads *acceptanceDataStore) Delete(dbTx model.DBTransaction, blockHash *externalapi.DomainHash) error {
	if _, ok := ads.staging[*blockHash]; ok {
		delete(ads.staging, *blockHash)
		return nil
	}
	return dbTx.Delete(ads.hashAsKey(blockHash))
}

func (ads *acceptanceDataStore) serializeAcceptanceData(acceptanceData model.AcceptanceData) ([]byte, error) {
	dbAcceptanceData := serialization.DomainAcceptanceDataToDbAcceptanceData(acceptanceData)
	return proto.Marshal(dbAcceptanceData)
}

func (ads *acceptanceDataStore) deserializeAcceptanceData(acceptanceDataBytes []byte) (model.AcceptanceData, error) {
	dbAcceptanceData := &serialization.DbAcceptanceData{}
	err := proto.Unmarshal(acceptanceDataBytes, dbAcceptanceData)
	if err != nil {
		return nil, err
	}
	return serialization.DbAcceptanceDataToDomainAcceptanceData(dbAcceptanceData)
}

func (ads *acceptanceDataStore) hashAsKey(hash *externalapi.DomainHash) model.DBKey {
	return bucket.Key(hash[:])
}
