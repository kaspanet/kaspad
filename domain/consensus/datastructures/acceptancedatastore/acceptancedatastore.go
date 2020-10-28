package acceptancedatastore

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/dbkeys"
)

var bucket = dbkeys.MakeBucket([]byte("acceptance-data"))

// acceptanceDataStore represents a store of AcceptanceData
type acceptanceDataStore struct {
	staging map[externalapi.DomainHash]*model.AcceptanceData
}

// New instantiates a new AcceptanceDataStore
func New() model.AcceptanceDataStore {
	return &acceptanceDataStore{
		staging: make(map[externalapi.DomainHash]*model.AcceptanceData),
	}
}

// Stage stages the given acceptanceData for the given blockHash
func (ads *acceptanceDataStore) Stage(blockHash *externalapi.DomainHash, acceptanceData *model.AcceptanceData) {
	ads.staging[*blockHash] = acceptanceData
}

func (ads *acceptanceDataStore) IsStaged() bool {
	return len(ads.staging) != 0
}

func (ads *acceptanceDataStore) Discard() {
	ads.staging = make(map[externalapi.DomainHash]*model.AcceptanceData)
}

func (ads *acceptanceDataStore) Commit(dbTx model.DBTransaction) error {
	for hash, acceptanceData := range ads.staging {
		err := dbTx.Put(ads.hashAsKey(&hash), ads.serializeAcceptanceData(acceptanceData))
		if err != nil {
			return err
		}
	}

	ads.Discard()
	return nil
}

// Get gets the acceptanceData associated with the given blockHash
func (ads *acceptanceDataStore) Get(dbContext model.DBReader, blockHash *externalapi.DomainHash) (*model.AcceptanceData, error) {
	if header, ok := ads.staging[*blockHash]; ok {
		return header, nil
	}

	headerBytes, err := dbContext.Get(ads.hashAsKey(blockHash))
	if err != nil {
		return nil, err
	}

	return ads.deserializeAcceptanceData(headerBytes)
}

// Delete deletes the acceptanceData associated with the given blockHash
func (ads *acceptanceDataStore) Delete(dbTx model.DBTransaction, blockHash *externalapi.DomainHash) error {
	return nil
}

func (bms *acceptanceDataStore) serializeAcceptanceData(acceptanceData *model.AcceptanceData) []byte {
	panic("implement me")
}

func (bms *acceptanceDataStore) deserializeAcceptanceData(acceptanceDataBytes []byte) (*model.AcceptanceData, error) {
	panic("implement me")
}

func (bms *acceptanceDataStore) hashAsKey(hash *externalapi.DomainHash) model.DBKey {
	return bucket.Key(hash[:])
}
