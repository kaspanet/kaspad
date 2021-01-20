package finalitystore

import (
	"github.com/kaspanet/kaspad/domain/consensus/database"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/lrucache"
)

var bucket = database.MakeBucket([]byte("finality-points"))

type finalityStore struct {
	staging  map[externalapi.DomainHash]*externalapi.DomainHash
	toDelete map[externalapi.DomainHash]struct{}
	cache    *lrucache.LRUCache
}

// New instantiates a new FinalityStore
func New(cacheSize int) model.FinalityStore {
	return &finalityStore{
		staging:  make(map[externalapi.DomainHash]*externalapi.DomainHash),
		toDelete: make(map[externalapi.DomainHash]struct{}),
		cache:    lrucache.New(cacheSize),
	}
}

func (fs *finalityStore) StageFinalityPoint(blockHash *externalapi.DomainHash, finalityPointHash *externalapi.DomainHash) {
	fs.staging[*blockHash] = finalityPointHash
}

func (fs *finalityStore) FinalityPoint(
	dbContext model.DBReader, blockHash *externalapi.DomainHash) (*externalapi.DomainHash, error) {
	if finalityPointHash, ok := fs.staging[*blockHash]; ok {
		return finalityPointHash, nil
	}

	if finalityPointHash, ok := fs.cache.Get(blockHash); ok {
		return finalityPointHash.(*externalapi.DomainHash), nil
	}

	finalityPointHashBytes, err := dbContext.Get(fs.hashAsKey(blockHash))
	if err != nil {
		return nil, err
	}
	finalityPointHash, err := externalapi.NewDomainHashFromByteSlice(finalityPointHashBytes)
	if err != nil {
		return nil, err
	}

	fs.cache.Add(blockHash, finalityPointHash)
	return finalityPointHash, nil
}

func (fs *finalityStore) Discard() {
	fs.staging = make(map[externalapi.DomainHash]*externalapi.DomainHash)
}

func (fs *finalityStore) Commit(dbTx model.DBTransaction) error {
	for hash, finalityPointHash := range fs.staging {
		err := dbTx.Put(fs.hashAsKey(&hash), finalityPointHash.ByteSlice())
		if err != nil {
			return err
		}
		fs.cache.Add(&hash, finalityPointHash)
	}

	fs.Discard()
	return nil
}

func (fs *finalityStore) IsStaged() bool {
	return len(fs.staging) == 0
}

func (fs *finalityStore) hashAsKey(hash *externalapi.DomainHash) model.DBKey {
	return bucket.Key(hash.ByteSlice())
}
