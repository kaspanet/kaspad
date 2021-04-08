package utxodiffstore

import (
	"github.com/golang/protobuf/proto"
	"github.com/kaspanet/kaspad/domain/consensus/database"
	"github.com/kaspanet/kaspad/domain/consensus/database/serialization"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/lrucache"
	"github.com/pkg/errors"
)

var utxoDiffBucket = database.MakeBucket([]byte("utxo-diffs"))
var utxoDiffChildBucket = database.MakeBucket([]byte("utxo-diff-children"))

// utxoDiffStore represents a store of UTXODiffs
type utxoDiffStore struct {
	utxoDiffCache      *lrucache.LRUCache
	utxoDiffChildCache *lrucache.LRUCache
}

// New instantiates a new UTXODiffStore
func New(cacheSize int, preallocate bool) model.UTXODiffStore {
	return &utxoDiffStore{
		utxoDiffCache:      lrucache.New(cacheSize, preallocate),
		utxoDiffChildCache: lrucache.New(cacheSize, preallocate),
	}
}

// Stage stages the given utxoDiff for the given blockHash
func (uds *utxoDiffStore) Stage(stagingArea *model.StagingArea, blockHash *externalapi.DomainHash, utxoDiff externalapi.UTXODiff, utxoDiffChild *externalapi.DomainHash) {
	stagingShard := uds.stagingShard(stagingArea)

	stagingShard.utxoDiffToAdd[*blockHash] = utxoDiff

	if utxoDiffChild != nil {
		stagingShard.utxoDiffChildToAdd[*blockHash] = utxoDiffChild
	}
}

func (uds *utxoDiffStore) IsStaged(stagingArea *model.StagingArea) bool {
	return uds.stagingShard(stagingArea).isStaged()
}

func (uds *utxoDiffStore) isBlockHashStaged(stagingShard *utxoDiffStagingShard, blockHash *externalapi.DomainHash) bool {
	if _, ok := stagingShard.utxoDiffToAdd[*blockHash]; ok {
		return true
	}
	_, ok := stagingShard.utxoDiffChildToAdd[*blockHash]
	return ok
}

// UTXODiff gets the utxoDiff associated with the given blockHash
func (uds *utxoDiffStore) UTXODiff(dbContext model.DBReader, stagingArea *model.StagingArea, blockHash *externalapi.DomainHash) (externalapi.UTXODiff, error) {
	stagingShard := uds.stagingShard(stagingArea)

	if utxoDiff, ok := stagingShard.utxoDiffToAdd[*blockHash]; ok {
		return utxoDiff, nil
	}

	if utxoDiff, ok := uds.utxoDiffCache.Get(blockHash); ok {
		return utxoDiff.(externalapi.UTXODiff), nil
	}

	utxoDiffBytes, err := dbContext.Get(uds.utxoDiffHashAsKey(blockHash))
	if err != nil {
		return nil, err
	}

	utxoDiff, err := uds.deserializeUTXODiff(utxoDiffBytes)
	if err != nil {
		return nil, err
	}
	uds.utxoDiffCache.Add(blockHash, utxoDiff)
	return utxoDiff, nil
}

// UTXODiffChild gets the utxoDiff child associated with the given blockHash
func (uds *utxoDiffStore) UTXODiffChild(dbContext model.DBReader, stagingArea *model.StagingArea, blockHash *externalapi.DomainHash) (*externalapi.DomainHash, error) {
	stagingShard := uds.stagingShard(stagingArea)

	if utxoDiffChild, ok := stagingShard.utxoDiffChildToAdd[*blockHash]; ok {
		return utxoDiffChild, nil
	}

	if utxoDiffChild, ok := uds.utxoDiffChildCache.Get(blockHash); ok {
		return utxoDiffChild.(*externalapi.DomainHash), nil
	}

	utxoDiffChildBytes, err := dbContext.Get(uds.utxoDiffChildHashAsKey(blockHash))
	if err != nil {
		return nil, err
	}

	utxoDiffChild, err := uds.deserializeUTXODiffChild(utxoDiffChildBytes)
	if err != nil {
		return nil, err
	}
	uds.utxoDiffChildCache.Add(blockHash, utxoDiffChild)
	return utxoDiffChild, nil
}

// HasUTXODiffChild returns true if the given blockHash has a UTXODiffChild
func (uds *utxoDiffStore) HasUTXODiffChild(dbContext model.DBReader, stagingArea *model.StagingArea, blockHash *externalapi.DomainHash) (bool, error) {
	stagingShard := uds.stagingShard(stagingArea)

	if _, ok := stagingShard.utxoDiffChildToAdd[*blockHash]; ok {
		return true, nil
	}

	if uds.utxoDiffChildCache.Has(blockHash) {
		return true, nil
	}

	return dbContext.Has(uds.utxoDiffChildHashAsKey(blockHash))
}

// Delete deletes the utxoDiff associated with the given blockHash
func (uds *utxoDiffStore) Delete(stagingArea *model.StagingArea, blockHash *externalapi.DomainHash) {
	stagingShard := uds.stagingShard(stagingArea)

	if uds.isBlockHashStaged(stagingShard, blockHash) {
		if _, ok := stagingShard.utxoDiffToAdd[*blockHash]; ok {
			delete(stagingShard.utxoDiffToAdd, *blockHash)
		}
		if _, ok := stagingShard.utxoDiffChildToAdd[*blockHash]; ok {
			delete(stagingShard.utxoDiffChildToAdd, *blockHash)
		}
		return
	}
	stagingShard.toDelete[*blockHash] = struct{}{}
}

func (uds *utxoDiffStore) utxoDiffHashAsKey(hash *externalapi.DomainHash) model.DBKey {
	return utxoDiffBucket.Key(hash.ByteSlice())
}

func (uds *utxoDiffStore) utxoDiffChildHashAsKey(hash *externalapi.DomainHash) model.DBKey {
	return utxoDiffChildBucket.Key(hash.ByteSlice())
}

func (uds *utxoDiffStore) serializeUTXODiff(utxoDiff externalapi.UTXODiff) ([]byte, error) {
	dbUtxoDiff, err := serialization.UTXODiffToDBUTXODiff(utxoDiff)
	if err != nil {
		return nil, err
	}

	bytes, err := proto.Marshal(dbUtxoDiff)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return bytes, nil
}

func (uds *utxoDiffStore) deserializeUTXODiff(utxoDiffBytes []byte) (externalapi.UTXODiff, error) {
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
