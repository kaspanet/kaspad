// Copyright (c) 2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package indexers

import (
	"fmt"

	"bytes"

	"github.com/daglabs/btcd/blockdag"
	"github.com/daglabs/btcd/dagconfig/daghash"
	"github.com/daglabs/btcd/database"
	"github.com/daglabs/btcd/util"
	"github.com/daglabs/btcd/wire"
)

var (
	// indexTipsBucketName is the name of the db bucket used to house the
	// current tip of each index.
	indexTipsBucketName = []byte("idxtips")
)

// dbIndexConnectBlock adds all of the index entries associated with the
// given block using the provided indexer and updates the tip of the indexer
// accordingly.  An error will be returned if the current tip for the indexer is
// not the previous block for the passed block.
func dbIndexConnectBlock(dbTx database.Tx, indexer Indexer, block *util.Block, virtual *blockdag.VirtualBlock) error {

	// Notify the indexer with the connected block so it can index it.
	if err := indexer.ConnectBlock(dbTx, block, virtual); err != nil {
		return err
	}

	return nil
}

// dbIndexDisconnectBlock removes all of the index entries associated with the
// given block using the provided indexer and updates the tip of the indexer
// accordingly.  An error will be returned if the current tip for the indexer is
// not the passed block.
func dbIndexDisconnectBlock(dbTx database.Tx, indexer Indexer, block *util.Block, virtual *blockdag.VirtualBlock) error {

	// Notify the indexer with the disconnected block so it can remove all
	// of the appropriate entries.
	if err := indexer.DisconnectBlock(dbTx, block, virtual); err != nil {
		return err
	}

	return nil
}

// Manager defines an index manager that manages multiple optional indexes and
// implements the blockchain.IndexManager interface so it can be seamlessly
// plugged into normal chain processing.
type Manager struct {
	db             database.DB
	enabledIndexes []Indexer
}

// Ensure the Manager type implements the blockchain.IndexManager interface.
var _ blockdag.IndexManager = (*Manager)(nil)

// indexDropKey returns the key for an index which indicates it is in the
// process of being dropped.
func indexDropKey(idxKey []byte) []byte {
	dropKey := make([]byte, len(idxKey)+1)
	dropKey[0] = 'd'
	copy(dropKey[1:], idxKey)
	return dropKey
}

// maybeFinishDrops determines if each of the enabled indexes are in the middle
// of being dropped and finishes dropping them when the are.  This is necessary
// because dropping and index has to be done in several atomic steps rather than
// one big atomic step due to the massive number of entries.
func (m *Manager) maybeFinishDrops(interrupt <-chan struct{}) error {
	indexNeedsDrop := make([]bool, len(m.enabledIndexes))
	err := m.db.View(func(dbTx database.Tx) error {
		// None of the indexes needs to be dropped if the index tips
		// bucket hasn't been created yet.
		indexesBucket := dbTx.Metadata().Bucket(indexTipsBucketName)
		if indexesBucket == nil {
			return nil
		}

		// Mark the indexer as requiring a drop if one is already in
		// progress.
		for i, indexer := range m.enabledIndexes {
			dropKey := indexDropKey(indexer.Key())
			if indexesBucket.Get(dropKey) != nil {
				indexNeedsDrop[i] = true
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	if interruptRequested(interrupt) {
		return errInterruptRequested
	}

	// Finish dropping any of the enabled indexes that are already in the
	// middle of being dropped.
	for i, indexer := range m.enabledIndexes {
		if !indexNeedsDrop[i] {
			continue
		}

		log.Infof("Resuming %s drop", indexer.Name())
		err := dropIndex(m.db, indexer.Key(), indexer.Name(), interrupt)
		if err != nil {
			return err
		}
	}

	return nil
}

// maybeCreateIndexes determines if each of the enabled indexes have already
// been created and creates them if not.
func (m *Manager) maybeCreateIndexes(dbTx database.Tx) error {
	indexesBucket := dbTx.Metadata().Bucket(indexTipsBucketName)
	for _, indexer := range m.enabledIndexes {
		// Nothing to do if the index tip already exists.
		idxKey := indexer.Key()
		if indexesBucket.Get(idxKey) != nil {
			continue
		}

		// The tip for the index does not exist, so create it and
		// invoke the create callback for the index so it can perform
		// any one-time initialization it requires.
		if err := indexer.Create(dbTx); err != nil {
			return err
		}
	}

	return nil
}

// Init initializes the enabled indexes.  This is called during chain
// initialization and primarily consists of catching up all indexes to the
// current best chain tip.  This is necessary since each index can be disabled
// and re-enabled at any time and attempting to catch-up indexes at the same
// time new blocks are being downloaded would lead to an overall longer time to
// catch up due to the I/O contention.
//
// This is part of the blockchain.IndexManager interface.
func (m *Manager) Init(blockDAG *blockdag.BlockDAG, interrupt <-chan struct{}) error {
	// Nothing to do when no indexes are enabled.
	if len(m.enabledIndexes) == 0 {
		return nil
	}

	if interruptRequested(interrupt) {
		return errInterruptRequested
	}

	// Finish and drops that were previously interrupted.
	if err := m.maybeFinishDrops(interrupt); err != nil {
		return err
	}

	// Create the initial state for the indexes as needed.
	err := m.db.Update(func(dbTx database.Tx) error {
		// Create the bucket for the current tips as needed.
		meta := dbTx.Metadata()
		_, err := meta.CreateBucketIfNotExists(indexTipsBucketName)
		if err != nil {
			return err
		}

		return m.maybeCreateIndexes(dbTx)
	})
	if err != nil {
		return err
	}

	// Initialize each of the enabled indexes.
	for _, indexer := range m.enabledIndexes {
		if err := indexer.Init(); err != nil {
			return err
		}
	}

	return nil
}

// indexNeedsInputs returns whether or not the index needs access to the txouts
// referenced by the transaction inputs being indexed.
func indexNeedsInputs(index Indexer) bool {
	if idx, ok := index.(NeedsInputser); ok {
		return idx.NeedsInputs()
	}

	return false
}

// dbFetchTx looks up the passed transaction hash in the transaction index and
// loads it from the database.
func dbFetchTx(dbTx database.Tx, hash *daghash.Hash) (*wire.MsgTx, error) {
	// Look up the location of the transaction.
	blockRegion, err := dbFetchTxIndexEntry(dbTx, hash)
	if err != nil {
		return nil, err
	}
	if blockRegion == nil {
		return nil, fmt.Errorf("transaction %v not found", hash)
	}

	// Load the raw transaction bytes from the database.
	txBytes, err := dbTx.FetchBlockRegion(blockRegion)
	if err != nil {
		return nil, err
	}

	// Deserialize the transaction.
	var msgTx wire.MsgTx
	err = msgTx.Deserialize(bytes.NewReader(txBytes))
	if err != nil {
		return nil, err
	}

	return &msgTx, nil
}

// makeUtxoView creates a mock unspent transaction output view by using the
// transaction index in order to look up all inputs referenced by the
// transactions in the block.  This is sometimes needed when catching indexes up
// because many of the txouts could actually already be spent however the
// associated scripts are still required to index them.
func makeUtxoView(dbTx database.Tx, block *util.Block, interrupt <-chan struct{}) (*blockdag.UTXOView, error) {
	view := blockdag.NewUTXOView()
	for txIdx, tx := range block.Transactions() {
		// Coinbases do not reference any inputs.  Since the block is
		// required to have already gone through full validation, it has
		// already been proven on the first transaction in the block is
		// a coinbase.
		if txIdx == 0 {
			continue
		}

		// Use the transaction index to load all of the referenced
		// inputs and add their outputs to the view.
		for _, txIn := range tx.MsgTx().TxIn {
			originOut := &txIn.PreviousOutPoint
			originTx, err := dbFetchTx(dbTx, &originOut.Hash)
			if err != nil {
				return nil, err
			}

			view.AddTxOuts(util.NewTx(originTx), 0)
		}

		if interruptRequested(interrupt) {
			return nil, errInterruptRequested
		}
	}

	return view, nil
}

// ConnectBlock must be invoked when a block is extending the main chain.  It
// keeps track of the state of each index it is managing, performs some sanity
// checks, and invokes each indexer.
//
// This is part of the blockchain.IndexManager interface.
func (m *Manager) ConnectBlock(dbTx database.Tx, block *util.Block, virtual *blockdag.VirtualBlock) error {
	// Call each of the currently active optional indexes with the block
	// being connected so they can update accordingly.
	for _, index := range m.enabledIndexes {
		err := dbIndexConnectBlock(dbTx, index, block, virtual)
		if err != nil {
			return err
		}
	}
	return nil
}

// DisconnectBlock must be invoked when a block is being disconnected from the
// end of the main chain.  It keeps track of the state of each index it is
// managing, performs some sanity checks, and invokes each indexer to remove
// the index entries associated with the block.
//
// This is part of the blockchain.IndexManager interface.
func (m *Manager) DisconnectBlock(dbTx database.Tx, block *util.Block, virtual *blockdag.VirtualBlock) error {
	// Call each of the currently active optional indexes with the block
	// being disconnected so they can update accordingly.
	for _, index := range m.enabledIndexes {
		err := dbIndexDisconnectBlock(dbTx, index, block, virtual)
		if err != nil {
			return err
		}
	}
	return nil
}

// NewManager returns a new index manager with the provided indexes enabled.
//
// The manager returned satisfies the blockchain.IndexManager interface and thus
// cleanly plugs into the normal blockchain processing path.
func NewManager(db database.DB, enabledIndexes []Indexer) *Manager {
	return &Manager{
		db:             db,
		enabledIndexes: enabledIndexes,
	}
}

// dropIndex drops the passed index from the database.  Since indexes can be
// massive, it deletes the index in multiple database transactions in order to
// keep memory usage to reasonable levels.  It also marks the drop in progress
// so the drop can be resumed if it is stopped before it is done before the
// index can be used again.
func dropIndex(db database.DB, idxKey []byte, idxName string, interrupt <-chan struct{}) error {
	// Nothing to do if the index doesn't already exist.
	var needsDelete bool
	err := db.View(func(dbTx database.Tx) error {
		indexesBucket := dbTx.Metadata().Bucket(indexTipsBucketName)
		if indexesBucket != nil && indexesBucket.Get(idxKey) != nil {
			needsDelete = true
		}
		return nil
	})
	if err != nil {
		return err
	}
	if !needsDelete {
		log.Infof("Not dropping %s because it does not exist", idxName)
		return nil
	}

	// Mark that the index is in the process of being dropped so that it
	// can be resumed on the next start if interrupted before the process is
	// complete.
	log.Infof("Dropping all %s entries.  This might take a while...",
		idxName)
	err = db.Update(func(dbTx database.Tx) error {
		indexesBucket := dbTx.Metadata().Bucket(indexTipsBucketName)
		return indexesBucket.Put(indexDropKey(idxKey), idxKey)
	})
	if err != nil {
		return err
	}

	// Since the indexes can be so large, attempting to simply delete
	// the bucket in a single database transaction would result in massive
	// memory usage and likely crash many systems due to ulimits.  In order
	// to avoid this, use a cursor to delete a maximum number of entries out
	// of the bucket at a time. Recurse buckets depth-first to delete any
	// sub-buckets.
	const maxDeletions = 2000000
	var totalDeleted uint64

	// Recurse through all buckets in the index, cataloging each for
	// later deletion.
	var subBuckets [][][]byte
	var subBucketClosure func(database.Tx, []byte, [][]byte) error
	subBucketClosure = func(dbTx database.Tx,
		subBucket []byte, tlBucket [][]byte) error {
		// Get full bucket name and append to subBuckets for later
		// deletion.
		var bucketName [][]byte
		if (tlBucket == nil) || (len(tlBucket) == 0) {
			bucketName = append(bucketName, subBucket)
		} else {
			bucketName = append(tlBucket, subBucket)
		}
		subBuckets = append(subBuckets, bucketName)
		// Recurse sub-buckets to append to subBuckets slice.
		bucket := dbTx.Metadata()
		for _, subBucketName := range bucketName {
			bucket = bucket.Bucket(subBucketName)
		}
		return bucket.ForEachBucket(func(k []byte) error {
			return subBucketClosure(dbTx, k, bucketName)
		})
	}

	// Call subBucketClosure with top-level bucket.
	err = db.View(func(dbTx database.Tx) error {
		return subBucketClosure(dbTx, idxKey, nil)
	})
	if err != nil {
		return nil
	}

	// Iterate through each sub-bucket in reverse, deepest-first, deleting
	// all keys inside them and then dropping the buckets themselves.
	for i := range subBuckets {
		bucketName := subBuckets[len(subBuckets)-1-i]
		// Delete maxDeletions key/value pairs at a time.
		for numDeleted := maxDeletions; numDeleted == maxDeletions; {
			numDeleted = 0
			err := db.Update(func(dbTx database.Tx) error {
				subBucket := dbTx.Metadata()
				for _, subBucketName := range bucketName {
					subBucket = subBucket.Bucket(subBucketName)
				}
				cursor := subBucket.Cursor()
				for ok := cursor.First(); ok; ok = cursor.Next() &&
					numDeleted < maxDeletions {

					if err := cursor.Delete(); err != nil {
						return err
					}
					numDeleted++
				}
				return nil
			})
			if err != nil {
				return err
			}

			if numDeleted > 0 {
				totalDeleted += uint64(numDeleted)
				log.Infof("Deleted %d keys (%d total) from %s",
					numDeleted, totalDeleted, idxName)
			}
		}

		if interruptRequested(interrupt) {
			return errInterruptRequested
		}

		// Drop the bucket itself.
		err = db.Update(func(dbTx database.Tx) error {
			bucket := dbTx.Metadata()
			for j := 0; j < len(bucketName)-1; j++ {
				bucket = bucket.Bucket(bucketName[j])
			}
			return bucket.DeleteBucket(bucketName[len(bucketName)-1])
		})
	}

	// Call extra index specific deinitialization for the transaction index.
	if idxName == txIndexName {
		if err := dropBlockIDIndex(db); err != nil {
			return err
		}
	}

	// Remove the index tip, index bucket, and in-progress drop flag now
	// that all index entries have been removed.
	err = db.Update(func(dbTx database.Tx) error {
		meta := dbTx.Metadata()
		indexesBucket := meta.Bucket(indexTipsBucketName)
		if err := indexesBucket.Delete(idxKey); err != nil {
			return err
		}

		return indexesBucket.Delete(indexDropKey(idxKey))
	})
	if err != nil {
		return err
	}

	log.Infof("Dropped %s", idxName)
	return nil
}
