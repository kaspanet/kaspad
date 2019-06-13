// Copyright (c) 2014-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package blockdag_test

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/daglabs/btcd/blockdag"
	"github.com/daglabs/btcd/dagconfig"
	"github.com/daglabs/btcd/database"
	_ "github.com/daglabs/btcd/database/ffldb"
	"github.com/daglabs/btcd/util"
)

// This example demonstrates how to create a new chain instance and use
// ProcessBlock to attempt to add a block to the chain.  As the package
// overview documentation describes, this includes all of the Bitcoin consensus
// rules.  This example intentionally attempts to insert a duplicate genesis
// block to illustrate how an invalid block is handled.
func ExampleBlockDAG_ProcessBlock() {
	// Create a new database to store the accepted blocks into.  Typically
	// this would be opening an existing database and would not be deleting
	// and creating a new database like this, but it is done here so this is
	// a complete working example and does not leave temporary files laying
	// around.
	dbPath := filepath.Join(os.TempDir(), "exampleprocessblock")
	_ = os.RemoveAll(dbPath)
	db, err := database.Create("ffldb", dbPath, dagconfig.MainNetParams.Net)
	if err != nil {
		fmt.Printf("Failed to create database: %v\n", err)
		return
	}
	defer os.RemoveAll(dbPath)
	defer db.Close()

	// Create a new BlockDAG instance using the underlying database for
	// the main bitcoin network.  This example does not demonstrate some
	// of the other available configuration options such as specifying a
	// notification callback and signature cache.  Also, the caller would
	// ordinarily keep a reference to the median time source and add time
	// values obtained from other peers on the network so the local time is
	// adjusted to be in agreement with other peers.
	chain, err := blockdag.New(&blockdag.Config{
		DB:         db,
		DAGParams:  &dagconfig.MainNetParams,
		TimeSource: blockdag.NewMedianTime(),
	})
	if err != nil {
		fmt.Printf("Failed to create chain instance: %v\n", err)
		return
	}

	// Process a block.  For this example, we are going to intentionally
	// cause an error by trying to process the genesis block which already
	// exists.
	genesisBlock := util.NewBlock(dagconfig.MainNetParams.GenesisBlock)
	isOrphan, err := chain.ProcessBlock(genesisBlock,
		blockdag.BFNone)
	if err != nil {
		fmt.Printf("Failed to process block: %v\n", err)
		return
	}
	fmt.Printf("Block accepted. Is it an orphan?: %v", isOrphan)

	// Output:
	// Failed to process block: already have block 09c081c17dbfb421fc015aa77ba17e278c1a1cf9a7311a2042ffb8932d2ff365
}
