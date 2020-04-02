package blockdag

import (
	"bytes"
	"github.com/kaspanet/go-secp256k1"
	"github.com/kaspanet/kaspad/dbaccess"
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
func (store *multisetStore) flushToDB(dbContext *dbaccess.TxContext) error {
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

		err = store.storeMultiset(dbContext, &hash, w.Bytes())
		if err != nil {
			return err
		}
	}
	return nil
}

func (store *multisetStore) clearNewEntries() {
	store.new = make(map[daghash.Hash]struct{})
}

func (store *multisetStore) init(dbContext dbaccess.Context) error {
	cursor, err := dbaccess.MultisetCursor(dbContext)
	if err != nil {
		return err
	}
	defer cursor.Close()

	for ok := cursor.First(); ok; ok = cursor.Next() {
		key, err := cursor.Key()
		if err != nil {
			return err
		}

		hash, err := daghash.NewHash(key)
		if err != nil {
			return err
		}

		serializedMS, err := cursor.Value()
		if err != nil {
			return err
		}

		ms, err := deserializeMultiset(bytes.NewReader(serializedMS))
		if err != nil {
			return err
		}

		store.loaded[*hash] = *ms
	}
	return nil
}

// storeMultiset stores the multiset data to the database.
func (store *multisetStore) storeMultiset(dbContext dbaccess.Context, blockHash *daghash.Hash, serializedMS []byte) error {
	exists, err := dbaccess.HasMultiset(dbContext, blockHash)
	if err != nil {
		return err
	}

	if exists {
		return errors.Errorf("Can't override an existing multiset database entry for block %s", blockHash)
	}

	return dbaccess.StoreMultiset(dbContext, blockHash, serializedMS)
}
