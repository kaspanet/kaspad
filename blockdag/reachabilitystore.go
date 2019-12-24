package blockdag

import (
	"github.com/kaspanet/kaspad/database"
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/kaspanet/kaspad/util/locks"
	"github.com/pkg/errors"
)

type reachabilityData struct {
	treeNode          *reachabilityTreeNode
	futureCoveringSet *futureCoveringBlockSet
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

func (store *reachabilityStore) setTreeNode(node *blockNode, treeNode *reachabilityTreeNode) error {
	store.mtx.HighPriorityWriteLock()
	defer store.mtx.HighPriorityWriteUnlock()
	// load the reachability data from DB to store.loaded
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

func (store *reachabilityStore) setFutureCoveringSet(node *blockNode, futureCoveringSet *futureCoveringBlockSet) error {
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

func (store *reachabilityStore) futureCoveringSetByBlockNode(node *blockNode) (*futureCoveringBlockSet, error) {
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

func deserializeReachabilityData(serializedReachabilityData []byte) (*reachabilityData, error) {
	return nil, nil
}

func serializeReachabilityData(reachabilityData *reachabilityData) ([]byte, error) {
	return nil, nil
}
