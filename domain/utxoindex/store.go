package utxoindex

import (
	"encoding/hex"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/infrastructure/db/database"
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

func (uis *utxoIndexStore) add(scriptPublicKey []byte, outpoint *externalapi.DomainOutpoint, utxoEntry *externalapi.UTXOEntry) {
	key := uis.convertScriptPublicKeyToHexString(scriptPublicKey)
	if _, ok := uis.toAdd[key]; !ok {
		uis.toAdd[key] = make(toAddUTXOOutpointEntryPairs)
	}
	toAddPair := uis.toAdd[key]
	toAddPair[*outpoint] = *utxoEntry
}

func (uis *utxoIndexStore) remove(scriptPublicKey []byte, outpoint *externalapi.DomainOutpoint) {
	key := uis.convertScriptPublicKeyToHexString(scriptPublicKey)
	uis.toRemove[key][*outpoint] = struct{}{}
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
