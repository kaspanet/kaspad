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
	toAdd          map[externalapi.DomainTransactionID]*externalapi.DomainHash
	virtualParents []*externalapi.DomainHash
	pruningPoint   *externalapi.DomainHash
}

func newTXIndexStore(database database.Database) *txIndexStore {
	return &txIndexStore{
		database:       database,
		toAdd:          make(map[externalapi.DomainTransactionID]*externalapi.DomainHash),
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

func (tis *txIndexStore) add(txID externalapi.DomainTransactionID, blockHash *externalapi.DomainHash) {
	log.Tracef("Adding %s Txs from blockHash %s", txID.String(), blockHash.String())
	tis.toAdd[txID] = blockHash
}

func (tis *txIndexStore) discard() {
	tis.toAdd = make(map[externalapi.DomainTransactionID]*externalapi.DomainHash)
	tis.virtualParents = nil
	tis.pruningPoint = nil
}

func (tis *txIndexStore) commitAndReturnRemoved() (
	removed map[externalapi.DomainTransactionID]*externalapi.DomainHash,
	err error) {
	onEnd := logger.LogAndMeasureExecutionTime(log, "txIndexStore.commit")
	defer onEnd()

	dbTransaction, err := tis.database.Begin()
	if err != nil {
		return nil, err
	}

	defer dbTransaction.RollbackUnlessClosed()

	removed = make(map[externalapi.DomainTransactionID]*externalapi.DomainHash)

	for txID, blockHash := range tis.toAdd {
		key := tis.convertTxIDToKey(txAcceptedIndexBucket, txID)
		found, err := dbTransaction.Has(key)
		if err != nil {
			return nil, err
		}

		if found {
			serializedRemovedBlockHash, err := dbTransaction.Get(key)
			if err != nil {
				return nil, err
			}
			removedBlockHash, err := externalapi.NewDomainHashFromByteSlice(serializedRemovedBlockHash)
			if err != nil {
				return nil, err
			}
			removed[txID] = removedBlockHash
		}

		dbTransaction.Put(key, blockHash.ByteSlice())
		if err != nil {
			return nil, err
		}
	}
	err = dbTransaction.Put(virtualParentsKey, serializeHashes(tis.virtualParents))
	if err != nil {
		return nil, err
	}
	err = dbTransaction.Put(pruningPointKey, tis.pruningPoint.ByteSlice())
	if err != nil {
		return nil, err
	}

	err = dbTransaction.Commit()
	if err != nil {
		return nil, err
	}

	tis.discard()

	return removed, nil
}

func (tis *txIndexStore) updateAndCommitVirtualParentsWithoutTransaction(virtualParents []*externalapi.DomainHash) error {
	serializeParentHashes := serializeHashes(virtualParents)
	return tis.database.Put(virtualParentsKey, serializeParentHashes)
}

func (tis *txIndexStore) updateAndCommitPruningPointWithoutTransaction(pruningPoint *externalapi.DomainHash) error {
	return tis.database.Put(pruningPointKey, pruningPoint.ByteSlice())

}

func (tis *txIndexStore) updateVirtualParents(virtualParents []*externalapi.DomainHash) {
	tis.virtualParents = virtualParents
}

func (tis *txIndexStore) CommitWithoutTransaction() error {
	for txID, blockHash := range tis.toAdd {
		key := tis.convertTxIDToKey(txAcceptedIndexBucket, txID)
		err := tis.database.Put(key, blockHash.ByteSlice())
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
	toAdd map[externalapi.DomainTransactionID]*externalapi.DomainHash,
	virtualParents []*externalapi.DomainHash,
	pruningPoint *externalapi.DomainHash) {
	toAddClone := make(map[externalapi.DomainTransactionID]*externalapi.DomainHash)
	for txID, blockHash := range tis.toAdd {
		toAddClone[txID] = blockHash

	}
	return toAddClone, tis.virtualParents, tis.pruningPoint
}

func (tis *txIndexStore) isAnythingStaged() bool {
	return len(tis.toAdd) > 0
}

func (tis *txIndexStore) getTxAcceptingBlockHash(txID *externalapi.DomainTransactionID) (blockHash *externalapi.DomainHash, found bool, err error) {

	if tis.isAnythingStaged() {
		return nil, false, errors.Errorf("cannot get TX accepting Block hash while staging isn't empty")
	}

	key := tis.convertTxIDToKey(txAcceptedIndexBucket, *txID)
	serializedAcceptingBlockHash, err := tis.database.Get(key)
	if err != nil {
		if err == database.ErrNotFound {
			return nil, false, nil
		}
		return nil, false, err
	}

	acceptingBlockHash, err := externalapi.NewDomainHashFromByteSlice(serializedAcceptingBlockHash)
	if err != nil {
		return nil, false, err
	}

	return acceptingBlockHash, true, nil
}
