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

		hashes := []*externalapi.DomainHash{consensusConfig.GenesisHash}

		// Create a chain of blocks
		const initialChainLength = 10
		previousBlockHash := consensusConfig.GenesisHash
		for i := 0; i < initialChainLength; i++ {
			previousBlockHash, _, err = tc.AddBlock([]*externalapi.DomainHash{previousBlockHash}, nil, nil)
			hashes = append(hashes, previousBlockHash)
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
			hashes = append(hashes, previousBlockHash)

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
		hashes = append(hashes, consensushashing.BlockHash(blockTemplate.Block))

		// Complete resolving virtual
		for !isCompletelyResolved {
			isCompletelyResolved, err = tc.ResolveVirtualWithMaxParam(2)
			if err != nil {
				t.Fatalf("Error resolving virtual in re-org chain: %+v", err)
			}
		}

		verifyUtxoDiffPaths(t, tc, hashes)
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

		hashes := []*externalapi.DomainHash{consensusConfig.GenesisHash}

		// Create a chain of blocks
		const initialChainLength = 6
		previousBlockHash := consensusConfig.GenesisHash
		for i := 0; i < initialChainLength; i++ {
			previousBlockHash, _, err = tc.AddBlock([]*externalapi.DomainHash{previousBlockHash}, nil, nil)
			hashes = append(hashes, previousBlockHash)
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
			hashes = append(hashes, previousBlockHash)

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

		verifyUtxoDiffPaths(t, tc, hashes)
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

		hashes := []*externalapi.DomainHash{consensusConfig.GenesisHash}

		// Create a chain of blocks
		const initialChainLength = 6
		previousBlockHash := consensusConfig.GenesisHash
		for i := 0; i < initialChainLength; i++ {
			previousBlockHash, _, err = tc.AddBlock([]*externalapi.DomainHash{previousBlockHash}, nil, nil)
			hashes = append(hashes, previousBlockHash)
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
			hashes = append(hashes, previousBlockHash)

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

		verifyUtxoDiffPaths(t, tc, hashes)
	})
}

func TestResolveVirtualBackAndForthReorgs(t *testing.T) {

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
		printfDebug("%s\n\n", consensusConfig.GenesisHash)

		// Create a chain of blocks
		const initialChainLength = 6
		previousBlockHash := consensusConfig.GenesisHash
		for i := 0; i < initialChainLength; i++ {
			previousBlockHash, _, err = tc.AddBlock([]*externalapi.DomainHash{previousBlockHash}, nil, nil)
			blocks[*previousBlockHash] = fmt.Sprintf("A_%d", i)
			hashes = append(hashes, previousBlockHash)
			printfDebug("A_%d: %s\n", i, previousBlockHash)

			if err != nil {
				t.Fatalf("Error mining block no. %d in initial chain: %+v", i, err)
			}
		}

		printfDebug("\n")
		verifyUtxoDiffPaths(t, tc, hashes)

		firstChainTip := previousBlockHash

		// Mine a chain with more blocks, to re-organize the DAG
		const reorgChainLength = 12 // initialChainLength + 1
		previousBlockHash = consensusConfig.GenesisHash
		for i := 0; i < reorgChainLength; i++ {
			previousBlock, _, err := tc.BuildBlockWithParents([]*externalapi.DomainHash{previousBlockHash}, nil, nil)
			if err != nil {
				t.Fatalf("Error mining block no. %d in re-org chain: %+v", i, err)
			}
			previousBlockHash = consensushashing.BlockHash(previousBlock)
			blocks[*previousBlockHash] = fmt.Sprintf("B_%d", i)
			hashes = append(hashes, previousBlockHash)
			printfDebug("B_%d: %s\n", i, previousBlockHash)

			// Do not UTXO validate in order to resolve virtual later
			err = tc.ValidateAndInsertBlock(previousBlock, false)
			if err != nil {
				t.Fatalf("Error mining block no. %d in re-org chain: %+v", i, err)
			}
		}

		printfDebug("\n")

		printUtxoDiffChildren(t, tc, hashes, blocks)
		verifyUtxoDiffPaths(t, tc, hashes)

		// Resolve one step
		_, err = tc.ResolveVirtualWithMaxParam(3)
		if err != nil {
			printUtxoDiffChildren(t, tc, hashes, blocks)
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

		printUtxoDiffChildren(t, tc, hashes, blocks)
		verifyUtxoDiffPaths(t, tc, hashes)

		// Now get the first chain back to the wining position
		previousBlockHash = firstChainTip
		for i := 0; i < reorgChainLength; i++ {
			previousBlockHash, _, err = tc.AddBlock([]*externalapi.DomainHash{previousBlockHash}, nil, nil)
			blocks[*previousBlockHash] = fmt.Sprintf("A_%d", initialChainLength+i)
			hashes = append(hashes, previousBlockHash)
			printfDebug("A_%d: %s\n", initialChainLength+i, previousBlockHash)

			if err != nil {
				t.Fatalf("Error mining block no. %d in initial chain: %+v", initialChainLength+i, err)
			}
		}

		printfDebug("\n")

		printUtxoDiffChildren(t, tc, hashes, blocks)
		verifyUtxoDiffPaths(t, tc, hashes)
	})
}

func verifyUtxoDiffPathToRoot(t *testing.T, tc testapi.TestConsensus, stagingArea *model.StagingArea, block, utxoDiffRoot *externalapi.DomainHash) {
	current := block
	for !current.Equal(utxoDiffRoot) {
		hasUTXODiffChild, err := tc.UTXODiffStore().HasUTXODiffChild(tc.DatabaseContext(), stagingArea, current)
		if err != nil {
			t.Fatalf("Error while reading utxo diff store: %+v", err)
		}
		if !hasUTXODiffChild {
			t.Fatalf("%s is expected to have a UTXO diff child", current)
		}
		current, err = tc.UTXODiffStore().UTXODiffChild(tc.DatabaseContext(), stagingArea, current)
		if err != nil {
			t.Fatalf("Error while reading utxo diff store: %+v", err)
		}
	}
}

func verifyUtxoDiffPaths(t *testing.T, tc testapi.TestConsensus, hashes []*externalapi.DomainHash) {
	stagingArea := model.NewStagingArea()

	virtualGHOSTDAGData, err := tc.GHOSTDAGDataStore().Get(tc.DatabaseContext(), stagingArea, model.VirtualBlockHash, false)
	if err != nil {
		t.Fatal(err)
	}

	utxoDiffRoot := virtualGHOSTDAGData.SelectedParent()
	hasUTXODiffChild, err := tc.UTXODiffStore().HasUTXODiffChild(tc.DatabaseContext(), stagingArea, utxoDiffRoot)
	if err != nil {
		t.Fatalf("Error while reading utxo diff store: %+v", err)
	}
	if hasUTXODiffChild {
		t.Fatalf("Virtual selected parent is not expected to have an explicit diff child")
	}
	_, err = tc.UTXODiffStore().UTXODiff(tc.DatabaseContext(), stagingArea, utxoDiffRoot)
	if err != nil {
		t.Fatalf("Virtual selected parent is expected to have a utxo diff: %+v", err)
	}

	for _, block := range hashes {
		hasUTXODiffChild, err = tc.UTXODiffStore().HasUTXODiffChild(tc.DatabaseContext(), stagingArea, block)
		if err != nil {
			t.Fatalf("Error while reading utxo diff store: %+v", err)
		}
		isOnVirtualSelectedChain, err := tc.DAGTopologyManager().IsInSelectedParentChainOf(stagingArea, block, utxoDiffRoot)
		if err != nil {
			t.Fatal(err)
		}
		// We expect a valid path to root in both cases: (i) block has a diff child, (ii) block is on the virtual selected chain
		if hasUTXODiffChild || isOnVirtualSelectedChain {
			verifyUtxoDiffPathToRoot(t, tc, stagingArea, block, utxoDiffRoot)
		}
	}
}

func printfDebug(format string, a ...any) {
	// Uncomment below when debugging the test
	//fmt.Printf(format, a...)
}

func printUtxoDiffChildren(t *testing.T, tc testapi.TestConsensus, hashes []*externalapi.DomainHash, blocks map[externalapi.DomainHash]string) {
	printfDebug("\n===============================\nBlock\t\tDiff child\n")
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
			printfDebug("%s\t\t\t%s\n", blocks[*block], blocks[*utxoDiffChild])
		} else {
			printfDebug("%s\n", blocks[*block])
		}
	}
	printfDebug("\n===============================\n")
}
