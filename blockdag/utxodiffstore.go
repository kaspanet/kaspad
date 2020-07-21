package blockdag

import (
	"bytes"

	"github.com/kaspanet/kaspad/dbaccess"
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/kaspanet/kaspad/util/locks"
)

type blockUTXODiffData struct {
	diff      *UTXODiff
	diffChild *blockNode
}

type utxoDiffStore struct {
	dag    *BlockDAG
	dirty  map[*blockNode]struct{}
	loaded map[*blockNode]*blockUTXODiffData
	mtx    *locks.PriorityMutex
}

func newUTXODiffStore(dag *BlockDAG) *utxoDiffStore {
	return &utxoDiffStore{
		dag:    dag,
		dirty:  make(map[*blockNode]struct{}),
		loaded: make(map[*blockNode]*blockUTXODiffData),
		mtx:    locks.NewPriorityMutex(),
	}
}

func (diffStore *utxoDiffStore) setBlockDiff(node *blockNode, diff *UTXODiff) error {
	diffStore.mtx.HighPriorityWriteLock()
	defer diffStore.mtx.HighPriorityWriteUnlock()
	// load the diff data from DB to diffStore.loaded
	_, err := diffStore.diffDataByBlockNode(node)
	if dbaccess.IsNotFoundError(err) {
		diffStore.loaded[node] = &blockUTXODiffData{}
	} else if err != nil {
		return err
	}

	diffStore.loaded[node].diff = diff
	diffStore.setBlockAsDirty(node)
	return nil
}

func (diffStore *utxoDiffStore) setBlockDiffChild(node *blockNode, diffChild *blockNode) error {
	diffStore.mtx.HighPriorityWriteLock()
	defer diffStore.mtx.HighPriorityWriteUnlock()
	// load the diff data from DB to diffStore.loaded
	_, err := diffStore.diffDataByBlockNode(node)
	if err != nil {
		return err
	}

	diffStore.loaded[node].diffChild = diffChild
	diffStore.setBlockAsDirty(node)
	return nil
}

func (diffStore *utxoDiffStore) removeBlocksDiffData(dbContext dbaccess.Context, nodes []*blockNode) error {
	for _, node := range nodes {
		err := diffStore.removeBlockDiffData(dbContext, node)
		if err != nil {
			return err
		}
	}
	return nil
}

func (diffStore *utxoDiffStore) removeBlockDiffData(dbContext dbaccess.Context, node *blockNode) error {
	diffStore.mtx.LowPriorityWriteLock()
	defer diffStore.mtx.LowPriorityWriteUnlock()
	delete(diffStore.loaded, node)
	err := dbaccess.RemoveDiffData(dbContext, node.hash)
	if err != nil {
		return err
	}
	return nil
}

func (diffStore *utxoDiffStore) setBlockAsDirty(node *blockNode) {
	diffStore.dirty[node] = struct{}{}
}

func (diffStore *utxoDiffStore) diffDataByBlockNode(node *blockNode) (*blockUTXODiffData, error) {
	if diffData, ok := diffStore.loaded[node]; ok {
		return diffData, nil
	}
	diffData, err := diffStore.diffDataFromDB(node.hash)
	if err != nil {
		return nil, err
	}
	diffStore.loaded[node] = diffData
	return diffData, nil
}

func (diffStore *utxoDiffStore) diffByNode(node *blockNode) (*UTXODiff, error) {
	diffStore.mtx.HighPriorityReadLock()
	defer diffStore.mtx.HighPriorityReadUnlock()
	diffData, err := diffStore.diffDataByBlockNode(node)
	if err != nil {
		return nil, err
	}
	return diffData.diff, nil
}

func (diffStore *utxoDiffStore) diffChildByNode(node *blockNode) (*blockNode, error) {
	diffStore.mtx.HighPriorityReadLock()
	defer diffStore.mtx.HighPriorityReadUnlock()
	diffData, err := diffStore.diffDataByBlockNode(node)
	if err != nil {
		return nil, err
	}
	return diffData.diffChild, nil
}

func (diffStore *utxoDiffStore) diffDataFromDB(hash *daghash.Hash) (*blockUTXODiffData, error) {
	serializedBlockDiffData, err := dbaccess.FetchUTXODiffData(diffStore.dag.databaseContext, hash)
	if err != nil {
		return nil, err
	}

	return diffStore.deserializeBlockUTXODiffData(serializedBlockDiffData)
}

// flushToDB writes all dirty diff data to the database.
func (diffStore *utxoDiffStore) flushToDB(dbContext *dbaccess.TxContext) error {
	diffStore.mtx.HighPriorityWriteLock()
	defer diffStore.mtx.HighPriorityWriteUnlock()
	if len(diffStore.dirty) == 0 {
		return nil
	}

	// Allocate a buffer here to avoid needless allocations/grows
	// while writing each entry.
	buffer := &bytes.Buffer{}
	for node := range diffStore.dirty {
		buffer.Reset()
		diffData := diffStore.loaded[node]
		err := storeDiffData(dbContext, buffer, node.hash, diffData)
		if err != nil {
			return err
		}
	}
	return nil
}

func (diffStore *utxoDiffStore) clearDirtyEntries() {
	diffStore.dirty = make(map[*blockNode]struct{})
}

// maxBlueScoreDifferenceToKeepLoaded is the maximum difference
// between the virtual's blueScore and a blockNode's blueScore
// under which to keep diff data loaded in memory.
var maxBlueScoreDifferenceToKeepLoaded uint64 = 100

// clearOldEntries removes entries whose blue score is lower than
// virtual.blueScore - maxBlueScoreDifferenceToKeepLoaded. Note
// that tips are not removed either even if their blue score is
// lower than the above.
func (diffStore *utxoDiffStore) clearOldEntries() {
	diffStore.mtx.HighPriorityWriteLock()
	defer diffStore.mtx.HighPriorityWriteUnlock()

	virtualBlueScore := diffStore.dag.VirtualBlueScore()
	minBlueScore := virtualBlueScore - maxBlueScoreDifferenceToKeepLoaded
	if maxBlueScoreDifferenceToKeepLoaded > virtualBlueScore {
		minBlueScore = 0
	}

	tips := diffStore.dag.virtual.tips()

	toRemove := make(map[*blockNode]struct{})
	for node := range diffStore.loaded {
		if node.blueScore < minBlueScore && !tips.contains(node) {
			toRemove[node] = struct{}{}
		}
	}
	for node := range toRemove {
		delete(diffStore.loaded, node)
	}
}

// storeDiffData stores the UTXO diff data to the database.
// This overwrites the current entry if there exists one.
func storeDiffData(dbContext dbaccess.Context, w *bytes.Buffer, hash *daghash.Hash, diffData *blockUTXODiffData) error {
	// To avoid a ton of allocs, use the io.Writer
	// instead of allocating one. We expect the buffer to
	// already be initialized and, in most cases, to already
	// be large enough to accommodate the serialized data
	// without growing.
	err := serializeBlockUTXODiffData(w, diffData)
	if err != nil {
		return err
	}

	return dbaccess.StoreUTXODiffData(dbContext, hash, w.Bytes())
}
