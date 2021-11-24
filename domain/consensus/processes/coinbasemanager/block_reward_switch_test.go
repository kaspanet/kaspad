package coinbasemanager_test

import (
	"github.com/kaspanet/kaspad/domain/consensus"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
	"github.com/kaspanet/kaspad/domain/consensus/utils/constants"
	"github.com/kaspanet/kaspad/domain/consensus/utils/testutils"
	"github.com/kaspanet/kaspad/domain/consensus/utils/transactionhelper"
	"github.com/kaspanet/kaspad/util/difficulty"
	"testing"
	"time"
)

func TestBlockRewardSwitch(t *testing.T) {
	testutils.ForAllNets(t, true, func(t *testing.T, consensusConfig *consensus.Config) {
		// Set the pruning depth to 10
		consensusConfig.MergeSetSizeLimit = 1
		consensusConfig.K = 1
		consensusConfig.FinalityDuration = 1 * time.Second
		consensusConfig.TargetTimePerBlock = 1 * time.Second

		// Disable difficulty adjustment so that we could reason about blue work
		consensusConfig.DisableDifficultyAdjustment = true

		// Disable pruning so that we could have access to all the blocks
		consensusConfig.IsArchival = true

		// Set the interval to 10
		consensusConfig.FixedSubsidySwitchPruningPointInterval = 10

		// Set the hash rate difference such that the switch would trigger exactly
		// on the `FixedSubsidySwitchPruningPointInterval + 1`th pruning point
		workToAcceptGenesis := difficulty.CalcWork(consensusConfig.GenesisBlock.Header.Bits())
		consensusConfig.FixedSubsidySwitchHashRateThreshold = workToAcceptGenesis

		// Set the min, max, and post-switch subsidies to values that would make it
		// easy to tell whether the switch happened
		consensusConfig.MinSubsidy = 2 * constants.SompiPerKaspa
		consensusConfig.MaxSubsidy = 2 * constants.SompiPerKaspa
		consensusConfig.SubsidyGenesisReward = 1 * constants.SompiPerKaspa

		factory := consensus.NewFactory()
		tc, teardown, err := factory.NewTestConsensus(consensusConfig, "TestBlockRewardSwitch")
		if err != nil {
			t.Fatalf("Error setting up consensus: %+v", err)
		}
		defer teardown(false)

		// Make the pruning point move FixedSubsidySwitchPruningPointInterval times
		tipHash := consensusConfig.GenesisHash
		for i := uint64(0); i < consensusConfig.PruningDepth()+consensusConfig.FixedSubsidySwitchPruningPointInterval; i++ {
			addedBlockHash, _, err := tc.AddBlock([]*externalapi.DomainHash{tipHash}, nil, nil)
			if err != nil {
				t.Fatalf("AddBlock: %+v", err)
			}
			tipHash = addedBlockHash
		}

		// We expect to see `FixedSubsidySwitchPruningPointInterval` pruning points + the genesis
		pruningPointHeaders, err := tc.PruningPointHeaders()
		if err != nil {
			t.Fatalf("PruningPointHeaders: %+v", pruningPointHeaders)
		}
		expectedPruningPointHeaderAmount := consensusConfig.FixedSubsidySwitchPruningPointInterval + 1
		if uint64(len(pruningPointHeaders)) != expectedPruningPointHeaderAmount {
			t.Fatalf("Unexpected amount of pruning point headers. "+
				"Want: %d, got: %d", expectedPruningPointHeaderAmount, len(pruningPointHeaders))
		}

		// Make sure that all the headers thus far had a non-fixed subsidies
		// Note that we skip the genesis, since that always has the post-switch
		// value
		for _, pruningPointHeader := range pruningPointHeaders[1:] {
			pruningPointHash := consensushashing.HeaderHash(pruningPointHeader)
			pruningPoint, err := tc.GetBlock(pruningPointHash)
			if err != nil {
				t.Fatalf("GetBlock: %+v", err)
			}
			pruningPointCoinbase := pruningPoint.Transactions[transactionhelper.CoinbaseTransactionIndex]
			_, _, subsidy, err := tc.CoinbaseManager().ExtractCoinbaseDataBlueScoreAndSubsidy(pruningPointCoinbase)
			if err != nil {
				t.Fatalf("ExtractCoinbaseDataBlueScoreAndSubsidy: %+v", err)
			}
			if subsidy != consensusConfig.MinSubsidy {
				t.Fatalf("Subsidy has unexpected value. Want: %d, got: %d", consensusConfig.MinSubsidy, subsidy)
			}
		}

		// Add another block. We expect it to be another pruning point
		//lastPruningPointHash, _, err := tc.AddBlock([]*externalapi.DomainHash{tipHash}, nil, nil)
		//if err != nil {
		//	t.Fatalf("AddBlock: %+v", err)
		//}

		// Make sure that another pruning point had been added
		//pruningPointHeaders, err = tc.PruningPointHeaders()
		//if err != nil {
		//	t.Fatalf("PruningPointHeaders: %+v", pruningPointHeaders)
		//}
		//expectedPruningPointHeaderAmount = expectedPruningPointHeaderAmount + 1
		//if uint64(len(pruningPointHeaders)) != expectedPruningPointHeaderAmount {
		//	t.Fatalf("Unexpected amount of pruning point headers. "+
		//		"Want: %d, got: %d", expectedPruningPointHeaderAmount, len(pruningPointHeaders))
		//}

		//// Make sure that the last pruning point has a post-switch subsidy
		//lastPruningPoint, err := tc.GetBlock(lastPruningPointHash)
		//if err != nil {
		//	t.Fatalf("GetBlock: %+v", err)
		//}
		//lastPruningPointCoinbase := lastPruningPoint.Transactions[transactionhelper.CoinbaseTransactionIndex]
		//_, _, subsidy, err := tc.CoinbaseManager().ExtractCoinbaseDataBlueScoreAndSubsidy(lastPruningPointCoinbase)
		//if err != nil {
		//	t.Fatalf("ExtractCoinbaseDataBlueScoreAndSubsidy: %+v", err)
		//}
		//if subsidy != consensusConfig.SubsidyGenesisReward {
		//	t.Fatalf("Subsidy has unexpected value. Want: %d, got: %d", consensusConfig.SubsidyGenesisReward, subsidy)
		//}
	})
}
