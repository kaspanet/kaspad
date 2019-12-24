package blockdag

import (
	"bytes"
	"github.com/kaspanet/kaspad/database"
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/kaspanet/kaspad/util/locks"
	"github.com/kaspanet/kaspad/wire"
	"github.com/pkg/errors"
	"io"
)

type reachabilityData struct {
	treeNode          *reachabilityTreeNode
	futureCoveringSet futureCoveringBlockSet
}

type reachabilityStore struct {
	dag    *BlockDAG
	dirty  map[daghash.Hash]struct{}
	loaded map[daghash.Hash]*reachabilityData
	mtx    *locks.PriorityMutex
}

func newReachabilityStore(dag *BlockDAG) *reachabilityStore {
	return &reachabilityStore{
		dag:    dag,
		dirty:  make(map[daghash.Hash]struct{}),
		loaded: make(map[daghash.Hash]*reachabilityData),
		mtx:    locks.NewPriorityMutex(),
	}
}

func (store *reachabilityStore) setTreeNode(treeNode *reachabilityTreeNode) error {
	store.mtx.HighPriorityWriteLock()
	defer store.mtx.HighPriorityWriteUnlock()
	// load the reachability data from DB to store.loaded
	node := treeNode.blockNode
	_, exists, err := store.reachabilityDataByHash(node.hash)
	if err != nil {
		return err
	}
	if !exists {
		store.loaded[*node.hash] = &reachabilityData{}
	}

	store.loaded[*node.hash].treeNode = treeNode
	store.setBlockAsDirty(node.hash)
	return nil
}

func (store *reachabilityStore) setFutureCoveringSet(node *blockNode, futureCoveringSet futureCoveringBlockSet) error {
	store.mtx.HighPriorityWriteLock()
	defer store.mtx.HighPriorityWriteUnlock()
	// load the reachability data from DB to store.loaded
	_, exists, err := store.reachabilityDataByHash(node.hash)
	if err != nil {
		return err
	}
	if !exists {
		return reachabilityNotFoundError(node)
	}

	store.loaded[*node.hash].futureCoveringSet = futureCoveringSet
	store.setBlockAsDirty(node.hash)
	return nil
}

func (store *reachabilityStore) reachabilityDataByHash(hash *daghash.Hash) (*reachabilityData, bool, error) {
	if reachabilityData, ok := store.loaded[*hash]; ok {
		return reachabilityData, true, nil
	}
	reachabilityData, err := store.reachabilityDataFromDB(hash)
	if err != nil {
		return nil, false, err
	}
	exists := reachabilityData != nil
	if exists {
		store.loaded[*hash] = reachabilityData
	}
	return reachabilityData, exists, nil
}

func (store *reachabilityStore) setBlockAsDirty(blockHash *daghash.Hash) {
	store.dirty[*blockHash] = struct{}{}
}

func reachabilityNotFoundError(node *blockNode) error {
	return errors.Errorf("Couldn't find reachability data for block %s", node.hash)
}

func (store *reachabilityStore) treeNodeByBlockNode(node *blockNode) (*reachabilityTreeNode, error) {
	store.mtx.HighPriorityReadLock()
	defer store.mtx.HighPriorityReadUnlock()
	reachabilityData, exists, err := store.reachabilityDataByHash(node.hash)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, reachabilityNotFoundError(node)
	}
	return reachabilityData.treeNode, nil
}

func (store *reachabilityStore) futureCoveringSetByBlockNode(node *blockNode) (futureCoveringBlockSet, error) {
	store.mtx.HighPriorityReadLock()
	defer store.mtx.HighPriorityReadUnlock()
	reachabilityData, exists, err := store.reachabilityDataByHash(node.hash)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, reachabilityNotFoundError(node)
	}
	return reachabilityData.futureCoveringSet, nil
}

func (store *reachabilityStore) reachabilityDataFromDB(hash *daghash.Hash) (*reachabilityData, error) {
	var reachabilityData *reachabilityData
	err := store.dag.db.View(func(dbTx database.Tx) error {
		bucket := dbTx.Metadata().Bucket(reachabilityDataBucketName)
		serializedReachabilityData := bucket.Get(hash[:])
		if serializedReachabilityData != nil {
			var err error
			reachabilityData, err = deserializeReachabilityData(serializedReachabilityData)
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return reachabilityData, nil
}

// flushToDB writes all dirty reachability data to the database. If all writes
// succeed, this clears the dirty set.
func (store *reachabilityStore) flushToDB(dbTx database.Tx) error {
	store.mtx.HighPriorityWriteLock()
	defer store.mtx.HighPriorityWriteUnlock()
	if len(store.dirty) == 0 {
		return nil
	}

	for hash := range store.dirty {
		reachabilityData := store.loaded[hash]
		err := dbStoreReachabilityData(dbTx, &hash, reachabilityData)
		if err != nil {
			return err
		}
	}
	return nil
}

func (store *reachabilityStore) clearDirtyEntries() {
	store.dirty = make(map[daghash.Hash]struct{})
}

// dbStoreReachabilityData stores the reachability data to the database.
// This overwrites the current entry if there exists one.
func dbStoreReachabilityData(dbTx database.Tx, hash *daghash.Hash, reachabilityData *reachabilityData) error {
	serializedReachabilyData, err := serializeReachabilityData(reachabilityData)
	if err != nil {
		return err
	}

	return dbTx.Metadata().Bucket(reachabilityDataBucketName).Put(hash[:], serializedReachabilyData)
}

func serializeReachabilityData(reachabilityData *reachabilityData) ([]byte, error) {
	w := &bytes.Buffer{}
	err := serializeTreeNode(w, reachabilityData.treeNode)
	if err != nil {
		return nil, err
	}
	err = serializeFutureCoveringSet(w, reachabilityData.futureCoveringSet)
	if err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

func serializeTreeNode(w io.Writer, treeNode *reachabilityTreeNode) error {
	// Serialize the interval
	err := serializeReachabilityInterval(w, &treeNode.interval)
	if err != nil {
		return err
	}

	// Serialize the remaining interval
	err = serializeReachabilityInterval(w, &treeNode.remainingInterval)
	if err != nil {
		return err
	}

	// Serialize the parent
	// If this is the genesis block, write the zero hash instead
	parentHash := &daghash.ZeroHash
	if treeNode.parent != nil {
		parentHash = treeNode.parent.blockNode.hash
	}
	err = wire.WriteElement(w, parentHash)
	if err != nil {
		return err
	}

	// Serialize the amount of children
	err = wire.WriteVarInt(w, uint64(len(treeNode.children)))
	if err != nil {
		return err
	}

	// Serialize the children
	for _, child := range treeNode.children {
		err = wire.WriteElement(w, child.blockNode.hash)
		if err != nil {
			return err
		}
	}

	return nil
}

func serializeReachabilityInterval(w io.Writer, interval *reachabilityInterval) error {
	// Serialize start
	err := wire.WriteElement(w, interval.start)
	if err != nil {
		return err
	}

	// Serialize end
	err = wire.WriteElement(w, interval.end)
	if err != nil {
		return err
	}

	return nil
}

func serializeFutureCoveringSet(w io.Writer, futureCoveringSet futureCoveringBlockSet) error {
	// Serialize the set size
	err := wire.WriteVarInt(w, uint64(len(futureCoveringSet)))
	if err != nil {
		return err
	}

	// Serialize each block in the set
	for _, block := range futureCoveringSet {
		err = wire.WriteElement(w, block.blockNode.hash)
		if err != nil {
			return err
		}
	}

	return nil
}

func deserializeReachabilityData(serializedReachabilityDataBytes []byte) (*reachabilityData, error) {
	serializedReachabilityData := bytes.NewBuffer(serializedReachabilityDataBytes)
	treeNode, err := deserializeTreeNode(serializedReachabilityData)
	if err != nil {
		return nil, err
	}
	futureCoveringSet, err := deserializeFutureConveringSet(serializedReachabilityData)
	if err != nil {
		return nil, err
	}
	return &reachabilityData{treeNode: treeNode, futureCoveringSet: futureCoveringSet}, nil
}

func deserializeTreeNode(r io.Reader) (*reachabilityTreeNode, error) {
	return nil, nil
}

func deserializeFutureConveringSet(r io.Reader) (futureCoveringBlockSet, error) {
	return nil, nil
}
