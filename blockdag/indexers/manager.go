// Copyright (c) 2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package indexers

import (
	"github.com/kaspanet/kaspad/blockdag"
	"github.com/kaspanet/kaspad/database"
	"github.com/kaspanet/kaspad/util"
)

// Manager defines an index manager that manages multiple optional indexes and
// implements the blockdag.IndexManager interface so it can be seamlessly
// plugged into normal DAG processing.
type Manager struct {
	enabledIndexes []Indexer
}

// Ensure the Manager type implements the blockdag.IndexManager interface.
var _ blockdag.IndexManager = (*Manager)(nil)

// Init initializes the enabled indexes.
// This is part of the blockdag.IndexManager interface.
func (m *Manager) Init(blockDAG *blockdag.BlockDAG) error {
	for _, indexer := range m.enabledIndexes {
		if err := indexer.Init(blockDAG); err != nil {
			return err
		}
	}

	return nil
}

// ConnectBlock must be invoked when a block is added to the DAG. It
// keeps track of the state of each index it is managing, performs some sanity
// checks, and invokes each indexer.
//
// This is part of the blockdag.IndexManager interface.
func (m *Manager) ConnectBlock(dbTx database.Tx, block *util.Block, dag *blockdag.BlockDAG,
	txsAcceptanceData blockdag.MultiBlockTxsAcceptanceData, virtualTxsAcceptanceData blockdag.MultiBlockTxsAcceptanceData) error {

	// Call each of the currently active optional indexes with the block
	// being connected so they can update accordingly.
	for _, index := range m.enabledIndexes {
		// Notify the indexer with the connected block so it can index it.
		if err := index.ConnectBlock(dbTx, block, dag, txsAcceptanceData, virtualTxsAcceptanceData); err != nil {
			return err
		}
	}
	return nil
}

// NewManager returns a new index manager with the provided indexes enabled.
//
// The manager returned satisfies the blockdag.IndexManager interface and thus
// cleanly plugs into the normal blockdag processing path.
func NewManager(enabledIndexes []Indexer) *Manager {
	return &Manager{
		enabledIndexes: enabledIndexes,
	}
}
