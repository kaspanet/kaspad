package rpchandlers_test

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/rpc/rpccontext"
	"github.com/kaspanet/kaspad/app/rpc/rpchandlers"
	"github.com/kaspanet/kaspad/domain/consensus"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/model/testapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/hashes"
	"github.com/kaspanet/kaspad/domain/consensus/utils/testutils"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/kaspanet/kaspad/domain/miningmanager"
	"github.com/kaspanet/kaspad/infrastructure/config"
	"reflect"
	"sort"
	"testing"
)

type fakeDomain struct {
	testapi.TestConsensus
}

func (d fakeDomain) Consensus() externalapi.Consensus           { return d }
func (d fakeDomain) MiningManager() miningmanager.MiningManager { return nil }

func TestHandleGetBlocks(t *testing.T) {
	testutils.ForAllNets(t, true, func(t *testing.T, params *dagconfig.Params) {
		factory := consensus.NewFactory()
		tc, teardown, err := factory.NewTestConsensus(params, false, "TestHandleGetBlocks")
		if err != nil {
			t.Fatalf("Error setting up consensus: %+v", err)
		}
		defer teardown(false)

		fakeContext := rpccontext.Context{
			Config: &config.Config{Flags: &config.Flags{NetworkFlags: config.NetworkFlags{ActiveNetParams: params}}},
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
				isInPastOfPovBlock, err := tc.DAGTopologyManager().IsAncestorOf(blockHash, povBlock)
				if err != nil {
					t.Fatalf("Failed doing reachability check: '%v'", err)
				}
				if !isInPastOfPovBlock {
					antipast = append(antipast, blockHash)
				}
			}
			return antipast
		}

		upBfsOrder := make([]*externalapi.DomainHash, 0, 30)
		selectedParent := params.GenesisHash
		upBfsOrder = append(upBfsOrder, selectedParent)
		for i := 0; i < 10; i++ {
			parents := make([]*externalapi.DomainHash, 0, 3)
			for j := 0; j < 4; j++ {
				blockHash, _, err := tc.AddBlock([]*externalapi.DomainHash{selectedParent}, nil, nil)
				if err != nil {
					t.Fatalf("Failed adding block: %v", err)
				}
				parents = append(parents, blockHash)
				upBfsOrder = append(upBfsOrder, blockHash)
			}
			selectedParent, _, err = tc.AddBlock(parents, nil, nil)
			if err != nil {
				t.Fatalf("Failed adding block: %v", err)
			}
			upBfsOrder = append(upBfsOrder, selectedParent)
		}

		virtualSelectedParent, err := tc.GetVirtualSelectedParent()
		if err != nil {
			t.Fatalf("Failed getting SelectedParent: %v", err)
		}
		if !virtualSelectedParent.Equal(upBfsOrder[len(upBfsOrder)-1]) {
			t.Fatalf("Expected %s to be selectedParent, instead found: %s", upBfsOrder[len(upBfsOrder)-1], virtualSelectedParent)
		}

		requestSelectedParent := getBlocks(virtualSelectedParent)
		if !reflect.DeepEqual(requestSelectedParent.BlockHashes, hashes.ToStrings([]*externalapi.DomainHash{virtualSelectedParent})) {
			t.Fatalf("TestSyncManager_GetHashesBetween expected %v\n == \n%v", requestSelectedParent.BlockHashes, virtualSelectedParent)
		}

		for i, blockHash := range upBfsOrder {
			expectedBlocks := filterAntiPast(blockHash, upBfsOrder)
			// sort the slice in the order GetBlocks is returning them
			sort.Sort(sort.Reverse(testutils.NewTestGhostDAGSorter(expectedBlocks, tc, t)))
			expectedBlocks = append([]*externalapi.DomainHash{blockHash}, expectedBlocks...)

			blocks := getBlocks(blockHash)
			if !reflect.DeepEqual(blocks.BlockHashes, hashes.ToStrings(expectedBlocks)) {
				t.Fatalf("TestSyncManager_GetHashesBetween %d expected %s\n == \n%s", i, blocks.BlockHashes, hashes.ToStrings(expectedBlocks))
			}
		}

		// Make explitly sure that if lowHash==highHash we get a slice with a single hash.
		blocks := getBlocks(virtualSelectedParent)
		if !reflect.DeepEqual(blocks.BlockHashes, []string{virtualSelectedParent.String()}) {
			t.Fatalf("TestSyncManager_GetHashesBetween expected blocks to contain just '%s', instead got: \n%s", virtualSelectedParent, blocks.BlockHashes)
		}

		sort.Sort(sort.Reverse(testutils.NewTestGhostDAGSorter(upBfsOrder, tc, t)))
		requestAllViaNil := getBlocks(nil)
		if !reflect.DeepEqual(requestAllViaNil.BlockHashes, hashes.ToStrings(upBfsOrder)) {
			t.Fatalf("TestSyncManager_GetHashesBetween expected %v\n == \n%v", requestAllViaNil.BlockHashes, upBfsOrder)
		}

		requestAllExplictly := getBlocks(params.GenesisHash)
		if !reflect.DeepEqual(requestAllExplictly.BlockHashes, hashes.ToStrings(upBfsOrder)) {
			t.Fatalf("TestSyncManager_GetHashesBetween expected %v\n == \n%v", requestAllExplictly.BlockHashes, upBfsOrder)
		}
	})
}
