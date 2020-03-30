package blockdag

import (
	"bytes"
	"github.com/kaspanet/kaspad/dbaccess"
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

func (diffStore *utxoDiffStore) removeBlocksDiffData(context dbaccess.Context, blockHashes []*daghash.Hash) error {
	for _, hash := range blockHashes {
		err := diffStore.removeBlockDiffData(context, hash)
		if err != nil {
			return err
		}
	}
	return nil
}

func (diffStore *utxoDiffStore) removeBlockDiffData(context dbaccess.Context, blockHash *daghash.Hash) error {
	diffStore.mtx.LowPriorityWriteLock()
	defer diffStore.mtx.LowPriorityWriteUnlock()
	delete(diffStore.loaded, *blockHash)
	err := dbaccess.RemoveDiffData(context, blockHash)
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
	serializedBlockDiffData, err := dbaccess.FetchUTXODiffData(dbaccess.NoTx(), hash)
	if err != nil {
		return nil, err
	}

	return diffStore.deserializeBlockUTXODiffData(serializedBlockDiffData)
}

// flushToDB writes all dirty diff data to the database. If all writes
// succeed, this clears the dirty set.
func (diffStore *utxoDiffStore) flushToDB(context dbaccess.Context) error {
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
		err := dbStoreDiffData(context, buffer, &hash, diffData)
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
func dbStoreDiffData(context dbaccess.Context, w *bytes.Buffer, hash *daghash.Hash, diffData *blockUTXODiffData) error {
	// To avoid a ton of allocs, use the io.Writer
	// instead of allocating one. We expect the buffer to
	// already be initalized and, in most cases, to already
	// be large enough to accommodate the serialized data
	// without growing.
	err := serializeBlockUTXODiffData(w, diffData)
	if err != nil {
		return err
	}

	return dbaccess.StoreUTXODiffData(context, hash, w.Bytes())
}
