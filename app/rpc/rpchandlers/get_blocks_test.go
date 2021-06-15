package rpchandlers_test

import (
	"reflect"
	"sort"
	"testing"

	"github.com/kaspanet/kaspad/domain/consensus/model"

	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/rpc/rpccontext"
	"github.com/kaspanet/kaspad/app/rpc/rpchandlers"
	"github.com/kaspanet/kaspad/domain/consensus"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/model/testapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/hashes"
	"github.com/kaspanet/kaspad/domain/consensus/utils/testutils"
	"github.com/kaspanet/kaspad/domain/miningmanager"
	"github.com/kaspanet/kaspad/infrastructure/config"
)

type fakeDomain struct {
	testapi.TestConsensus
}

func (d fakeDomain) DeleteStagingConsensus() error {
	panic("implement me")
}

func (d fakeDomain) StagingConsensus() externalapi.Consensus {
	panic("implement me")
}

func (d fakeDomain) InitStagingConsensus() error {
	panic("implement me")
}

func (d fakeDomain) CommitStagingConsensus() error {
	panic("implement me")
}

func (d fakeDomain) Consensus() externalapi.Consensus           { return d }
func (d fakeDomain) MiningManager() miningmanager.MiningManager { return nil }

func TestHandleGetBlocks(t *testing.T) {
	testutils.ForAllNets(t, true, func(t *testing.T, consensusConfig *consensus.Config) {
		stagingArea := model.NewStagingArea()

		factory := consensus.NewFactory()
		tc, teardown, err := factory.NewTestConsensus(consensusConfig, "TestHandleGetBlocks")
		if err != nil {
			t.Fatalf("Error setting up consensus: %+v", err)
		}
		defer teardown(false)

		fakeContext := rpccontext.Context{
			Config: &config.Config{Flags: &config.Flags{NetworkFlags: config.NetworkFlags{ActiveNetParams: &consensusConfig.Params}}},
			Domain: fakeDomain{tc},
		}

		getBlocks := func(lowHash *externalapi.DomainHash) *appmessage.GetBlocksResponseMessage {
			request := appmessage.GetBlocksRequestMessage{}
			if lowHash != nil {
				request.LowHash = lowHash.String()
			}
			response, err := rpchandlers.HandleGetBlocks(&fakeContext, nil, &request)
			if err != nil {
				t.Fatalf("Expected empty request to not fail, instead: '%v'", err)
			}
			return response.(*appmessage.GetBlocksResponseMessage)
		}

		filterAntiPast := func(povBlock *externalapi.DomainHash, slice []*externalapi.DomainHash) []*externalapi.DomainHash {
			antipast := make([]*externalapi.DomainHash, 0, len(slice))

			for _, blockHash := range slice {
				isInPastOfPovBlock, err := tc.DAGTopologyManager().IsAncestorOf(stagingArea, blockHash, povBlock)
				if err != nil {
					t.Fatalf("Failed doing reachability check: '%v'", err)
				}
				if !isInPastOfPovBlock {
					antipast = append(antipast, blockHash)
				}
			}
			return antipast
		}

		// Create a DAG with the following structure:
		//          merging block
		//         /      |      \
		//      split1  split2   split3
		//        \       |      /
		//         merging block
		//         /      |      \
		//      split1  split2   split3
		//        \       |      /
		//               etc.
		expectedOrder := make([]*externalapi.DomainHash, 0, 40)
		mergingBlock := consensusConfig.GenesisHash
		for i := 0; i < 10; i++ {
			splitBlocks := make([]*externalapi.DomainHash, 0, 3)
			for j := 0; j < 3; j++ {
				blockHash, _, err := tc.AddBlock([]*externalapi.DomainHash{mergingBlock}, nil, nil)
				if err != nil {
					t.Fatalf("Failed adding block: %v", err)
				}
				splitBlocks = append(splitBlocks, blockHash)
			}
			sort.Sort(sort.Reverse(testutils.NewTestGhostDAGSorter(stagingArea, splitBlocks, tc, t)))
			restOfSplitBlocks, selectedParent := splitBlocks[:len(splitBlocks)-1], splitBlocks[len(splitBlocks)-1]
			expectedOrder = append(expectedOrder, selectedParent)
			expectedOrder = append(expectedOrder, restOfSplitBlocks...)

			mergingBlock, _, err = tc.AddBlock(splitBlocks, nil, nil)
			if err != nil {
				t.Fatalf("Failed adding block: %v", err)
			}
			expectedOrder = append(expectedOrder, mergingBlock)
		}

		virtualSelectedParent, err := tc.GetVirtualSelectedParent()
		if err != nil {
			t.Fatalf("Failed getting SelectedParent: %v", err)
		}
		if !virtualSelectedParent.Equal(expectedOrder[len(expectedOrder)-1]) {
			t.Fatalf("Expected %s to be selectedParent, instead found: %s", expectedOrder[len(expectedOrder)-1], virtualSelectedParent)
		}

		requestSelectedParent := getBlocks(virtualSelectedParent)
		if !reflect.DeepEqual(requestSelectedParent.BlockHashes, hashes.ToStrings([]*externalapi.DomainHash{virtualSelectedParent})) {
			t.Fatalf("TestHandleGetBlocks expected:\n%v\nactual:\n%v", virtualSelectedParent, requestSelectedParent.BlockHashes)
		}

		for i, blockHash := range expectedOrder {
			expectedBlocks := filterAntiPast(blockHash, expectedOrder)
			expectedBlocks = append([]*externalapi.DomainHash{blockHash}, expectedBlocks...)

			actualBlocks := getBlocks(blockHash)
			if !reflect.DeepEqual(actualBlocks.BlockHashes, hashes.ToStrings(expectedBlocks)) {
				t.Fatalf("TestHandleGetBlocks %d \nexpected: \n%v\nactual:\n%v", i,
					hashes.ToStrings(expectedBlocks), actualBlocks.BlockHashes)
			}
		}

		// Make explicitly sure that if lowHash==highHash we get a slice with a single hash.
		actualBlocks := getBlocks(virtualSelectedParent)
		if !reflect.DeepEqual(actualBlocks.BlockHashes, []string{virtualSelectedParent.String()}) {
			t.Fatalf("TestHandleGetBlocks expected blocks to contain just '%s', instead got: \n%v",
				virtualSelectedParent, actualBlocks.BlockHashes)
		}

		expectedOrder = append([]*externalapi.DomainHash{consensusConfig.GenesisHash}, expectedOrder...)
		actualOrder := getBlocks(nil)
		if !reflect.DeepEqual(actualOrder.BlockHashes, hashes.ToStrings(expectedOrder)) {
			t.Fatalf("TestHandleGetBlocks \nexpected: %v \nactual:\n%v", expectedOrder, actualOrder.BlockHashes)
		}

		requestAllExplictly := getBlocks(consensusConfig.GenesisHash)
		if !reflect.DeepEqual(requestAllExplictly.BlockHashes, hashes.ToStrings(expectedOrder)) {
			t.Fatalf("TestHandleGetBlocks \nexpected: \n%v\n. actual:\n%v", expectedOrder, requestAllExplictly.BlockHashes)
		}
	})
}
