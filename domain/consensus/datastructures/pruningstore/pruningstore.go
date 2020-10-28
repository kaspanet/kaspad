package pruningstore

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/dbkeys"
)

var pruningBlockHashKey = dbkeys.MakeBucket().Key([]byte("pruning-block-hash"))
var pruningSerializedUTXOSetkey = dbkeys.MakeBucket().Key([]byte("pruning-utxo-set"))

// pruningStore represents a store for the current pruning state
type pruningStore struct {
	blockHashStaging         *externalapi.DomainHash
	serializedUTXOSetStaging []byte
}

// New instantiates a new PruningStore
func New() model.PruningStore {
	return &pruningStore{
		blockHashStaging:         nil,
		serializedUTXOSetStaging: nil,
	}
}

// Stage stages the pruning state
func (ps *pruningStore) Stage(pruningPointBlockHash *externalapi.DomainHash, pruningPointUTXOSet model.ReadOnlyUTXOSet) {
	ps.blockHashStaging = pruningPointBlockHash
	ps.serializedUTXOSetStaging = ps.serializeUTXOSet(pruningPointUTXOSet)
}

func (ps *pruningStore) IsStaged() bool {
	return ps.blockHashStaging != nil || ps.serializedUTXOSetStaging != nil
}

func (ps *pruningStore) Discard() {
	ps.blockHashStaging = nil
	ps.serializedUTXOSetStaging = nil
}

func (ps *pruningStore) Commit(dbTx model.DBTransaction) error {
	err := dbTx.Put(pruningBlockHashKey, ps.serializePruningPoint(ps.blockHashStaging))
	if err != nil {
		return err
	}
	err = dbTx.Put(pruningSerializedUTXOSetkey, ps.serializedUTXOSetStaging)
	if err != nil {
		return err
	}
	ps.Discard()
	return nil
}

// PruningPoint gets the current pruning point
func (ps *pruningStore) PruningPoint(dbContext model.DBReader) (*externalapi.DomainHash, error) {
	if ps.blockHashStaging != nil {
		return ps.blockHashStaging, nil
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
	return dbContext.Get(pruningSerializedUTXOSetkey)
}

func (ps *pruningStore) serializePruningPoint(pruningPoint *externalapi.DomainHash) []byte {
	panic("implement me")
}

func (ps *pruningStore) deserializePruningPoint(pruningPointBytes []byte) (*externalapi.DomainHash, error) {
	panic("implement me")
}

func (ps *pruningStore) serializeUTXOSet(utxoSet model.ReadOnlyUTXOSet) []byte {
	panic("implement me")
}
