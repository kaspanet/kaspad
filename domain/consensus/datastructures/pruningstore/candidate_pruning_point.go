package pruningstore

import (
	"github.com/golang/protobuf/proto"
	"github.com/kaspanet/kaspad/domain/consensus/database/serialization"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/dbkeys"
)

var candidatePruningPointUTXOsBucket = dbkeys.MakeBucket([]byte("candidate-pruning-point-utxos"))
var candidatePruningPointMultiset = dbkeys.MakeBucket().Key([]byte("candidate-pruning-point-multiset"))

func (ps *pruningStore) ClearCandidatePruningPointUTXOs(dbTx model.DBTransaction) error {
	cursor, err := dbTx.Cursor(candidatePruningPointUTXOsBucket)
	if err != nil {
		return err
	}

	for cursor.Next() {
		key, err := cursor.Key()
		if err != nil {
			return err
		}
		err = dbTx.Delete(key)
		if err != nil {
			return err
		}
	}
	return nil
}

func (ps *pruningStore) InsertCandidatePruningPointUTXOs(dbTx model.DBTransaction,
	outpointAndUTXOEntryPairs []*externalapi.OutpointAndUTXOEntryPair) error {

	for _, outpointAndUTXOEntryPair := range outpointAndUTXOEntryPairs {
		key, err := ps.candidatePruningPointUTXOKey(outpointAndUTXOEntryPair.Outpoint)
		if err != nil {
			return err
		}
		serializedUTXOEntry, err := ps.serializeUTXOEntry(outpointAndUTXOEntryPair.UTXOEntry)
		if err != nil {
			return err
		}
		err = dbTx.Put(key, serializedUTXOEntry)
		if err != nil {
			return err
		}
	}

	return nil
}

func (ps *pruningStore) CandidatePruningPointUTXOIterator(dbContext model.DBReader) (model.ReadOnlyUTXOSetIterator, error) {
	cursor, err := dbContext.Cursor(candidatePruningPointUTXOsBucket)
	if err != nil {
		return nil, err
	}
	return ps.newCursorUTXOSetIterator(cursor), nil
}

type utxoSetIterator struct {
	pruningStore *pruningStore
	cursor       model.DBCursor
}

func (ps *pruningStore) newCursorUTXOSetIterator(cursor model.DBCursor) model.ReadOnlyUTXOSetIterator {
	return &utxoSetIterator{pruningStore: ps, cursor: cursor}
}

func (u utxoSetIterator) First() {
	u.cursor.First()
}

func (u utxoSetIterator) Next() bool {
	return u.cursor.Next()
}

func (u utxoSetIterator) Get() (outpoint *externalapi.DomainOutpoint, utxoEntry externalapi.UTXOEntry, err error) {
	key, err := u.cursor.Key()
	if err != nil {
		panic(err)
	}

	utxoEntryBytes, err := u.cursor.Value()
	if err != nil {
		return nil, nil, err
	}

	outpoint, err = u.pruningStore.deserializeOutpoint(key.Suffix())
	if err != nil {
		return nil, nil, err
	}

	utxoEntry, err = u.pruningStore.deserializeUTXOEntry(utxoEntryBytes)
	if err != nil {
		return nil, nil, err
	}

	return outpoint, utxoEntry, nil
}

func (ps *pruningStore) candidatePruningPointUTXOKey(outpoint *externalapi.DomainOutpoint) (model.DBKey, error) {
	serializedOutpoint, err := ps.serializeOutpoint(outpoint)
	if err != nil {
		return nil, err
	}

	return candidatePruningPointUTXOsBucket.Key(serializedOutpoint), nil
}

func (ps *pruningStore) serializeOutpoint(outpoint *externalapi.DomainOutpoint) ([]byte, error) {
	return proto.Marshal(serialization.DomainOutpointToDbOutpoint(outpoint))
}

func (ps *pruningStore) serializeUTXOEntry(entry externalapi.UTXOEntry) ([]byte, error) {
	return proto.Marshal(serialization.UTXOEntryToDBUTXOEntry(entry))
}

func (ps *pruningStore) deserializeOutpoint(outpointBytes []byte) (*externalapi.DomainOutpoint, error) {
	dbOutpoint := &serialization.DbOutpoint{}
	err := proto.Unmarshal(outpointBytes, dbOutpoint)
	if err != nil {
		return nil, err
	}

	return serialization.DbOutpointToDomainOutpoint(dbOutpoint)
}

func (ps *pruningStore) deserializeUTXOEntry(entryBytes []byte) (externalapi.UTXOEntry, error) {
	dbEntry := &serialization.DbUtxoEntry{}
	err := proto.Unmarshal(entryBytes, dbEntry)
	if err != nil {
		return nil, err
	}
	return serialization.DBUTXOEntryToUTXOEntry(dbEntry)
}

func (ps *pruningStore) ClearCandidatePruningPointMultiset(dbTx model.DBTransaction) error {
	return dbTx.Delete(candidatePruningPointMultiset)
}

func (ps *pruningStore) CandidatePruningPointMultiset(dbContext model.DBReader) (model.Multiset, error) {
	multisetBytes, err := dbContext.Get(candidatePruningPointMultiset)
	if err != nil {
		return nil, err
	}
	return ps.deserializeMultiset(multisetBytes)
}

func (ps *pruningStore) UpdateCandidatePruningPointMultiset(dbTx model.DBTransaction, multiset model.Multiset) error {
	multisetBytes, err := ps.serializeMultiset(multiset)
	if err != nil {
		return err
	}
	return dbTx.Put(candidatePruningPointMultiset, multisetBytes)
}

func (ps *pruningStore) serializeMultiset(multiset model.Multiset) ([]byte, error) {
	return proto.Marshal(serialization.MultisetToDBMultiset(multiset))
}

func (ps *pruningStore) deserializeMultiset(multisetBytes []byte) (model.Multiset, error) {
	dbMultiset := &serialization.DbMultiset{}
	err := proto.Unmarshal(multisetBytes, dbMultiset)
	if err != nil {
		return nil, err
	}

	return serialization.DBMultisetToMultiset(dbMultiset)
}
