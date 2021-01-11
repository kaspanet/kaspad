package headersselectedtipmanager_test

import (
	"github.com/kaspanet/kaspad/domain/consensus"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/testutils"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"testing"
)

func TestAddHeaderTip(t *testing.T) {
	testutils.ForAllNets(t, true, func(t *testing.T, params *dagconfig.Params) {
		factory := consensus.NewFactory()
		tc, tearDown, err := factory.NewTestConsensus(params, false, "TestAddHeaderTip")
		if err != nil {
			t.Fatalf("NewTestConsensus: %s", err)
		}
		defer tearDown(false)

		checkExpectedSelectedChain := func(expectedSelectedChain []*externalapi.DomainHash) {
			for i, blockHash := range expectedSelectedChain {
				chainBlockHash, exists, err := tc.HeadersSelectedChainStore().GetHashByIndex(tc.DatabaseContext(), uint64(i))
				if err != nil {
					t.Fatalf("GetHashByIndex: %+v", err)
				}

				if !exists {
					t.Fatalf("index %d is expected to exist", i)
				}

				if !blockHash.Equal(chainBlockHash) {
					t.Fatalf("chain block %d is expected to be %s but got %s", i, blockHash, chainBlockHash)
				}

				index, exists, err := tc.HeadersSelectedChainStore().GetIndexByHash(tc.DatabaseContext(), blockHash)
				if err != nil {
					t.Fatalf("GetIndexByHash: %+v", err)
				}

				if !exists {
					t.Fatalf("hash %s is expected to exist", blockHash)
				}

				if uint64(i) != index {
					t.Fatalf("chain block %s is expected to be %d but got %d", blockHash, i, index)
				}
			}

			_, exists, err := tc.HeadersSelectedChainStore().GetHashByIndex(tc.DatabaseContext(),
				uint64(len(expectedSelectedChain)+1))
			if err != nil {
				t.Fatalf("GetIndexByHash: %+v", err)
			}

			if exists {
				t.Fatalf("index %d is not expected to exist", uint64(len(expectedSelectedChain)+1))
			}
		}

		expectedSelectedChain := []*externalapi.DomainHash{params.GenesisHash}
		tipHash := params.GenesisHash
		for i := 0; i < 10; i++ {
			var err error
			tipHash, _, err = tc.AddBlock([]*externalapi.DomainHash{tipHash}, nil, nil)
			if err != nil {
				t.Fatalf("AddBlock: %+v", err)
			}

			expectedSelectedChain = append(expectedSelectedChain, tipHash)
			checkExpectedSelectedChain(expectedSelectedChain)
		}

		expectedSelectedChain = []*externalapi.DomainHash{params.GenesisHash}
		tipHash = params.GenesisHash
		for i := 0; i < 11; i++ {
			var err error
			tipHash, _, err = tc.AddBlock([]*externalapi.DomainHash{tipHash}, nil, nil)
			if err != nil {
				t.Fatalf("AddBlock: %+v", err)
			}

			expectedSelectedChain = append(expectedSelectedChain, tipHash)
		}
		checkExpectedSelectedChain(expectedSelectedChain)
	})
}
