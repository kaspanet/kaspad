// Copyright (c) 2015-2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package blockdag

import (
	"sync"

	"github.com/daglabs/btcd/dagconfig"
	"github.com/daglabs/btcd/dagconfig/daghash"
	"github.com/daglabs/btcd/database"
)

// blockIndex provides facilities for keeping track of an in-memory index of the
// block chain.  Although the name block chain suggests a single chain of
// blocks, it is actually a tree-shaped structure where any node can have
// multiple children.  However, there can only be one active branch which does
// indeed form a chain from the tip all the way back to the genesis block.
type blockIndex struct {
	// The following fields are set when the instance is created and can't
	// be changed afterwards, so there is no need to protect them with a
	// separate mutex.
	db        database.DB
	dagParams *dagconfig.Params

	sync.RWMutex
	index map[daghash.Hash]*blockNode
	dirty map[*blockNode]struct{}
}

// newBlockIndex returns a new empty instance of a block index.  The index will
// be dynamically populated as block nodes are loaded from the database and
// manually added.
func newBlockIndex(db database.DB, dagParams *dagconfig.Params) *blockIndex {
	return &blockIndex{
		db:        db,
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
	_, hasBlock := bi.index[*hash]
	bi.RUnlock()
	return hasBlock
}

// LookupNode returns the block node identified by the provided hash.  It will
// return nil if there is no entry for the hash.
//
// This function is safe for concurrent access.
func (bi *blockIndex) LookupNode(hash *daghash.Hash) *blockNode {
	bi.RLock()
	node := bi.index[*hash]
	bi.RUnlock()
	return node
}

// AddNode adds the provided node to the block index and marks it as dirty.
// Duplicate entries are not checked so it is up to caller to avoid adding them.
//
// This function is safe for concurrent access.
func (bi *blockIndex) AddNode(node *blockNode) {
	bi.Lock()
	bi.addNode(node)
	bi.dirty[node] = struct{}{}
	bi.Unlock()
}

// addNode adds the provided node to the block index, but does not mark it as
// dirty. This can be used while initializing the block index.
//
// This function is NOT safe for concurrent access.
func (bi *blockIndex) addNode(node *blockNode) {
	bi.index[node.hash] = node
}

// NodeStatus provides concurrent-safe access to the status field of a node.
//
// This function is safe for concurrent access.
func (bi *blockIndex) NodeStatus(node *blockNode) blockStatus {
	bi.RLock()
	status := node.status
	bi.RUnlock()
	return status
}

// SetStatusFlags flips the provided status flags on the block node to on,
// regardless of whether they were on or off previously. This does not unset any
// flags currently on.
//
// This function is safe for concurrent access.
func (bi *blockIndex) SetStatusFlags(node *blockNode, flags blockStatus) {
	bi.Lock()
	node.status |= flags
	bi.dirty[node] = struct{}{}
	bi.Unlock()
}

// UnsetStatusFlags flips the provided status flags on the block node to off,
// regardless of whether they were on or off previously.
//
// This function is safe for concurrent access.
func (bi *blockIndex) UnsetStatusFlags(node *blockNode, flags blockStatus) {
	bi.Lock()
	node.status &^= flags
	bi.dirty[node] = struct{}{}
	bi.Unlock()
}

// flushToDB writes all dirty block nodes to the database. If all writes
// succeed, this clears the dirty set.
func (bi *blockIndex) flushToDB() error {
	bi.Lock()
	if len(bi.dirty) == 0 {
		bi.Unlock()
		return nil
	}

	err := bi.db.Update(func(dbTx database.Tx) error {
		for node := range bi.dirty {
			err := dbStoreBlockNode(dbTx, node)
			if err != nil {
				return err
			}
		}
		return nil
	})

	// If write was successful, clear the dirty set.
	if err == nil {
		bi.dirty = make(map[*blockNode]struct{})
	}

	bi.Unlock()
	return err
}
