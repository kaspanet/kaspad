package utxoindex

import (
	"encoding/hex"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/infrastructure/db/database"
	"github.com/pkg/errors"
)

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
	uis.discard()
	return nil
}

func (uis *utxoIndexStore) convertScriptPublicKeyToHexString(scriptPublicKey []byte) scriptPublicKeyHexString {
	return scriptPublicKeyHexString(hex.EncodeToString(scriptPublicKey))
}
