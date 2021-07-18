package utxoindex

import (
	"encoding/binary"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/infrastructure/db/database"
)

var utxoIndexBucket = database.MakeBucket([]byte("utxo-index"))
var virtualParentsKey = database.MakeBucket([]byte("")).Key([]byte("utxo-index-virtual-parents"))

type utxoIndexStore struct {
	database database.Database
	batch    database.Transaction
}

func newUTXOIndexStore(database database.Database) *utxoIndexStore {
	return &utxoIndexStore{database: database}
}

func (uis *utxoIndexStore) StartBatch() error {
	var err error
	uis.batch, err = uis.database.Begin()
	return err
}

func (uis *utxoIndexStore) RollbackUnlessClosed() {
	uis.batch.RollbackUnlessClosed()
	uis.batch = nil
}

func (uis *utxoIndexStore) Commit() error {
	err := uis.batch.Commit()
	return err
}

func (uis *utxoIndexStore) put(pair *externalapi.OutpointAndUTXOEntryPair) error {
	key, err := uis.makeKey(pair)
	if err != nil {
		return err
	}
	value, err := serializeUTXOEntry(pair.UTXOEntry)
	if err != nil {
		return err
	}
	return uis.batch.Put(key, value)
}

func (uis *utxoIndexStore) delete(pair *externalapi.OutpointAndUTXOEntryPair) error {
	key, err := uis.makeKey(pair)
	if err != nil {
		return err
	}
	return uis.batch.Delete(key)
}

func (uis *utxoIndexStore) putVirtualParents(virtualParents []*externalapi.DomainHash) error {
	return uis.database.Put(virtualParentsKey, serializeHashes(virtualParents))
}

func (uis *utxoIndexStore) makeSPKBucket(scriptPublicKey *externalapi.ScriptPublicKey) *database.Bucket {
	var scriptPublicKeyBytes = make([]byte, 2+len(scriptPublicKey.Script)) // uint16
	binary.LittleEndian.PutUint16(scriptPublicKeyBytes[:2], scriptPublicKey.Version)
	copy(scriptPublicKeyBytes[2:], scriptPublicKey.Script)
	return utxoIndexBucket.Bucket(scriptPublicKeyBytes)
}

func (uis *utxoIndexStore) makeKey(pair *externalapi.OutpointAndUTXOEntryPair) (*database.Key, error) {
	serializedOutpoint, err := serializeOutpoint(pair.Outpoint)
	if err != nil {
		return nil, err
	}
	return uis.makeSPKBucket(pair.UTXOEntry.ScriptPublicKey()).Key(serializedOutpoint), nil
}

func (uis *utxoIndexStore) getUTXOsByScriptPublicKey(scriptPublicKey *externalapi.ScriptPublicKey) (UTXOMap, error) {

	cursor, err := uis.database.Cursor(uis.makeSPKBucket(scriptPublicKey))
	if err != nil {
		return nil, err
	}
	defer cursor.Close()

	utxoOutpointEntryPairs := make(UTXOMap)
	for cursor.Next() {
		key, err := cursor.Key()
		if err != nil {
			return nil, err
		}
		outpoint, err := deserializeOutpoint(key.Suffix())
		if err != nil {
			return nil, err
		}
		serializedUTXOEntry, err := cursor.Value()
		if err != nil {
			return nil, err
		}
		utxoEntry, err := deserializeUTXOEntry(serializedUTXOEntry)
		if err != nil {
			return nil, err
		}
		utxoOutpointEntryPairs[*outpoint] = utxoEntry
	}
	return utxoOutpointEntryPairs, nil
}

func (uis *utxoIndexStore) getVirtualParents() ([]*externalapi.DomainHash, error) {

	serializedHashes, err := uis.database.Get(virtualParentsKey)
	if err != nil {
		return nil, err
	}

	return deserializeHashes(serializedHashes)
}

func (uis *utxoIndexStore) clear() error {
	// First we delete the virtual parents, so if anything goes wrong, the UTXO index will be marked as "not synced"
	// and will be reset.
	barch := uis.batch
	cursor, err := barch.Cursor(utxoIndexBucket)
	if err != nil {
		return err
	}
	defer cursor.Close()
	for cursor.Next() {
		key, err := cursor.Key()
		if err != nil {
			return err
		}
		err = barch.Delete(key)
		if err != nil {
			return err
		}
	}
	return nil
}
