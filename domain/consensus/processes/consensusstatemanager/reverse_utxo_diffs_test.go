package consensusstatemanager_test

import (
	"testing"

	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"

	"github.com/kaspanet/kaspad/domain/consensus/model"

	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

	"github.com/kaspanet/kaspad/domain/consensus"
	"github.com/kaspanet/kaspad/domain/consensus/utils/testutils"
)

func TestReverseUTXODiffs(t *testing.T) {
	// This test doesn't check ReverseUTXODiffs directly, since that would be quite complicated,
	// instead, it creates a situation where a reversal would defenitely happen - a reorg of 5 blocks,
	// then verifies that the resulting utxo-diffs and utxo-diff-children are all correct.

	testutils.ForAllNets(t, true, func(t *testing.T, consensusConfig *consensus.Config) {
		factory := consensus.NewFactory()

		tc, teardown, err := factory.NewTestConsensus(consensusConfig, "TestUTXOCommitment")
		if err != nil {
			t.Fatalf("Error setting up consensus: %+v", err)
		}
		defer teardown(false)

		// Create a chain of 5 blocks
		const initialChainLength = 5
		previousBlockHash := consensusConfig.GenesisHash
		for i := 0; i < initialChainLength; i++ {
			previousBlockHash, _, err = tc.AddBlock([]*externalapi.DomainHash{previousBlockHash}, nil, nil)
			if err != nil {
				t.Fatalf("Error mining block no. %d in initial chain: %+v", i, err)
			}
		}

		// Mine a chain of 6 blocks, to re-organize the DAG
		const reorgChainLength = initialChainLength + 1
		reorgChain := make([]*externalapi.DomainHash, reorgChainLength)
		previousBlockHash = consensusConfig.GenesisHash
		for i := 0; i < reorgChainLength; i++ {
			previousBlockHash, _, err = tc.AddBlock([]*externalapi.DomainHash{previousBlockHash}, nil, nil)
			reorgChain[i] = previousBlockHash
			if err != nil {
				t.Fatalf("Error mining block no. %d in re-org chain: %+v", i, err)
			}
		}

		stagingArea := model.NewStagingArea()
		// Check that every block in the reorg chain has the next block as it's UTXODiffChild,
		// except that tip that has virtual, And that the diff is only `{ toRemove: { coinbase } }`
		for i, currentBlockHash := range reorgChain {
			if i == reorgChainLength-1 {
				hasUTXODiffChild, err := tc.UTXODiffStore().HasUTXODiffChild(tc.DatabaseContext(), stagingArea, currentBlockHash)
				if err != nil {
					t.Fatalf("Error getting HasUTXODiffChild of %s: %+v", currentBlockHash, err)
				}
				if hasUTXODiffChild {
					t.Errorf("Block %s expected utxoDiffChild is virtual, but HasUTXODiffChild returned true",
						currentBlockHash)
				}
			} else {
				utxoDiffChild, err := tc.UTXODiffStore().UTXODiffChild(tc.DatabaseContext(), stagingArea, currentBlockHash)
				if err != nil {
					t.Fatalf("Error getting utxoDiffChild of block No. %d, %s: %+v", i, currentBlockHash, err)
				}
				expectedUTXODiffChild := reorgChain[i+1]
				if !expectedUTXODiffChild.Equal(utxoDiffChild) {
					t.Errorf("Block %s expected utxoDiffChild is %s, but got %s instead",
						currentBlockHash, expectedUTXODiffChild, utxoDiffChild)
					continue
				}
			}

			// skip the first block, since it's coinbase doesn't create outputs
			if i == 0 {
				continue
			}

			currentBlock, err := tc.BlockStore().Block(tc.DatabaseContext(), stagingArea, currentBlockHash)
			if err != nil {
				t.Fatalf("Error getting block %s: %+v", currentBlockHash, err)
			}
			utxoDiff, err := tc.UTXODiffStore().UTXODiff(tc.DatabaseContext(), stagingArea, currentBlockHash)
			if err != nil {
				t.Fatalf("Error getting utxoDiffChild of %s: %+v", currentBlockHash, err)
			}
			if !checkIsUTXODiffOnlyRemoveCoinbase(t, utxoDiff, currentBlock) {
				t.Errorf("Expected %s to only have toRemove: {%s}, but got %s instead",
					currentBlockHash, consensushashing.TransactionID(currentBlock.Transactions[0]), utxoDiff)
			}
		}
	})
}

func checkIsUTXODiffOnlyRemoveCoinbase(t *testing.T, utxoDiff externalapi.UTXODiff, currentBlock *externalapi.DomainBlock) bool {
	if utxoDiff.ToAdd().Len() > 0 || utxoDiff.ToRemove().Len() > 1 {
		return false
	}

	iterator := utxoDiff.ToRemove().Iterator()
	iterator.First()
	outpoint, _, err := iterator.Get()
	if err != nil {
		t.Fatalf("Error getting from UTXODiff's iterator: %+v", err)
	}
	if !outpoint.TransactionID.Equal(consensushashing.TransactionID(currentBlock.Transactions[0])) {
		return false
	}

	return true
}
