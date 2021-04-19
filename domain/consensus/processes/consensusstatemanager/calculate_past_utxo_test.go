package consensusstatemanager_test

import (
	"testing"

	"github.com/kaspanet/kaspad/domain/consensus"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/model/testapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/multiset"
	"github.com/kaspanet/kaspad/domain/consensus/utils/testutils"
	"github.com/kaspanet/kaspad/domain/consensus/utils/transactionhelper"
)

func TestUTXOCommitment(t *testing.T) {
	testutils.ForAllNets(t, true, func(t *testing.T, consensusConfig *consensus.Config) {
		consensusConfig.BlockCoinbaseMaturity = 0
		factory := consensus.NewFactory()

		consensus, teardown, err := factory.NewTestConsensus(consensusConfig, "TestUTXOCommitment")
		if err != nil {
			t.Fatalf("Error setting up consensus: %+v", err)
		}
		defer teardown(false)

		// Build the following DAG:
		// G <- A <- B <- C <- E
		//             <- D <-
		// Where block D has a non-coinbase transaction
		genesisHash := consensusConfig.GenesisHash

		// Block A:
		blockAHash, _, err := consensus.AddBlock([]*externalapi.DomainHash{genesisHash}, nil, nil)
		if err != nil {
			t.Fatalf("Error creating block A: %+v", err)
		}
		checkBlockUTXOCommitment(t, consensus, blockAHash, "A")
		// Block B:
		blockBHash, _, err := consensus.AddBlock([]*externalapi.DomainHash{blockAHash}, nil, nil)
		if err != nil {
			t.Fatalf("Error creating block B: %+v", err)
		}
		blockB, err := consensus.GetBlock(blockBHash)
		if err != nil {
			t.Fatalf("Error getting block B: %+v", err)
		}
		checkBlockUTXOCommitment(t, consensus, blockBHash, "B")
		// Block C:
		blockCHash, _, err := consensus.AddBlock([]*externalapi.DomainHash{blockBHash}, nil, nil)
		if err != nil {
			t.Fatalf("Error creating block C: %+v", err)
		}
		checkBlockUTXOCommitment(t, consensus, blockCHash, "C")
		// Block D:
		blockDTransaction, err := testutils.CreateTransaction(
			blockB.Transactions[transactionhelper.CoinbaseTransactionIndex], 1)
		if err != nil {
			t.Fatalf("Error creating transaction: %+v", err)
		}
		blockDHash, _, err := consensus.AddBlock([]*externalapi.DomainHash{blockBHash}, nil,
			[]*externalapi.DomainTransaction{blockDTransaction})
		if err != nil {
			t.Fatalf("Error creating block D: %+v", err)
		}
		checkBlockUTXOCommitment(t, consensus, blockDHash, "D")
		// Block E:
		blockEHash, _, err := consensus.AddBlock([]*externalapi.DomainHash{blockCHash, blockDHash}, nil, nil)
		if err != nil {
			t.Fatalf("Error creating block E: %+v", err)
		}
		checkBlockUTXOCommitment(t, consensus, blockEHash, "E")
	})
}

func checkBlockUTXOCommitment(t *testing.T, consensus testapi.TestConsensus, blockHash *externalapi.DomainHash, blockName string) {
	block, err := consensus.GetBlock(blockHash)
	if err != nil {
		t.Fatalf("Error getting block %s: %+v", blockName, err)
	}

	// Get the past UTXO set of block
	csm := consensus.ConsensusStateManager()
	utxoSetIterator, err := csm.RestorePastUTXOSetIterator(model.NewStagingArea(), blockHash)
	if err != nil {
		t.Fatalf("Error restoring past UTXO of block %s: %+v", blockName, err)
	}
	defer utxoSetIterator.Close()

	// Build a Multiset
	ms := multiset.New()
	for ok := utxoSetIterator.First(); ok; ok = utxoSetIterator.Next() {
		outpoint, entry, err := utxoSetIterator.Get()
		if err != nil {
			t.Fatalf("Error getting from UTXOSet iterator: %+v", err)
		}
		err = consensus.ConsensusStateManager().AddUTXOToMultiset(ms, entry, outpoint)
		if err != nil {
			t.Fatalf("Error adding utxo to multiset: %+v", err)
		}
	}

	// Turn the multiset into a UTXO commitment
	utxoCommitment := ms.Hash()

	// Make sure that the two commitments are equal
	if !utxoCommitment.Equal(block.Header.UTXOCommitment()) {
		t.Fatalf("TestUTXOCommitment: calculated UTXO commitment for block %s and "+
			"actual UTXO commitment don't match. Want: %s, got: %s", blockName,
			utxoCommitment, block.Header.UTXOCommitment())
	}
}

func TestPastUTXOMultiset(t *testing.T) {
	testutils.ForAllNets(t, true, func(t *testing.T, consensusConfig *consensus.Config) {
		stagingArea := model.NewStagingArea()

		factory := consensus.NewFactory()

		consensus, teardown, err := factory.NewTestConsensus(consensusConfig, "TestUTXOCommitment")
		if err != nil {
			t.Fatalf("Error setting up consensus: %+v", err)
		}
		defer teardown(false)

		// Build a short chain
		currentHash := consensusConfig.GenesisHash
		for i := 0; i < 3; i++ {
			currentHash, _, err = consensus.AddBlock([]*externalapi.DomainHash{currentHash}, nil, nil)
			if err != nil {
				t.Fatalf("Error creating block A: %+v", err)
			}
		}

		// Save the current tip's hash to be used lated
		testedBlockHash := currentHash

		// Take testedBlock's multiset and hash
		firstMultiset, err := consensus.MultisetStore().Get(consensus.DatabaseContext(), stagingArea, testedBlockHash)
		if err != nil {
			return
		}
		firstMultisetHash := firstMultiset.Hash()

		// Add another block on top of testedBlock
		_, _, err = consensus.AddBlock([]*externalapi.DomainHash{testedBlockHash}, nil, nil)
		if err != nil {
			t.Fatalf("Error creating block A: %+v", err)
		}

		// Take testedBlock's multiset and hash again
		secondMultiset, err := consensus.MultisetStore().Get(consensus.DatabaseContext(), stagingArea, testedBlockHash)
		if err != nil {
			return
		}
		secondMultisetHash := secondMultiset.Hash()

		// Make sure the multiset hasn't changed
		if !firstMultisetHash.Equal(secondMultisetHash) {
			t.Fatalf("TestPastUTXOMultiSet: selectedParentMultiset appears to have changed!")
		}
	})
}
