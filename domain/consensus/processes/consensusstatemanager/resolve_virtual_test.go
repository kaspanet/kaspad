package consensusstatemanager_test

import (
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
	"testing"

	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

	"github.com/kaspanet/kaspad/domain/consensus"
	"github.com/kaspanet/kaspad/domain/consensus/utils/testutils"
)

func TestAddBlockBetweenResolveVirtualCalls(t *testing.T) {

	testutils.ForAllNets(t, true, func(t *testing.T, consensusConfig *consensus.Config) {
		factory := consensus.NewFactory()

		tc, teardown, err := factory.NewTestConsensus(consensusConfig, "TestAddBlockBetweenResolveVirtualCalls")
		if err != nil {
			t.Fatalf("Error setting up consensus: %+v", err)
		}
		defer teardown(false)

		// Create a chain of blocks
		const initialChainLength = 10
		previousBlockHash := consensusConfig.GenesisHash
		for i := 0; i < initialChainLength; i++ {
			previousBlockHash, _, err = tc.AddBlock([]*externalapi.DomainHash{previousBlockHash}, nil, nil)
			if err != nil {
				t.Fatalf("Error mining block no. %d in initial chain: %+v", i, err)
			}
		}

		// Mine a chain with more blocks, to re-organize the DAG
		const reorgChainLength = initialChainLength + 1
		previousBlockHash = consensusConfig.GenesisHash
		for i := 0; i < reorgChainLength; i++ {
			previousBlock, _, err := tc.BuildBlockWithParents([]*externalapi.DomainHash{previousBlockHash}, nil, nil)
			if err != nil {
				t.Fatalf("Error mining block no. %d in re-org chain: %+v", i, err)
			}
			previousBlockHash = consensushashing.BlockHash(previousBlock)

			// Do not UTXO validate in order to resolve virtual later
			_, err = tc.ValidateAndInsertBlock(previousBlock, false)
			if err != nil {
				t.Fatalf("Error mining block no. %d in re-org chain: %+v", i, err)
			}
		}

		// Resolve one step
		_, _, err = tc.ResolveVirtualWithMaxParam(2)
		if err != nil {
			t.Fatalf("Error resolving virtual in re-org chain: %+v", err)
		}

		emptyCoinbase := &externalapi.DomainCoinbaseData{
			ScriptPublicKey: &externalapi.ScriptPublicKey{
				Script:  nil,
				Version: 0,
			},
		}

		// Get template based on current resolve state
		blockTemplate, err := tc.BuildBlockTemplate(emptyCoinbase, nil)
		if err != nil {
			t.Fatalf("Error building block template during virtual resolution of reorg: %+v", err)
		}

		// Resolve one more step
		_, isCompletelyResolved, err := tc.ResolveVirtualWithMaxParam(2)
		if err != nil {
			t.Fatalf("Error resolving virtual in re-org chain: %+v", err)
		}

		// Add the mined block (now virtual was modified)
		_, err = tc.ValidateAndInsertBlock(blockTemplate.Block, true)
		if err != nil {
			t.Fatalf("Error mining block during virtual resolution of reorg: %+v", err)
		}

		// Complete resolving virtual
		for !isCompletelyResolved {
			_, isCompletelyResolved, err = tc.ResolveVirtualWithMaxParam(2)
			if err != nil {
				t.Fatalf("Error resolving virtual in re-org chain: %+v", err)
			}
		}
	})
}

func TestAddGenesisChildAfterOneResolveVirtualCall(t *testing.T) {

	testutils.ForAllNets(t, true, func(t *testing.T, consensusConfig *consensus.Config) {
		factory := consensus.NewFactory()

		tc, teardown, err := factory.NewTestConsensus(consensusConfig, "TestAddBlockBetweenResolveVirtualCalls")
		if err != nil {
			t.Fatalf("Error setting up consensus: %+v", err)
		}
		defer teardown(false)

		// Create a chain of blocks
		const initialChainLength = 10
		previousBlockHash := consensusConfig.GenesisHash
		for i := 0; i < initialChainLength; i++ {
			previousBlockHash, _, err = tc.AddBlock([]*externalapi.DomainHash{previousBlockHash}, nil, nil)
			if err != nil {
				t.Fatalf("Error mining block no. %d in initial chain: %+v", i, err)
			}
		}

		// Mine a chain with more blocks, to re-organize the DAG
		const reorgChainLength = initialChainLength + 1
		previousBlockHash = consensusConfig.GenesisHash
		for i := 0; i < reorgChainLength; i++ {
			previousBlock, _, err := tc.BuildBlockWithParents([]*externalapi.DomainHash{previousBlockHash}, nil, nil)
			if err != nil {
				t.Fatalf("Error mining block no. %d in re-org chain: %+v", i, err)
			}
			previousBlockHash = consensushashing.BlockHash(previousBlock)

			// Do not UTXO validate in order to resolve virtual later
			_, err = tc.ValidateAndInsertBlock(previousBlock, false)
			if err != nil {
				t.Fatalf("Error mining block no. %d in re-org chain: %+v", i, err)
			}
		}

		// Resolve one step
		_, isCompletelyResolved, err := tc.ResolveVirtualWithMaxParam(2)
		if err != nil {
			t.Fatalf("Error resolving virtual in re-org chain: %+v", err)
		}

		_, _, err = tc.AddBlock([]*externalapi.DomainHash{consensusConfig.GenesisHash}, nil, nil)
		if err != nil {
			t.Fatalf("Error adding block during virtual resolution of reorg: %+v", err)
		}

		// Complete resolving virtual
		for !isCompletelyResolved {
			_, isCompletelyResolved, err = tc.ResolveVirtualWithMaxParam(2)
			if err != nil {
				t.Fatalf("Error resolving virtual in re-org chain: %+v", err)
			}
		}
	})
}

func TestAddGenesisChildAfterTwoResolveVirtualCalls(t *testing.T) {

	testutils.ForAllNets(t, true, func(t *testing.T, consensusConfig *consensus.Config) {
		factory := consensus.NewFactory()

		tc, teardown, err := factory.NewTestConsensus(consensusConfig, "TestAddBlockBetweenResolveVirtualCalls")
		if err != nil {
			t.Fatalf("Error setting up consensus: %+v", err)
		}
		defer teardown(false)

		// Create a chain of blocks
		const initialChainLength = 10
		previousBlockHash := consensusConfig.GenesisHash
		for i := 0; i < initialChainLength; i++ {
			previousBlockHash, _, err = tc.AddBlock([]*externalapi.DomainHash{previousBlockHash}, nil, nil)
			if err != nil {
				t.Fatalf("Error mining block no. %d in initial chain: %+v", i, err)
			}
		}

		// Mine a chain with more blocks, to re-organize the DAG
		const reorgChainLength = initialChainLength + 1
		previousBlockHash = consensusConfig.GenesisHash
		for i := 0; i < reorgChainLength; i++ {
			previousBlock, _, err := tc.BuildBlockWithParents([]*externalapi.DomainHash{previousBlockHash}, nil, nil)
			if err != nil {
				t.Fatalf("Error mining block no. %d in re-org chain: %+v", i, err)
			}
			previousBlockHash = consensushashing.BlockHash(previousBlock)

			// Do not UTXO validate in order to resolve virtual later
			_, err = tc.ValidateAndInsertBlock(previousBlock, false)
			if err != nil {
				t.Fatalf("Error mining block no. %d in re-org chain: %+v", i, err)
			}
		}

		// Resolve one step
		_, _, err = tc.ResolveVirtualWithMaxParam(2)
		if err != nil {
			t.Fatalf("Error resolving virtual in re-org chain: %+v", err)
		}

		// Resolve one more step
		_, isCompletelyResolved, err := tc.ResolveVirtualWithMaxParam(2)
		if err != nil {
			t.Fatalf("Error resolving virtual in re-org chain: %+v", err)
		}

		_, _, err = tc.AddBlock([]*externalapi.DomainHash{consensusConfig.GenesisHash}, nil, nil)
		if err != nil {
			t.Fatalf("Error adding block during virtual resolution of reorg: %+v", err)
		}

		// Complete resolving virtual
		for !isCompletelyResolved {
			_, isCompletelyResolved, err = tc.ResolveVirtualWithMaxParam(2)
			if err != nil {
				t.Fatalf("Error resolving virtual in re-org chain: %+v", err)
			}
		}
	})
}
