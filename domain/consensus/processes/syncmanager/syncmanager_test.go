package syncmanager_test

import (
	"github.com/kaspanet/kaspad/domain/consensus"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/testutils"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"math"
	"reflect"
	"sort"
	"testing"
)

func TestSyncManager_GetHashesBetween(t *testing.T) {
	testutils.ForAllNets(t, true, func(t *testing.T, params *dagconfig.Params) {
		factory := consensus.NewFactory()
		tc, teardown, err := factory.NewTestConsensus(params, false, "TestSyncManager_GetHashesBetween")
		if err != nil {
			t.Fatalf("Error setting up consensus: %+v", err)
		}
		defer teardown(false)

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

		for i, blockHash := range upBfsOrder {
			empty, err := tc.SyncManager().GetHashesBetween(blockHash, blockHash, math.MaxUint64)
			if err != nil {
				t.Fatalf("TestSyncManager_GetHashesBetween failed returning 0 hashes on the %d'th block: %v", i, err)
			}
			if len(empty) != 0 {
				t.Fatalf("Expected lowHash=highHash to return empty on the %d'th block, instead found: %v", i, empty)
			}
		}

		allHashes, err := tc.SyncManager().GetHashesBetween(upBfsOrder[0], upBfsOrder[len(upBfsOrder)-1], math.MaxUint64)
		if err != nil {
			t.Fatalf("TestSyncManager_GetHashesBetween failed returning allHashes: %v", err)
		}

		sort.Sort(sort.Reverse(testutils.NewTestGhostDAGSorter(upBfsOrder, tc, t)))
		upBfsOrderExcludingGenesis := upBfsOrder[1:]
		if !reflect.DeepEqual(allHashes, upBfsOrderExcludingGenesis) {
			t.Fatalf("TestSyncManager_GetHashesBetween expected %v\n == \n%v", allHashes, upBfsOrder)
		}
	})
}
