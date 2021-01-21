package pruningstore

import (
	"github.com/golang/protobuf/proto"
	"github.com/kaspanet/kaspad/domain/consensus/database"
	"github.com/kaspanet/kaspad/domain/consensus/database/serialization"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

var importedPruningPointUTXOsBucket = database.MakeBucket([]byte("imported-pruning-point-utxos"))
var importedPruningPointMultiset = database.MakeBucket(nil).Key([]byte("imported-pruning-point-multiset"))

func (ps *pruningStore) ClearImportedPruningPointUTXOs(dbContext model.DBWriter) error {
	cursor, err := dbContext.Cursor(importedPruningPointUTXOsBucket)
	if err != nil {
		return err
	}

	for cursor.Next() {
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

func (ps *pruningStore) InsertImportedPruningPointUTXOs(dbTx model.DBTransaction,
	outpointAndUTXOEntryPairs []*externalapi.OutpointAndUTXOEntryPair) error {

	for _, outpointAndUTXOEntryPair := range outpointAndUTXOEntryPairs {
		key, err := ps.importedPruningPointUTXOKey(outpointAndUTXOEntryPair.Outpoint)
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

func (ps *pruningStore) ImportedPruningPointUTXOIterator(dbContext model.DBReader) (model.ReadOnlyUTXOSetIterator, error) {
	cursor, err := dbContext.Cursor(importedPruningPointUTXOsBucket)
	if err != nil {
		return nil, err
	}
	return ps.newCursorUTXOSetIterator(cursor), nil
}

type utxoSetIterator struct {
	cursor model.DBCursor
}

func (ps *pruningStore) newCursorUTXOSetIterator(cursor model.DBCursor) model.ReadOnlyUTXOSetIterator {
	return &utxoSetIterator{cursor: cursor}
}

func (u *utxoSetIterator) First() bool {
	return u.cursor.First()
}

func (u *utxoSetIterator) Next() bool {
	return u.cursor.Next()
}

func (u *utxoSetIterator) Get() (outpoint *externalapi.DomainOutpoint, utxoEntry externalapi.UTXOEntry, err error) {
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

func (ps *pruningStore) importedPruningPointUTXOKey(outpoint *externalapi.DomainOutpoint) (model.DBKey, error) {
	serializedOutpoint, err := ps.serializeOutpoint(outpoint)
	if err != nil {
		return nil, err
	}

	return importedPruningPointUTXOsBucket.Key(serializedOutpoint), nil
}

func (ps *pruningStore) serializeOutpoint(outpoint *externalapi.DomainOutpoint) ([]byte, error) {
	return proto.Marshal(serialization.DomainOutpointToDbOutpoint(outpoint))
}

func (ps *pruningStore) serializeUTXOEntry(entry externalapi.UTXOEntry) ([]byte, error) {
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
	return dbContext.Delete(importedPruningPointMultiset)
}

func (ps *pruningStore) ImportedPruningPointMultiset(dbContext model.DBReader) (model.Multiset, error) {
	multisetBytes, err := dbContext.Get(importedPruningPointMultiset)
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
	return dbTx.Put(importedPruningPointMultiset, multisetBytes)
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
	deleteCursor, err := dbContext.Cursor(pruningPointUTXOSetBucket)
	if err != nil {
		return err
	}
	for deleteCursor.Next() {
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
	insertCursor, err := dbContext.Cursor(importedPruningPointUTXOsBucket)
	if err != nil {
		return err
	}
	for insertCursor.Next() {
		importedPruningPointUTXOSetKey, err := insertCursor.Key()
		if err != nil {
			return err
		}
		pruningPointUTXOSetKey := pruningPointUTXOSetBucket.Key(importedPruningPointUTXOSetKey.Suffix())

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
