// Copyright (c) 2015-2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package blocknode

import (
	"encoding/binary"
	"sync"

	"github.com/kaspanet/kaspad/infrastructure/db/dbaccess"
	"github.com/pkg/errors"

	"github.com/kaspanet/kaspad/util/daghash"
)

// Index provides facilities for keeping track of an in-memory Index of the
// block DAG.
type Index struct {
	// The following fields are set when the instance is created and can't
	// be changed afterwards, so there is no need to protect them with a
	// separate mutex.

	sync.RWMutex
	Index map[daghash.Hash]*Node
	dirty map[*Node]struct{}
}

// NewIndex returns a new empty instance of a block Index. The Index will
// be dynamically populated as block nodes are loaded from the database and
// manually added.
func NewIndex() *Index {
	return &Index{
		Index: make(map[daghash.Hash]*Node),
		dirty: make(map[*Node]struct{}),
	}
}

// HaveBlock returns whether or not the block Index contains the provided hash.
//
// This function is safe for concurrent access.
func (bi *Index) HaveBlock(hash *daghash.Hash) bool {
	bi.RLock()
	defer bi.RUnlock()
	_, hasBlock := bi.Index[*hash]
	return hasBlock
}

// LookupNode returns the block node identified by the provided hash. It will
// return nil if there is no entry for the hash.
//
// This function is safe for concurrent access.
func (bi *Index) LookupNode(hash *daghash.Hash) (*Node, bool) {
	bi.RLock()
	defer bi.RUnlock()
	node, ok := bi.Index[*hash]
	return node, ok
}

// LookupNodes returns the list of block nodes identified by provided hashes.
func (bi *Index) LookupNodes(hashes []*daghash.Hash) ([]*Node, error) {
	blocks := make([]*Node, 0, len(hashes))
	for _, hash := range hashes {
		node, ok := bi.LookupNode(hash)
		if !ok {
			return nil, errors.Errorf("Couldn't find block with hash %s", hash)
		}
		blocks = append(blocks, node)
	}
	return blocks, nil
}

// AddNode adds the provided node to the block Index and marks it as dirty.
// Duplicate entries are not checked so it is up to caller to avoid adding them.
//
// This function is safe for concurrent access.
func (bi *Index) AddNode(node *Node) {
	bi.Lock()
	defer bi.Unlock()
	bi.AddNodeNoLock(node)
	bi.dirty[node] = struct{}{}
}

// AddNodeNoLock adds the provided node to the block Index, but does not mark it as
// dirty. This can be used while initializing the block Index.
//
// This function is NOT safe for concurrent access.
func (bi *Index) AddNodeNoLock(node *Node) {
	bi.Index[*node.Hash] = node
}

// BlockNodeStatus provides concurrent-safe access to the status field of a node.
//
// This function is safe for concurrent access.
func (bi *Index) BlockNodeStatus(node *Node) Status {
	bi.RLock()
	defer bi.RUnlock()
	status := node.Status
	return status
}

// SetBlockNodeStatus changes the status of a Node
//
// This function is safe for concurrent access.
func (bi *Index) SetBlockNodeStatus(node *Node, newStatus Status) {
	bi.Lock()
	defer bi.Unlock()
	node.Status = newStatus
	bi.dirty[node] = struct{}{}
}

// FlushToDB writes all dirty block nodes to the database.
func (bi *Index) FlushToDB(dbContext *dbaccess.TxContext) error {
	bi.Lock()
	defer bi.Unlock()
	if len(bi.dirty) == 0 {
		return nil
	}

	for node := range bi.dirty {
		serializedBlockNode, err := SerializeNode(node)
		if err != nil {
			return err
		}
		key := BlockIndexKey(node.Hash, node.BlueScore)
		err = dbaccess.StoreIndexBlock(dbContext, key, serializedBlockNode)
		if err != nil {
			return err
		}
	}
	return nil
}

// ClearDirtyEntries clears all existing dirty entries
func (bi *Index) ClearDirtyEntries() {
	bi.dirty = make(map[*Node]struct{})
}

// BlockIndexKey generates the binary key for an entry in the block Index
// bucket. The key is composed of the block blue score encoded as a big-endian
// 64-bit unsigned int followed by the 32 byte block hash.
// The blue score component is important for iteration order.
func BlockIndexKey(blockHash *daghash.Hash, blueScore uint64) []byte {
	indexKey := make([]byte, daghash.HashSize+8)
	binary.BigEndian.PutUint64(indexKey[0:8], blueScore)
	copy(indexKey[8:daghash.HashSize+8], blockHash[:])
	return indexKey
}

// BlockHashFromBlockIndexKey generates the block hash for the given binary block index key.
func BlockHashFromBlockIndexKey(blockIndexKey []byte) (*daghash.Hash, error) {
	return daghash.NewHash(blockIndexKey[8 : daghash.HashSize+8])
}
