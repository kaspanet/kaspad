package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/domain/consensus"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/infrastructure/config"
	"github.com/kaspanet/kaspad/infrastructure/db/database"
	"github.com/kaspanet/kaspad/infrastructure/network/rpcclient"
	"github.com/kaspanet/kaspad/util/panics"
	"github.com/kaspanet/kaspad/util/profiling"
	"github.com/kaspanet/kaspad/version"
	"github.com/pkg/errors"
)

func main() {
	defer panics.HandlePanic(log, "MAIN", nil)

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
	dbPath := filepath.Join(dataDir, "db")
	consensusConfig := &consensus.Config{Params: *cfg.NetParams()}
	factory := consensus.NewFactory()
	factory.SetTestDataDir(dbPath)
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
		rootHash, err := externalapi.NewDomainHashFromString(root.Root)
		if err != nil {
			return err
		}

		rootHeader, err := tc.BlockHeaderStore().BlockHeader(tc.DatabaseContext(), model.NewStagingArea(), rootHash)
		if database.IsNotFoundError(err) {
			continue
		}

		if err != nil {
			return err
		}

		nextPP := ppHeaders[root.PPIndex-1]

		blockToChild := make(map[externalapi.DomainHash]externalapi.DomainHash)

		// TODO: Since GD data is not always available, we should extract the blue work from the header and use that for topological traversal
		heap := tc.DAGTraversalManager().NewDownHeap(model.NewStagingArea())
		for _, parent := range rootHeader.DirectParents() {
			heap.Push(parent)
			blockToChild[*parent] = *rootHash
		}

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

			chunk = append(chunk, &appmessage.ArchivalBlock{
				Block: appmessage.DomainBlockToRPCBlock(block),
				Child: blockToChild[*hash].String(),
			})

			if len(chunk) == cap(chunk) {
				_, err := rpcClient.AddArchivalBlocks(chunk)
				if err != nil {
					return err
				}

				chunk = chunk[:0]
			}

			for _, parent := range block.Header.DirectParents() {
				heap.Push(parent)
				blockToChild[*parent] = *hash
			}
		}
	}

	return nil
}

func printErrorAndExit(err error) {
	fmt.Fprintf(os.Stderr, "%+v\n", err)
	os.Exit(1)
}
