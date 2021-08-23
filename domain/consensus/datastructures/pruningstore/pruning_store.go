package pruningstore

import (
	"encoding/binary"
	"github.com/golang/protobuf/proto"
	"github.com/kaspanet/kaspad/domain/consensus/database"
	"github.com/kaspanet/kaspad/domain/consensus/database/binaryserialization"
	"github.com/kaspanet/kaspad/domain/consensus/database/serialization"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/lrucacheuint64tohash"
	"github.com/kaspanet/kaspad/domain/prefixmanager/prefix"
)

var pruningBlockIndexKeyName = []byte("pruning-block-index")
var candidatePruningPointHashKeyName = []byte("candidate-pruning-point-hash")
var pruningPointUTXOSetBucketName = []byte("pruning-point-utxo-set")
var updatingPruningPointUTXOSetKeyName = []byte("updating-pruning-point-utxo-set")
var pruningPointByIndexBucketName = []byte("pruning-point-by-index")

// pruningStore represents a store for the current pruning state
type pruningStore struct {
	pruningPointByIndexCache   *lrucacheuint64tohash.LRUCache
	pruningPointIndexCache     *uint64
	pruningPointCandidateCache *externalapi.DomainHash

	pruningBlockIndexKey            model.DBKey
	candidatePruningPointHashKey    model.DBKey
	pruningPointUTXOSetBucket       model.DBBucket
	updatingPruningPointUTXOSetKey  model.DBKey
	importedPruningPointUTXOsBucket model.DBBucket
	importedPruningPointMultisetKey model.DBKey
	pruningPointByIndexBucket       model.DBBucket
}

// New instantiates a new PruningStore
func New(prefix *prefix.Prefix, cacheSize int, preallocate bool) model.PruningStore {
	return &pruningStore{
		pruningPointByIndexCache:        lrucacheuint64tohash.New(cacheSize, preallocate),
		pruningBlockIndexKey:            database.MakeBucket(prefix.Serialize()).Key(pruningBlockIndexKeyName),
		candidatePruningPointHashKey:    database.MakeBucket(prefix.Serialize()).Key(candidatePruningPointHashKeyName),
		pruningPointUTXOSetBucket:       database.MakeBucket(prefix.Serialize()).Bucket(pruningPointUTXOSetBucketName),
		importedPruningPointUTXOsBucket: database.MakeBucket(prefix.Serialize()).Bucket(importedPruningPointUTXOsBucketName),
		updatingPruningPointUTXOSetKey:  database.MakeBucket(prefix.Serialize()).Key(updatingPruningPointUTXOSetKeyName),
		importedPruningPointMultisetKey: database.MakeBucket(prefix.Serialize()).Key(importedPruningPointMultisetKeyName),
		pruningPointByIndexBucket:       database.MakeBucket(prefix.Serialize()).Bucket(pruningPointByIndexBucketName),
	}
}

func (ps *pruningStore) StagePruningPointCandidate(stagingArea *model.StagingArea, candidate *externalapi.DomainHash) {
	stagingShard := ps.stagingShard(stagingArea)

	stagingShard.newPruningPointCandidate = candidate
}

func (ps *pruningStore) PruningPointCandidate(dbContext model.DBReader, stagingArea *model.StagingArea) (*externalapi.DomainHash, error) {
	stagingShard := ps.stagingShard(stagingArea)

	if stagingShard.newPruningPointCandidate != nil {
		return stagingShard.newPruningPointCandidate, nil
	}

	if ps.pruningPointCandidateCache != nil {
		return ps.pruningPointCandidateCache, nil
	}

	candidateBytes, err := dbContext.Get(ps.candidatePruningPointHashKey)
	if err != nil {
		return nil, err
	}

	candidate, err := ps.deserializePruningPoint(candidateBytes)
	if err != nil {
		return nil, err
	}
	ps.pruningPointCandidateCache = candidate
	return candidate, nil
}

func (ps *pruningStore) HasPruningPointCandidate(dbContext model.DBReader, stagingArea *model.StagingArea) (bool, error) {
	stagingShard := ps.stagingShard(stagingArea)

	if stagingShard.newPruningPointCandidate != nil {
		return true, nil
	}

	if ps.pruningPointCandidateCache != nil {
		return true, nil
	}

	return dbContext.Has(ps.candidatePruningPointHashKey)
}

// StagePruningPoint stages the pruning state
func (ps *pruningStore) StagePruningPoint(dbContext model.DBWriter, stagingArea *model.StagingArea, pruningPointBlockHash *externalapi.DomainHash) error {
	newPruningPointIndex := uint64(0)
	pruningPointIndex, err := ps.PruningPointIndex(dbContext, stagingArea)
	if database.IsNotFoundError(err) {
		newPruningPointIndex = 0
	} else if err != nil {
		return err
	} else {
		newPruningPointIndex = pruningPointIndex + 1
	}

	err = ps.StagePruningPointByIndex(dbContext, stagingArea, pruningPointBlockHash, newPruningPointIndex)
	if err != nil {
		return err
	}

	return nil
}

func (ps *pruningStore) IsStaged(stagingArea *model.StagingArea) bool {
	return ps.stagingShard(stagingArea).isStaged()
}

func (ps *pruningStore) UpdatePruningPointUTXOSet(dbContext model.DBWriter, diff externalapi.UTXODiff) error {
	toRemoveIterator := diff.ToRemove().Iterator()
	defer toRemoveIterator.Close()
	for ok := toRemoveIterator.First(); ok; ok = toRemoveIterator.Next() {
		toRemoveOutpoint, _, err := toRemoveIterator.Get()
		if err != nil {
			return err
		}
		serializedOutpoint, err := serializeOutpoint(toRemoveOutpoint)
		if err != nil {
			return err
		}
		err = dbContext.Delete(ps.pruningPointUTXOSetBucket.Key(serializedOutpoint))
		if err != nil {
			return err
		}
	}

	toAddIterator := diff.ToAdd().Iterator()
	defer toAddIterator.Close()
	for ok := toAddIterator.First(); ok; ok = toAddIterator.Next() {
		toAddOutpoint, entry, err := toAddIterator.Get()
		if err != nil {
			return err
		}
		serializedOutpoint, err := serializeOutpoint(toAddOutpoint)
		if err != nil {
			return err
		}
		serializedUTXOEntry, err := serializeUTXOEntry(entry)
		if err != nil {
			return err
		}
		err = dbContext.Put(ps.pruningPointUTXOSetBucket.Key(serializedOutpoint), serializedUTXOEntry)
		if err != nil {
			return err
		}
	}
	return nil
}

// PruningPoint gets the current pruning point
func (ps *pruningStore) PruningPoint(dbContext model.DBReader, stagingArea *model.StagingArea) (*externalapi.DomainHash, error) {
	pruningPointIndex, err := ps.PruningPointIndex(dbContext, stagingArea)
	if err != nil {
		return nil, err
	}

	return ps.PruningPointByIndex(dbContext, stagingArea, pruningPointIndex)
}

func (ps *pruningStore) PruningPointByIndex(dbContext model.DBReader, stagingArea *model.StagingArea, index uint64) (*externalapi.DomainHash, error) {
	stagingShard := ps.stagingShard(stagingArea)

	if hash, exists := stagingShard.pruningPointByIndex[index]; exists {
		return hash, nil
	}

	if hash, exists := ps.pruningPointByIndexCache.Get(index); exists {
		return hash, nil
	}

	pruningPointBytes, err := dbContext.Get(ps.indexAsKey(index))
	if err != nil {
		return nil, err
	}

	pruningPoint, err := ps.deserializePruningPoint(pruningPointBytes)
	if err != nil {
		return nil, err
	}
	ps.pruningPointByIndexCache.Add(index, pruningPoint)
	return pruningPoint, nil
}

func (ps *pruningStore) serializeHash(hash *externalapi.DomainHash) ([]byte, error) {
	return proto.Marshal(serialization.DomainHashToDbHash(hash))
}

func (ps *pruningStore) deserializePruningPoint(pruningPointBytes []byte) (*externalapi.DomainHash, error) {
	dbHash := &serialization.DbHash{}
	err := proto.Unmarshal(pruningPointBytes, dbHash)
	if err != nil {
		return nil, err
	}

	return serialization.DbHashToDomainHash(dbHash)
}

func (ps *pruningStore) deserializeIndex(indexBytes []byte) (uint64, error) {
	return binaryserialization.DeserializeUint64(indexBytes)
}

func (ps *pruningStore) serializeIndex(index uint64) []byte {
	return binaryserialization.SerializeUint64(index)
}

func (ps *pruningStore) HasPruningPoint(dbContext model.DBReader, stagingArea *model.StagingArea) (bool, error) {
	stagingShard := ps.stagingShard(stagingArea)

	if stagingShard.pruningPointIndex != nil {
		return true, nil
	}

	if ps.pruningPointIndexCache != nil {
		return true, nil
	}

	return dbContext.Has(ps.pruningBlockIndexKey)
}

func (ps *pruningStore) PruningPointUTXOIterator(dbContext model.DBReader) (externalapi.ReadOnlyUTXOSetIterator, error) {
	cursor, err := dbContext.Cursor(ps.pruningPointUTXOSetBucket)
	if err != nil {
		return nil, err
	}
	return ps.newCursorUTXOSetIterator(cursor), nil
}

func (ps *pruningStore) PruningPointUTXOs(dbContext model.DBReader,
	fromOutpoint *externalapi.DomainOutpoint, limit int) ([]*externalapi.OutpointAndUTXOEntryPair, error) {

	cursor, err := dbContext.Cursor(ps.pruningPointUTXOSetBucket)
	if err != nil {
		return nil, err
	}
	defer cursor.Close()

	if fromOutpoint != nil {
		serializedFromOutpoint, err := serializeOutpoint(fromOutpoint)
		if err != nil {
			return nil, err
		}
		seekKey := ps.pruningPointUTXOSetBucket.Key(serializedFromOutpoint)
		err = cursor.Seek(seekKey)
		if err != nil {
			return nil, err
		}
	}

	pruningPointUTXOIterator := ps.newCursorUTXOSetIterator(cursor)
	defer pruningPointUTXOIterator.Close()

	outpointAndUTXOEntryPairs := make([]*externalapi.OutpointAndUTXOEntryPair, 0, limit)
	for len(outpointAndUTXOEntryPairs) < limit && pruningPointUTXOIterator.Next() {
		outpoint, utxoEntry, err := pruningPointUTXOIterator.Get()
		if err != nil {
			return nil, err
		}
		outpointAndUTXOEntryPairs = append(outpointAndUTXOEntryPairs, &externalapi.OutpointAndUTXOEntryPair{
			Outpoint:  outpoint,
			UTXOEntry: utxoEntry,
		})
	}
	return outpointAndUTXOEntryPairs, nil
}

func (ps *pruningStore) StageStartUpdatingPruningPointUTXOSet(stagingArea *model.StagingArea) {
	stagingShard := ps.stagingShard(stagingArea)

	stagingShard.startUpdatingPruningPointUTXOSet = true
}

func (ps *pruningStore) HadStartedUpdatingPruningPointUTXOSet(dbContext model.DBWriter) (bool, error) {
	return dbContext.Has(ps.updatingPruningPointUTXOSetKey)
}

func (ps *pruningStore) FinishUpdatingPruningPointUTXOSet(dbContext model.DBWriter) error {
	return dbContext.Delete(ps.updatingPruningPointUTXOSetKey)
}

func (ps *pruningStore) indexAsKey(index uint64) model.DBKey {
	var keyBytes [8]byte
	binary.BigEndian.PutUint64(keyBytes[:], index)
	return ps.pruningPointByIndexBucket.Key(keyBytes[:])
}

func (ps *pruningStore) StagePruningPointByIndex(dbContext model.DBReader, stagingArea *model.StagingArea,
	pruningPointBlockHash *externalapi.DomainHash, index uint64) error {

	stagingShard := ps.stagingShard(stagingArea)
	stagingShard.pruningPointByIndex[index] = pruningPointBlockHash

	pruningPointIndex, err := ps.PruningPointIndex(dbContext, stagingArea)
	isNotFoundError := database.IsNotFoundError(err)
	if !isNotFoundError && err != nil {
		return err
	}

	if stagingShard.pruningPointIndex == nil {
		var zero uint64
		stagingShard.pruningPointIndex = &zero
	}

	if isNotFoundError || index > pruningPointIndex {
		*stagingShard.pruningPointIndex = index
	}

	return nil
}

func (ps *pruningStore) PruningPointIndex(dbContext model.DBReader, stagingArea *model.StagingArea) (uint64, error) {
	stagingShard := ps.stagingShard(stagingArea)

	if stagingShard.pruningPointIndex != nil {
		return *stagingShard.pruningPointIndex, nil
	}

	if ps.pruningPointIndexCache != nil {
		return *ps.pruningPointIndexCache, nil
	}

	pruningPointIndexBytes, err := dbContext.Get(ps.pruningBlockIndexKey)
	if err != nil {
		return 0, err
	}

	index, err := ps.deserializeIndex(pruningPointIndexBytes)
	if err != nil {
		return 0, err
	}
	*ps.pruningPointIndexCache = index
	return index, nil
}
