package multisetstore

import (
	"github.com/golang/protobuf/proto"
	"github.com/kaspanet/kaspad/domain/consensus/database/serialization"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/dbkeys"
)

var bucket = dbkeys.MakeBucket([]byte("multisets"))

// multisetStore represents a store of Multisets
type multisetStore struct {
	staging  map[externalapi.DomainHash]model.Multiset
	toDelete map[externalapi.DomainHash]struct{}
}

// New instantiates a new MultisetStore
func New() model.MultisetStore {
	return &multisetStore{
		staging:  make(map[externalapi.DomainHash]model.Multiset),
		toDelete: make(map[externalapi.DomainHash]struct{}),
	}
}

// Stage stages the given multiset for the given blockHash
func (ms *multisetStore) Stage(blockHash *externalapi.DomainHash, multiset model.Multiset) {
	ms.staging[*blockHash] = multiset
}

func (ms *multisetStore) IsStaged() bool {
	return len(ms.staging) != 0 || len(ms.toDelete) != 0
}

func (ms *multisetStore) Discard() {
	ms.staging = make(map[externalapi.DomainHash]model.Multiset)
	ms.toDelete = make(map[externalapi.DomainHash]struct{})
}

func (ms *multisetStore) Commit(dbTx model.DBTransaction) error {
	for hash, multiset := range ms.staging {
		multisetBytes, err := ms.serializeMultiset(multiset)
		if err != nil {
			return err
		}

		err = dbTx.Put(ms.hashAsKey(&hash), multisetBytes)
		if err != nil {
			return err
		}
	}

	for hash := range ms.toDelete {
		err := dbTx.Delete(ms.hashAsKey(&hash))
		if err != nil {
			return err
		}
	}

	ms.Discard()
	return nil
}

// Get gets the multiset associated with the given blockHash
func (ms *multisetStore) Get(dbContext model.DBReader, blockHash *externalapi.DomainHash) (model.Multiset, error) {
	if multiset, ok := ms.staging[*blockHash]; ok {
		return multiset, nil
	}

	multisetBytes, err := dbContext.Get(ms.hashAsKey(blockHash))
	if err != nil {
		return nil, err
	}

	return ms.deserializeMultiset(multisetBytes)
}

// Delete deletes the multiset associated with the given blockHash
func (ms *multisetStore) Delete(blockHash *externalapi.DomainHash) {
	if _, ok := ms.staging[*blockHash]; ok {
		delete(ms.staging, *blockHash)
		return
	}
	ms.toDelete[*blockHash] = struct{}{}
}

func (ms *multisetStore) hashAsKey(hash *externalapi.DomainHash) model.DBKey {
	return bucket.Key(hash[:])
}

func (ms *multisetStore) serializeMultiset(multiset model.Multiset) ([]byte, error) {
	return proto.Marshal(serialization.MultisetToDBMultiset(multiset))
}

func (ms *multisetStore) deserializeMultiset(multisetBytes []byte) (model.Multiset, error) {
	dbMultiset := &serialization.DbMultiset{}
	err := proto.Unmarshal(multisetBytes, dbMultiset)
	if err != nil {
		return nil, err
	}

	return serialization.DBMultisetToMultiset(dbMultiset)
}
