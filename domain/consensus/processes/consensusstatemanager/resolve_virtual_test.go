package consensusstatemanager_test

import (
	"fmt"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/testapi"
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
			err = tc.ValidateAndInsertBlock(previousBlock, false)
			if err != nil {
				t.Fatalf("Error mining block no. %d in re-org chain: %+v", i, err)
			}
		}

		// Resolve one step
		_, err = tc.ResolveVirtualWithMaxParam(2)
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
		isCompletelyResolved, err := tc.ResolveVirtualWithMaxParam(2)
		if err != nil {
			t.Fatalf("Error resolving virtual in re-org chain: %+v", err)
		}

		// Add the mined block (now virtual was modified)
		err = tc.ValidateAndInsertBlock(blockTemplate.Block, true)
		if err != nil {
			t.Fatalf("Error mining block during virtual resolution of reorg: %+v", err)
		}

		// Complete resolving virtual
		for !isCompletelyResolved {
			isCompletelyResolved, err = tc.ResolveVirtualWithMaxParam(2)
			if err != nil {
				t.Fatalf("Error resolving virtual in re-org chain: %+v", err)
			}
		}
	})
}

func TestAddGenesisChildAfterOneResolveVirtualCall(t *testing.T) {

	testutils.ForAllNets(t, true, func(t *testing.T, consensusConfig *consensus.Config) {
		factory := consensus.NewFactory()

		tc, teardown, err := factory.NewTestConsensus(consensusConfig, "TestAddGenesisChildAfterOneResolveVirtualCall")
		if err != nil {
			t.Fatalf("Error setting up consensus: %+v", err)
		}
		defer teardown(false)

		// Create a chain of blocks
		const initialChainLength = 6
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
			err = tc.ValidateAndInsertBlock(previousBlock, false)
			if err != nil {
				t.Fatalf("Error mining block no. %d in re-org chain: %+v", i, err)
			}
		}

		// Resolve one step
		isCompletelyResolved, err := tc.ResolveVirtualWithMaxParam(2)
		if err != nil {
			t.Fatalf("Error resolving virtual in re-org chain: %+v", err)
		}

		_, _, err = tc.AddBlock([]*externalapi.DomainHash{consensusConfig.GenesisHash}, nil, nil)
		if err != nil {
			t.Fatalf("Error adding block during virtual resolution of reorg: %+v", err)
		}

		// Complete resolving virtual
		for !isCompletelyResolved {
			isCompletelyResolved, err = tc.ResolveVirtualWithMaxParam(2)
			if err != nil {
				t.Fatalf("Error resolving virtual in re-org chain: %+v", err)
			}
		}
	})
}

func TestAddGenesisChildAfterTwoResolveVirtualCalls(t *testing.T) {

	testutils.ForAllNets(t, true, func(t *testing.T, consensusConfig *consensus.Config) {
		factory := consensus.NewFactory()

		tc, teardown, err := factory.NewTestConsensus(consensusConfig, "TestAddGenesisChildAfterTwoResolveVirtualCalls")
		if err != nil {
			t.Fatalf("Error setting up consensus: %+v", err)
		}
		defer teardown(false)

		// Create a chain of blocks
		const initialChainLength = 6
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
			err = tc.ValidateAndInsertBlock(previousBlock, false)
			if err != nil {
				t.Fatalf("Error mining block no. %d in re-org chain: %+v", i, err)
			}
		}

		// Resolve one step
		_, err = tc.ResolveVirtualWithMaxParam(2)
		if err != nil {
			t.Fatalf("Error resolving virtual in re-org chain: %+v", err)
		}

		// Resolve one more step
		isCompletelyResolved, err := tc.ResolveVirtualWithMaxParam(2)
		if err != nil {
			t.Fatalf("Error resolving virtual in re-org chain: %+v", err)
		}

		_, _, err = tc.AddBlock([]*externalapi.DomainHash{consensusConfig.GenesisHash}, nil, nil)
		if err != nil {
			t.Fatalf("Error adding block during virtual resolution of reorg: %+v", err)
		}

		// Complete resolving virtual
		for !isCompletelyResolved {
			isCompletelyResolved, err = tc.ResolveVirtualWithMaxParam(2)
			if err != nil {
				t.Fatalf("Error resolving virtual in re-org chain: %+v", err)
			}
		}
	})
}

func TestResolveVirtualMess(t *testing.T) {

	testutils.ForAllNets(t, true, func(t *testing.T, consensusConfig *consensus.Config) {
		factory := consensus.NewFactory()

		tc, teardown, err := factory.NewTestConsensus(consensusConfig, "TestAddGenesisChildAfterTwoResolveVirtualCalls")
		if err != nil {
			t.Fatalf("Error setting up consensus: %+v", err)
		}
		defer teardown(false)

		hashes := []*externalapi.DomainHash{consensusConfig.GenesisHash}
		blocks := make(map[externalapi.DomainHash]string)
		blocks[*consensusConfig.GenesisHash] = "g"
		blocks[*model.VirtualBlockHash] = "v"
		//fmt.Printf("%s\n\n", consensusConfig.GenesisHash)

		// Create a chain of blocks
		const initialChainLength = 6
		previousBlockHash := consensusConfig.GenesisHash
		for i := 0; i < initialChainLength; i++ {
			previousBlockHash, _, err = tc.AddBlock([]*externalapi.DomainHash{previousBlockHash}, nil, nil)
			blocks[*previousBlockHash] = fmt.Sprintf("A_%d", i)
			hashes = append(hashes, previousBlockHash)
			//fmt.Printf("A_%d: %s\n", i, previousBlockHash)

			if err != nil {
				t.Fatalf("Error mining block no. %d in initial chain: %+v", i, err)
			}
		}

		//fmt.Printf("\n")

		// Mine a chain with more blocks, to re-organize the DAG
		const reorgChainLength = 20 // initialChainLength + 1
		previousBlockHash = consensusConfig.GenesisHash
		for i := 0; i < reorgChainLength; i++ {
			previousBlock, _, err := tc.BuildBlockWithParents([]*externalapi.DomainHash{previousBlockHash}, nil, nil)
			if err != nil {
				t.Fatalf("Error mining block no. %d in re-org chain: %+v", i, err)
			}
			previousBlockHash = consensushashing.BlockHash(previousBlock)
			blocks[*previousBlockHash] = fmt.Sprintf("B_%d", i)
			hashes = append(hashes, previousBlockHash)
			//fmt.Printf("B_%d: %s\n", i, previousBlockHash)

			// Do not UTXO validate in order to resolve virtual later
			err = tc.ValidateAndInsertBlock(previousBlock, false)
			if err != nil {
				t.Fatalf("Error mining block no. %d in re-org chain: %+v", i, err)
			}
		}

		//fmt.Printf("\n")

		//printUtxoDiffChildren(t, hashes, tc, blocks)

		// Resolve one step
		_, err = tc.ResolveVirtualWithMaxParam(3)
		if err != nil {
			printUtxoDiffChildren(t, hashes, tc, blocks)
			t.Fatalf("Error resolving virtual in re-org chain: %+v", err)
		}

		// Resolve one more step
		isCompletelyResolved, err := tc.ResolveVirtualWithMaxParam(3)
		if err != nil {
			t.Fatalf("Error resolving virtual in re-org chain: %+v", err)
		}

		// Complete resolving virtual
		for !isCompletelyResolved {
			isCompletelyResolved, err = tc.ResolveVirtualWithMaxParam(3)
			if err != nil {
				t.Fatalf("Error resolving virtual in re-org chain: %+v", err)
			}
		}

		//printUtxoDiffChildren(t, hashes, tc, blocks)
	})
}

func printUtxoDiffChildren(t *testing.T, hashes []*externalapi.DomainHash, tc testapi.TestConsensus, blocks map[externalapi.DomainHash]string) {
	fmt.Printf("\n===============================\nBlock\t\tDiff child\n")
	stagingArea := model.NewStagingArea()
	for _, block := range hashes {
		hasUTXODiffChild, err := tc.UTXODiffStore().HasUTXODiffChild(tc.DatabaseContext(), stagingArea, block)
		if err != nil {
			t.Fatalf("Error while reading utxo diff store: %+v", err)
		}
		if hasUTXODiffChild {
			utxoDiffChild, err := tc.UTXODiffStore().UTXODiffChild(tc.DatabaseContext(), stagingArea, block)
			if err != nil {
				t.Fatalf("Error while reading utxo diff store: %+v", err)
			}
			fmt.Printf("%s\t\t\t%s\n", blocks[*block], blocks[*utxoDiffChild])
		} else {
			fmt.Printf("%s\n", blocks[*block])
		}
	}
	fmt.Printf("\n===============================\n")
}
