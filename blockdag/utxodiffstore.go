package blockdag

import (
	"bytes"
	"encoding/binary"
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
	sync.RWMutex
}

func newUTXODiffStore(dag *BlockDAG) *utxoDiffStore {
	return &utxoDiffStore{
		dag:    dag,
		dirty:  make(map[daghash.Hash]struct{}),
		loaded: make(map[daghash.Hash]*blockUTXODiffData),
	}
}

func (diffStore *utxoDiffStore) setBlockDiff(node *blockNode, diff *UTXODiff) error {
	diffStore.Lock()
	defer diffStore.Unlock()
	// load the diff data from DB to diffStore.loaded
	_, err := diffStore.get(node.hash)
	if err != nil {
		return err
	}

	diffStore.loaded[*node.hash].diff = diff
	diffStore.setBlockAsDirty(node.hash)
	return nil
}

func (diffStore *utxoDiffStore) setBlockDiffChild(node *blockNode, diffChild *blockNode) error {
	diffStore.Lock()
	defer diffStore.Unlock()
	// load the diff data from DB to diffStore.loaded
	_, err := diffStore.get(node.hash)
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

func (diffStore *utxoDiffStore) get(hash *daghash.Hash) (*blockUTXODiffData, error) {
	if diffData, ok := diffStore.loaded[*hash]; ok {
		return diffData, nil
	}
	diffData, err := diffStore.getFromDB(hash)
	if err != nil {
		return nil, err
	}
	diffStore.loaded[*hash] = diffData
	return diffData, nil
}

func (diffStore *utxoDiffStore) getBlockDiff(node *blockNode) (*UTXODiff, error) {
	diffStore.RLock()
	defer diffStore.RUnlock()
	diffData, err := diffStore.get(node.hash)
	if err != nil {
		return nil, err
	}
	return diffData.diff, nil
}

func (diffStore *utxoDiffStore) getBlockDiffChild(node *blockNode) (*blockNode, error) {
	diffStore.RLock()
	defer diffStore.RUnlock()
	diffData, err := diffStore.get(node.hash)
	if err != nil {
		return nil, err
	}
	return diffData.diffChild, nil
}

func (diffStore *utxoDiffStore) getFromDB(hash *daghash.Hash) (*blockUTXODiffData, error) {
	var diffData *blockUTXODiffData
	err := diffStore.dag.db.View(func(dbTx database.Tx) error {
		bucket := dbTx.Metadata().Bucket(utxoDiffsBucketName)
		serializedBlockDiffData := bucket.Get(hash[:])
		if serializedBlockDiffData != nil {
			var err error
			diffData, err = diffStore.deserializeBlockUTXODiffData(serializedBlockDiffData)
			return err
		}
		diffData = &blockUTXODiffData{}
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

	diffData.diff = NewUTXODiff()

	err = deserializeDiffEntriesAndAddToUTXOCollection(serializedDiffData, diffData.diff.toAdd)
	if err != nil {
		return nil, err
	}

	err = deserializeDiffEntriesAndAddToUTXOCollection(serializedDiffData, diffData.diff.toRemove)
	if err != nil {
		return nil, err
	}

	return diffData, nil
}

func deserializeDiffEntriesAndAddToUTXOCollection(r io.Reader, collection utxoCollection) error {
	count, err := wire.ReadVarInt(r)
	if err != nil {
		return err
	}
	for i := uint64(0); i < count; i++ {
		outPointSize, err := wire.ReadVarInt(r)
		if err != nil {
			return err
		}

		serializedOutPoint := make([]byte, outPointSize)
		err = binary.Read(r, byteOrder, serializedOutPoint)
		if err != nil {
			return err
		}
		outPoint, err := deserializeOutPoint(serializedOutPoint)
		if err != nil {
			return err
		}

		utxoEntrySize, err := wire.ReadVarInt(r)
		if err != nil {
			return err
		}
		serializedEntry := make([]byte, utxoEntrySize)
		err = binary.Read(r, byteOrder, serializedEntry)
		if err != nil {
			return err
		}
		utxoEntry, err := deserializeUTXOEntry(serializedEntry)
		if err != nil {
			return err
		}
		collection.add(*outPoint, utxoEntry)
	}
	return nil
}

func serializeBlockUTXODiffData(diffData *blockUTXODiffData) ([]byte, error) {
	serializedDiffData := &bytes.Buffer{}
	hasDiffChild := diffData.diffChild != nil
	err := wire.WriteElement(serializedDiffData, hasDiffChild)
	if err != nil {
		return nil, err
	}
	if hasDiffChild {
		err := wire.WriteElement(serializedDiffData, diffData.diffChild.hash)
		if err != nil {
			return nil, err
		}
	}

	err = serializeUTXOCollection(serializedDiffData, diffData.diff.toAdd)
	if err != nil {
		return nil, err
	}

	err = serializeUTXOCollection(serializedDiffData, diffData.diff.toRemove)
	if err != nil {
		return nil, err
	}

	return serializedDiffData.Bytes(), nil
}

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
	diffStore.Lock()
	defer diffStore.Unlock()
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
