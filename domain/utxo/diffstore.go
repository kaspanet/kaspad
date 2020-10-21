package utxo

import (
	"bytes"

	"github.com/kaspanet/kaspad/domain/blocknode"

	"github.com/kaspanet/kaspad/infrastructure/db/dbaccess"
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/kaspanet/kaspad/util/locks"
)

type blockUTXODiffData struct {
	Diff      *Diff
	diffChild *blocknode.Node
}

// DiffStore provides functions to operate with UTXO diffs and to interact with its storage
type DiffStore struct {
	databaseContext *dbaccess.DatabaseContext
	Dirty           map[*blocknode.Node]struct{}
	Loaded          map[*blocknode.Node]*blockUTXODiffData
	mtx             *locks.PriorityMutex
}

// NewDiffStore returns a new DiffStore instance setting the provided database context
func NewDiffStore(databaseContext *dbaccess.DatabaseContext) *DiffStore {
	return &DiffStore{
		databaseContext: databaseContext,
		Dirty:           make(map[*blocknode.Node]struct{}),
		Loaded:          make(map[*blocknode.Node]*blockUTXODiffData),
		mtx:             locks.NewPriorityMutex(),
	}
}

// SetBlockDiff sets a block diff for the provided block node
func (diffStore *DiffStore) SetBlockDiff(node *blocknode.Node, diff *Diff) error {
	diffStore.mtx.HighPriorityWriteLock()
	defer diffStore.mtx.HighPriorityWriteUnlock()
	// load the Diff data from DB to diffStore.Loaded
	_, err := diffStore.diffDataByBlockNode(node)
	if dbaccess.IsNotFoundError(err) {
		diffStore.Loaded[node] = &blockUTXODiffData{}
	} else if err != nil {
		return err
	}

	diffStore.Loaded[node].Diff = diff
	diffStore.setBlockAsDirty(node)
	return nil
}

// SetBlockDiffChild sets a block diff child for the provided block node
func (diffStore *DiffStore) SetBlockDiffChild(node *blocknode.Node, diffChild *blocknode.Node) error {
	diffStore.mtx.HighPriorityWriteLock()
	defer diffStore.mtx.HighPriorityWriteUnlock()
	// load the Diff data from DB to diffStore.Loaded
	_, err := diffStore.diffDataByBlockNode(node)
	if err != nil {
		return err
	}

	diffStore.Loaded[node].diffChild = diffChild
	diffStore.setBlockAsDirty(node)
	return nil
}

// RemoveBlocksDiffData removes blocks diff data for the specified node using provided database context
func (diffStore *DiffStore) RemoveBlocksDiffData(dbContext dbaccess.Context, nodes []*blocknode.Node) error {
	for _, node := range nodes {
		err := diffStore.removeBlockDiffData(dbContext, node)
		if err != nil {
			return err
		}
	}
	return nil
}

func (diffStore *DiffStore) removeBlockDiffData(dbContext dbaccess.Context, node *blocknode.Node) error {
	diffStore.mtx.LowPriorityWriteLock()
	defer diffStore.mtx.LowPriorityWriteUnlock()
	delete(diffStore.Loaded, node)
	err := dbaccess.RemoveDiffData(dbContext, node.Hash)
	if err != nil {
		return err
	}
	return nil
}

func (diffStore *DiffStore) setBlockAsDirty(node *blocknode.Node) {
	diffStore.Dirty[node] = struct{}{}
}

func (diffStore *DiffStore) diffDataByBlockNode(node *blocknode.Node) (*blockUTXODiffData, error) {
	if diffData, ok := diffStore.Loaded[node]; ok {
		return diffData, nil
	}
	diffData, err := diffStore.diffDataFromDB(node.Hash)
	if err != nil {
		return nil, err
	}
	diffStore.Loaded[node] = diffData
	return diffData, nil
}

// DiffByNode returns the diff for the provided block node
func (diffStore *DiffStore) DiffByNode(node *blocknode.Node) (*Diff, error) {
	diffStore.mtx.HighPriorityReadLock()
	defer diffStore.mtx.HighPriorityReadUnlock()
	diffData, err := diffStore.diffDataByBlockNode(node)
	if err != nil {
		return nil, err
	}
	return diffData.Diff, nil
}

// DiffChildByNode returns the diff child for the provided block node
func (diffStore *DiffStore) DiffChildByNode(node *blocknode.Node) (*blocknode.Node, error) {
	diffStore.mtx.HighPriorityReadLock()
	defer diffStore.mtx.HighPriorityReadUnlock()
	diffData, err := diffStore.diffDataByBlockNode(node)
	if err != nil {
		return nil, err
	}
	return diffData.diffChild, nil
}

// diffDataFromDB returns fetched UTXO diff data from the database for the provided block hash
func (diffStore *DiffStore) diffDataFromDB(hash *daghash.Hash) (*blockUTXODiffData, error) {
	serializedBlockDiffData, err := dbaccess.FetchUTXODiffData(diffStore.databaseContext, hash)
	if err != nil {
		return nil, err
	}

	return diffStore.deserializeBlockUTXODiffData(serializedBlockDiffData)
}

// FlushToDB writes all Dirty Diff data to the database.
func (diffStore *DiffStore) FlushToDB(dbContext *dbaccess.TxContext) error {
	diffStore.mtx.HighPriorityWriteLock()
	defer diffStore.mtx.HighPriorityWriteUnlock()
	if len(diffStore.Dirty) == 0 {
		return nil
	}

	// Allocate a buffer here to avoid needless allocations/grows
	// while writing each entry.
	buffer := &bytes.Buffer{}
	for node := range diffStore.Dirty {
		buffer.Reset()
		diffData := diffStore.Loaded[node]
		err := storeDiffData(dbContext, buffer, node.Hash, diffData)
		if err != nil {
			return err
		}
	}
	return nil
}

// ClearDirtyEntries clears all existing dirty entries
func (diffStore *DiffStore) ClearDirtyEntries() {
	diffStore.Dirty = make(map[*blocknode.Node]struct{})
}

// MaxBlueScoreDifferenceToKeepLoaded is the maximum difference
// between the virtual's blueScore and a blocknode.Node's blueScore
// under which to keep Diff data Loaded in memory.
var MaxBlueScoreDifferenceToKeepLoaded uint64 = 100

// ClearOldEntries removes entries whose blue score is lower than
// virtual.blueScore - MaxBlueScoreDifferenceToKeepLoaded.
// Note that parents of virtual are not removed even
// if their blue score is lower than the above.
func (diffStore *DiffStore) ClearOldEntries(virtualBlueScore uint64, virtualParents blocknode.Set) {
	diffStore.mtx.HighPriorityWriteLock()
	defer diffStore.mtx.HighPriorityWriteUnlock()

	minBlueScore := virtualBlueScore - MaxBlueScoreDifferenceToKeepLoaded
	if MaxBlueScoreDifferenceToKeepLoaded > virtualBlueScore {
		minBlueScore = 0
	}

	toRemove := make(map[*blocknode.Node]struct{})
	for node := range diffStore.Loaded {
		if node.BlueScore < minBlueScore && !virtualParents.Contains(node) {
			toRemove[node] = struct{}{}
		}
	}
	for node := range toRemove {
		delete(diffStore.Loaded, node)
	}
}

// storeDiffData stores the UTXO Diff data to the database.
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
