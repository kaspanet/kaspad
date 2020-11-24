package utxodiffstore

import (
	"github.com/golang/protobuf/proto"
	"github.com/kaspanet/golang-lru/simplelru"
	"github.com/kaspanet/kaspad/domain/consensus/database/serialization"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/dbkeys"
	"github.com/pkg/errors"
)

var utxoDiffBucket = dbkeys.MakeBucket([]byte("utxo-diffs"))
var utxoDiffChildBucket = dbkeys.MakeBucket([]byte("utxo-diff-children"))

// utxoDiffStore represents a store of UTXODiffs
type utxoDiffStore struct {
	utxoDiffStaging      map[externalapi.DomainHash]*model.UTXODiff
	utxoDiffChildStaging map[externalapi.DomainHash]*externalapi.DomainHash
	toDelete             map[externalapi.DomainHash]struct{}
	cache                simplelru.LRUCache
}

// New instantiates a new UTXODiffStore
func New(cacheSize int) (model.UTXODiffStore, error) {
	utxoDiffStore := &utxoDiffStore{
		utxoDiffStaging:      make(map[externalapi.DomainHash]*model.UTXODiff),
		utxoDiffChildStaging: make(map[externalapi.DomainHash]*externalapi.DomainHash),
		toDelete:             make(map[externalapi.DomainHash]struct{}),
	}

	cache, err := simplelru.NewLRU(cacheSize, nil)
	if err != nil {
		return nil, err
	}
	utxoDiffStore.cache = cache

	return utxoDiffStore, nil
}

// Stage stages the given utxoDiff for the given blockHash
func (uds *utxoDiffStore) Stage(blockHash *externalapi.DomainHash, utxoDiff *model.UTXODiff, utxoDiffChild *externalapi.DomainHash) error {
	utxoDiffClone, err := uds.cloneUTXODiff(utxoDiff)
	if err != nil {
		return err
	}
	uds.utxoDiffStaging[*blockHash] = utxoDiffClone

	if utxoDiffChild != nil {
		utxoDiffChildClone := uds.cloneUTXODiffChild(utxoDiffChild)
		uds.utxoDiffChildStaging[*blockHash] = utxoDiffChildClone
	}
	return nil
}

func (uds *utxoDiffStore) IsStaged() bool {
	return len(uds.utxoDiffStaging) != 0 || len(uds.utxoDiffChildStaging) != 0 || len(uds.toDelete) != 0
}

func (uds *utxoDiffStore) IsBlockHashStaged(blockHash *externalapi.DomainHash) bool {
	if _, ok := uds.utxoDiffStaging[*blockHash]; ok {
		return true
	}
	_, ok := uds.utxoDiffChildStaging[*blockHash]
	return ok
}

func (uds *utxoDiffStore) Discard() {
	uds.utxoDiffStaging = make(map[externalapi.DomainHash]*model.UTXODiff)
	uds.utxoDiffChildStaging = make(map[externalapi.DomainHash]*externalapi.DomainHash)
	uds.toDelete = make(map[externalapi.DomainHash]struct{})
}

func (uds *utxoDiffStore) Commit(dbTx model.DBTransaction) error {
	for hash, utxoDiff := range uds.utxoDiffStaging {
		utxoDiffBytes, err := uds.serializeUTXODiff(utxoDiff)
		if err != nil {
			return err
		}

		err = dbTx.Put(uds.utxoDiffHashAsKey(&hash), utxoDiffBytes)
		if err != nil {
			return err
		}
	}
	for hash, utxoDiffChild := range uds.utxoDiffChildStaging {
		if utxoDiffChild == nil {
			continue
		}

		utxoDiffChildBytes, err := uds.serializeUTXODiffChild(utxoDiffChild)
		if err != nil {
			return err
		}
		err = dbTx.Put(uds.utxoDiffChildHashAsKey(&hash), utxoDiffChildBytes)
		if err != nil {
			return err
		}
	}

	for hash := range uds.toDelete {
		err := dbTx.Delete(uds.utxoDiffHashAsKey(&hash))
		if err != nil {
			return err
		}

		err = dbTx.Delete(uds.utxoDiffChildHashAsKey(&hash))
		if err != nil {
			return err
		}
	}

	uds.Discard()
	return nil
}

// UTXODiff gets the utxoDiff associated with the given blockHash
func (uds *utxoDiffStore) UTXODiff(dbContext model.DBReader, blockHash *externalapi.DomainHash) (*model.UTXODiff, error) {
	if utxoDiff, ok := uds.utxoDiffStaging[*blockHash]; ok {
		return utxoDiff, nil
	}

	utxoDiffBytes, err := dbContext.Get(uds.utxoDiffHashAsKey(blockHash))
	if err != nil {
		return nil, err
	}

	return uds.deserializeUTXODiff(utxoDiffBytes)
}

// UTXODiffChild gets the utxoDiff child associated with the given blockHash
func (uds *utxoDiffStore) UTXODiffChild(dbContext model.DBReader, blockHash *externalapi.DomainHash) (*externalapi.DomainHash, error) {
	if utxoDiffChild, ok := uds.utxoDiffChildStaging[*blockHash]; ok {
		return utxoDiffChild, nil
	}

	utxoDiffChildBytes, err := dbContext.Get(uds.utxoDiffChildHashAsKey(blockHash))
	if err != nil {
		return nil, err
	}

	utxoDiffChild, err := uds.deserializeUTXODiffChild(utxoDiffChildBytes)
	if err != nil {
		return nil, err
	}
	return utxoDiffChild, nil
}

// HasUTXODiffChild returns true if the given blockHash has a UTXODiffChild
func (uds *utxoDiffStore) HasUTXODiffChild(dbContext model.DBReader, blockHash *externalapi.DomainHash) (bool, error) {
	if _, ok := uds.utxoDiffChildStaging[*blockHash]; ok {
		return true, nil
	}

	return dbContext.Has(uds.utxoDiffChildHashAsKey(blockHash))
}

// Delete deletes the utxoDiff associated with the given blockHash
func (uds *utxoDiffStore) Delete(blockHash *externalapi.DomainHash) {
	if uds.IsBlockHashStaged(blockHash) {
		if _, ok := uds.utxoDiffStaging[*blockHash]; ok {
			delete(uds.utxoDiffStaging, *blockHash)
		}
		if _, ok := uds.utxoDiffChildStaging[*blockHash]; ok {
			delete(uds.utxoDiffChildStaging, *blockHash)
		}
		return
	}
	uds.toDelete[*blockHash] = struct{}{}
}

func (uds *utxoDiffStore) utxoDiffHashAsKey(hash *externalapi.DomainHash) model.DBKey {
	return utxoDiffBucket.Key(hash[:])
}

func (uds *utxoDiffStore) utxoDiffChildHashAsKey(hash *externalapi.DomainHash) model.DBKey {
	return utxoDiffChildBucket.Key(hash[:])
}

func (uds *utxoDiffStore) serializeUTXODiff(utxoDiff *model.UTXODiff) ([]byte, error) {
	dbUtxoDiff := serialization.UTXODiffToDBUTXODiff(utxoDiff)
	bytes, err := proto.Marshal(dbUtxoDiff)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return bytes, nil
}

func (uds *utxoDiffStore) deserializeUTXODiff(utxoDiffBytes []byte) (*model.UTXODiff, error) {
	dbUTXODiff := &serialization.DbUtxoDiff{}
	err := proto.Unmarshal(utxoDiffBytes, dbUTXODiff)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return serialization.DBUTXODiffToUTXODiff(dbUTXODiff)
}

func (uds *utxoDiffStore) serializeUTXODiffChild(utxoDiffChild *externalapi.DomainHash) ([]byte, error) {
	bytes, err := proto.Marshal(serialization.DomainHashToDbHash(utxoDiffChild))
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return bytes, nil
}

func (uds *utxoDiffStore) deserializeUTXODiffChild(utxoDiffChildBytes []byte) (*externalapi.DomainHash, error) {
	dbHash := &serialization.DbHash{}
	err := proto.Unmarshal(utxoDiffChildBytes, dbHash)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return serialization.DbHashToDomainHash(dbHash)
}

func (uds *utxoDiffStore) cloneUTXODiff(diff *model.UTXODiff) (*model.UTXODiff, error) {
	serialized, err := uds.serializeUTXODiff(diff)
	if err != nil {
		return nil, err
	}

	return uds.deserializeUTXODiff(serialized)
}

func (uds *utxoDiffStore) cloneUTXODiffChild(diffChild *externalapi.DomainHash) *externalapi.DomainHash {
	return diffChild.Clone()
}
