package blockdag

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"sync"

	"github.com/daglabs/btcd/dagconfig/daghash"
	"github.com/daglabs/btcd/database"
	"github.com/daglabs/btcd/wire"
)

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
	_, err := diffStore.diffDataByHashAllowNotFound(node.hash)
	if err != nil {
		return err
	}

	diffStore.loaded[*node.hash].diff = diff
	diffStore.setBlockAsDirty(node.hash)
	return nil
}

func (diffStore *utxoDiffStore) setBlockDiffChild(node *blockNode, diffChild *blockNode) error {
	diffStore.mtx.Lock()
	defer diffStore.mtx.Unlock()
	// load the diff data from DB to diffStore.loaded
	_, err := diffStore.diffDataByHashDisallowNotFound(node.hash)
	if err != nil {
		return err
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

func (diffStore *utxoDiffStore) diffDataByHashAllowNotFound(hash *daghash.Hash) (*blockUTXODiffData, error) {
	diffData, exists, err := diffStore.diffDataByHash(hash)
	if err != nil {
		return nil, err
	}
	if !exists {
		diffStore.loaded[*hash] = &blockUTXODiffData{}
	}
	return diffData, nil
}

func (diffStore *utxoDiffStore) diffDataByHashDisallowNotFound(hash *daghash.Hash) (*blockUTXODiffData, error) {
	diffData, exists, err := diffStore.diffDataByHash(hash)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("Couldn't find diff data for block %s", hash)
	}
	return diffData, nil
}

func (diffStore *utxoDiffStore) diffByNode(node *blockNode) (*UTXODiff, error) {
	diffStore.mtx.RLock()
	defer diffStore.mtx.RUnlock()
	diffData, err := diffStore.diffDataByHashDisallowNotFound(node.hash)
	if err != nil {
		return nil, err
	}
	return diffData.diff, nil
}

func (diffStore *utxoDiffStore) diffChildByNode(node *blockNode) (*blockNode, error) {
	diffStore.mtx.RLock()
	defer diffStore.mtx.RUnlock()
	diffData, err := diffStore.diffDataByHashDisallowNotFound(node.hash)
	if err != nil {
		return nil, err
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

func (diffStore *utxoDiffStore) deserializeBlockUTXODiffData(serializedDiffDataBytes []byte) (*blockUTXODiffData, error) {
	diffData := &blockUTXODiffData{}
	serializedDiffData := bytes.NewBuffer(serializedDiffDataBytes)

	var hasDiffChild bool
	err := wire.ReadElement(serializedDiffData, &hasDiffChild)
	if err != nil {
		return nil, err
	}

	if hasDiffChild {
		hash := &daghash.Hash{}
		err := wire.ReadElement(serializedDiffData, hash)
		if err != nil {
			return nil, err
		}
		diffData.diffChild = diffStore.dag.index.LookupNode(hash)
	}

	diffData.diff = &UTXODiff{}

	diffData.diff.toAdd, err = deserializeDiffEntries(serializedDiffData)
	if err != nil {
		return nil, err
	}

	diffData.diff.toRemove, err = deserializeDiffEntries(serializedDiffData)
	if err != nil {
		return nil, err
	}

	return diffData, nil
}

func deserializeDiffEntries(r io.Reader) (utxoCollection, error) {
	count, err := wire.ReadVarInt(r)
	if err != nil {
		return nil, err
	}
	collection := utxoCollection{}
	for i := uint64(0); i < count; i++ {
		outPointSize, err := wire.ReadVarInt(r)
		if err != nil {
			return nil, err
		}

		serializedOutPoint := make([]byte, outPointSize)
		err = binary.Read(r, byteOrder, serializedOutPoint)
		if err != nil {
			return nil, err
		}
		outPoint, err := deserializeOutPoint(serializedOutPoint)
		if err != nil {
			return nil, err
		}

		utxoEntrySize, err := wire.ReadVarInt(r)
		if err != nil {
			return nil, err
		}
		serializedEntry := make([]byte, utxoEntrySize)
		err = binary.Read(r, byteOrder, serializedEntry)
		if err != nil {
			return nil, err
		}
		utxoEntry, err := deserializeUTXOEntry(serializedEntry)
		if err != nil {
			return nil, err
		}
		collection.add(*outPoint, utxoEntry)
	}
	return collection, nil
}

// serializeBlockUTXODiffData serializes diff data in the following format:
// the first byte indicates if the diffData has a diff child, and if it
// has, its hash will be written after. After that the utxo diff is serialized.
func serializeBlockUTXODiffData(diffData *blockUTXODiffData) ([]byte, error) {
	w := &bytes.Buffer{}
	hasDiffChild := diffData.diffChild != nil
	err := wire.WriteElement(w, hasDiffChild)
	if err != nil {
		return nil, err
	}
	if hasDiffChild {
		err := wire.WriteElement(w, diffData.diffChild.hash)
		if err != nil {
			return nil, err
		}
	}

	err = serializeUTXODiff(w, diffData.diff)
	if err != nil {
		return nil, err
	}

	return w.Bytes(), nil
}

// serializeUTXODiff serializes UTXODiff by serializing
// UTXODiff.toAdd and UTXODiff.toRemove one after the other.
func serializeUTXODiff(w io.Writer, diff *UTXODiff) error {
	err := serializeUTXOCollection(w, diff.toAdd)
	if err != nil {
		return err
	}

	err = serializeUTXOCollection(w, diff.toRemove)
	if err != nil {
		return err
	}
	return nil
}

// serializeUTXOCollection serializes utxoCollection by iterating over
// the utxo entries and serializing them and their corresponding outpoint
// prefixed by a varint that indicates their size.
func serializeUTXOCollection(w io.Writer, collection utxoCollection) error {
	err := wire.WriteVarInt(w, uint64(len(collection)))
	if err != nil {
		return err
	}
	for outPoint, utxoEntry := range collection {
		serializedOutPoint := *outpointKey(outPoint)
		err = wire.WriteVarInt(w, uint64(len(serializedOutPoint)))
		if err != nil {
			return err
		}

		err := binary.Write(w, byteOrder, serializedOutPoint)
		if err != nil {
			return err
		}

		serializedUTXOEntry, err := serializeUTXOEntry(utxoEntry)
		if err != nil {
			return err
		}
		err = wire.WriteVarInt(w, uint64(len(serializedUTXOEntry)))
		if err != nil {
			return err
		}
		err = binary.Write(w, byteOrder, serializedUTXOEntry)
		if err != nil {
			return err
		}
	}
	return nil
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
