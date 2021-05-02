package stability_tests

import (
	"github.com/kaspanet/kaspad/domain/consensus"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/testutils"
	"testing"
	"time"
)

// TestManyTips checks how much time and how many blocks need to add until we get only one
// tip in a DAG that contains 10,000 tips.
func TestManyTips(t *testing.T) {
	testutils.ForAllNets(t, true, func(t *testing.T, consensusConfig *consensus.Config) {
		factory := consensus.NewFactory()
		testConsensus, teardown, err := factory.NewTestConsensus(consensusConfig, "TestManyTips")
		if err != nil {
			t.Fatalf("Error setting up testConsensus: %+v", err)
		}
		defer teardown(false)

		// Mines a chain of 1k blocks
		chainLength := 1000
		chainHash := consensusConfig.GenesisHash
		for i := 0; i < chainLength; i++ {
			chainHash, _, err = testConsensus.AddBlock([]*externalapi.DomainHash{chainHash}, nil, nil)
			if err != nil {
				t.Fatalf("Error creating a block: %+v", err)
			}
		}
		emptyCoinbase := externalapi.DomainCoinbaseData{
			ScriptPublicKey: &externalapi.ScriptPublicKey{
				Script:  nil,
				Version: 0,
			},
		}
		// Mines on top of the chain 10k tips
		startTimeCreateTips := time.Now()
		numOfTips := 10000
		for i := 0; i < numOfTips; i++ {
			_, _, err := testConsensus.AddBlock([]*externalapi.DomainHash{chainHash}, nil, nil)
			if err != nil {
				t.Fatalf("Error creating a block: %+v", err)
			}
		}
		durationCreateTips := time.Since(startTimeCreateTips)
		t.Logf("Took %s to create 1k tips.", durationCreateTips)

		// Starts mining with BuildBlock.
		currentTips, err := testConsensus.Tips()
		if err != nil {
			t.Fatalf("Failed get the current tips : %v", err)
		}
		startTime := time.Now()
		addedBlock := 0
		for len(currentTips) != 1 {
			addedBlock++
			block, err := testConsensus.BuildBlock(&emptyCoinbase, nil)
			if err != nil {
				t.Fatalf("BuildBlock: %+v", err)
			}
			startTimeValidate := time.Now()
			_, err = testConsensus.ValidateAndInsertBlock(block)
			if err != nil {
				t.Fatalf("testConsensus.ValidateAndInsertBlock with a block "+
					"straight from BuildBlock should not fail: %v", err)
			}
			durationValidate := time.Since(startTimeValidate)
			if durationValidate.Seconds() > 1 {
				t.Errorf("Took %s to validate one block \n", durationValidate)
			}
			currentTips, err = testConsensus.Tips()
		}
		t.Logf("We added %d blocks to reach a state where we have a single tip(after having 10k tips),"+
			" and it took %s\n", addedBlock, time.Since(startTime))
	})
}

// Conclusions TestManyTips :
// #kaspa-simnet : addedBlocks: 1111, times: 1m6.157471884s, 1m2.377676168s, 1m6.284468597s
// #kaspa-devnet : addedBlocks: 1111, times: 1m4.256895465s, 1m1.469580168s, 1m4.663453651s
// #kaspa-mainet : addedBlocks: 1111, times: 1m4.179190308s , 59.763526509s, 1m3.062770231s
// #kaspa-testnet-2 : addedBlocks: 1111, times: 1m4.33514436s, 1m1.432405061s, 1m4.796397035s
