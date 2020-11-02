package pruningstore

import (
	"github.com/golang/protobuf/proto"
	"github.com/kaspanet/kaspad/domain/consensus/database/serialization"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/dbkeys"
)

var pruningBlockHashKey = dbkeys.MakeBucket().Key([]byte("pruning-block-hash"))
var pruningSerializedUTXOSetkey = dbkeys.MakeBucket().Key([]byte("pruning-utxo-set"))

// pruningStore represents a store for the current pruning state
type pruningStore struct {
	pruningPointStaging      *externalapi.DomainHash
	serializedUTXOSetStaging []byte
}

// New instantiates a new PruningStore
func New() model.PruningStore {
	return &pruningStore{
		pruningPointStaging:      nil,
		serializedUTXOSetStaging: nil,
	}
}

// Stage stages the pruning state
func (ps *pruningStore) Stage(pruningPointBlockHash *externalapi.DomainHash, pruningPointUTXOSetBytes []byte) {
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
	pruningPointBytes, err := ps.serializePruningPoint(ps.pruningPointStaging)
	if err != nil {
		return err
	}

	err = dbTx.Put(pruningBlockHashKey, pruningPointBytes)
	if err != nil {
		return err
	}

	utxoSetBytes, err := ps.serializeUTXOSetBytes(ps.serializedUTXOSetStaging)
	if err != nil {
		return err
	}

	err = dbTx.Put(pruningSerializedUTXOSetkey, utxoSetBytes)
	if err != nil {
		return err
	}

	ps.Discard()
	return nil
}

// PruningPoint gets the current pruning point
func (ps *pruningStore) PruningPoint(dbContext model.DBReader) (*externalapi.DomainHash, error) {
	if ps.pruningPointStaging != nil {
		return ps.pruningPointStaging, nil
	}

	blockHashBytes, err := dbContext.Get(pruningBlockHashKey)
	if err != nil {
		return nil, err
	}

	blockHash, err := ps.deserializePruningPoint(blockHashBytes)
	if err != nil {
		return nil, err
	}
	return blockHash, nil
}

// PruningPointSerializedUTXOSet returns the serialized UTXO set of the current pruning point
func (ps *pruningStore) PruningPointSerializedUTXOSet(dbContext model.DBReader) ([]byte, error) {
	if ps.serializedUTXOSetStaging != nil {
		return ps.serializedUTXOSetStaging, nil
	}

	dbPruningPointUTXOSetBytes, err := dbContext.Get(pruningSerializedUTXOSetkey)
	if err != nil {
		return nil, err
	}

	return ps.deserializeUTXOSetBytes(dbPruningPointUTXOSetBytes)
}

func (ps *pruningStore) serializePruningPoint(pruningPoint *externalapi.DomainHash) ([]byte, error) {
	return proto.Marshal(serialization.DomainHashToDbHash(pruningPoint))
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
	return proto.Marshal(&serialization.PruningPointUTXOSetBytes{Bytes: pruningPointUTXOSetBytes})
}

func (ps *pruningStore) deserializeUTXOSetBytes(dbPruningPointUTXOSetBytes []byte) ([]byte, error) {
	dbPruningPointUTXOSet := &serialization.PruningPointUTXOSetBytes{}
	err := proto.Unmarshal(dbPruningPointUTXOSetBytes, dbPruningPointUTXOSet)
	if err != nil {
		return nil, err
	}

	return dbPruningPointUTXOSet.Bytes, nil
}
