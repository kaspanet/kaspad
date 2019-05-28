package blockdag

import (
	"fmt"
	"sync"

	"github.com/daglabs/btcd/database"
	"github.com/daglabs/btcd/util/daghash"
)

var multisetPointSize = 32

type blockUTXODiffData struct {
	diff      *UTXODiff
	diffChild *blockNode
}

type utxoDiffStore struct {
	dag    *BlockDAG
	dirty  map[daghash.Hash]struct{}
	loaded map[daghash.Hash]*blockUTXODiffData
	mtx    sync.RWMutex
}

func newUTXODiffStore(dag *BlockDAG) *utxoDiffStore {
	return &utxoDiffStore{
		dag:    dag,
		dirty:  make(map[daghash.Hash]struct{}),
		loaded: make(map[daghash.Hash]*blockUTXODiffData),
	}
}

func (diffStore *utxoDiffStore) setBlockDiff(node *blockNode, diff *UTXODiff) error {
	diffStore.mtx.Lock()
	defer diffStore.mtx.Unlock()
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
	diffStore.mtx.Lock()
	defer diffStore.mtx.Unlock()
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
	return fmt.Errorf("Couldn't find diff data for block %s", node.hash)
}

func (diffStore *utxoDiffStore) diffByNode(node *blockNode) (*UTXODiff, error) {
	diffStore.mtx.RLock()
	defer diffStore.mtx.RUnlock()
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
	diffStore.mtx.RLock()
	defer diffStore.mtx.RUnlock()
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
	diffStore.mtx.Lock()
	defer diffStore.mtx.Unlock()
	if len(diffStore.dirty) == 0 {
		return nil
	}

	for hash := range diffStore.dirty {
		diffData := diffStore.loaded[hash]
		err := dbStoreDiffData(dbTx, &hash, diffData)
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
func dbStoreDiffData(dbTx database.Tx, hash *daghash.Hash, diffData *blockUTXODiffData) error {
	serializedDiffData, err := serializeBlockUTXODiffData(diffData)
	if err != nil {
		return err
	}

	return dbTx.Metadata().Bucket(utxoDiffsBucketName).Put(hash[:], serializedDiffData)
}
