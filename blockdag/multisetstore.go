package blockdag

import (
	"bytes"
	"github.com/kaspanet/go-secp256k1"
	"github.com/kaspanet/kaspad/database"
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/kaspanet/kaspad/util/locks"
	"github.com/pkg/errors"
)

type multisetStore struct {
	dag    *BlockDAG
	new    map[daghash.Hash]struct{}
	loaded map[daghash.Hash]secp256k1.MultiSet
	mtx    *locks.PriorityMutex
}

func newMultisetStore(dag *BlockDAG) *multisetStore {
	return &multisetStore{
		dag:    dag,
		new:    make(map[daghash.Hash]struct{}),
		loaded: make(map[daghash.Hash]secp256k1.MultiSet),
	}
}

func (store *multisetStore) setMultiset(node *blockNode, ms *secp256k1.MultiSet) {
	store.loaded[*node.hash] = *ms
	store.addToNewBlocks(node.hash)
}

func (store *multisetStore) addToNewBlocks(blockHash *daghash.Hash) {
	store.new[*blockHash] = struct{}{}
}

func multisetNotFoundError(blockHash *daghash.Hash) error {
	return errors.Errorf("Couldn't find multiset data for block %s", blockHash)
}

func (store *multisetStore) multisetByBlockNode(node *blockNode) (*secp256k1.MultiSet, error) {
	ms, exists := store.multisetByBlockHash(node.hash)
	if !exists {
		return nil, multisetNotFoundError(node.hash)
	}
	return ms, nil
}

func (store *multisetStore) multisetByBlockHash(hash *daghash.Hash) (*secp256k1.MultiSet, bool) {
	ms, ok := store.loaded[*hash]
	return &ms, ok
}

// flushToDB writes all new multiset data to the database.
func (store *multisetStore) flushToDB(dbTx database.Tx) error {
	if len(store.new) == 0 {
		return nil
	}

	w := &bytes.Buffer{}
	for hash := range store.new {
		hash := hash // Copy hash to a new variable to avoid passing the same pointer

		w.Reset()
		ms, exists := store.loaded[hash]
		if !exists {
			return multisetNotFoundError(&hash)
		}

		err := serializeMultiset(w, &ms)
		if err != nil {
			return err
		}

		err = store.dbStoreMultiset(dbTx, &hash, w.Bytes())
		if err != nil {
			return err
		}
	}
	return nil
}

func (store *multisetStore) clearNewEntries() {
	store.new = make(map[daghash.Hash]struct{})
}

func (store *multisetStore) init(dbTx database.Tx) error {
	bucket := dbTx.Metadata().Bucket(multisetBucketName)
	cursor := bucket.Cursor()
	for ok := cursor.First(); ok; ok = cursor.Next() {
		hash, err := daghash.NewHash(cursor.Key())
		if err != nil {
			return err
		}

		ms, err := deserializeMultiset(bytes.NewReader(cursor.Value()))
		if err != nil {
			return err
		}

		store.loaded[*hash] = *ms
	}
	return nil
}

// dbStoreMultiset stores the multiset data to the database.
func (store *multisetStore) dbStoreMultiset(dbTx database.Tx, blockHash *daghash.Hash, serializedMS []byte) error {
	bucket := dbTx.Metadata().Bucket(multisetBucketName)
	if bucket.Get(blockHash[:]) != nil {
		return errors.Errorf("Can't override an existing multiset database entry for block %s", blockHash)
	}
	return bucket.Put(blockHash[:], serializedMS)
}
