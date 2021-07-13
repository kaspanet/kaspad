package utxoindex

import (
	"encoding/binary"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/infrastructure/db/database"
	"github.com/kaspanet/kaspad/infrastructure/logger"
	"github.com/pkg/errors"
)

var utxoIndexBucket = database.MakeBucket([]byte("utxo-index"))
var virtualParentsKey = database.MakeBucket([]byte("")).Key([]byte("utxo-index-virtual-parents"))

type utxoIndexStore struct {
	database       database.Database
	toAdd          AddressesUTXOMap
	toRemove       AddressesUTXOMap
	virtualParents []*externalapi.DomainHash
}

func newUTXOIndexStore(database database.Database) *utxoIndexStore {
	return &utxoIndexStore{
		database: database,
		toAdd:    make(AddressesUTXOMap),
		toRemove: make(AddressesUTXOMap),
	}
}

func (uis *utxoIndexStore) updateVirtualParents(virtualParents []*externalapi.DomainHash) {
	uis.virtualParents = virtualParents
}

func (uis *utxoIndexStore) discard() {
	uis.toAdd = make(AddressesUTXOMap)
	uis.toRemove = make(AddressesUTXOMap)
	uis.virtualParents = nil
}

func (uis *utxoIndexStore) commit() error {
	onEnd := logger.LogAndMeasureExecutionTime(log, "utxoIndexStore.commit")
	defer onEnd()

	dbTransaction, err := uis.database.Begin()
	if err != nil {
		return err
	}
	defer dbTransaction.RollbackUnlessClosed()

	for _, set := range []*AddressesUTXOMap{&uis.toRemove, &uis.toAdd} {
		for KeyString, utxoMapOrNils := range *set {
			bucket := uis.bucketForScriptPublicKey(ConvertStringToScriptPublicKey(KeyString))
			for outpoint, utxoEntry := range utxoMapOrNils {
				key, err := uis.convertOutpointToKey(bucket, &outpoint)
				if err != nil {
					return err
				}
				if set == &uis.toRemove {
					err = dbTransaction.Delete(key)
					if err != nil {
						return err
					}
				} else {
					serializedUTXOEntry, err := serializeUTXOEntry(utxoEntry)
					if err != nil {
						return err
					}
					err = dbTransaction.Put(key, serializedUTXOEntry)
					if err != nil {
						return err
					}
				}
			}
		}
	}

	err = dbTransaction.Put(virtualParentsKey, serializeHashes(uis.virtualParents))
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

func (uis *utxoIndexStore) addAndCommitOutpointsWithoutTransaction(utxoPairs []*externalapi.OutpointAndUTXOEntryPair) error {
	for _, pair := range utxoPairs {
		bucket := uis.bucketForScriptPublicKey(pair.UTXOEntry.ScriptPublicKey())
		key, err := uis.convertOutpointToKey(bucket, pair.Outpoint)
		if err != nil {
			return err
		}
		serializedUTXOEntry, err := serializeUTXOEntry(pair.UTXOEntry)
		if err != nil {
			return err
		}
		err = uis.database.Put(key, serializedUTXOEntry)
		if err != nil {
			return err
		}
	}

	return nil
}

func (uis *utxoIndexStore) updateAndCommitVirtualParentsWithoutTransaction(virtualParents []*externalapi.DomainHash) error {
	serializeParentHashes := serializeHashes(virtualParents)
	return uis.database.Put(virtualParentsKey, serializeParentHashes)
}

func (uis *utxoIndexStore) bucketForScriptPublicKey(scriptPublicKey *externalapi.ScriptPublicKey) *database.Bucket {
	var scriptPublicKeyBytes = make([]byte, 2+len(scriptPublicKey.Script)) // uint16
	binary.LittleEndian.PutUint16(scriptPublicKeyBytes[:2], scriptPublicKey.Version)
	copy(scriptPublicKeyBytes[2:], scriptPublicKey.Script)
	return utxoIndexBucket.Bucket(scriptPublicKeyBytes)
}

func (uis *utxoIndexStore) convertOutpointToKey(bucket *database.Bucket, outpoint *externalapi.DomainOutpoint) (*database.Key, error) {
	serializedOutpoint, err := serializeOutpoint(outpoint)
	if err != nil {
		return nil, err
	}
	return bucket.Key(serializedOutpoint), nil
}

func (uis *utxoIndexStore) convertKeyToOutpoint(key *database.Key) (*externalapi.DomainOutpoint, error) {
	serializedOutpoint := key.Suffix()
	return deserializeOutpoint(serializedOutpoint)
}

func (uis *utxoIndexStore) stagedData() (
	toAdd AddressesUTXOMap,
	toRemove AddressesUTXOMap,
	virtualParents []*externalapi.DomainHash) {

	toAddClone := make(map[KeyString]UTXOMap, len(uis.toAdd))
	for keyString, toAddUTXOOutpointEntryPairs := range uis.toAdd {
		toAddUTXOOutpointEntryPairsClone := make(UTXOMap, len(toAddUTXOOutpointEntryPairs))
		for outpoint, utxoEntry := range toAddUTXOOutpointEntryPairs {
			toAddUTXOOutpointEntryPairsClone[outpoint] = utxoEntry
		}
		toAddClone[keyString] = toAddUTXOOutpointEntryPairsClone
	}

	toRemoveClone := make(AddressesUTXOMap, len(uis.toRemove))
	for keyString, toRemoveOutpoints := range uis.toRemove {
		toRemoveOutpointsClone := make(UTXOMap, len(toRemoveOutpoints))
		for outpoint := range toRemoveOutpoints {
			toRemoveOutpointsClone[outpoint] = nil
		}
		toRemoveClone[keyString] = toRemoveOutpointsClone
	}

	return toAddClone, toRemoveClone, uis.virtualParents
}

func (uis *utxoIndexStore) isAnythingStaged() bool {
	return len(uis.toAdd) > 0 || len(uis.toRemove) > 0
}

func (uis *utxoIndexStore) getUTXOOutpointEntryPairs(scriptPublicKey *externalapi.ScriptPublicKey) (UTXOMap, error) {
	if uis.isAnythingStaged() {
		return nil, errors.Errorf("cannot get utxo outpoint entry pairs while staging isn't empty")
	}

	bucket := uis.bucketForScriptPublicKey(scriptPublicKey)
	cursor, err := uis.database.Cursor(bucket)
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
		outpoint, err := uis.convertKeyToOutpoint(key)
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
	if uis.isAnythingStaged() {
		return nil, errors.Errorf("cannot get the virtual parents while staging isn't empty")
	}

	serializedHashes, err := uis.database.Get(virtualParentsKey)
	if err != nil {
		return nil, err
	}

	return deserializeHashes(serializedHashes)
}

func (uis *utxoIndexStore) deleteAll() error {
	// First we delete the virtual parents, so if anything goes wrong, the UTXO index will be marked as "not synced"
	// and will be reset.
	err := uis.database.Delete(virtualParentsKey)
	if err != nil {
		return err
	}

	cursor, err := uis.database.Cursor(utxoIndexBucket)
	if err != nil {
		return err
	}
	defer cursor.Close()
	for cursor.Next() {
		key, err := cursor.Key()
		if err != nil {
			return err
		}

		err = uis.database.Delete(key)
		if err != nil {
			return err
		}
	}

	return nil
}
