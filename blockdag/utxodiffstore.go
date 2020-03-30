package blockdag

import (
	"bytes"
	"github.com/kaspanet/kaspad/database"
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/kaspanet/kaspad/util/locks"
	"github.com/pkg/errors"
)

type blockUTXODiffData struct {
	diff      *UTXODiff
	diffChild *blockNode
}

type utxoDiffStore struct {
	dag    *BlockDAG
	dirty  map[daghash.Hash]struct{}
	loaded map[daghash.Hash]*blockUTXODiffData
	mtx    *locks.PriorityMutex
}

func newUTXODiffStore(dag *BlockDAG) *utxoDiffStore {
	return &utxoDiffStore{
		dag:    dag,
		dirty:  make(map[daghash.Hash]struct{}),
		loaded: make(map[daghash.Hash]*blockUTXODiffData),
		mtx:    locks.NewPriorityMutex(),
	}
}

func (diffStore *utxoDiffStore) setBlockDiff(node *blockNode, diff *UTXODiff) error {
	diffStore.mtx.HighPriorityWriteLock()
	defer diffStore.mtx.HighPriorityWriteUnlock()
	// load the diff data from DB to diffStore.loaded
	_, exists, err := diffStore.diffDataByHash(node.hash)
	if err != nil {
		return err
	}
	if !exists {
		diffStore.loaded[*node.hash] = &blockUTXODiffData{}
	}

	diffStore.loaded[*node.hash].diff = diff
	diffStore.setBlockAsDirty(node.hash)
	return nil
}

func (diffStore *utxoDiffStore) setBlockDiffChild(node *blockNode, diffChild *blockNode) error {
	diffStore.mtx.HighPriorityWriteLock()
	defer diffStore.mtx.HighPriorityWriteUnlock()
	// load the diff data from DB to diffStore.loaded
	_, exists, err := diffStore.diffDataByHash(node.hash)
	if err != nil {
		return err
	}
	if !exists {
		return diffNotFoundError(node)
	}

	diffStore.loaded[*node.hash].diffChild = diffChild
	diffStore.setBlockAsDirty(node.hash)
	return nil
}

func (diffStore *utxoDiffStore) removeBlocksDiffData(dbTx database.Tx, blockHashes []*daghash.Hash) error {
	for _, hash := range blockHashes {
		err := diffStore.removeBlockDiffData(dbTx, hash)
		if err != nil {
			return err
		}
	}
	return nil
}

func (diffStore *utxoDiffStore) removeBlockDiffData(dbTx database.Tx, blockHash *daghash.Hash) error {
	diffStore.mtx.LowPriorityWriteLock()
	defer diffStore.mtx.LowPriorityWriteUnlock()
	delete(diffStore.loaded, *blockHash)
	err := dbRemoveDiffData(dbTx, blockHash)
	if err != nil {
		return err
	}
	return nil
}

func (diffStore *utxoDiffStore) setBlockAsDirty(blockHash *daghash.Hash) {
	diffStore.dirty[*blockHash] = struct{}{}
}

func (diffStore *utxoDiffStore) diffDataByHash(hash *daghash.Hash) (*blockUTXODiffData, bool, error) {
	if diffData, ok := diffStore.loaded[*hash]; ok {
		return diffData, true, nil
	}
	diffData, err := diffStore.diffDataFromDB(hash)
	if err != nil {
		return nil, false, err
	}
	exists := diffData != nil
	if exists {
		diffStore.loaded[*hash] = diffData
	}
	return diffData, exists, nil
}

func diffNotFoundError(node *blockNode) error {
	return errors.Errorf("Couldn't find diff data for block %s", node.hash)
}

func (diffStore *utxoDiffStore) diffByNode(node *blockNode) (*UTXODiff, error) {
	diffStore.mtx.HighPriorityReadLock()
	defer diffStore.mtx.HighPriorityReadUnlock()
	diffData, exists, err := diffStore.diffDataByHash(node.hash)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, diffNotFoundError(node)
	}
	return diffData.diff, nil
}

func (diffStore *utxoDiffStore) diffChildByNode(node *blockNode) (*blockNode, error) {
	diffStore.mtx.HighPriorityReadLock()
	defer diffStore.mtx.HighPriorityReadUnlock()
	diffData, exists, err := diffStore.diffDataByHash(node.hash)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, diffNotFoundError(node)
	}
	return diffData.diffChild, nil
}

func (diffStore *utxoDiffStore) diffDataFromDB(hash *daghash.Hash) (*blockUTXODiffData, error) {
	var diffData *blockUTXODiffData
	err := diffStore.dag.db.View(func(dbTx database.Tx) error {
		bucket := dbTx.Metadata().Bucket(utxoDiffsBucketName)
		serializedBlockDiffData := bucket.Get(hash[:])
		if serializedBlockDiffData != nil {
			var err error
			diffData, err = diffStore.deserializeBlockUTXODiffData(serializedBlockDiffData)
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return diffData, nil
}

// flushToDB writes all dirty diff data to the database. If all writes
// succeed, this clears the dirty set.
func (diffStore *utxoDiffStore) flushToDB(dbTx database.Tx) error {
	diffStore.mtx.HighPriorityWriteLock()
	defer diffStore.mtx.HighPriorityWriteUnlock()
	if len(diffStore.dirty) == 0 {
		return nil
	}

	// Allocate a buffer here to avoid needless allocations/grows
	// while writing each entry.
	buffer := &bytes.Buffer{}
	for hash := range diffStore.dirty {
		hash := hash // Copy hash to a new variable to avoid passing the same pointer
		buffer.Reset()
		diffData := diffStore.loaded[hash]
		err := dbStoreDiffData(dbTx, buffer, &hash, diffData)
		if err != nil {
			return err
		}
	}
	return nil
}

func (diffStore *utxoDiffStore) clearDirtyEntries() {
	diffStore.dirty = make(map[daghash.Hash]struct{})
}

// dbStoreDiffData stores the UTXO diff data to the database.
// This overwrites the current entry if there exists one.
func dbStoreDiffData(dbTx database.Tx, writeBuffer *bytes.Buffer, hash *daghash.Hash, diffData *blockUTXODiffData) error {
	// To avoid a ton of allocs, use the given writeBuffer
	// instead of allocating one. We expect the buffer to
	// already be initalized and, in most cases, to already
	// be large enough to accommodate the serialized data
	// without growing.
	err := serializeBlockUTXODiffData(writeBuffer, diffData)
	if err != nil {
		return err
	}

	// Bucket.Put doesn't copy on its own, so we manually
	// copy here. We do so because we expect the buffer
	// to be reused once we're done with it.
	serializedDiffData := make([]byte, writeBuffer.Len())
	copy(serializedDiffData, writeBuffer.Bytes())

	return dbTx.Metadata().Bucket(utxoDiffsBucketName).Put(hash[:], serializedDiffData)
}

func dbRemoveDiffData(dbTx database.Tx, hash *daghash.Hash) error {
	return dbTx.Metadata().Bucket(utxoDiffsBucketName).Delete(hash[:])
}
