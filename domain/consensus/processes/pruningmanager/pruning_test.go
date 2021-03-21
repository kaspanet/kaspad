package pruningmanager_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/kaspanet/kaspad/domain/consensus"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/testutils"
	"github.com/kaspanet/kaspad/domain/dagconfig"
)

type jsonBlock struct {
	ID      string   `json:"ID"`
	Parents []string `json:"Parents"`
}

type testJSON struct {
	MergeSetSizeLimit uint64       `json:"mergeSetSizeLimit"`
	FinalityDepth     uint64       `json:"finalityDepth"`
	Blocks            []*jsonBlock `json:"blocks"`
}

func TestPruning(t *testing.T) {
	expectedPruningPointByNet := map[string]map[string]string{
		"chain-for-test-pruning.json": {
			dagconfig.MainnetParams.Name: "1582",
			dagconfig.TestnetParams.Name: "1582",
			dagconfig.DevnetParams.Name:  "1582",
			dagconfig.SimnetParams.Name:  "1582",
		},
		"dag-for-test-pruning.json": {
			dagconfig.MainnetParams.Name: "502",
			dagconfig.TestnetParams.Name: "503",
			dagconfig.DevnetParams.Name:  "503",
			dagconfig.SimnetParams.Name:  "503",
		},
	}

	testutils.ForAllNets(t, true, func(t *testing.T, params *dagconfig.Params) {
		err := filepath.Walk("./testdata", func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}

			jsonFile, err := os.Open(path)
			if err != nil {
				t.Fatalf("TestPruning : failed opening json file %s: %s", path, err)
			}
			defer jsonFile.Close()

			test := &testJSON{}
			decoder := json.NewDecoder(jsonFile)
			decoder.DisallowUnknownFields()
			err = decoder.Decode(&test)
			if err != nil {
				t.Fatalf("TestPruning: failed decoding json: %v", err)
			}

			params.FinalityDuration = time.Duration(test.FinalityDepth) * params.TargetTimePerBlock
			params.MergeSetSizeLimit = test.MergeSetSizeLimit

			factory := consensus.NewFactory()
			tc, teardown, err := factory.NewTestConsensus(params, false, "TestPruning")
			if err != nil {
				t.Fatalf("Error setting up consensus: %+v", err)
			}
			defer teardown(false)

			blockIDToHash := map[string]*externalapi.DomainHash{
				"0": params.GenesisHash,
			}

			blockHashToID := map[externalapi.DomainHash]string{
				*params.GenesisHash: "0",
			}

			for _, dagBlock := range test.Blocks {
				if dagBlock.ID == "0" {
					continue
				}
				parentHashes := make([]*externalapi.DomainHash, 0, len(dagBlock.Parents))
				for _, parentID := range dagBlock.Parents {
					parentHash, ok := blockIDToHash[parentID]
					if !ok {
						t.Fatalf("No hash was found for block with ID %s", parentID)
					}
					parentHashes = append(parentHashes, parentHash)
				}

				blockHash, _, err := tc.AddBlock(parentHashes, nil, nil)
				if err != nil {
					t.Fatalf("AddBlock: %+v", err)
				}

				blockIDToHash[dagBlock.ID] = blockHash
				blockHashToID[*blockHash] = dagBlock.ID

				pruningPoint, err := tc.PruningPoint()
				if err != nil {
					return err
				}

				pruningPointCandidate, err := tc.PruningStore().PruningPointCandidate(tc.DatabaseContext())
				if err != nil {
					return err
				}

				isValidPruningPoint, err := tc.IsValidPruningPoint(pruningPointCandidate)
				if err != nil {
					return err
				}

				shouldBeValid := *pruningPoint == *pruningPointCandidate
				if isValidPruningPoint != shouldBeValid {
					t.Fatalf("isValidPruningPoint is %t while expected %t", isValidPruningPoint, shouldBeValid)
				}
			}

			pruningPoint, err := tc.PruningPoint()
			if err != nil {
				t.Fatalf("PruningPoint: %+v", err)
			}

			pruningPointID := blockHashToID[*pruningPoint]
			expectedPruningPoint := expectedPruningPointByNet[info.Name()][params.Name]
			if expectedPruningPoint != pruningPointID {
				t.Fatalf("%s: Expected pruning point to be %s but got %s", info.Name(), expectedPruningPoint, pruningPointID)
			}

			for _, jsonBlock := range test.Blocks {
				id := jsonBlock.ID
				blockHash := blockIDToHash[id]

				isPruningPointAncestorOfBlock, err := tc.DAGTopologyManager().IsAncestorOf(pruningPoint, blockHash)
				if err != nil {
					t.Fatalf("IsAncestorOf: %+v", err)
				}

				expectsBlock := true
				if !isPruningPointAncestorOfBlock {
					isBlockAncestorOfPruningPoint, err := tc.DAGTopologyManager().IsAncestorOf(blockHash, pruningPoint)
					if err != nil {
						t.Fatalf("IsAncestorOf: %+v", err)
					}

					if isBlockAncestorOfPruningPoint {
						expectsBlock = false
					} else {
						virtualInfo, err := tc.GetVirtualInfo()
						if err != nil {
							t.Fatalf("GetVirtualInfo: %+v", err)
						}

						isInPastOfVirtual := false
						for _, virtualParent := range virtualInfo.ParentHashes {
							isAncestorOfVirtualParent, err := tc.DAGTopologyManager().IsAncestorOf(blockHash, virtualParent)
							if err != nil {
								t.Fatalf("IsAncestorOf: %+v", err)
							}

							if isAncestorOfVirtualParent {
								isInPastOfVirtual = true
								break
							}
						}

						if !isInPastOfVirtual {
							expectsBlock = false
						}
					}

				}

				hasBlock, err := tc.BlockStore().HasBlock(tc.DatabaseContext(),, blockHash)
				if err != nil {
					t.Fatalf("HasBlock: %+v", err)
				}

				if expectsBlock != hasBlock {
					t.Fatalf("expected hasBlock to be %t for block %s but got %t", expectsBlock, id, hasBlock)
				}
			}

			return nil
		})
		if err != nil {
			t.Fatalf("Walk: %+v", err)
		}
	})
}
