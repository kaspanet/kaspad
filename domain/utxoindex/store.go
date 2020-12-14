package utxoindex

import (
	"github.com/golang/protobuf/proto"
	"github.com/kaspanet/kaspad/domain/consensus/database/serialization"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/infrastructure/db/database"
	"github.com/pkg/errors"
)

var utxoIndexBucket = database.MakeBucket([]byte("utxo-index"))

type utxoIndexStore struct {
	database database.Database
	toAdd    map[ScriptPublicKeyHexString]UTXOOutpointEntryPairs
	toRemove map[ScriptPublicKeyHexString]UTXOOutpoints
}

func newUTXOIndexStore(database database.Database) *utxoIndexStore {
	return &utxoIndexStore{
		database: database,
		toAdd:    make(map[ScriptPublicKeyHexString]UTXOOutpointEntryPairs),
		toRemove: make(map[ScriptPublicKeyHexString]UTXOOutpoints),
	}
}

func (uis *utxoIndexStore) add(scriptPublicKey []byte, outpoint *externalapi.DomainOutpoint, utxoEntry *externalapi.UTXOEntry) error {
	key := ConvertScriptPublicKeyToHexString(scriptPublicKey)

	// If the outpoint exists in `toRemove` simply remove it from there and return
	if toRemoveOutpointsOfKey, ok := uis.toRemove[key]; ok {
		if _, ok := toRemoveOutpointsOfKey[*outpoint]; ok {
			delete(toRemoveOutpointsOfKey, *outpoint)
			return nil
		}
	}

	// Create a UTXOOutpointEntryPairs entry in `toAdd` if it doesn't exist
	if _, ok := uis.toAdd[key]; !ok {
		uis.toAdd[key] = make(UTXOOutpointEntryPairs)
	}

	// Return an error if the outpoint already exists in `toAdd`
	toAddPairsOfKey := uis.toAdd[key]
	if _, ok := toAddPairsOfKey[*outpoint]; ok {
		return errors.Errorf("cannot add outpoint %s because it's being added already", outpoint)
	}

	toAddPairsOfKey[*outpoint] = *utxoEntry
	return nil
}

func (uis *utxoIndexStore) remove(scriptPublicKey []byte, outpoint *externalapi.DomainOutpoint) error {
	key := ConvertScriptPublicKeyToHexString(scriptPublicKey)

	// If the outpoint exists in `toAdd` simply remove it from there and return
	if toAddPairsOfKey, ok := uis.toAdd[key]; ok {
		if _, ok := toAddPairsOfKey[*outpoint]; ok {
			delete(toAddPairsOfKey, *outpoint)
			return nil
		}
	}

	// Create a UTXOOutpoints entry in `toRemove` if it doesn't exist
	if _, ok := uis.toRemove[key]; !ok {
		uis.toRemove[key] = make(UTXOOutpoints)
	}

	// Return an error if the outpoint already exists in `toRemove`
	toRemoveOutpointsOfKey := uis.toRemove[key]
	if _, ok := toRemoveOutpointsOfKey[*outpoint]; ok {
		return errors.Errorf("cannot remove outpoint %s because it's being removed already", outpoint)
	}

	toRemoveOutpointsOfKey[*outpoint] = struct{}{}
	return nil
}

func (uis *utxoIndexStore) discard() {
	uis.toAdd = make(map[ScriptPublicKeyHexString]UTXOOutpointEntryPairs)
	uis.toRemove = make(map[ScriptPublicKeyHexString]UTXOOutpoints)
}

func (uis *utxoIndexStore) commit() error {
	dbTransaction, err := uis.database.Begin()
	if err != nil {
		return err
	}
	defer dbTransaction.RollbackUnlessClosed()

	for scriptPublicKeyHexString, toRemoveOutpointsOfKey := range uis.toRemove {
		scriptPublicKey, err := ConvertHexStringToScriptPublicKey(scriptPublicKeyHexString)
		if err != nil {
			return err
		}
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

	for scriptPublicKeyHexString, toAddUTXOOutpointEntryPairs := range uis.toAdd {
		scriptPublicKey, err := ConvertHexStringToScriptPublicKey(scriptPublicKeyHexString)
		if err != nil {
			return err
		}
		bucket := uis.bucketForScriptPublicKey(scriptPublicKey)
		for outpointToAdd, utxoEntryToAdd := range toAddUTXOOutpointEntryPairs {
			key, err := uis.convertOutpointToKey(bucket, &outpointToAdd)
			if err != nil {
				return err
			}
			serializedUTXOEntry, err := uis.serializeUTXOEntry(&utxoEntryToAdd)
			if err != nil {
				return err
			}
			err = dbTransaction.Put(key, serializedUTXOEntry)
			if err != nil {
				return err
			}
		}
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

func (uis *utxoIndexStore) serializeOutpoint(outpoint *externalapi.DomainOutpoint) ([]byte, error) {
	dbOutpoint := serialization.DomainOutpointToDbOutpoint(outpoint)
	return proto.Marshal(dbOutpoint)
}

func (uis *utxoIndexStore) serializeUTXOEntry(utxoEntry *externalapi.UTXOEntry) ([]byte, error) {
	dbUTXOEntry := serialization.UTXOEntryToDBUTXOEntry(*utxoEntry)
	return proto.Marshal(dbUTXOEntry)
}

func (uis *utxoIndexStore) stagedData() (
	toAdd map[ScriptPublicKeyHexString]UTXOOutpointEntryPairs,
	toRemove map[ScriptPublicKeyHexString]UTXOOutpoints) {

	toAddClone := make(map[ScriptPublicKeyHexString]UTXOOutpointEntryPairs, len(uis.toAdd))
	for scriptPublicKeyHexString, toAddUTXOOutpointEntryPairs := range uis.toAdd {
		toAddUTXOOutpointEntryPairsClone := make(UTXOOutpointEntryPairs, len(toAddUTXOOutpointEntryPairs))
		for outpoint, utxoEntry := range toAddUTXOOutpointEntryPairs {
			toAddUTXOOutpointEntryPairsClone[outpoint] = utxoEntry
		}
		toAddClone[scriptPublicKeyHexString] = toAddUTXOOutpointEntryPairsClone
	}

	toRemoveClone := make(map[ScriptPublicKeyHexString]UTXOOutpoints, len(uis.toRemove))
	for scriptPublicKeyHexString, toRemoveOutpoints := range uis.toRemove {
		toRemoveOutpointsClone := make(UTXOOutpoints, len(toRemoveOutpoints))
		for outpoint := range toRemoveOutpoints {
			toRemoveOutpointsClone[outpoint] = struct{}{}
		}
		toRemoveClone[scriptPublicKeyHexString] = toRemoveOutpointsClone
	}

	return toAddClone, toRemoveClone
}

func (uis *utxoIndexStore) getUTXOOutpointEntryPairs(ScriptPublicKeyHexString) (UTXOOutpointEntryPairs, error) {
	return nil, nil
}
