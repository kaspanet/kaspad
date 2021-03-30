package pruningstore

import (
	"github.com/golang/protobuf/proto"
	"github.com/kaspanet/kaspad/domain/consensus/database"
	"github.com/kaspanet/kaspad/domain/consensus/database/serialization"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

var pruningBlockHashKey = database.MakeBucket(nil).Key([]byte("pruning-block-hash"))
var candidatePruningPointHashKey = database.MakeBucket(nil).Key([]byte("candidate-pruning-point-hash"))
var pruningPointUTXOSetBucket = database.MakeBucket([]byte("pruning-point-utxo-set"))
var updatingPruningPointUTXOSetKey = database.MakeBucket(nil).Key([]byte("updating-pruning-point-utxo-set"))

// pruningStore represents a store for the current pruning state
type pruningStore struct {
	pruningPointCache          *externalapi.DomainHash
	pruningPointCandidateCache *externalapi.DomainHash
}

// New instantiates a new PruningStore
func New() model.PruningStore {
	return &pruningStore{}
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

	candidateBytes, err := dbContext.Get(pruningBlockHashKey)
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

	return dbContext.Has(candidatePruningPointHashKey)
}

// Stage stages the pruning state
func (ps *pruningStore) StagePruningPoint(stagingArea *model.StagingArea, pruningPointBlockHash *externalapi.DomainHash) {
	stagingShard := ps.stagingShard(stagingArea)

	stagingShard.newPruningPoint = pruningPointBlockHash
}

func (ps *pruningStore) IsStaged(stagingArea *model.StagingArea) bool {
	return ps.stagingShard(stagingArea).isStaged()
}

func (ps *pruningStore) UpdatePruningPointUTXOSet(dbContext model.DBWriter,
	utxoSetIterator externalapi.ReadOnlyUTXOSetIterator) error {

	// Delete all the old UTXOs from the database
	deleteCursor, err := dbContext.Cursor(pruningPointUTXOSetBucket)
	if err != nil {
		return err
	}
	defer deleteCursor.Close()
	for ok := deleteCursor.First(); ok; ok = deleteCursor.Next() {
		key, err := deleteCursor.Key()
		if err != nil {
			return err
		}
		err = dbContext.Delete(key)
		if err != nil {
			return err
		}
	}

	// Insert all the new UTXOs into the database
	for ok := utxoSetIterator.First(); ok; ok = utxoSetIterator.Next() {
		outpoint, entry, err := utxoSetIterator.Get()
		if err != nil {
			return err
		}
		serializedOutpoint, err := serializeOutpoint(outpoint)
		if err != nil {
			return err
		}
		key := pruningPointUTXOSetBucket.Key(serializedOutpoint)
		serializedUTXOEntry, err := serializeUTXOEntry(entry)
		if err != nil {
			return err
		}
		err = dbContext.Put(key, serializedUTXOEntry)
		if err != nil {
			return err
		}
	}

	return nil
}

// PruningPoint gets the current pruning point
func (ps *pruningStore) PruningPoint(dbContext model.DBReader, stagingArea *model.StagingArea) (*externalapi.DomainHash, error) {
	stagingShard := ps.stagingShard(stagingArea)

	if stagingShard.newPruningPoint != nil {
		return stagingShard.newPruningPoint, nil
	}

	if ps.pruningPointCache != nil {
		return ps.pruningPointCache, nil
	}

	pruningPointBytes, err := dbContext.Get(pruningBlockHashKey)
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

	if stagingShard.newPruningPoint != nil {
		return true, nil
	}

	if ps.pruningPointCache != nil {
		return true, nil
	}

	return dbContext.Has(pruningBlockHashKey)
}

func (ps *pruningStore) PruningPointUTXOs(dbContext model.DBReader,
	fromOutpoint *externalapi.DomainOutpoint, limit int) ([]*externalapi.OutpointAndUTXOEntryPair, error) {

	cursor, err := dbContext.Cursor(pruningPointUTXOSetBucket)
	if err != nil {
		return nil, err
	}
	defer cursor.Close()

	if fromOutpoint != nil {
		serializedFromOutpoint, err := serializeOutpoint(fromOutpoint)
		if err != nil {
			return nil, err
		}
		seekKey := pruningPointUTXOSetBucket.Key(serializedFromOutpoint)
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
	return dbContext.Has(updatingPruningPointUTXOSetKey)
}

func (ps *pruningStore) FinishUpdatingPruningPointUTXOSet(dbContext model.DBWriter) error {
	return dbContext.Delete(updatingPruningPointUTXOSetKey)
}
