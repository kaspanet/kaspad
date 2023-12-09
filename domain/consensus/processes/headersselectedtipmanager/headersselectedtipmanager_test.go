package headersselectedtipmanager_test

import (
	"testing"

	"github.com/zoomy-network/zoomyd/domain/consensus/model"

	"github.com/pkg/errors"
	"github.com/zoomy-network/zoomyd/domain/consensus"
	"github.com/zoomy-network/zoomyd/domain/consensus/model/externalapi"
	"github.com/zoomy-network/zoomyd/domain/consensus/utils/testutils"
	"github.com/zoomy-network/zoomyd/infrastructure/db/database"
)

func TestAddHeaderTip(t *testing.T) {
	testutils.ForAllNets(t, true, func(t *testing.T, consensusConfig *consensus.Config) {
		factory := consensus.NewFactory()
		tc, tearDown, err := factory.NewTestConsensus(consensusConfig, "TestAddHeaderTip")
		if err != nil {
			t.Fatalf("NewTestConsensus: %s", err)
		}
		defer tearDown(false)

		stagingArea := model.NewStagingArea()
		checkExpectedSelectedChain := func(expectedSelectedChain []*externalapi.DomainHash) {
			for i, blockHash := range expectedSelectedChain {
				chainBlockHash, err := tc.HeadersSelectedChainStore().GetHashByIndex(tc.DatabaseContext(), stagingArea, uint64(i))
				if err != nil {
					t.Fatalf("GetHashByIndex: %+v", err)
				}

				if !blockHash.Equal(chainBlockHash) {
					t.Fatalf("chain block %d is expected to be %s but got %s", i, blockHash, chainBlockHash)
				}

				index, err := tc.HeadersSelectedChainStore().GetIndexByHash(tc.DatabaseContext(), stagingArea, blockHash)
				if err != nil {
					t.Fatalf("GetIndexByHash: %+v", err)
				}

				if uint64(i) != index {
					t.Fatalf("chain block %s is expected to be %d but got %d", blockHash, i, index)
				}
			}

			_, err := tc.HeadersSelectedChainStore().GetHashByIndex(tc.DatabaseContext(), stagingArea, uint64(len(expectedSelectedChain)+1))
			if !errors.Is(err, database.ErrNotFound) {
				t.Fatalf("index %d is not expected to exist, but instead got error: %+v",
					uint64(len(expectedSelectedChain)+1), err)
			}
		}

		expectedSelectedChain := []*externalapi.DomainHash{consensusConfig.GenesisHash}
		tipHash := consensusConfig.GenesisHash
		for i := 0; i < 10; i++ {
			var err error
			tipHash, _, err = tc.AddBlock([]*externalapi.DomainHash{tipHash}, nil, nil)
			if err != nil {
				t.Fatalf("AddBlock: %+v", err)
			}

			expectedSelectedChain = append(expectedSelectedChain, tipHash)
			checkExpectedSelectedChain(expectedSelectedChain)
		}

		expectedSelectedChain = []*externalapi.DomainHash{consensusConfig.GenesisHash}
		tipHash = consensusConfig.GenesisHash
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
