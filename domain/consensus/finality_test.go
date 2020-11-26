package consensus

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensusserialization"
	"github.com/kaspanet/kaspad/domain/consensus/utils/testutils"
	"github.com/kaspanet/kaspad/domain/dagconfig"

	"testing"
)

func TestFinality(t *testing.T) {
	testutils.ForAllNets(t, true, func(t *testing.T, params *dagconfig.Params) {
		// Set finalityInterval to 50 blocks, so that test runs quickly
		params.FinalityDuration = 50 * params.TargetTimePerBlock

		factory := NewFactory()
		consensus, teardown, err := factory.NewTestConsensus(params, "TestFinality")
		if err != nil {
			t.Fatalf("Error setting up consensus: %+v", err)
		}
		defer teardown()

		buildAndInsertBlock := func(parentHashes []*externalapi.DomainHash) (*externalapi.DomainBlock, error) {
			block, err := consensus.BuildBlockWithParents(parentHashes, nil, nil)
			if err != nil {
				return nil, err
			}

			err = consensus.ValidateAndInsertBlock(block)
			if err != nil {
				return nil, err
			}
			return block, nil
		}

		// Build a chain of `finalityInterval - 1` blocks
		finalityInterval := params.FinalityDepth()
		var mainChainTip *externalapi.DomainBlock
		mainChainTipHash := params.GenesisHash

		for i := uint64(0); i < finalityInterval-1; i++ {
			mainChainTip, err = buildAndInsertBlock([]*externalapi.DomainHash{mainChainTipHash})
			if err != nil {
				t.Fatalf("TestFinality: Failed to process Block #%d: %v", i, err)
			}
			mainChainTipHash = consensusserialization.BlockHash(mainChainTip)

			blockInfo, err := consensus.GetBlockInfo(mainChainTipHash)
			if err != nil {
				t.Fatalf("TestFinality: Block #%d failed to get info: %v", i, err)
			}
			if blockInfo.BlockStatus != externalapi.StatusValid {
				t.Fatalf("Block #%d in main chain expected to have status '%s', but got '%s'",
					i, externalapi.StatusValid, blockInfo.BlockStatus)
			}
		}

		// Mine another chain of `finality-Interval - 2` blocks
		var sideChainTip *externalapi.DomainBlock
		sideChainTipHash := params.GenesisHash
		for i := uint64(0); i < finalityInterval-2; i++ {
			sideChainTip, err = buildAndInsertBlock([]*externalapi.DomainHash{sideChainTipHash})
			if err != nil {
				t.Fatalf("TestFinality: Failed to process sidechain Block #%d: %v", i, err)
			}
			sideChainTipHash = consensusserialization.BlockHash(sideChainTip)

			blockInfo, err := consensus.GetBlockInfo(sideChainTipHash)
			if err != nil {
				t.Fatalf("TestFinality: Block #%d failed to get info: %v", i, err)
			} else if !blockInfo.Exists {
				t.Fatalf("TestFinality: Failed getting block info, doesn't exists")
			}
			if blockInfo.BlockStatus != externalapi.StatusUTXOPendingVerification {
				t.Fatalf("Block #%d in side chain expected to have status '%s', but got '%s'",
					i, externalapi.StatusUTXOPendingVerification, blockInfo.BlockStatus)
			}
		}

		// Add two more blocks in the side-chain until it becomes the selected chain
		for i := uint64(0); i < 2; i++ {
			sideChainTip, err = buildAndInsertBlock([]*externalapi.DomainHash{sideChainTipHash})
			if err != nil {
				t.Fatalf("TestFinality: Failed to process sidechain Block #%d: %v", i, err)
			}
			sideChainTipHash = consensusserialization.BlockHash(sideChainTip)
		}

		// Make sure that now the sideChainTip is valid and selectedTip
		blockInfo, err := consensus.GetBlockInfo(sideChainTipHash)
		if err != nil {
			t.Fatalf("TestFinality: Failed to get block info: %v", err)
		} else if !blockInfo.Exists {
			t.Fatalf("TestFinality: Failed getting block info, doesn't exists")
		}
		if blockInfo.BlockStatus != externalapi.StatusValid {
			t.Fatalf("TestFinality: Overtaking block in side-chain expected to have status '%s', but got '%s'",
				externalapi.StatusValid, blockInfo.BlockStatus)
		}
		selectedTip, err := consensus.GetVirtualSelectedParent()
		if err != nil {
			t.Fatalf("TestFinality: Failed getting virtual selectedParent: %v", err)
		}
		if !consensusserialization.BlockHash(selectedTip).Equal(sideChainTipHash) {
			t.Fatalf("Overtaking block in side-chain is not selectedTip")
		}

		// Add two more blocks to main chain, to move finality point to first non-genesis block in mainChain
		for i := uint64(0); i < 2; i++ {
			mainChainTip, err = buildAndInsertBlock([]*externalapi.DomainHash{mainChainTipHash})
			if err != nil {
				t.Fatalf("TestFinality: Failed to process sidechain Block #%d: %v", i, err)
			}
			mainChainTipHash = consensusserialization.BlockHash(mainChainTip)
		}

		virtualFinality, err := consensus.ConsensusStateManager().VirtualFinalityPoint()
		if err != nil {
			t.Fatalf("TestFinality: Failed getting the virtual's finality point: %v", err)
		}

		if virtualFinality.Equal(params.GenesisHash) {
			t.Fatalf("virtual's finalityPoint is still genesis after adding finalityInterval + 1 blocks to the main chain")
		}

		// TODO: Make sure that a finality conflict notification is sent
		// Add two more blocks to the side chain, so that it violates finality and gets status UTXOPendingVerification even
		// though it is the block with the highest blue score.
		for i := uint64(0); i < 2; i++ {
			sideChainTip, err = buildAndInsertBlock([]*externalapi.DomainHash{sideChainTipHash})
			if err != nil {
				t.Fatalf("TestFinality: Failed to process sidechain Block #%d: %v", i, err)
			}
			sideChainTipHash = consensusserialization.BlockHash(sideChainTip)
		}

		// Check that sideChainTip hash higher blue score than the selected parent
		selectedTip, err = consensus.GetVirtualSelectedParent()
		if err != nil {
			t.Fatalf("TestFinality: Failed getting virtual selectedParent: %v", err)
		}
		selectedTipGhostDagData, err := consensus.GHOSTDAGDataStore().Get(consensus.DatabaseContext(), consensusserialization.BlockHash(selectedTip))
		if err != nil {
			t.Fatalf("TestFinality: Failed getting the ghost dag data of the selected tip: %v", err)
		}

		sideChainTipGhostDagData, err := consensus.GHOSTDAGDataStore().Get(consensus.DatabaseContext(), sideChainTipHash)
		if err != nil {
			t.Fatalf("TestFinality: Failed getting the ghost dag data of the sidechain tip: %v", err)
		}

		if selectedTipGhostDagData.BlueScore > sideChainTipGhostDagData.BlueScore {
			t.Fatalf("sideChainTip is not the bluest tip when it is expected to be")
		}

		// Blocks violating finality should have a UTXOPendingVerification status
		blockInfo, err = consensus.GetBlockInfo(sideChainTipHash)
		if err != nil {
			t.Fatalf("TestFinality: Failed to get block info: %v", err)
		} else if !blockInfo.Exists {
			t.Fatalf("TestFinality: Failed getting block info, doesn't exists")
		}
		if blockInfo.BlockStatus != externalapi.StatusUTXOPendingVerification {
			t.Fatalf("TestFinality: Finality violating block expected to have status '%s', but got '%s'",
				externalapi.StatusUTXOPendingVerification, blockInfo.BlockStatus)
		}
	})
}
