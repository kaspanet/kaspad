package txindex

import (
	"encoding/binary"

	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/infrastructure/db/database"
	"github.com/kaspanet/kaspad/infrastructure/logger"
	"github.com/pkg/errors"
	"golang.org/x/exp/maps"
)

var txAcceptedIndexBucket = database.MakeBucket([]byte("tx-index"))
var addrIndexSentBucket = database.MakeBucket([]byte("addr-index-sent"))
var addrIndexReceivedBucket = database.MakeBucket([]byte("addr-index-received"))

var virtualParentsKey = database.MakeBucket([]byte("")).Key([]byte("tx-index-virtual-parent"))
var virtualBlueScoreKey = database.MakeBucket([]byte("")).Key([]byte("tx-index-virtual-bluescore"))
var pruningPointKey = database.MakeBucket([]byte("")).Key([]byte("tx-index-prunning-point"))

type txIndexStore struct {
	database       		database.Database
	toAddTxs          	TxChange
	toRemoveTxs     	TxChange
	toAddSent		AddrsChange
	toRemoveSent		AddrsChange
	toAddReceived		AddrsChange
	toRemoveReceived	AddrsChange
	virtualParents 		[]*externalapi.DomainHash
	virtualBlueScore 	VirtualBlueScore
	pruningPoint   		*externalapi.DomainHash
}

func newTXIndexStore(database database.Database) *txIndexStore {
	return &txIndexStore{
		database:       database,
		toAddTxs:          make(TxChange),
		toRemoveTxs:       make(TxChange),
		toAddSent:	   make(AddrsChange),
		toRemoveSent:	   make(AddrsChange),
		toAddReceived:	   make(AddrsChange),
		toRemoveReceived:   make(AddrsChange),
		virtualParents: nil,
		pruningPoint:   nil,
	}
}

func (tis *txIndexStore) deleteAll() error {
	err := tis.database.Delete(virtualParentsKey)
	if err != nil {
		return err
	}

	err = tis.database.Delete(pruningPointKey)
	if err != nil {
		return err
	}

	err = tis.database.Delete(virtualBlueScoreKey)
	if err != nil {
		return err
	}

	cursor, err := tis.database.Cursor(txAcceptedIndexBucket)
	if err != nil {
		return err
	}
	defer cursor.Close()
	for cursor.Next() {
		key, err := cursor.Key()
		if err != nil {
			return err
		}

		err = tis.database.Delete(key)
		if err != nil {
			return err
		}
	}

	return nil
}

func (tis *txIndexStore) add(txID externalapi.DomainTransactionID, includingIndex uint32,
	includingBlockHash *externalapi.DomainHash, acceptingBlockHash *externalapi.DomainHash, 
	sentAddrs []*externalapi.ScriptPublicKey, receivedAddrs  []*externalapi.ScriptPublicKey) {
	log.Tracef("Adding %s Txs from blockHash %s", txID.String(), includingBlockHash.String())
	delete(tis.toRemoveTxs, txID) //adding takes precedence
	tis.toAddTxs[txID] = &TxData{
		IncludingBlockHash: includingBlockHash,
		IncludingIndex:     includingIndex,
		AcceptingBlockHash: acceptingBlockHash,
	}
	for _, sentAddrs := range sentAddrs{
		txIDs, found := tis.toAddSent[ScriptPublicKeyString(sentAddrs.String())] 
		if found {
			tis.toAddSent[ScriptPublicKeyString(sentAddrs.String())] = append(txIDs, &txID)
		} else {
			tis.toAddSent[ScriptPublicKeyString(sentAddrs.String())] = []*externalapi.DomainTransactionID{&txID}
		}
	}
	for _, receivedAddr := range receivedAddrs{
		txIDs, found := tis.toAddReceived[ScriptPublicKeyString(receivedAddr.String())] 
		if found {
			tis.toAddReceived[ScriptPublicKeyString(receivedAddr.String())] = append(txIDs, &txID)
		} else {
			tis.toAddReceived[ScriptPublicKeyString(receivedAddr.String())] = []*externalapi.DomainTransactionID{&txID}
		}
	}
}

func (tis *txIndexStore) remove(txID externalapi.DomainTransactionID, includingIndex uint32,
	includingBlockHash *externalapi.DomainHash, acceptingBlockHash *externalapi.DomainHash,
	sentAddrs []*externalapi.ScriptPublicKey, receivedAddrs []*externalapi.ScriptPublicKey) {
	log.Tracef("Removing %s Txs from blockHash %s", txID.String(), includingBlockHash.String())
	if _, found := tis.toAddTxs[txID]; !found { //adding takes precedence
		tis.toRemoveTxs[txID] = &TxData{
			IncludingBlockHash: includingBlockHash,
			IncludingIndex:     includingIndex,
			AcceptingBlockHash: acceptingBlockHash,
		}
		for _, sentAddrs := range sentAddrs{
			txIDs, found := tis.toAddSent[ScriptPublicKeyString(sentAddrs.String())] 
			if found {
				tis.toAddSent[ScriptPublicKeyString(sentAddrs.String())] = append(txIDs, &txID)
			} else {
				tis.toAddSent[ScriptPublicKeyString(sentAddrs.String())] = []*externalapi.DomainTransactionID{&txID}
			}
		}
		for _, receivedAddr := range receivedAddrs{
			txIDs, found := tis.toAddReceived[ScriptPublicKeyString(receivedAddr.String())] 
			if found {
				tis.toAddReceived[ScriptPublicKeyString(receivedAddr.String())] = append(txIDs, &txID)
			} else {
				tis.toAddReceived[ScriptPublicKeyString(receivedAddr.String())] = []*externalapi.DomainTransactionID{&txID}
			}
		}
	}
}

func (tis *txIndexStore) discardAllButPruningPoint() {
	tis.toAddTxs = make(TxChange)
	tis.toRemoveTxs = make(TxChange)
	tis.toAddSent =	   make(AddrsChange)
	tis.toRemoveSent =	   make(AddrsChange)
	tis.toAddReceived =	   make(AddrsChange)
	tis.toRemoveReceived =   make(AddrsChange)
	tis.virtualParents = nil
}

func (tis *txIndexStore) commit() error {
	onEnd := logger.LogAndMeasureExecutionTime(log, "txIndexStore.commit")
	defer onEnd()

	dbTransaction, err := tis.database.Begin()
	if err != nil {
		return err
	}

	defer dbTransaction.RollbackUnlessClosed()

	for toAddTxID, txData := range tis.toAddTxs {
		delete(tis.toRemoveTxs, toAddTxID) //safeguard
		key := tis.convertTxIDToKey(txAcceptedIndexBucket, toAddTxID)
		dbTransaction.Put(key, serializeTxIndexData(txData))
		if err != nil {
			return err
		}
	}

	for toRemoveTxID := range tis.toRemoveTxs {
		key := tis.convertTxIDToKey(txAcceptedIndexBucket, toRemoveTxID)
		err := dbTransaction.Delete(key)
		if err != nil {
			return err
		}
	}

	for scriptPublicKey, receivedTxsRemoved := range tis.toRemoveSent {
		scriptPublicKey := externalapi.NewScriptPublicKeyFromString(string(scriptPublicKey))
		key := tis.convertScriptPublicKeyToKey(addrIndexSentBucket, scriptPublicKey)
		serializedTxIds, err := tis.database.Get(key)
		if err != nil {
			return err
		}
		TxIdsSet, err := deserializeTxIdsToMap(serializedTxIds)
		if err != nil {
			return err
		}
		for _, receivedTxIdToRemove := range receivedTxsRemoved {
			delete(TxIdsSet, receivedTxIdToRemove)
		}

		serializedTxIds = serializeTxIdsFromMap(TxIdsSet)
		
		err = tis.database.Put(key, serializedTxIds)
		if err != nil {
			return err
		}	
	}

	for scriptPublicKey, receivedTxsRemoved := range tis.toRemoveReceived {
		scriptPublicKey := externalapi.NewScriptPublicKeyFromString(string(scriptPublicKey))
		key := tis.convertScriptPublicKeyToKey(addrIndexReceivedBucket, scriptPublicKey)
		serializedTxIds, err := tis.database.Get(key)
		if err != nil {
			return err
		}
		TxIdsSet, err := deserializeTxIdsToMap(serializedTxIds)
		if err != nil {
			return err
		}
		for _, receivedTxIdToRemove := range receivedTxsRemoved {
			delete(TxIdsSet, receivedTxIdToRemove)
		}

		serializedTxIds = serializeTxIdsFromMap(TxIdsSet)
		
		err = tis.database.Put(key, serializedTxIds)
		if err != nil {
			return err
		}	
	}

	for scriptPublicKey, sentTxsAdded := range tis.toAddSent {
		scriptPublicKey := externalapi.NewScriptPublicKeyFromString(string(scriptPublicKey))
		key := tis.convertScriptPublicKeyToKey(addrIndexSentBucket, scriptPublicKey)
		found, err := tis.database.Has(key) 
		if err != nil {
			return err
		} 
		if found{
			serializedTxIds, err := tis.database.Get(key)
			if err != nil {
				return err
			}
			TxIdsSet, err := deserializeTxIdsToMap(serializedTxIds)
			if err != nil {
				return err
			}
			newTxIds := make([]*externalapi.DomainTransactionID, 0)
			for _, sentTxId := range sentTxsAdded {
				if _, found := TxIdsSet[sentTxId]; found {
					continue
				}
				newTxIds = append(newTxIds, sentTxId)
			}
			sentTxsAdded = newTxIds

		}
		serializedTxIds := serializeTxIds(sentTxsAdded)
		
		err = tis.database.Put(key, serializedTxIds)
		if err != nil {
			return err
		}	
	}

	for scriptPublicKey, receivedTxsAdded := range tis.toAddReceived {
		scriptPublicKey := externalapi.NewScriptPublicKeyFromString(string(scriptPublicKey))
		key := tis.convertScriptPublicKeyToKey(addrIndexReceivedBucket, scriptPublicKey)
		found, err := tis.database.Has(key)
		if err != nil {
			return err
		} 
		if found {
			serializedTxIds, err := tis.database.Get(key)
			if err != nil {
				return err
			}
			TxIdsSet, err := deserializeTxIdsToMap(serializedTxIds)
			if err != nil {
				return err
			}
			newTxIds := make([]*externalapi.DomainTransactionID, 0)
			for _, recivedTxIdToAdd := range receivedTxsAdded {
				if _, found := TxIdsSet[recivedTxIdToAdd]; found {
					continue
				}
				newTxIds = append(newTxIds, recivedTxIdToAdd)
			}
			receivedTxsAdded = newTxIds

		}
		serializedTxIds := serializeTxIds(receivedTxsAdded)
		
		err = tis.database.Put(key, serializedTxIds)
		if err != nil {
			return err
		}	
	}

	err = dbTransaction.Put(virtualParentsKey, serializeHashes(tis.virtualParents))
	if err != nil {
		return err
	}

	err = dbTransaction.Commit()
	if err != nil {
		return err
	}

	tis.discardAllButPruningPoint()

	return nil
}

func (tis *txIndexStore) commitVirtualParentsWithoutTransaction(virtualParents []*externalapi.DomainHash) error {
	serializeParentHashes := serializeHashes(virtualParents)
	return tis.database.Put(virtualParentsKey, serializeParentHashes)
}

func (tis *txIndexStore) updateVirtualParents(virtualParents []*externalapi.DomainHash) {
	tis.virtualParents = virtualParents
}

func (tis *txIndexStore) updateAndCommitPruningPointWithoutTransaction(pruningPoint *externalapi.DomainHash) error {
	tis.pruningPoint = pruningPoint

	return tis.database.Put(pruningPointKey, pruningPoint.ByteSlice())
}

func (tis *txIndexStore) commitTxIDsWithoutTransaction() error {
	for txID, txData := range tis.toAddTxs {
		delete(tis.toRemoveTxs, txID) //adding takes precedence
		key := tis.convertTxIDToKey(txAcceptedIndexBucket, txID)
		err := tis.database.Put(key, serializeTxIndexData(txData))
		if err != nil {
			return err
		}
	}

	for txID := range tis.toRemoveTxs { //safer to remove first
		key := tis.convertTxIDToKey(txAcceptedIndexBucket, txID)
		err := tis.database.Delete(key)
		if err != nil {
			return err
		}
	}

	return nil
}

func (tis *txIndexStore) getVirtualParents() ([]*externalapi.DomainHash, error) {
	if tis.isAnythingStaged() {
		return nil, errors.Errorf("cannot get the virtual parent while staging isn't empty")
	}

	serializedVirtualParentHash, err := tis.database.Get(virtualParentsKey)
	if err != nil {
		return nil, err
	}

	return deserializeHashes(serializedVirtualParentHash)
}

func (tis *txIndexStore) getPruningPoint() (*externalapi.DomainHash, error) {
	if tis.isAnythingStaged() {
		return nil, errors.Errorf("cannot get the Pruning point while staging isn't empty")
	}

	serializedPruningPointHash, err := tis.database.Get(pruningPointKey)
	if err != nil {
		return nil, err
	}

	return externalapi.NewDomainHashFromByteSlice(serializedPruningPointHash)
}

func (tis *txIndexStore) getBlueScore() (uint64, error) {
	if tis.isAnythingStaged() {
		return 0, errors.Errorf("cannot get the virtual bluescore while staging isn't empty")
	}

	serializedVirtualBlueScore, err := tis.database.Get(virtualBlueScoreKey)
	if err != nil {
		return 0, err
	}

	return binary.BigEndian.Uint64(serializedVirtualBlueScore), nil
}

func (tis *txIndexStore) convertTxIDToKey(bucket *database.Bucket, txID externalapi.DomainTransactionID) *database.Key {
	return bucket.Key(txID.ByteSlice())
}

func (tis *txIndexStore) convertScriptPublicKeyToKey(bucket *database.Bucket, scriptPublicKey *externalapi.ScriptPublicKey) *database.Key {
	var scriptPublicKeyBytes = make([]byte, 2+len(scriptPublicKey.Script)) // uint16
	binary.LittleEndian.PutUint16(scriptPublicKeyBytes[:2], scriptPublicKey.Version)
	copy(scriptPublicKeyBytes[2:], scriptPublicKey.Script)
	return bucket.Key(scriptPublicKeyBytes)
}

func (tis *txIndexStore) stagedData() (
	toAddTxs 		TxChange,
	toRemoveTxs 		TxChange,
	toAddSent		AddrsChange,
	toRemoveSent		AddrsChange,
	toAddReceived		AddrsChange,
	toRemoveReceived	AddrsChange,
	virtualParents 		[]*externalapi.DomainHash,
	pruningPoint 		*externalapi.DomainHash ) {

	toAddClone := make(TxChange)
	toRemoveClone := make(TxChange)
	toAddSentClone := make(AddrsChange)
	toRemoveSentClone := make(AddrsChange)
	toAddReceivedClone := make(AddrsChange)
	toRemoveReceivedClone := make(AddrsChange)

	maps.Copy(toAddClone, tis.toAddTxs)
	maps.Copy(toRemoveClone, tis.toRemoveTxs)
	maps.Copy(toAddSentClone, tis.toAddSent)
	maps.Copy(toRemoveSentClone, tis.toRemoveSent)
	maps.Copy(toAddReceivedClone, tis.toAddReceived)
	maps.Copy(toRemoveReceivedClone, tis.toRemoveReceived)

	return toAddClone, toRemoveClone, toAddSentClone, 
		toRemoveSentClone, toAddReceivedClone, 
		toRemoveReceived, tis.virtualParents, tis.pruningPoint
}

func (tis *txIndexStore) isAnythingStaged() bool {
	return len(tis.toAddTxs) > 0 || len(tis.toRemoveTxs) > 0
}

func (tis *txIndexStore) getTxData(txID *externalapi.DomainTransactionID) (txData *TxData, found bool, err error) {

	if tis.isAnythingStaged() {
		return nil, false, errors.Errorf("cannot get TX accepting Block hash while staging isn't empty")
	}

	key := tis.convertTxIDToKey(txAcceptedIndexBucket, *txID)
	serializedTxData, err := tis.database.Get(key)
	if err != nil {
		if database.IsNotFoundError(err) {
			return nil, false, nil
		}
		return nil, false, err
	}

	deserializedTxData, err := deserializeTxIndexData(serializedTxData)
	if err != nil {
		return nil, false, err
	}

	return deserializedTxData, true, nil
}

func (tis *txIndexStore) getTxsData(txIDs []*externalapi.DomainTransactionID) (
	txsData TxIDsToTxIndexData, notFoundTxIDs []*externalapi.DomainTransactionID, err error) {

	if tis.isAnythingStaged() {
		return nil, nil, errors.Errorf("cannot get TX accepting Block hash while staging isn't empty")
	}

	keys := make([]*database.Key, len(txIDs))

	txsData = make(TxIDsToTxIndexData)
	notFoundTxIDs = make([]*externalapi.DomainTransactionID, 0)

	for i, key := range keys {
		key = tis.convertTxIDToKey(txAcceptedIndexBucket, *txIDs[i])
		serializedTxData, err := tis.database.Get(key)
		if err != nil {
			if database.IsNotFoundError(err) {
				notFoundTxIDs = append(notFoundTxIDs, txIDs[i])
			} else {
				return nil, nil, err
			}
		}
		deserializedTxData, err := deserializeTxIndexData(serializedTxData)
		if err != nil {
			return nil, nil, err
		}

		txsData[*txIDs[i]] = deserializedTxData
	}

	return txsData, notFoundTxIDs, nil
}

func (tis *txIndexStore) getTxIdsFromScriptPublicKey(scriptPublicKey *externalapi.ScriptPublicKey, includeReceived bool, includeSent bool) (
	received []*externalapi.DomainTransactionID, sent []*externalapi.DomainTransactionID, err error) {

	if tis.isAnythingStaged() {
		return nil, nil, errors.Errorf("cannot get TX accepting Block hash while staging isn't empty")
	}

	if includeReceived {
		key := tis.convertScriptPublicKeyToKey(addrIndexReceivedBucket, scriptPublicKey)
		serializedTxIds, err := tis.database.Get(key)
		if err != nil && !database.IsNotFoundError(err){
			return nil, nil, err
		}
		received, err = deserializeTxIds(serializedTxIds)
		if err != nil {
			return nil, nil, err
		}

	}
	if includeSent {
		key := tis.convertScriptPublicKeyToKey(addrIndexSentBucket, scriptPublicKey)
		serializedTxIds, err := tis.database.Get(key)
		if err != nil && !database.IsNotFoundError(err) {
			return nil, nil, err
		}
		
		sent, err = deserializeTxIds(serializedTxIds)
		if err != nil {
			return nil, nil, err
		}		
	}

	return received, sent, nil
}

func (tis *txIndexStore) getTxIdsOfScriptPublicKeys(scriptPublicKeys []*externalapi.ScriptPublicKey, includeReceived bool, includeSent bool) (
	AddrsChange, AddrsChange, error) {

	if tis.isAnythingStaged() {
		return nil, nil, errors.Errorf("cannot get TXs of scriptPublicKeys while staging isn't empty")
	}

	AddressesToReceivedTxIds := make(AddrsChange)
	AddressesToSentTxIds := make(AddrsChange)

	for _, scriptPublicKey := range scriptPublicKeys {
		if includeReceived {
			key := tis.convertScriptPublicKeyToKey(addrIndexReceivedBucket, scriptPublicKey)
			serializedTxIds, err := tis.database.Get(key)
			if err != nil {
				if database.IsNotFoundError(err) {
					continue
				} else {
					return nil, nil, err
				}
			}
			deserializedTxIds, err := deserializeTxIds(serializedTxIds)
			if err != nil {
				return nil, nil, err
			}

			AddressesToReceivedTxIds[ScriptPublicKeyString(scriptPublicKey.String())] = deserializedTxIds
		}
		if includeSent {
			key := tis.convertScriptPublicKeyToKey(addrIndexSentBucket, scriptPublicKey)
			serializedTxIds, err := tis.database.Get(key)
			if err != nil {
				if database.IsNotFoundError(err) {
					continue
				} else {
					return nil, nil, err
				}
			}
			deserializedTxIds, err := deserializeTxIds(serializedTxIds)
			if err != nil {
				return nil, nil, err
			}
			
			AddressesToSentTxIds[ScriptPublicKeyString(scriptPublicKey.String())] = deserializedTxIds
		}
	}

	return AddressesToReceivedTxIds, AddressesToSentTxIds, nil
}
