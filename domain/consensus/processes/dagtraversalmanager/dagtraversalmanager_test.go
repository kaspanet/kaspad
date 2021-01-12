package dagtraversalmanager_test

import (
	"github.com/kaspanet/kaspad/domain/consensus"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/testutils"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"testing"
)

func TestLowestChainBlockAboveOrEqualToBlueScore(t *testing.T) {
	testutils.ForAllNets(t, true, func(t *testing.T, params *dagconfig.Params) {
		params.FinalityDuration = 10 * params.TargetTimePerBlock
		factory := consensus.NewFactory()
		tc, tearDown, err := factory.NewTestConsensus(params, false,
			"TestLowestChainBlockAboveOrEqualToBlueScore")
		if err != nil {
			t.Fatalf("NewTestConsensus: %s", err)
		}
		defer tearDown(false)

		checkExpectedBlock := func(highHash *externalapi.DomainHash, blueScore uint64, expected *externalapi.DomainHash) {
			blockHash, err := tc.DAGTraversalManager().LowestChainBlockAboveOrEqualToBlueScore(highHash, blueScore)
			if err != nil {
				t.Fatalf("LowestChainBlockAboveOrEqualToBlueScore: %+v", err)
			}

			if !blockHash.Equal(expected) {
				t.Fatalf("Expected block %s but got %s", expected, blockHash)
			}
		}

		checkBlueScore := func(blockHash *externalapi.DomainHash, expectedBlueScoe uint64) {
			ghostdagData, err := tc.GHOSTDAGDataStore().Get(tc.DatabaseContext(), blockHash)
			if err != nil {
				t.Fatalf("GHOSTDAGDataStore().Get: %+v", err)
			}

			if ghostdagData.BlueScore() != expectedBlueScoe {
				t.Fatalf("Expected blue score %d but got %d", expectedBlueScoe, ghostdagData.BlueScore())
			}
		}

		chain := []*externalapi.DomainHash{params.GenesisHash}
		tipHash := params.GenesisHash
		for i := 0; i < 9; i++ {
			var err error
			tipHash, _, err = tc.AddBlock([]*externalapi.DomainHash{tipHash}, nil, nil)
			if err != nil {
				t.Fatalf("AddBlock: %+v", err)
			}

			chain = append(chain, tipHash)
		}

		sideChain1TipHash, _, err := tc.AddBlock([]*externalapi.DomainHash{params.GenesisHash}, nil, nil)
		if err != nil {
			t.Fatalf("AddBlock: %+v", err)
		}

		tipHash, _, err = tc.AddBlock([]*externalapi.DomainHash{sideChain1TipHash, tipHash}, nil, nil)
		if err != nil {
			t.Fatalf("AddBlock: %+v", err)
		}

		chain = append(chain, tipHash)
		blueScore11BlockHash := tipHash
		checkBlueScore(blueScore11BlockHash, 11)

		for i := 0; i < 5; i++ {
			var err error
			tipHash, _, err = tc.AddBlock([]*externalapi.DomainHash{tipHash}, nil, nil)
			if err != nil {
				t.Fatalf("AddBlock: %+v", err)
			}

			chain = append(chain, tipHash)
		}

		sideChain2TipHash, _, err := tc.AddBlock([]*externalapi.DomainHash{params.GenesisHash}, nil, nil)
		if err != nil {
			t.Fatalf("AddBlock: %+v", err)
		}

		tipHash, _, err = tc.AddBlock([]*externalapi.DomainHash{sideChain2TipHash, tipHash}, nil, nil)
		if err != nil {
			t.Fatalf("AddBlock: %+v", err)
		}
		chain = append(chain, tipHash)

		blueScore18BlockHash := tipHash
		checkBlueScore(blueScore18BlockHash, 18)

		for i := 0; i < 3; i++ {
			var err error
			tipHash, _, err = tc.AddBlock([]*externalapi.DomainHash{tipHash}, nil, nil)
			if err != nil {
				t.Fatalf("AddBlock: %+v", err)
			}

			chain = append(chain, tipHash)
		}

		// Check by exact blue score
		checkExpectedBlock(tipHash, 0, params.GenesisHash)
		checkExpectedBlock(tipHash, 5, chain[5])
		checkExpectedBlock(tipHash, 19, chain[len(chain)-3])

		// Check by non exact blue score
		checkExpectedBlock(tipHash, 17, blueScore18BlockHash)
		checkExpectedBlock(tipHash, 10, blueScore11BlockHash)
	})
}
