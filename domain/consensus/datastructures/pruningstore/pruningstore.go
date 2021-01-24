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

// pruningStore represents a store for the current pruning state
type pruningStore struct {
	pruningPointStaging          *externalapi.DomainHash
	pruningPointCache            *externalapi.DomainHash
	pruningPointCandidateStaging *externalapi.DomainHash
	pruningPointCandidateCache   *externalapi.DomainHash
}

// New instantiates a new PruningStore
func New() model.PruningStore {
	return &pruningStore{}
}

func (ps *pruningStore) StagePruningPointCandidate(candidate *externalapi.DomainHash) {
	ps.pruningPointCandidateStaging = candidate
}

func (ps *pruningStore) PruningPointCandidate(dbContext model.DBReader) (*externalapi.DomainHash, error) {
	if ps.pruningPointCandidateStaging != nil {
		return ps.pruningPointCandidateStaging, nil
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

func (ps *pruningStore) HasPruningPointCandidate(dbContext model.DBReader) (bool, error) {
	if ps.pruningPointCandidateStaging != nil {
		return true, nil
	}

	if ps.pruningPointCandidateCache != nil {
		return true, nil
	}

	return dbContext.Has(candidatePruningPointHashKey)
}

// Stage stages the pruning state
func (ps *pruningStore) StagePruningPoint(pruningPointBlockHash *externalapi.DomainHash) {
	ps.pruningPointStaging = pruningPointBlockHash
}

func (ps *pruningStore) IsStaged() bool {
	return ps.pruningPointStaging != nil
}

func (ps *pruningStore) Discard() {
	ps.pruningPointStaging = nil
}

func (ps *pruningStore) Commit(dbTx model.DBTransaction) error {
	if ps.pruningPointStaging != nil {
		pruningPointBytes, err := ps.serializeHash(ps.pruningPointStaging)
		if err != nil {
			return err
		}
		err = dbTx.Put(pruningBlockHashKey, pruningPointBytes)
		if err != nil {
			return err
		}
		ps.pruningPointCache = ps.pruningPointStaging
	}

	if ps.pruningPointCandidateStaging != nil {
		candidateBytes, err := ps.serializeHash(ps.pruningPointCandidateStaging)
		if err != nil {
			return err
		}
		err = dbTx.Put(candidatePruningPointHashKey, candidateBytes)
		if err != nil {
			return err
		}
		ps.pruningPointCandidateCache = ps.pruningPointCandidateStaging
	}

	ps.Discard()
	return nil
}

func (ps *pruningStore) UpdatePruningPointUTXOSet(dbContext model.DBWriter,
	utxoSetIterator model.ReadOnlyUTXOSetIterator) error {

	// Delete all the old UTXOs from the database
	deleteCursor, err := dbContext.Cursor(pruningPointUTXOSetBucket)
	if err != nil {
		return err
	}
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
func (ps *pruningStore) PruningPoint(dbContext model.DBReader) (*externalapi.DomainHash, error) {
	if ps.pruningPointStaging != nil {
		return ps.pruningPointStaging, nil
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

func (ps *pruningStore) serializeUTXOSetBytes(pruningPointUTXOSetBytes []byte) ([]byte, error) {
	return proto.Marshal(&serialization.DbPruningPointUTXOSetBytes{Bytes: pruningPointUTXOSetBytes})
}

func (ps *pruningStore) deserializeUTXOSetBytes(dbPruningPointUTXOSetBytes []byte) ([]byte, error) {
	dbPruningPointUTXOSet := &serialization.DbPruningPointUTXOSetBytes{}
	err := proto.Unmarshal(dbPruningPointUTXOSetBytes, dbPruningPointUTXOSet)
	if err != nil {
		return nil, err
	}

	return dbPruningPointUTXOSet.Bytes, nil
}

func (ps *pruningStore) HasPruningPoint(dbContext model.DBReader) (bool, error) {
	if ps.pruningPointStaging != nil {
		return true, nil
	}

	if ps.pruningPointCache != nil {
		return true, nil
	}

	return dbContext.Has(pruningBlockHashKey)
}

func (ps *pruningStore) PruningPointUTXOs(dbContext model.DBReader,
	offset int, limit int) ([]*externalapi.OutpointAndUTXOEntryPair, error) {

	cursor, err := dbContext.Cursor(pruningPointUTXOSetBucket)
	if err != nil {
		return nil, err
	}
	pruningPointUTXOIterator := ps.newCursorUTXOSetIterator(cursor)

	offsetIndex := 0
	for offsetIndex < offset && pruningPointUTXOIterator.Next() {
		offsetIndex++
	}

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
