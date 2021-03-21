package headersselectedtipmanager_test

import (
	"testing"

	"github.com/kaspanet/kaspad/domain/consensus"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/testutils"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/kaspanet/kaspad/infrastructure/db/database"
	"github.com/pkg/errors"
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
				chainBlockHash, err := tc.HeadersSelectedChainStore().GetHashByIndex(tc.DatabaseContext(), nil, uint64(i))
				if err != nil {
					t.Fatalf("GetHashByIndex: %+v", err)
				}

				if !blockHash.Equal(chainBlockHash) {
					t.Fatalf("chain block %d is expected to be %s but got %s", i, blockHash, chainBlockHash)
				}

				index, err := tc.HeadersSelectedChainStore().GetIndexByHash(tc.DatabaseContext(), nil, blockHash)
				if err != nil {
					t.Fatalf("GetIndexByHash: %+v", err)
				}

				if uint64(i) != index {
					t.Fatalf("chain block %s is expected to be %d but got %d", blockHash, i, index)
				}
			}

			_, err := tc.HeadersSelectedChainStore().GetHashByIndex(tc.DatabaseContext(), nil, uint64(len(expectedSelectedChain)+1))
			if !errors.Is(err, database.ErrNotFound) {
				t.Fatalf("index %d is not expected to exist, but instead got error: %+v",
					uint64(len(expectedSelectedChain)+1), err)
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
