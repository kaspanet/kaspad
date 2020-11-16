package pastmediantimemanager_test

import (
	"github.com/kaspanet/kaspad/domain/consensus"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/testutils"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"testing"
)

func TestPastMedianTime(t *testing.T) {
	testutils.ForAllNets(t, true, func(t *testing.T, params *dagconfig.Params) {
		factory := consensus.NewFactory()
		tc, tearDown, err := factory.NewTestConsensus(params, "TestUpdateReindexRoot")
		if err != nil {
			t.Fatalf("NewTestConsensus: %s", err)
		}
		defer tearDown()

		numBlocks := uint32(300)
		blockHashes := make([]*externalapi.DomainHash, numBlocks)
		blockHashes[0] = params.GenesisHash
		blockTime := params.GenesisBlock.Header.TimeInMilliseconds
		for i := uint32(1); i < numBlocks; i++ {
			blockTime += 1000
			block, err := tc.BuildBlockWithParents([]*externalapi.DomainHash{blockHashes[i-1]}, nil, nil)
			if err != nil {
				t.Fatalf("BuildBlockWithParents: %s", err)
			}

			block.Header.TimeInMilliseconds = blockTime
			blockHash, err := tc.SolveAndAddBlock(block)
			if err != nil {
				t.Fatalf("SolveAndAddBlock: %s", err)
			}

			blockHashes[i] = blockHash
		}

		tests := []struct {
			blockNumber                      uint32
			expectedMillisecondsSinceGenesis int64
		}{
			{
				blockNumber:                      263,
				expectedMillisecondsSinceGenesis: 130000,
			},
			{
				blockNumber:                      271,
				expectedMillisecondsSinceGenesis: 138000,
			},
			{
				blockNumber:                      241,
				expectedMillisecondsSinceGenesis: 108000,
			},
			{
				blockNumber:                      5,
				expectedMillisecondsSinceGenesis: 0,
			},
		}

		for _, test := range tests {
			pastMedianTime, err := tc.PastMedianTimeManager().PastMedianTime(blockHashes[test.blockNumber])
			if err != nil {
				t.Fatalf("PastMedianTime: %s", err)
			}

			millisecondsSinceGenesis := pastMedianTime -
				params.GenesisBlock.Header.TimeInMilliseconds

			if millisecondsSinceGenesis != test.expectedMillisecondsSinceGenesis {
				t.Errorf("TestCalcPastMedianTime: expected past median time of block %v to be %v milliseconds "+
					"from genesis but got %v",
					test.blockNumber, test.expectedMillisecondsSinceGenesis, millisecondsSinceGenesis)
			}
		}
	})

}
