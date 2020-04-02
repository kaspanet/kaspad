// Copyright (c) 2015-2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package blockdag

import (
	"github.com/kaspanet/kaspad/dbaccess"
	"sync"

	"github.com/kaspanet/kaspad/dagconfig"
	"github.com/kaspanet/kaspad/util/daghash"
)

// blockIndex provides facilities for keeping track of an in-memory index of the
// block DAG.
type blockIndex struct {
	// The following fields are set when the instance is created and can't
	// be changed afterwards, so there is no need to protect them with a
	// separate mutex.
	dagParams *dagconfig.Params

	sync.RWMutex
	index map[daghash.Hash]*blockNode
	dirty map[*blockNode]struct{}
}

// newBlockIndex returns a new empty instance of a block index. The index will
// be dynamically populated as block nodes are loaded from the database and
// manually added.
func newBlockIndex(dagParams *dagconfig.Params) *blockIndex {
	return &blockIndex{
		dagParams: dagParams,
		index:     make(map[daghash.Hash]*blockNode),
		dirty:     make(map[*blockNode]struct{}),
	}
}

// HaveBlock returns whether or not the block index contains the provided hash.
//
// This function is safe for concurrent access.
func (bi *blockIndex) HaveBlock(hash *daghash.Hash) bool {
	bi.RLock()
	defer bi.RUnlock()
	_, hasBlock := bi.index[*hash]
	return hasBlock
}

// LookupNode returns the block node identified by the provided hash. It will
// return nil if there is no entry for the hash.
//
// This function is safe for concurrent access.
func (bi *blockIndex) LookupNode(hash *daghash.Hash) *blockNode {
	bi.RLock()
	defer bi.RUnlock()
	node := bi.index[*hash]
	return node
}

// AddNode adds the provided node to the block index and marks it as dirty.
// Duplicate entries are not checked so it is up to caller to avoid adding them.
//
// This function is safe for concurrent access.
func (bi *blockIndex) AddNode(node *blockNode) {
	bi.Lock()
	defer bi.Unlock()
	bi.addNode(node)
	bi.dirty[node] = struct{}{}
}

// addNode adds the provided node to the block index, but does not mark it as
// dirty. This can be used while initializing the block index.
//
// This function is NOT safe for concurrent access.
func (bi *blockIndex) addNode(node *blockNode) {
	bi.index[*node.hash] = node
}

// NodeStatus provides concurrent-safe access to the status field of a node.
//
// This function is safe for concurrent access.
func (bi *blockIndex) NodeStatus(node *blockNode) blockStatus {
	bi.RLock()
	defer bi.RUnlock()
	status := node.status
	return status
}

// SetStatusFlags flips the provided status flags on the block node to on,
// regardless of whether they were on or off previously. This does not unset any
// flags currently on.
//
// This function is safe for concurrent access.
func (bi *blockIndex) SetStatusFlags(node *blockNode, flags blockStatus) {
	bi.Lock()
	defer bi.Unlock()
	node.status |= flags
	bi.dirty[node] = struct{}{}
}

// UnsetStatusFlags flips the provided status flags on the block node to off,
// regardless of whether they were on or off previously.
//
// This function is safe for concurrent access.
func (bi *blockIndex) UnsetStatusFlags(node *blockNode, flags blockStatus) {
	bi.Lock()
	defer bi.Unlock()
	node.status &^= flags
	bi.dirty[node] = struct{}{}
}

// flushToDB writes all dirty block nodes to the database.
func (bi *blockIndex) flushToDB(dbContext *dbaccess.TxContext) error {
	bi.Lock()
	defer bi.Unlock()
	if len(bi.dirty) == 0 {
		return nil
	}

	for node := range bi.dirty {
		serializedBlockNode, err := serializeBlockNode(node)
		if err != nil {
			return err
		}
		key := blockIndexKey(node.hash, node.blueScore)
		err = dbaccess.StoreIndexBlock(dbContext, key, serializedBlockNode)
		if err != nil {
			return err
		}
	}
	return nil
}

func (bi *blockIndex) clearDirtyEntries() {
	bi.dirty = make(map[*blockNode]struct{})
}
