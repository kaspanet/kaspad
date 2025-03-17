package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/domain/consensus"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
	"github.com/kaspanet/kaspad/infrastructure/config"
	"github.com/kaspanet/kaspad/infrastructure/db/database"
	"github.com/kaspanet/kaspad/infrastructure/network/rpcclient"
	"github.com/kaspanet/kaspad/util/profiling"
	"github.com/kaspanet/kaspad/version"
	"github.com/pkg/errors"
)

func main() {
	// defer panics.HandlePanic(log, "MAIN", nil)

	cfg, err := parseConfig()
	if err != nil {
		printErrorAndExit(errors.Errorf("Error parsing command-line arguments: %s", err))
	}
	defer backendLog.Close()

	// Show version at startup.
	log.Infof("Version %s", version.Version())

	// Enable http profiling server if requested.
	if cfg.Profile != "" {
		profiling.Start(cfg.Profile, log)
	}

	err = mainImpl(cfg)
	if err != nil {
		printErrorAndExit(err)
	}
}

func mainImpl(cfg *configFlags) error {
	dataDir := filepath.Join(config.DefaultAppDir)
	dbPath := filepath.Join(dataDir, "kaspa-mainnet/datadir2")
	consensusConfig := &consensus.Config{Params: *cfg.NetParams()}
	factory := consensus.NewFactory()
	factory.SetTestDataDir(dbPath)
	factory.AutoSetActivePrefix(true)
	tc, tearDownFunc, err := factory.NewTestConsensus(consensusConfig, "archiveexport")
	if err != nil {
		return err
	}
	defer tearDownFunc(true)

	rpcAddress, err := cfg.NetParams().NormalizeRPCServerAddress(cfg.RPCServer)
	if err != nil {
		return err
	}
	rpcClient, err := rpcclient.NewRPCClient(rpcAddress)
	if err != nil {
		return err
	}

	rootsResp, err := rpcClient.GetPruningWindowRoots()
	if err != nil {
		return err
	}

	ppHeaders, err := tc.PruningPointHeaders()
	if err != nil {
		return err
	}

	for _, root := range rootsResp.Roots {
		log.Infof("Got root %s", root.Root)
	}

	counterStart := time.Now()
	counter := 0
	for _, root := range rootsResp.Roots {
		rootHash, err := externalapi.NewDomainHashFromString(root.Root)
		if err != nil {
			return err
		}

		log.Infof("Adding past of %s", rootHash)

		if err != nil {
			return err
		}

		nextPP := ppHeaders[root.PPIndex-1]

		blockToChild := make(map[externalapi.DomainHash]externalapi.DomainHash)

		// TODO: Since GD data is not always available, we should extract the blue work from the header and use that for topological traversal
		heap := tc.DAGTraversalManager().NewDownHeap(model.NewStagingArea())
		heap.Push(rootHash)

		visited := make(map[externalapi.DomainHash]struct{})
		chunk := make([]*appmessage.ArchivalBlock, 0, 1000)
		for heap.Len() > 0 {
			hash := heap.Pop()

			if _, ok := visited[*hash]; ok {
				continue
			}
			visited[*hash] = struct{}{}

			// TODO: Use header data instead of GD data
			blockGHOSTDAGData, err := tc.GHOSTDAGDataStore().Get(tc.DatabaseContext(), model.NewStagingArea(), hash, false)
			if err != nil {
				return err
			}

			if blockGHOSTDAGData.BlueWork().Cmp(nextPP.BlueWork()) <= 0 {
				break
			}

			block, err := tc.BlockStore().Block(tc.DatabaseContext(), model.NewStagingArea(), hash)
			if database.IsNotFoundError(err) {
				continue
			}

			if err != nil {
				return err
			}

			archivalBlock := &appmessage.ArchivalBlock{
				Block: appmessage.DomainBlockToRPCBlock(block),
			}
			if child, ok := blockToChild[*hash]; ok {
				archivalBlock.Child = child.String()
			}

			chunk = append(chunk, archivalBlock)

			if len(chunk) == 1 {
				log.Infof("Added %s to chunk", consensushashing.BlockHash(block))
			}

			if len(chunk) == cap(chunk) {
				err := sendChunk(rpcClient, chunk)
				if err != nil {
					return err
				}
				counter += len(chunk)
				counterDuration := time.Since(counterStart)
				if counterDuration > 10*time.Second {
					rate := float64(counter) / counterDuration.Seconds()
					log.Infof("Sent %d blocks in the last %.2f seconds (%.2f blocks/second)", counter, counterDuration.Seconds(), rate)
					counterStart = time.Now()
					counter = 0
				}

				chunk = chunk[:0]
			}

			for _, parent := range block.Header.DirectParents() {
				heap.Push(parent)
				blockToChild[*parent] = *hash
			}
		}

		if len(chunk) > 0 {
			sendChunk(rpcClient, chunk)
		}
	}

	return nil
}

func sendChunk(rpcClient *rpcclient.RPCClient, chunk []*appmessage.ArchivalBlock) error {
	log.Infof("Sending chunk")
	_, err := rpcClient.AddArchivalBlocks(chunk)
	if err != nil {
		return err
	}
	log.Infof("Sent chunk")

	// Checking existence of first block for sanity
	block := chunk[0]
	domainBlock, err := appmessage.RPCBlockToDomainBlock(block.Block)
	if err != nil {
		return err
	}

	blockHash := consensushashing.BlockHash(domainBlock)
	log.Infof("Checking block %s", blockHash)
	resp, err := rpcClient.GetBlock(blockHash.String(), true)
	if err != nil {
		return err
	}

	if len(resp.Block.Transactions) == 0 {
		return errors.Errorf("Block %s has no transactions on the server", blockHash)
	}

	return nil
}

func printErrorAndExit(err error) {
	fmt.Fprintf(os.Stderr, "%+v\n", err)
	os.Exit(1)
}
