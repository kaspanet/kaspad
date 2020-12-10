package utxoindex

import (
	"encoding/hex"
	"github.com/golang/protobuf/proto"
	"github.com/kaspanet/kaspad/domain/consensus/database/serialization"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/infrastructure/db/database"
	"github.com/pkg/errors"
)

var utxoIndexBucket = database.MakeBucket([]byte("utxo-index"))

type scriptPublicKeyHexString string
type toAddUTXOOutpointEntryPairs map[externalapi.DomainOutpoint]externalapi.UTXOEntry
type toRemoveUTXOOutpoints map[externalapi.DomainOutpoint]interface{}

type utxoIndexStore struct {
	database database.Database
	toAdd    map[scriptPublicKeyHexString]toAddUTXOOutpointEntryPairs
	toRemove map[scriptPublicKeyHexString]toRemoveUTXOOutpoints
}

func newUTXOIndexStore(database database.Database) *utxoIndexStore {
	return &utxoIndexStore{
		database: database,
		toAdd:    make(map[scriptPublicKeyHexString]toAddUTXOOutpointEntryPairs),
		toRemove: make(map[scriptPublicKeyHexString]toRemoveUTXOOutpoints),
	}
}

func (uis *utxoIndexStore) add(scriptPublicKey []byte, outpoint *externalapi.DomainOutpoint, utxoEntry *externalapi.UTXOEntry) error {
	key := uis.convertScriptPublicKeyToHexString(scriptPublicKey)

	// If the outpoint exists in `toRemove` simply remove it from there and return
	if toRemoveOutpointsOfKey, ok := uis.toRemove[key]; ok {
		if _, ok := toRemoveOutpointsOfKey[*outpoint]; ok {
			delete(toRemoveOutpointsOfKey, *outpoint)
			return nil
		}
	}

	// Create a toAddUTXOOutpointEntryPairs entry in `toAdd` if it doesn't exist
	if _, ok := uis.toAdd[key]; !ok {
		uis.toAdd[key] = make(toAddUTXOOutpointEntryPairs)
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
	key := uis.convertScriptPublicKeyToHexString(scriptPublicKey)

	// If the outpoint exists in `toAdd` simply remove it from there and return
	if toAddPairsOfKey, ok := uis.toAdd[key]; ok {
		if _, ok := toAddPairsOfKey[*outpoint]; ok {
			delete(toAddPairsOfKey, *outpoint)
			return nil
		}
	}

	// Create a toRemoveUTXOOutpoints entry in `toRemove` if it doesn't exist
	if _, ok := uis.toRemove[key]; !ok {
		uis.toRemove[key] = make(toRemoveUTXOOutpoints)
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
	uis.toAdd = make(map[scriptPublicKeyHexString]toAddUTXOOutpointEntryPairs)
	uis.toRemove = make(map[scriptPublicKeyHexString]toRemoveUTXOOutpoints)
}

func (uis *utxoIndexStore) commit() error {
	dbTransaction, err := uis.database.Begin()
	if err != nil {
		return err
	}
	defer dbTransaction.RollbackUnlessClosed()

	for scriptPublicKeyHexString, toRemoveOutpointsOfKey := range uis.toRemove {
		scriptPublicKey, err := uis.convertHexStringToScriptPublicKey(scriptPublicKeyHexString)
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
		scriptPublicKey, err := uis.convertHexStringToScriptPublicKey(scriptPublicKeyHexString)
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

func (uis *utxoIndexStore) convertScriptPublicKeyToHexString(scriptPublicKey []byte) scriptPublicKeyHexString {
	return scriptPublicKeyHexString(hex.EncodeToString(scriptPublicKey))
}

func (uis *utxoIndexStore) convertHexStringToScriptPublicKey(hexString scriptPublicKeyHexString) ([]byte, error) {
	return hex.DecodeString(string(hexString))
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
