package blockdag

import (
	"bytes"
	"github.com/kaspanet/kaspad/database"
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/kaspanet/kaspad/wire"
	"github.com/pkg/errors"
	"io"
	"sync"
)

type reachabilityData struct {
	treeNode          *reachabilityTreeNode
	futureCoveringSet futureCoveringBlockSet
}

type reachabilityStore struct {
	dag    *BlockDAG
	dirty  map[daghash.Hash]struct{}
	loaded map[daghash.Hash]*reachabilityData
	mtx    sync.RWMutex
}

func newReachabilityStore(dag *BlockDAG) *reachabilityStore {
	return &reachabilityStore{
		dag:    dag,
		dirty:  make(map[daghash.Hash]struct{}),
		loaded: make(map[daghash.Hash]*reachabilityData),
	}
}

func (store *reachabilityStore) setTreeNode(treeNode *reachabilityTreeNode) error {
	store.mtx.Lock()
	defer store.mtx.Unlock()
	// load the reachability data from DB to store.loaded
	node := treeNode.blockNode
	_, exists := store.reachabilityDataByHash(node.hash)
	if !exists {
		store.loaded[*node.hash] = &reachabilityData{}
	}

	store.loaded[*node.hash].treeNode = treeNode
	store.setBlockAsDirty(node.hash)
	return nil
}

func (store *reachabilityStore) setFutureCoveringSet(node *blockNode, futureCoveringSet futureCoveringBlockSet) error {
	store.mtx.Lock()
	defer store.mtx.Unlock()
	// load the reachability data from DB to store.loaded
	_, exists := store.reachabilityDataByHash(node.hash)
	if !exists {
		return reachabilityNotFoundError(node)
	}

	store.loaded[*node.hash].futureCoveringSet = futureCoveringSet
	store.setBlockAsDirty(node.hash)
	return nil
}

func (store *reachabilityStore) setBlockAsDirty(blockHash *daghash.Hash) {
	store.dirty[*blockHash] = struct{}{}
}

func reachabilityNotFoundError(node *blockNode) error {
	return errors.Errorf("Couldn't find reachability data for block %s", node.hash)
}

func (store *reachabilityStore) treeNodeByBlockNode(node *blockNode) (*reachabilityTreeNode, error) {
	store.mtx.RLock()
	defer store.mtx.RUnlock()
	reachabilityData, exists := store.reachabilityDataByHash(node.hash)
	if !exists {
		return nil, reachabilityNotFoundError(node)
	}
	return reachabilityData.treeNode, nil
}

func (store *reachabilityStore) futureCoveringSetByBlockNode(node *blockNode) (futureCoveringBlockSet, error) {
	store.mtx.RLock()
	defer store.mtx.RUnlock()
	reachabilityData, exists := store.reachabilityDataByHash(node.hash)
	if !exists {
		return nil, reachabilityNotFoundError(node)
	}
	return reachabilityData.futureCoveringSet, nil
}

func (store *reachabilityStore) reachabilityDataByHash(hash *daghash.Hash) (*reachabilityData, bool) {
	reachabilityData, ok := store.loaded[*hash]
	return reachabilityData, ok
}

// flushToDB writes all dirty reachability data to the database. If all writes
// succeed, this clears the dirty set.
func (store *reachabilityStore) flushToDB(dbTx database.Tx) error {
	store.mtx.Lock()
	defer store.mtx.Unlock()
	if len(store.dirty) == 0 {
		return nil
	}

	for hash := range store.dirty {
		reachabilityData := store.loaded[hash]
		err := store.dbStoreReachabilityData(dbTx, &hash, reachabilityData)
		if err != nil {
			return err
		}
	}
	return nil
}

func (store *reachabilityStore) clearDirtyEntries() {
	store.dirty = make(map[daghash.Hash]struct{})
}

func (store *reachabilityStore) init(dbTx database.Tx) error {
	bucket := dbTx.Metadata().Bucket(reachabilityDataBucketName)
	cursor := bucket.Cursor()
	for ok := cursor.First(); ok; ok = cursor.Next() {
		err := store.loadReachabilityDataFromCursor(cursor, false)
		if err != nil {
			return err
		}
	}
	cursor = bucket.Cursor()
	for ok := cursor.First(); ok; ok = cursor.Next() {
		err := store.loadReachabilityDataFromCursor(cursor, true)
		if err != nil {
			return err
		}
	}
	return nil
}

func (store *reachabilityStore) loadReachabilityDataFromCursor(cursor database.Cursor, withConnections bool) error {
	var hash daghash.Hash
	err := hash.SetBytes(cursor.Key())
	if err != nil {
		return err
	}
	reachabilityData, err := store.deserializeReachabilityData(cursor.Value(), withConnections)
	if err != nil {
		return err
	}

	// Connect the treeNode with its blockNode
	reachabilityData.treeNode.blockNode = store.dag.index.LookupNode(&hash)

	store.loaded[hash] = reachabilityData
	return nil
}

// dbStoreReachabilityData stores the reachability data to the database.
// This overwrites the current entry if there exists one.
func (store *reachabilityStore) dbStoreReachabilityData(dbTx database.Tx, hash *daghash.Hash, reachabilityData *reachabilityData) error {
	serializedReachabilyData, err := store.serializeReachabilityData(reachabilityData)
	if err != nil {
		return err
	}

	return dbTx.Metadata().Bucket(reachabilityDataBucketName).Put(hash[:], serializedReachabilyData)
}

func (store *reachabilityStore) serializeReachabilityData(reachabilityData *reachabilityData) ([]byte, error) {
	w := &bytes.Buffer{}
	err := store.serializeTreeNode(w, reachabilityData.treeNode)
	if err != nil {
		return nil, err
	}
	err = store.serializeFutureCoveringSet(w, reachabilityData.futureCoveringSet)
	if err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

func (store *reachabilityStore) serializeTreeNode(w io.Writer, treeNode *reachabilityTreeNode) error {
	// Serialize the interval
	err := store.serializeReachabilityInterval(w, treeNode.interval)
	if err != nil {
		return err
	}

	// Serialize the remaining interval
	err = store.serializeReachabilityInterval(w, treeNode.remainingInterval)
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

func (store *reachabilityStore) serializeReachabilityInterval(w io.Writer, interval *reachabilityInterval) error {
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

func (store *reachabilityStore) serializeFutureCoveringSet(w io.Writer, futureCoveringSet futureCoveringBlockSet) error {
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

func (store *reachabilityStore) deserializeReachabilityData(serializedReachabilityDataBytes []byte, withConnections bool) (*reachabilityData, error) {
	data := &reachabilityData{}
	serializedReachabilityData := bytes.NewBuffer(serializedReachabilityDataBytes)

	// Deserialize the tree node
	treeNode, err := store.deserializeTreeNode(serializedReachabilityData, withConnections)
	if err != nil {
		return nil, err
	}
	data.treeNode = treeNode

	// Deserialize the future covering set
	futureCoveringSet, err := store.deserializeFutureConveringSet(serializedReachabilityData, withConnections)
	if err != nil {
		return nil, err
	}
	data.futureCoveringSet = futureCoveringSet

	return data, nil
}

func (store *reachabilityStore) deserializeTreeNode(r io.Reader, withConnections bool) (*reachabilityTreeNode, error) {
	treeNode := &reachabilityTreeNode{}

	// Deserialize the interval
	interval, err := store.deserializeReachabilityInterval(r)
	if err != nil {
		return nil, err
	}
	treeNode.interval = interval

	// Deserialize the remaining interval
	remainingInterval, err := store.deserializeReachabilityInterval(r)
	if err != nil {
		return nil, err
	}
	treeNode.remainingInterval = remainingInterval

	// Deserialize the parent
	// If this is the zero hash, this node is the genesis and as such doesn't have a parent
	parentHash := &daghash.Hash{}
	err = wire.ReadElement(r, parentHash)
	if err != nil {
		return nil, err
	}
	if !daghash.ZeroHash.IsEqual(parentHash) {
		if withConnections {
			parentReachabilityData, ok := store.reachabilityDataByHash(parentHash)
			if !ok {
				return nil, errors.Errorf("parent reachability data not found for hash: %s", parentHash)
			}
			treeNode.parent = parentReachabilityData.treeNode
		}
	}

	// Deserialize the amount of children
	childAmount, err := wire.ReadVarInt(r)
	if err != nil {
		return nil, err
	}

	// Deserialize the children
	children := make([]*reachabilityTreeNode, childAmount)
	for i := uint64(0); i < childAmount; i++ {
		childHash := &daghash.Hash{}
		err = wire.ReadElement(r, childHash)
		if err != nil {
			return nil, err
		}

		if withConnections {
			childReachabilityData, ok := store.reachabilityDataByHash(childHash)
			if !ok {
				return nil, errors.Errorf("child reachability data not found for hash: %s", parentHash)
			}
			children[i] = childReachabilityData.treeNode
		}
	}
	treeNode.children = children

	return treeNode, nil
}

func (store *reachabilityStore) deserializeReachabilityInterval(r io.Reader) (*reachabilityInterval, error) {
	interval := &reachabilityInterval{}

	// Deserialize start
	start := uint64(0)
	err := wire.ReadElement(r, &start)
	if err != nil {
		return nil, err
	}
	interval.start = start

	// Deserialize end
	end := uint64(0)
	err = wire.ReadElement(r, &end)
	if err != nil {
		return nil, err
	}
	interval.end = end

	return interval, nil
}

func (store *reachabilityStore) deserializeFutureConveringSet(r io.Reader, withConnections bool) (futureCoveringBlockSet, error) {
	// Deserialize the set size
	setSize, err := wire.ReadVarInt(r)
	if err != nil {
		return nil, err
	}

	// Deserialize each block in the set
	futureCoveringSet := make(futureCoveringBlockSet, setSize)
	for i := uint64(0); i < setSize; i++ {
		blockHash := &daghash.Hash{}
		err = wire.ReadElement(r, blockHash)
		if err != nil {
			return nil, err
		}

		if withConnections {
			blockNode := store.dag.index.LookupNode(blockHash)
			if blockNode == nil {
				return nil, errors.Errorf("blockNode not found for hash %s", blockHash)
			}
			blockReachabilityData, ok := store.reachabilityDataByHash(blockHash)
			if !ok {
				return nil, errors.Errorf("block reachability data not found for hash: %s", blockHash)
			}
			futureCoveringSet[i] = &futureCoveringBlock{blockNode: blockNode, treeNode: blockReachabilityData.treeNode}
		}
	}

	return futureCoveringSet, nil
}
