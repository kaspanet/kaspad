// Copyright (c) 2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

/*
Package indexers implements optional block DAG indexes.
*/
package indexers

import (
	"github.com/kaspanet/kaspad/blockdag"
	"github.com/kaspanet/kaspad/database"
	"github.com/kaspanet/kaspad/util"
)

// Indexer provides a generic interface for an indexer that is managed by an
// index manager such as the Manager type provided by this package.
type Indexer interface {
	// Init is invoked when the index manager is first initializing the
	// index.
	Init(dag *blockdag.BlockDAG) error

	// ConnectBlock is invoked when the index manager is notified that a new
	// block has been connected to the DAG.
	ConnectBlock(dbTx database.Tx,
		block *util.Block,
		dag *blockdag.BlockDAG,
		acceptedTxsData blockdag.MultiBlockTxsAcceptanceData,
		virtualTxsAcceptanceData blockdag.MultiBlockTxsAcceptanceData) error
}
