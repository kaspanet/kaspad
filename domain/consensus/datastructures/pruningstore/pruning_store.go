package pruningstore

import (
	"github.com/golang/protobuf/proto"
	"github.com/kaspanet/kaspad/domain/consensus/database"
	"github.com/kaspanet/kaspad/domain/consensus/database/serialization"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/prefixmanager/prefix"
)

var pruningBlockHashKeyName = []byte("pruning-block-hash")
var previousPruningBlockHashKeyName = []byte("previous-pruning-block-hash")
var candidatePruningPointHashKeyName = []byte("candidate-pruning-point-hash")
var pruningPointUTXOSetBucketName = []byte("pruning-point-utxo-set")
var updatingPruningPointUTXOSetKeyName = []byte("updating-pruning-point-utxo-set")

// pruningStore represents a store for the current pruning state
type pruningStore struct {
	pruningPointCache          *externalapi.DomainHash
	oldPruningPointCache       *externalapi.DomainHash
	pruningPointCandidateCache *externalapi.DomainHash

	pruningBlockHashKey             model.DBKey
	previousPruningBlockHashKey     model.DBKey
	candidatePruningPointHashKey    model.DBKey
	pruningPointUTXOSetBucket       model.DBBucket
	updatingPruningPointUTXOSetKey  model.DBKey
	importedPruningPointUTXOsBucket model.DBBucket
	importedPruningPointMultisetKey model.DBKey
}

// New instantiates a new PruningStore
func New(prefix *prefix.Prefix) model.PruningStore {
	return &pruningStore{
		pruningBlockHashKey:             database.MakeBucket(prefix.Serialize()).Key(pruningBlockHashKeyName),
		previousPruningBlockHashKey:     database.MakeBucket(prefix.Serialize()).Key(previousPruningBlockHashKeyName),
		candidatePruningPointHashKey:    database.MakeBucket(prefix.Serialize()).Key(candidatePruningPointHashKeyName),
		pruningPointUTXOSetBucket:       database.MakeBucket(prefix.Serialize()).Bucket(pruningPointUTXOSetBucketName),
		importedPruningPointUTXOsBucket: database.MakeBucket(prefix.Serialize()).Bucket(importedPruningPointUTXOsBucketName),
		updatingPruningPointUTXOSetKey:  database.MakeBucket(prefix.Serialize()).Key(updatingPruningPointUTXOSetKeyName),
		importedPruningPointMultisetKey: database.MakeBucket(prefix.Serialize()).Key(importedPruningPointMultisetKeyName),
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

// Stage stages the pruning state
func (ps *pruningStore) StagePruningPoint(stagingArea *model.StagingArea, pruningPointBlockHash *externalapi.DomainHash) {
	stagingShard := ps.stagingShard(stagingArea)

	stagingShard.currentPruningPoint = pruningPointBlockHash
}

func (ps *pruningStore) StagePreviousPruningPoint(stagingArea *model.StagingArea, oldPruningPointBlockHash *externalapi.DomainHash) {
	stagingShard := ps.stagingShard(stagingArea)
	stagingShard.previousPruningPoint = oldPruningPointBlockHash
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
	stagingShard := ps.stagingShard(stagingArea)

	if stagingShard.currentPruningPoint != nil {
		return stagingShard.currentPruningPoint, nil
	}

	if ps.pruningPointCache != nil {
		return ps.pruningPointCache, nil
	}

	pruningPointBytes, err := dbContext.Get(ps.pruningBlockHashKey)
	if err != nil {
		return nil, err
	}

	pruningPoint, err := ps.deserializePruningPoint(pruningPointBytes)
	if err != nil {
		return nil, err
	}
	ps.pruningPointCache = pruningPoint
	return pruningPoint, nil
}

// OldPruningPoint returns the pruning point *before* the current one
func (ps *pruningStore) PreviousPruningPoint(dbContext model.DBReader, stagingArea *model.StagingArea) (*externalapi.DomainHash, error) {
	stagingShard := ps.stagingShard(stagingArea)

	if stagingShard.previousPruningPoint != nil {
		return stagingShard.previousPruningPoint, nil
	}
	if ps.oldPruningPointCache != nil {
		return ps.oldPruningPointCache, nil
	}

	oldPruningPointBytes, err := dbContext.Get(ps.previousPruningBlockHashKey)
	if err != nil {
		return nil, err
	}

	oldPruningPoint, err := ps.deserializePruningPoint(oldPruningPointBytes)
	if err != nil {
		return nil, err
	}
	ps.oldPruningPointCache = oldPruningPoint
	return oldPruningPoint, nil
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

func (ps *pruningStore) HasPruningPoint(dbContext model.DBReader, stagingArea *model.StagingArea) (bool, error) {
	stagingShard := ps.stagingShard(stagingArea)

	if stagingShard.currentPruningPoint != nil {
		return true, nil
	}

	if ps.pruningPointCache != nil {
		return true, nil
	}

	return dbContext.Has(ps.pruningBlockHashKey)
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
