package pruningstore

import (
	"runtime"

	"github.com/golang/protobuf/proto"
	"github.com/kaspanet/kaspad/domain/consensus/database"
	"github.com/kaspanet/kaspad/domain/consensus/database/serialization"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/infrastructure/logger"
)

var pruningBlockHashKey = database.MakeBucket(nil).Key([]byte("pruning-block-hash"))
var candidatePruningPointHashKey = database.MakeBucket(nil).Key([]byte("candidate-pruning-point-hash"))
var pruningSerializedUTXOSetKey = database.MakeBucket(nil).Key([]byte("pruning-utxo-set"))

var log, _ = logger.Get(logger.SubsystemTags.BDAG)

// pruningStore represents a store for the current pruning state
type pruningStore struct {
	pruningPointStaging          *externalapi.DomainHash
	pruningPointCandidateStaging *externalapi.DomainHash
	serializedUTXOSetStaging     []byte
	pruningPointCandidateCache   *externalapi.DomainHash
	pruningPointCache            *externalapi.DomainHash
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

// New instantiates a new PruningStore
func New() model.PruningStore {
	return &pruningStore{}
}

// Stage stages the pruning state
func (ps *pruningStore) StagePruningPoint(pruningPointBlockHash *externalapi.DomainHash, pruningPointUTXOSetBytes []byte) {
	ps.pruningPointStaging = pruningPointBlockHash
	ps.serializedUTXOSetStaging = pruningPointUTXOSetBytes
}

func (ps *pruningStore) IsStaged() bool {
	return ps.pruningPointStaging != nil || ps.serializedUTXOSetStaging != nil
}

func (ps *pruningStore) Discard() {
	ps.pruningPointStaging = nil
	ps.serializedUTXOSetStaging = nil
}

func (ps *pruningStore) Commit(dbTx model.DBTransaction) error {
	stats := runtime.MemStats{}
	runtime.ReadMemStats(&stats)
	log.Debugf("Starting pruningStore.Commit, used memory: %d bytes, total: %d bytes\n", stats.Alloc, stats.HeapIdle-stats.HeapReleased+stats.HeapInuse)

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

	if ps.serializedUTXOSetStaging != nil {
		err := dbTx.Put(pruningSerializedUTXOSetKey, ps.serializedUTXOSetStaging)
		if err != nil {
			return err
		}

		runtime.ReadMemStats(&stats)
		log.Debugf("pruningStore.Commit Staged UTXO set, used memory: %d bytes, total: %d bytes\n", stats.Alloc, stats.HeapIdle-stats.HeapReleased+stats.HeapInuse)
		ps.serializedUTXOSetStaging = nil
		runtime.GC() // Clear out all the used memory by pruningSerializedUTXOSetKey.
	}

	runtime.ReadMemStats(&stats)
	log.Debugf("Finished pruningStore.Commit, used memory: %d bytes, total: %d bytes\n", stats.Alloc, stats.HeapIdle-stats.HeapReleased+stats.HeapInuse)

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

// PruningPointSerializedUTXOSet returns the serialized UTXO set of the current pruning point
func (ps *pruningStore) PruningPointSerializedUTXOSet(dbContext model.DBReader) ([]byte, error) {
	if ps.serializedUTXOSetStaging != nil {
		return ps.serializedUTXOSetStaging, nil
	}

	pruningPointUTXOSet, err := dbContext.Get(pruningSerializedUTXOSetKey)
	if err != nil {
		return nil, err
	}
	return pruningPointUTXOSet, nil
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

func (ps *pruningStore) HasPruningPoint(dbContext model.DBReader) (bool, error) {
	if ps.pruningPointStaging != nil {
		return true, nil
	}

	if ps.pruningPointCache != nil {
		return true, nil
	}

	return dbContext.Has(pruningBlockHashKey)
}
