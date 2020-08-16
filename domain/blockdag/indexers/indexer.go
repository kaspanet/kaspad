// Copyright (c) 2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

/*
Package indexers implements optional block DAG indexes.
*/
package indexers

import (
	"github.com/kaspanet/kaspad/domain/blockdag"
	"github.com/kaspanet/kaspad/infrastructure/dbaccess"
	"github.com/kaspanet/kaspad/util/daghash"
)

// Indexer provides a generic interface for an indexer that is managed by an
// index manager such as the Manager type provided by this package.
type Indexer interface {
	// Init is invoked when the index manager is first initializing the
	// index.
	Init(dag *blockdag.BlockDAG, databaseContext *dbaccess.DatabaseContext) error

	// ConnectBlock is invoked when the index manager is notified that a new
	// block has been connected to the DAG.
	ConnectBlock(dbContext *dbaccess.TxContext,
		blockHash *daghash.Hash,
		acceptedTxsData blockdag.MultiBlockTxsAcceptanceData) error
}
