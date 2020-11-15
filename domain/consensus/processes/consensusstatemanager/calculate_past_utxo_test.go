package consensusstatemanager_test

import (
	"testing"

	"github.com/kaspanet/kaspad/domain/consensus"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/multiset"
	"github.com/kaspanet/kaspad/domain/consensus/utils/testutils"
	"github.com/kaspanet/kaspad/domain/dagconfig"
)

func TestUTXOCommitment(t *testing.T) {
	testutils.ForAllNets(t, true, func(t *testing.T, params *dagconfig.Params) {
		params.BlockCoinbaseMaturity = 0
		factory := consensus.NewFactory()

		consensus, teardown, err := factory.NewTestConsensus(params, "TestUTXOCommitment")
		if err != nil {
			t.Fatalf("Error setting up consensus: %+v", err)
		}
		defer teardown()

		// Build the following DAG:
		// G <- A <- B <- D
		//        <- C <-
		genesisHash := params.GenesisHash

		// Block A:
		blockAHash, err := consensus.AddBlock([]*externalapi.DomainHash{genesisHash}, nil, nil)
		if err != nil {
			t.Fatalf("Error creating block A: %+v", err)
		}
		blockA, err := consensus.GetBlock(blockAHash)
		if err != nil {
			t.Fatalf("Error getting block A: %+v", err)
		}
		// Block B:
		blockBHash, err := consensus.AddBlock([]*externalapi.DomainHash{blockAHash}, nil, nil)
		if err != nil {
			t.Fatalf("Error creating block B: %+v", err)
		}
		// Block C:
		blockCTransaction, err := testutils.CreateTransaction(blockA.Transactions[0])
		if err != nil {
			t.Fatalf("Error creating transaction: %+v", err)
		}
		blockCHash, err := consensus.AddBlock([]*externalapi.DomainHash{blockAHash}, nil,
			[]*externalapi.DomainTransaction{blockCTransaction})
		if err != nil {
			t.Fatalf("Error creating block C: %+v", err)
		}
		// Block D:
		blockDHash, err := consensus.AddBlock([]*externalapi.DomainHash{blockBHash, blockCHash}, nil, nil)
		if err != nil {
			t.Fatalf("Error creating block D: %+v", err)
		}
		blockD, err := consensus.GetBlock(blockDHash)
		if err != nil {
			t.Fatalf("Error getting block D: %+v", err)
		}

		// Get the past UTXO set of block D
		csm := consensus.ConsensusStateManager()
		utxoSetIterator, err := csm.RestorePastUTXOSetIterator(blockDHash)
		if err != nil {
			t.Fatalf("Error restoring past UTXO of block D: %+v", err)
		}

		// Build a Multiset for block D
		ms := multiset.New()
		for utxoSetIterator.Next() {
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
		if *utxoCommitment != blockD.Header.UTXOCommitment {
			t.Fatalf("TestUTXOCommitment: calculated UTXO commitment and "+
				"actual UTXO commitment don't match. Want: %s, got: %s",
				utxoCommitment, blockD.Header.UTXOCommitment)
		}
	})
}

func TestPastUTXOMultiset(t *testing.T) {
	testutils.ForAllNets(t, true, func(t *testing.T, params *dagconfig.Params) {
		factory := consensus.NewFactory()

		consensus, teardown, err := factory.NewTestConsensus(params, "TestUTXOCommitment")
		if err != nil {
			t.Fatalf("Error setting up consensus: %+v", err)
		}
		defer teardown()

		// Build a short chain
		currentHash := params.GenesisHash
		for i := 0; i < 3; i++ {
			currentHash, err = consensus.AddBlock([]*externalapi.DomainHash{currentHash}, nil, nil)
			if err != nil {
				t.Fatalf("Error creating block A: %+v", err)
			}
		}

		// Save the current tip's hash to be used lated
		testedBlockHash := currentHash

		// Take testedBlock's multiset and hash
		firstMultiset, err := consensus.MultisetStore().Get(consensus.DatabaseContext(), testedBlockHash)
		if err != nil {
			return
		}
		firstMultisetHash := firstMultiset.Hash()

		// Add another block on top of testedBlock
		_, err = consensus.AddBlock([]*externalapi.DomainHash{testedBlockHash}, nil, nil)
		if err != nil {
			t.Fatalf("Error creating block A: %+v", err)
		}

		// Take testedBlock's multiset and hash again
		secondMultiset, err := consensus.MultisetStore().Get(consensus.DatabaseContext(), testedBlockHash)
		if err != nil {
			return
		}
		secondMultisetHash := secondMultiset.Hash()

		// Make sure the multiset hasn't changed
		if *firstMultisetHash != *secondMultisetHash {
			t.Fatalf("TestPastUTXOMultiSet: selectedParentMultiset appears to have changed!")
		}
	})
}
