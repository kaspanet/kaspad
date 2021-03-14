package pastmediantimemanager_test

import (
	"testing"

	"github.com/kaspanet/kaspad/domain/consensus"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
	"github.com/kaspanet/kaspad/domain/consensus/utils/testutils"
	"github.com/kaspanet/kaspad/domain/dagconfig"
)

func TestPastMedianTime(t *testing.T) {
	testutils.ForAllNets(t, true, func(t *testing.T, params *dagconfig.Params) {
		factory := consensus.NewFactory()
		tc, tearDown, err := factory.NewTestConsensus(params, false, "TestUpdateReindexRoot")
		if err != nil {
			t.Fatalf("NewTestConsensus: %s", err)
		}
		defer tearDown(false)

		numBlocks := uint32(300)
		blockHashes := make([]*externalapi.DomainHash, numBlocks)
		blockHashes[0] = params.GenesisHash
		blockTime := params.GenesisBlock.Header.TimeInMilliseconds()
		for i := uint32(1); i < numBlocks; i++ {
			blockTime += 1000
			block, _, err := tc.BuildBlockWithParents([]*externalapi.DomainHash{blockHashes[i-1]}, nil, nil)
			if err != nil {
				t.Fatalf("BuildBlockWithParents: %s", err)
			}

			newHeader := block.Header.ToMutable()
			newHeader.SetTimeInMilliseconds(blockTime)
			block.Header = newHeader.ToImmutable()
			_, err = tc.ValidateAndInsertBlock(block)
			if err != nil {
				t.Fatalf("ValidateAndInsertBlock: %+v", err)
			}

			blockHashes[i] = consensushashing.BlockHash(block)
		}

		tests := []struct {
			blockNumber                      uint32
			expectedMillisecondsSinceGenesis int64
		}{
			{
				blockNumber:                      263,
				expectedMillisecondsSinceGenesis: 132000,
			},
			{
				blockNumber:                      271,
				expectedMillisecondsSinceGenesis: 139000,
			},
			{
				blockNumber:                      241,
				expectedMillisecondsSinceGenesis: 121000,
			},
			{
				blockNumber:                      5,
				expectedMillisecondsSinceGenesis: 3000,
			},
		}

		for _, test := range tests {
			pastMedianTime, err := tc.PastMedianTimeManager().PastMedianTime(blockHashes[test.blockNumber])
			if err != nil {
				t.Fatalf("PastMedianTime: %s", err)
			}

			millisecondsSinceGenesis := pastMedianTime -
				params.GenesisBlock.Header.TimeInMilliseconds()

			if millisecondsSinceGenesis != test.expectedMillisecondsSinceGenesis {
				t.Errorf("TestCalcPastMedianTime: expected past median time of block %v to be %v milliseconds "+
					"from genesis but got %v",
					test.blockNumber, test.expectedMillisecondsSinceGenesis, millisecondsSinceGenesis)
			}
		}
	})

}
