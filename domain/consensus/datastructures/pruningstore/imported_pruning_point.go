package pruningstore

import (
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
	"github.com/zoomy-network/zoomyd/domain/consensus/database/serialization"
	"github.com/zoomy-network/zoomyd/domain/consensus/model"
	"github.com/zoomy-network/zoomyd/domain/consensus/model/externalapi"
)

var importedPruningPointUTXOsBucketName = []byte("imported-pruning-point-utxos")
var importedPruningPointMultisetKeyName = []byte("imported-pruning-point-multiset")

func (ps *pruningStore) ClearImportedPruningPointUTXOs(dbContext model.DBWriter) error {
	cursor, err := dbContext.Cursor(ps.importedPruningPointUTXOsBucket)
	if err != nil {
		return err
	}
	defer cursor.Close()

	for ok := cursor.First(); ok; ok = cursor.Next() {
		key, err := cursor.Key()
		if err != nil {
			return err
		}
		err = dbContext.Delete(key)
		if err != nil {
			return err
		}
	}
	return nil
}

func (ps *pruningStore) AppendImportedPruningPointUTXOs(dbTx model.DBTransaction,
	outpointAndUTXOEntryPairs []*externalapi.OutpointAndUTXOEntryPair) error {

	for _, outpointAndUTXOEntryPair := range outpointAndUTXOEntryPairs {
		key, err := ps.importedPruningPointUTXOKey(outpointAndUTXOEntryPair.Outpoint)
		if err != nil {
			return err
		}
		serializedUTXOEntry, err := serializeUTXOEntry(outpointAndUTXOEntryPair.UTXOEntry)
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

func (ps *pruningStore) ImportedPruningPointUTXOIterator(dbContext model.DBReader) (externalapi.ReadOnlyUTXOSetIterator, error) {
	cursor, err := dbContext.Cursor(ps.importedPruningPointUTXOsBucket)
	if err != nil {
		return nil, err
	}
	return ps.newCursorUTXOSetIterator(cursor), nil
}

type utxoSetIterator struct {
	cursor   model.DBCursor
	isClosed bool
}

func (ps *pruningStore) newCursorUTXOSetIterator(cursor model.DBCursor) externalapi.ReadOnlyUTXOSetIterator {
	return &utxoSetIterator{cursor: cursor}
}

func (u *utxoSetIterator) First() bool {
	if u.isClosed {
		panic("Tried using a closed utxoSetIterator")
	}
	return u.cursor.First()
}

func (u *utxoSetIterator) Next() bool {
	if u.isClosed {
		panic("Tried using a closed utxoSetIterator")
	}
	return u.cursor.Next()
}

func (u *utxoSetIterator) Get() (outpoint *externalapi.DomainOutpoint, utxoEntry externalapi.UTXOEntry, err error) {
	if u.isClosed {
		return nil, nil, errors.New("Tried using a closed utxoSetIterator")
	}
	key, err := u.cursor.Key()
	if err != nil {
		panic(err)
	}

	utxoEntryBytes, err := u.cursor.Value()
	if err != nil {
		return nil, nil, err
	}

	outpoint, err = deserializeOutpoint(key.Suffix())
	if err != nil {
		return nil, nil, err
	}

	utxoEntry, err = deserializeUTXOEntry(utxoEntryBytes)
	if err != nil {
		return nil, nil, err
	}

	return outpoint, utxoEntry, nil
}

func (u *utxoSetIterator) Close() error {
	if u.isClosed {
		return errors.New("Tried using a closed utxoSetIterator")
	}
	u.isClosed = true
	err := u.cursor.Close()
	if err != nil {
		return err
	}
	u.cursor = nil
	return nil
}

func (ps *pruningStore) importedPruningPointUTXOKey(outpoint *externalapi.DomainOutpoint) (model.DBKey, error) {
	serializedOutpoint, err := serializeOutpoint(outpoint)
	if err != nil {
		return nil, err
	}

	return ps.importedPruningPointUTXOsBucket.Key(serializedOutpoint), nil
}

func serializeOutpoint(outpoint *externalapi.DomainOutpoint) ([]byte, error) {
	return proto.Marshal(serialization.DomainOutpointToDbOutpoint(outpoint))
}

func serializeUTXOEntry(entry externalapi.UTXOEntry) ([]byte, error) {
	return proto.Marshal(serialization.UTXOEntryToDBUTXOEntry(entry))
}

func deserializeOutpoint(outpointBytes []byte) (*externalapi.DomainOutpoint, error) {
	dbOutpoint := &serialization.DbOutpoint{}
	err := proto.Unmarshal(outpointBytes, dbOutpoint)
	if err != nil {
		return nil, err
	}

	return serialization.DbOutpointToDomainOutpoint(dbOutpoint)
}

func deserializeUTXOEntry(entryBytes []byte) (externalapi.UTXOEntry, error) {
	dbEntry := &serialization.DbUtxoEntry{}
	err := proto.Unmarshal(entryBytes, dbEntry)
	if err != nil {
		return nil, err
	}
	return serialization.DBUTXOEntryToUTXOEntry(dbEntry)
}

func (ps *pruningStore) ClearImportedPruningPointMultiset(dbContext model.DBWriter) error {
	return dbContext.Delete(ps.importedPruningPointMultisetKey)
}

func (ps *pruningStore) ImportedPruningPointMultiset(dbContext model.DBReader) (model.Multiset, error) {
	multisetBytes, err := dbContext.Get(ps.importedPruningPointMultisetKey)
	if err != nil {
		return nil, err
	}
	return ps.deserializeMultiset(multisetBytes)
}

func (ps *pruningStore) UpdateImportedPruningPointMultiset(dbTx model.DBTransaction, multiset model.Multiset) error {
	multisetBytes, err := ps.serializeMultiset(multiset)
	if err != nil {
		return err
	}
	return dbTx.Put(ps.importedPruningPointMultisetKey, multisetBytes)
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

func (ps *pruningStore) CommitImportedPruningPointUTXOSet(dbContext model.DBWriter) error {
	// Delete all the old UTXOs from the database
	deleteCursor, err := dbContext.Cursor(ps.pruningPointUTXOSetBucket)
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
	insertCursor, err := dbContext.Cursor(ps.importedPruningPointUTXOsBucket)
	if err != nil {
		return err
	}
	defer insertCursor.Close()
	for ok := insertCursor.First(); ok; ok = insertCursor.Next() {
		importedPruningPointUTXOSetKey, err := insertCursor.Key()
		if err != nil {
			return err
		}
		pruningPointUTXOSetKey := ps.pruningPointUTXOSetBucket.Key(importedPruningPointUTXOSetKey.Suffix())

		serializedUTXOEntry, err := insertCursor.Value()
		if err != nil {
			return err
		}

		err = dbContext.Put(pruningPointUTXOSetKey, serializedUTXOEntry)
		if err != nil {
			return err
		}
	}

	return nil
}
