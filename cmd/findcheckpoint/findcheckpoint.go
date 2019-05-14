// Copyright (c) 2013-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/daglabs/btcd/blockdag"
	"github.com/daglabs/btcd/dagconfig"
	"github.com/daglabs/btcd/database"
	"github.com/daglabs/btcd/util/daghash"
)

const blockDbNamePrefix = "blocks"

var (
	cfg *config
)

// loadBlockDB opens the block database and returns a handle to it.
func loadBlockDB() (database.DB, error) {
	// The database name is based on the database type.
	dbName := blockDbNamePrefix + "_" + cfg.DbType
	dbPath := filepath.Join(cfg.DataDir, dbName)
	fmt.Printf("Loading block database from '%s'\n", dbPath)
	db, err := database.Open(cfg.DbType, dbPath, activeNetParams.Net)
	if err != nil {
		return nil, err
	}
	return db, nil
}

// findCandidates searches the DAG backwards for checkpoint candidates and
// returns a slice of found candidates, if any.  It also stops searching for
// candidates at the last checkpoint that is already hard coded since there
// is no point in finding candidates before already existing checkpoints.
func findCandidates(dag *blockdag.BlockDAG, highestTipHash *daghash.Hash) ([]*dagconfig.Checkpoint, error) {
	// Start with the selected tip.
	block, err := dag.BlockByHash(highestTipHash)
	if err != nil {
		return nil, err
	}

	// Get the latest known checkpoint.
	latestCheckpoint := dag.LatestCheckpoint()
	if latestCheckpoint == nil {
		// Set the latest checkpoint to the genesis block if there isn't
		// already one.
		latestCheckpoint = &dagconfig.Checkpoint{
			Hash:   activeNetParams.GenesisHash,
			Height: 0,
		}
	}

	// The latest known block must be at least the last known checkpoint
	// plus required checkpoint confirmations.
	checkpointConfirmations := uint64(blockdag.CheckpointConfirmations)
	requiredHeight := latestCheckpoint.Height + checkpointConfirmations
	if block.Height() < requiredHeight {
		return nil, fmt.Errorf("the block database is only at height "+
			"%d which is less than the latest checkpoint height "+
			"of %d plus required confirmations of %d",
			block.Height(), latestCheckpoint.Height,
			checkpointConfirmations)
	}

	// For the first checkpoint, the required height is any block after the
	// genesis block, so long as the DAG has at least the required number
	// of confirmations (which is enforced above).
	if len(activeNetParams.Checkpoints) == 0 {
		requiredHeight = 1
	}

	// Indeterminate progress setup.
	numBlocksToTest := block.Height() - requiredHeight
	progressInterval := (numBlocksToTest / 100) + 1 // min 1
	fmt.Print("Searching for candidates")
	defer fmt.Println()

	// Loop backwards through the DAG to find checkpoint candidates.
	candidates := make([]*dagconfig.Checkpoint, 0, cfg.NumCandidates)
	numTested := uint64(0)
	for len(candidates) < cfg.NumCandidates && block.Height() > requiredHeight {
		// Display progress.
		if numTested%progressInterval == 0 {
			fmt.Print(".")
		}

		// Determine if this block is a checkpoint candidate.
		isCandidate, err := dag.IsCheckpointCandidate(block)
		if err != nil {
			return nil, err
		}

		// All checks passed, so this node seems like a reasonable
		// checkpoint candidate.
		if isCandidate {
			checkpoint := dagconfig.Checkpoint{
				Height: block.Height(),
				Hash:   block.Hash(),
			}
			candidates = append(candidates, &checkpoint)
		}

		parentHashes := block.MsgBlock().Header.ParentHashes
		selectedBlockHash := parentHashes[0]
		block, err = dag.BlockByHash(selectedBlockHash)
		if err != nil {
			return nil, err
		}
		numTested++
	}
	return candidates, nil
}

// showCandidate display a checkpoint candidate using and output format
// determined by the configuration parameters.  The Go syntax output
// uses the format the btcchain code expects for checkpoints added to the list.
func showCandidate(candidateNum int, checkpoint *dagconfig.Checkpoint) {
	if cfg.UseGoOutput {
		fmt.Printf("Candidate %d -- {%d, newShaHashFromStr(\"%s\")},\n",
			candidateNum, checkpoint.Height, checkpoint.Hash)
		return
	}

	fmt.Printf("Candidate %d -- Height: %d, Hash: %s\n", candidateNum,
		checkpoint.Height, checkpoint.Hash)

}

func main() {
	// Load configuration and parse command line.
	tcfg, _, err := loadConfig()
	if err != nil {
		return
	}
	cfg = tcfg

	// Load the block database.
	db, err := loadBlockDB()
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to load database:", err)
		return
	}
	defer db.Close()

	// Setup chain.  Ignore notifications since they aren't needed for this
	// util.
	dag, err := blockdag.New(&blockdag.Config{
		DB:         db,
		DAGParams:  activeNetParams,
		TimeSource: blockdag.NewMedianTime(),
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to initialize chain: %s\n", err)
		return
	}

	// Get the latest block hash and height from the database and report
	// status.
	fmt.Printf("Block database loaded with block height %d\n", dag.Height())

	// Find checkpoint candidates.
	highestTipHash := dag.HighestTipHash()
	candidates, err := findCandidates(dag, highestTipHash)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Unable to identify candidates:", err)
		return
	}

	// No candidates.
	if len(candidates) == 0 {
		fmt.Println("No candidates found.")
		return
	}

	// Show the candidates.
	for i, checkpoint := range candidates {
		showCandidate(i+1, checkpoint)
	}
}
