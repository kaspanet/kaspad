package pruningstore

import (
	"github.com/golang/protobuf/proto"
	"github.com/kaspanet/kaspad/domain/consensus/database/serialization"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/dbkeys"
)

var pruningBlockHashKey = dbkeys.MakeBucket().Key([]byte("pruning-block-hash"))
var candidatePruningPointHashKey = dbkeys.MakeBucket().Key([]byte("candidate-pruning-point-hash"))
var pruningPointUTXOSetBucket = dbkeys.MakeBucket([]byte("pruning-point-utxo-set"))

// pruningStore represents a store for the current pruning state
type pruningStore struct {
	databaseContext model.DBManager

	pruningPointStaging          *externalapi.DomainHash
	pruningPointCache            *externalapi.DomainHash
	pruningPointCandidateStaging *externalapi.DomainHash
	pruningPointCandidateCache   *externalapi.DomainHash

	pruningPointUTXOSetStaging model.ReadOnlyUTXOSetIterator
}

// New instantiates a new PruningStore
func New(databaseContext model.DBManager) model.PruningStore {
	return &pruningStore{
		databaseContext: databaseContext,
	}
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

func (ps *pruningStore) StagePruningPointUTXOSet(pruningPointUTXOSetIterator model.ReadOnlyUTXOSetIterator) {
	ps.pruningPointUTXOSetStaging = pruningPointUTXOSetIterator
}

func (ps *pruningStore) IsStaged() bool {
	return ps.pruningPointStaging != nil || ps.pruningPointUTXOSetStaging != nil
}

func (ps *pruningStore) Discard() {
	ps.pruningPointStaging = nil
	ps.pruningPointUTXOSetStaging = nil
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

	if ps.pruningPointUTXOSetStaging != nil {
		// Delete all the old UTXOs from the database
		deleteCursor, err := dbTx.Cursor(pruningPointUTXOSetBucket)
		if err != nil {
			return err
		}
		for deleteCursor.Next() {
			key, err := deleteCursor.Key()
			if err != nil {
				return err
			}
			err = dbTx.Delete(key)
			if err != nil {
				return err
			}
		}

		// Insert all the new UTXOs into the database
		ps.pruningPointUTXOSetStaging.First()
		for ps.pruningPointUTXOSetStaging.Next() {
			outpoint, entry, err := ps.pruningPointUTXOSetStaging.Get()
			if err != nil {
				return err
			}
			serializedOutpoint, err := ps.serializeOutpoint(outpoint)
			if err != nil {
				return err
			}
			key := pruningPointUTXOSetBucket.Key(serializedOutpoint)
			serializedUTXOEntry, err := ps.serializeUTXOEntry(entry)
			if err != nil {
				return err
			}
			err = dbTx.Put(key, serializedUTXOEntry)
			if err != nil {
				return err
			}
		}
	}

	ps.Discard()
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
	for offset < offsetIndex && pruningPointUTXOIterator.Next() {
		offsetIndex++
	}

	outpointAndUTXOEntryPairs := make([]*externalapi.OutpointAndUTXOEntryPair, 0, limit)
	limitIndex := 0
	for limitIndex < limit && pruningPointUTXOIterator.Next() {
		outpoint, utxoEntry, err := pruningPointUTXOIterator.Get()
		if err != nil {
			return nil, err
		}
		outpointAndUTXOEntryPairs = append(outpointAndUTXOEntryPairs, &externalapi.OutpointAndUTXOEntryPair{
			Outpoint:  outpoint,
			UTXOEntry: utxoEntry,
		})
		limitIndex++
	}
	return outpointAndUTXOEntryPairs, nil
}
