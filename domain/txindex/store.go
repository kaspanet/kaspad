package txindex

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/infrastructure/db/database"
	"github.com/kaspanet/kaspad/infrastructure/logger"
	"github.com/pkg/errors"
)

var txAcceptedIndexBucket = database.MakeBucket([]byte("tx-index"))
var virtualParentsKey = database.MakeBucket([]byte("")).Key([]byte("tx-index-virtual-parent"))
var pruningPointKey = database.MakeBucket([]byte("")).Key([]byte("tx-index-prunning-point"))

type txIndexStore struct {
	database       database.Database
	toAdd          map[externalapi.DomainTransactionID]*TxData
	toRemove       map[externalapi.DomainTransactionID]*TxData
	virtualParents []*externalapi.DomainHash
	pruningPoint   *externalapi.DomainHash
}

func newTXIndexStore(database database.Database) *txIndexStore {
	return &txIndexStore{
		database:       database,
		toAdd:          make(map[externalapi.DomainTransactionID]*TxData),
		toRemove:       make(map[externalapi.DomainTransactionID]*TxData),
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
	includingBlockHash *externalapi.DomainHash, acceptingBlockHash *externalapi.DomainHash) {
	log.Tracef("Adding %s Txs from blockHash %s", txID.String(), includingBlockHash.String())
	delete(tis.toRemove, txID) //adding takes precedence
	tis.toAdd[txID] = &TxData{
		IncludingBlockHash: includingBlockHash,
		IncludingIndex:     includingIndex,
		AcceptingBlockHash: acceptingBlockHash,
	}
}

func (tis *txIndexStore) remove(txID externalapi.DomainTransactionID, includingIndex uint32,
	includingBlockHash *externalapi.DomainHash, acceptingBlockHash *externalapi.DomainHash) {
	log.Tracef("Removing %s Txs from blockHash %s", txID.String(), includingBlockHash.String())
	if _, found := tis.toAdd[txID]; !found { //adding takes precedence
		tis.toRemove[txID] = &TxData{
			IncludingBlockHash: includingBlockHash,
			IncludingIndex:     includingIndex,
			AcceptingBlockHash: acceptingBlockHash,
		}
	}
}

func (tis *txIndexStore) discardAllButPruningPoint() {
	tis.toAdd = make(map[externalapi.DomainTransactionID]*TxData)
	tis.toRemove = make(map[externalapi.DomainTransactionID]*TxData)
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

	for toAddTxID, txData := range tis.toAdd {
		delete(tis.toRemove, toAddTxID) //safeguard
		key := tis.convertTxIDToKey(txAcceptedIndexBucket, toAddTxID)
		dbTransaction.Put(key, serializeTxIndexData(txData))
		if err != nil {
			return err
		}
	}

	for toRemoveTxID := range tis.toRemove {
		key := tis.convertTxIDToKey(txAcceptedIndexBucket, toRemoveTxID)
		err := dbTransaction.Delete(key)
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
	for txID, txData := range tis.toAdd {
		delete(tis.toRemove, txID) //adding takes precedence
		key := tis.convertTxIDToKey(txAcceptedIndexBucket, txID)
		err := tis.database.Put(key, serializeTxIndexData(txData))
		if err != nil {
			return err
		}
	}

	for txID := range tis.toRemove { //safer to remove first
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

func (tis *txIndexStore) convertTxIDToKey(bucket *database.Bucket, txID externalapi.DomainTransactionID) *database.Key {
	return bucket.Key(txID.ByteSlice())
}

func (tis *txIndexStore) stagedData() (
	toAdd map[externalapi.DomainTransactionID]*TxData,
	toRemove map[externalapi.DomainTransactionID]*TxData,
	virtualParents []*externalapi.DomainHash,
	pruningPoint *externalapi.DomainHash) {
	toAddClone := make(map[externalapi.DomainTransactionID]*TxData)
	toRemoveClone := make(map[externalapi.DomainTransactionID]*TxData)
	for txID, txData := range tis.toAdd {
		toAddClone[txID] = txData

	}
	for txID, txData := range tis.toRemove {
		toRemoveClone[txID] = txData
	}
	return toAddClone, toRemoveClone, tis.virtualParents, tis.pruningPoint
}

func (tis *txIndexStore) isAnythingStaged() bool {
	return len(tis.toAdd) > 0 || len(tis.toRemove) > 0
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
