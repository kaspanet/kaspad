package utxoindex

import (
	"github.com/golang/protobuf/proto"
	"github.com/kaspanet/kaspad/domain/consensus/database/serialization"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/infrastructure/db/database"
	"github.com/kaspanet/kaspad/infrastructure/logger"
	"github.com/pkg/errors"
)

var utxoIndexBucket = database.MakeBucket([]byte("utxo-index"))
var utxoIndexLastVirtualSelectedParentKey = database.MakeBucket().Key([]byte("utxo-index-last-virtual-selected-parent"))

type utxoIndexStore struct {
	database              database.Database
	toAdd                 map[ScriptPublicKeyString]UTXOOutpointEntryPairs
	toRemove              map[ScriptPublicKeyString]UTXOOutpoints
	virtualSelectedParent *externalapi.DomainHash
}

func newUTXOIndexStore(database database.Database) *utxoIndexStore {
	return &utxoIndexStore{
		database: database,
		toAdd:    make(map[ScriptPublicKeyString]UTXOOutpointEntryPairs),
		toRemove: make(map[ScriptPublicKeyString]UTXOOutpoints),
	}
}

func (uis *utxoIndexStore) add(scriptPublicKey []byte, outpoint *externalapi.DomainOutpoint, utxoEntry externalapi.UTXOEntry) error {
	key := ConvertScriptPublicKeyToString(scriptPublicKey)
	log.Tracef("Adding outpoint %s:%d to scriptPublicKey %s",
		outpoint.TransactionID, outpoint.Index, key)

	// If the outpoint exists in `toRemove` simply remove it from there and return
	if toRemoveOutpointsOfKey, ok := uis.toRemove[key]; ok {
		if _, ok := toRemoveOutpointsOfKey[*outpoint]; ok {
			log.Tracef("Outpoint %s:%d exists in `toRemove`. Deleting it from there",
				outpoint.TransactionID, outpoint.Index)
			delete(toRemoveOutpointsOfKey, *outpoint)
			return nil
		}
	}

	// Create a UTXOOutpointEntryPairs entry in `toAdd` if it doesn't exist
	if _, ok := uis.toAdd[key]; !ok {
		log.Tracef("Creating key %s in `toAdd`", key)
		uis.toAdd[key] = make(UTXOOutpointEntryPairs)
	}

	// Return an error if the outpoint already exists in `toAdd`
	toAddPairsOfKey := uis.toAdd[key]
	if _, ok := toAddPairsOfKey[*outpoint]; ok {
		return errors.Errorf("cannot add outpoint %s because it's being added already", outpoint)
	}

	toAddPairsOfKey[*outpoint] = utxoEntry

	log.Tracef("Added outpoint %s:%d to scriptPublicKey %s",
		outpoint.TransactionID, outpoint.Index, key)
	return nil
}

func (uis *utxoIndexStore) remove(scriptPublicKey []byte, outpoint *externalapi.DomainOutpoint) error {
	key := ConvertScriptPublicKeyToString(scriptPublicKey)
	log.Tracef("Removing outpoint %s:%d from scriptPublicKey %s",
		outpoint.TransactionID, outpoint.Index, key)

	// If the outpoint exists in `toAdd` simply remove it from there and return
	if toAddPairsOfKey, ok := uis.toAdd[key]; ok {
		if _, ok := toAddPairsOfKey[*outpoint]; ok {
			log.Tracef("Outpoint %s:%d exists in `toAdd`. Deleting it from there",
				outpoint.TransactionID, outpoint.Index)
			delete(toAddPairsOfKey, *outpoint)
			return nil
		}
	}

	// Create a UTXOOutpoints entry in `toRemove` if it doesn't exist
	if _, ok := uis.toRemove[key]; !ok {
		log.Tracef("Creating key %s in `toRemove`", key)
		uis.toRemove[key] = make(UTXOOutpoints)
	}

	// Return an error if the outpoint already exists in `toRemove`
	toRemoveOutpointsOfKey := uis.toRemove[key]
	if _, ok := toRemoveOutpointsOfKey[*outpoint]; ok {
		return errors.Errorf("cannot remove outpoint %s because it's being removed already", outpoint)
	}

	toRemoveOutpointsOfKey[*outpoint] = struct{}{}

	log.Tracef("Removed outpoint %s:%d from scriptPublicKey %s",
		outpoint.TransactionID, outpoint.Index, key)
	return nil
}

func (uis *utxoIndexStore) discard() {
	uis.toAdd = make(map[ScriptPublicKeyString]UTXOOutpointEntryPairs)
	uis.toRemove = make(map[ScriptPublicKeyString]UTXOOutpoints)
	uis.virtualSelectedParent = nil
}

func (uis *utxoIndexStore) commit() error {
	onEnd := logger.LogAndMeasureExecutionTime(log, "utxoIndexStore.commit")
	defer onEnd()

	dbTransaction, err := uis.database.Begin()
	if err != nil {
		return err
	}
	defer dbTransaction.RollbackUnlessClosed()

	for scriptPublicKeyString, toRemoveOutpointsOfKey := range uis.toRemove {
		scriptPublicKey := ConvertStringToScriptPublicKey(scriptPublicKeyString)
		bucket := uis.bucketForScriptPublicKey(scriptPublicKey)
		for outpointToRemove := range toRemoveOutpointsOfKey {
			key, err := uis.convertOutpointToKey(bucket, &outpointToRemove)
			if err != nil {
				return err
			}
			err = dbTransaction.Delete(key)
			if err != nil {
				return err
			}
		}
	}

	for scriptPublicKeyString, toAddUTXOOutpointEntryPairs := range uis.toAdd {
		scriptPublicKey := ConvertStringToScriptPublicKey(scriptPublicKeyString)
		bucket := uis.bucketForScriptPublicKey(scriptPublicKey)
		for outpointToAdd, utxoEntryToAdd := range toAddUTXOOutpointEntryPairs {
			key, err := uis.convertOutpointToKey(bucket, &outpointToAdd)
			if err != nil {
				return err
			}
			serializedUTXOEntry, err := uis.serializeUTXOEntry(utxoEntryToAdd)
			if err != nil {
				return err
			}
			err = dbTransaction.Put(key, serializedUTXOEntry)
			if err != nil {
				return err
			}
		}
	}

	err = dbTransaction.Put(utxoIndexLastVirtualSelectedParentKey, uis.virtualSelectedParent.ByteSlice())
	if err != nil {
		return err
	}

	err = dbTransaction.Commit()
	if err != nil {
		return err
	}

	uis.discard()
	return nil
}

func (uis *utxoIndexStore) bucketForScriptPublicKey(scriptPublicKey []byte) *database.Bucket {
	return utxoIndexBucket.Bucket(scriptPublicKey)
}

func (uis *utxoIndexStore) convertOutpointToKey(bucket *database.Bucket, outpoint *externalapi.DomainOutpoint) (*database.Key, error) {
	serializedOutpoint, err := uis.serializeOutpoint(outpoint)
	if err != nil {
		return nil, err
	}
	return bucket.Key(serializedOutpoint), nil
}

func (uis *utxoIndexStore) convertKeyToOutpoint(key *database.Key) (*externalapi.DomainOutpoint, error) {
	serializedOutpoint := key.Suffix()
	return uis.deserializeOutpoint(serializedOutpoint)
}

func (uis *utxoIndexStore) serializeOutpoint(outpoint *externalapi.DomainOutpoint) ([]byte, error) {
	dbOutpoint := serialization.DomainOutpointToDbOutpoint(outpoint)
	return proto.Marshal(dbOutpoint)
}

func (uis *utxoIndexStore) deserializeOutpoint(serializedOutpoint []byte) (*externalapi.DomainOutpoint, error) {
	var dbOutpoint serialization.DbOutpoint
	err := proto.Unmarshal(serializedOutpoint, &dbOutpoint)
	if err != nil {
		return nil, err
	}
	return serialization.DbOutpointToDomainOutpoint(&dbOutpoint)
}

func (uis *utxoIndexStore) serializeUTXOEntry(utxoEntry externalapi.UTXOEntry) ([]byte, error) {
	dbUTXOEntry := serialization.UTXOEntryToDBUTXOEntry(utxoEntry)
	return proto.Marshal(dbUTXOEntry)
}

func (uis *utxoIndexStore) deserializeUTXOEntry(serializedUTXOEntry []byte) (externalapi.UTXOEntry, error) {
	var dbUTXOEntry serialization.DbUtxoEntry
	err := proto.Unmarshal(serializedUTXOEntry, &dbUTXOEntry)
	if err != nil {
		return nil, err
	}
	return serialization.DBUTXOEntryToUTXOEntry(&dbUTXOEntry), nil
}

func (uis *utxoIndexStore) stagedData() (
	toAdd map[ScriptPublicKeyString]UTXOOutpointEntryPairs,
	toRemove map[ScriptPublicKeyString]UTXOOutpoints) {

	toAddClone := make(map[ScriptPublicKeyString]UTXOOutpointEntryPairs, len(uis.toAdd))
	for scriptPublicKeyString, toAddUTXOOutpointEntryPairs := range uis.toAdd {
		toAddUTXOOutpointEntryPairsClone := make(UTXOOutpointEntryPairs, len(toAddUTXOOutpointEntryPairs))
		for outpoint, utxoEntry := range toAddUTXOOutpointEntryPairs {
			toAddUTXOOutpointEntryPairsClone[outpoint] = utxoEntry
		}
		toAddClone[scriptPublicKeyString] = toAddUTXOOutpointEntryPairsClone
	}

	toRemoveClone := make(map[ScriptPublicKeyString]UTXOOutpoints, len(uis.toRemove))
	for scriptPublicKeyString, toRemoveOutpoints := range uis.toRemove {
		toRemoveOutpointsClone := make(UTXOOutpoints, len(toRemoveOutpoints))
		for outpoint := range toRemoveOutpoints {
			toRemoveOutpointsClone[outpoint] = struct{}{}
		}
		toRemoveClone[scriptPublicKeyString] = toRemoveOutpointsClone
	}

	return toAddClone, toRemoveClone
}

func (uis *utxoIndexStore) getUTXOOutpointEntryPairs(scriptPublicKey []byte) (UTXOOutpointEntryPairs, error) {
	if uis.isAnythingStaged() {
		return nil, errors.Errorf("cannot get utxo outpoint entry pairs while staging isn't empty")
	}

	bucket := uis.bucketForScriptPublicKey(scriptPublicKey)
	cursor, err := uis.database.Cursor(bucket)
	if err != nil {
		return nil, err
	}
	utxoOutpointEntryPairs := make(UTXOOutpointEntryPairs)
	for cursor.Next() {
		key, err := cursor.Key()
		if err != nil {
			return nil, err
		}
		outpoint, err := uis.convertKeyToOutpoint(key)
		if err != nil {
			return nil, err
		}
		serializedUTXOEntry, err := cursor.Value()
		if err != nil {
			return nil, err
		}
		utxoEntry, err := uis.deserializeUTXOEntry(serializedUTXOEntry)
		if err != nil {
			return nil, err
		}
		utxoOutpointEntryPairs[*outpoint] = utxoEntry
	}
	return utxoOutpointEntryPairs, nil
}

func (uis *utxoIndexStore) getLastVirtualSelectedParent() (*externalapi.DomainHash, bool, error) {
	if uis.isAnythingStaged() {
		return nil, false, errors.Errorf("cannot get last selected tip while staging isn't empty")
	}

	hasLastVirtualSelectedParent, err := uis.database.Has(utxoIndexLastVirtualSelectedParentKey)
	if err != nil {
		return nil, false, err
	}

	if !hasLastVirtualSelectedParent {
		return nil, false, nil
	}

	lastVirtualSelectedParentBytes, err := uis.database.Get(utxoIndexLastVirtualSelectedParentKey)
	if err != nil {
		return nil, false, err
	}

	lastVirtualSelectedParent, err := externalapi.NewDomainHashFromByteSlice(lastVirtualSelectedParentBytes)
	if err != nil {
		return nil, false, err
	}

	return lastVirtualSelectedParent, true, nil
}

func (uis *utxoIndexStore) isAnythingStaged() bool {
	return len(uis.toAdd) > 0 || len(uis.toRemove) > 0 || uis.virtualSelectedParent != nil
}

func (uis *utxoIndexStore) replaceUTXOSet(utxoSet []*externalapi.OutpointUTXOPair,
	virtualSelectedParent *externalapi.DomainHash) error {

	onEnd := logger.LogAndMeasureExecutionTime(log, "utxoIndexStore.replaceUTXOSet")
	defer onEnd()

	if uis.isAnythingStaged() {
		return errors.Errorf("cannot replace utxo set while something is staged")
	}

	err := uis.resetStore()
	if err != nil {
		return err
	}

	uis.virtualSelectedParent = virtualSelectedParent
	for _, pair := range utxoSet {
		err := uis.add(pair.Entry.ScriptPublicKey(), pair.Outpoint, pair.Entry)
		if err != nil {
			return err
		}
	}

	return uis.commit()
}

func (uis *utxoIndexStore) resetStore() error {
	onEnd := logger.LogAndMeasureExecutionTime(log, "utxoIndexStore.resetStore")
	defer onEnd()

	cursor, err := uis.database.Cursor(utxoIndexBucket)
	if err != nil {
		return err
	}

	keysToDelete := make([]*database.Key, 0)
	for cursor.Next() {
		key, err := cursor.Key()
		if err != nil {
			return err
		}

		keysToDelete = append(keysToDelete, key)
	}

	for _, key := range keysToDelete {
		err = uis.database.Delete(key)
		if err != nil {
			return err
		}
	}

	err = uis.database.Delete(utxoIndexLastVirtualSelectedParentKey)
	if err != nil {
		return err
	}

	return nil
}
