package consensus_test

import (
	"fmt"
	"github.com/kaspanet/kaspad/domain/consensus"
	"testing"

	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/model/testapi"
	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
	"github.com/kaspanet/kaspad/domain/consensus/utils/testutils"
	"github.com/pkg/errors"
)

func TestFinality(t *testing.T) {
	testutils.ForAllNets(t, true, func(t *testing.T, consensusConfig *consensus.Config) {
		// Set finalityInterval to 50 blocks, so that test runs quickly
		consensusConfig.FinalityDuration = 50 * consensusConfig.TargetTimePerBlock

		factory := consensus.NewFactory()
		consensus, teardown, err := factory.NewTestConsensus(consensusConfig, "TestFinality")
		if err != nil {
			t.Fatalf("Error setting up consensus: %+v", err)
		}
		defer teardown(false)

		buildAndInsertBlock := func(parentHashes []*externalapi.DomainHash) (*externalapi.DomainBlock, error) {
			block, _, err := consensus.BuildBlockWithParents(parentHashes, nil, nil)
			if err != nil {
				return nil, err
			}

			_, err = consensus.ValidateAndInsertBlock(block)
			if err != nil {
				return nil, err
			}
			return block, nil
		}

		// Build a chain of `finalityInterval - 1` blocks
		finalityInterval := consensusConfig.FinalityDepth()
		var mainChainTip *externalapi.DomainBlock
		mainChainTipHash := consensusConfig.GenesisHash

		for i := uint64(0); i < finalityInterval-1; i++ {
			mainChainTip, err = buildAndInsertBlock([]*externalapi.DomainHash{mainChainTipHash})
			if err != nil {
				t.Fatalf("TestFinality: Failed to process Block #%d: %+v", i, err)
			}
			mainChainTipHash = consensushashing.BlockHash(mainChainTip)

			blockInfo, err := consensus.GetBlockInfo(mainChainTipHash)
			if err != nil {
				t.Fatalf("TestFinality: Block #%d failed to get info: %+v", i, err)
			}
			if blockInfo.BlockStatus != externalapi.StatusUTXOValid {
				t.Fatalf("Block #%d in main chain expected to have status '%s', but got '%s'",
					i, externalapi.StatusUTXOValid, blockInfo.BlockStatus)
			}
		}

		// Mine another chain of `finality-Interval - 2` blocks
		var sideChainTip *externalapi.DomainBlock
		sideChainTipHash := consensusConfig.GenesisHash
		for i := uint64(0); i < finalityInterval-2; i++ {
			sideChainTip, err = buildAndInsertBlock([]*externalapi.DomainHash{sideChainTipHash})
			if err != nil {
				t.Fatalf("TestFinality: Failed to process sidechain Block #%d: %+v", i, err)
			}
			sideChainTipHash = consensushashing.BlockHash(sideChainTip)

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

		stagingArea := model.NewStagingArea()

		// Add two more blocks in the side-chain until it becomes the selected chain
		for i := uint64(0); i < 2; i++ {
			sideChainTip, err = buildAndInsertBlock([]*externalapi.DomainHash{sideChainTipHash})
			if err != nil {
				t.Fatalf("TestFinality: Failed to process sidechain Block #%d: %v", i, err)
			}
			sideChainTipHash = consensushashing.BlockHash(sideChainTip)
		}

		// Make sure that now the sideChainTip is valid and selectedTip
		blockInfo, err := consensus.GetBlockInfo(sideChainTipHash)
		if err != nil {
			t.Fatalf("TestFinality: Failed to get block info: %v", err)
		} else if !blockInfo.Exists {
			t.Fatalf("TestFinality: Failed getting block info, doesn't exists")
		}
		if blockInfo.BlockStatus != externalapi.StatusUTXOValid {
			t.Fatalf("TestFinality: Overtaking block in side-chain expected to have status '%s', but got '%s'",
				externalapi.StatusUTXOValid, blockInfo.BlockStatus)
		}
		selectedTip, err := consensus.GetVirtualSelectedParent()
		if err != nil {
			t.Fatalf("TestFinality: Failed getting virtual selectedParent: %v", err)
		}
		if !selectedTip.Equal(sideChainTipHash) {
			t.Fatalf("Overtaking block in side-chain is not selectedTip")
		}

		// Add two more blocks to main chain, to move finality point to first non-genesis block in mainChain
		for i := uint64(0); i < 2; i++ {
			mainChainTip, err = buildAndInsertBlock([]*externalapi.DomainHash{mainChainTipHash})
			if err != nil {
				t.Fatalf("TestFinality: Failed to process sidechain Block #%d: %v", i, err)
			}
			mainChainTipHash = consensushashing.BlockHash(mainChainTip)
		}

		virtualFinality, err := consensus.FinalityManager().VirtualFinalityPoint(stagingArea)
		if err != nil {
			t.Fatalf("TestFinality: Failed getting the virtual's finality point: %v", err)
		}

		if virtualFinality.Equal(consensusConfig.GenesisHash) {
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
			sideChainTipHash = consensushashing.BlockHash(sideChainTip)
		}

		// Check that sideChainTip hash higher blue score than the selected parent
		selectedTip, err = consensus.GetVirtualSelectedParent()
		if err != nil {
			t.Fatalf("TestFinality: Failed getting virtual selectedParent: %v", err)
		}
		selectedTipGhostDagData, err :=
			consensus.GHOSTDAGDataStore().Get(consensus.DatabaseContext(), stagingArea, selectedTip)
		if err != nil {
			t.Fatalf("TestFinality: Failed getting the ghost dag data of the selected tip: %v", err)
		}

		sideChainTipGhostDagData, err :=
			consensus.GHOSTDAGDataStore().Get(consensus.DatabaseContext(), stagingArea, sideChainTipHash)
		if err != nil {
			t.Fatalf("TestFinality: Failed getting the ghost dag data of the sidechain tip: %v", err)
		}

		if selectedTipGhostDagData.BlueWork().Cmp(sideChainTipGhostDagData.BlueWork()) == 1 {
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

func TestBoundedMergeDepth(t *testing.T) {
	testutils.ForAllNets(t, true, func(t *testing.T, consensusConfig *consensus.Config) {
		// Set finalityInterval to 50 blocks, so that test runs quickly
		consensusConfig.FinalityDuration = 50 * consensusConfig.TargetTimePerBlock
		finalityInterval := int(consensusConfig.FinalityDepth())

		if int(consensusConfig.K) >= finalityInterval {
			t.Fatal("K must be smaller than finality duration for this test to run")
		}

		checkViolatingMergeDepth := func(consensus testapi.TestConsensus, parents []*externalapi.DomainHash) (*externalapi.DomainBlock, bool) {
			block, _, err := consensus.BuildBlockWithParents(parents, nil, nil)
			if err != nil {
				t.Fatalf("TestBoundedMergeDepth: BuildBlockWithParents failed: %+v", err)
				return nil, false // fo some reason go doesn't recognize that t.Fatalf never returns
			}

			_, err = consensus.ValidateAndInsertBlock(block)
			if err == nil {
				return block, false
			} else if errors.Is(err, ruleerrors.ErrViolatingBoundedMergeDepth) {
				return block, true
			} else {
				t.Fatalf("TestBoundedMergeDepth: expected err: %v, found err: %v", ruleerrors.ErrViolatingBoundedMergeDepth, err)
				return nil, false // fo some reason go doesn't recognize that t.Fatalf never returns
			}
		}

		processBlock := func(consensus testapi.TestConsensus, block *externalapi.DomainBlock, name string) {
			_, err := consensus.ValidateAndInsertBlock(block)
			if err != nil {
				t.Fatalf("TestBoundedMergeDepth: %s got unexpected error from ProcessBlock: %+v", name, err)

			}
		}

		buildAndInsertBlock := func(consensus testapi.TestConsensus, parentHashes []*externalapi.DomainHash) *externalapi.DomainBlock {
			block, _, err := consensus.BuildBlockWithParents(parentHashes, nil, nil)
			if err != nil {
				t.Fatalf("TestBoundedMergeDepth: Failed building block: %v", err)
			}
			_, err = consensus.ValidateAndInsertBlock(block)
			if err != nil {
				t.Fatalf("TestBoundedMergeDepth: Failed Inserting block to consensus: %v", err)
			}
			return block
		}

		getStatus := func(consensus testapi.TestConsensus, block *externalapi.DomainBlock) externalapi.BlockStatus {
			blockInfo, err := consensus.GetBlockInfo(consensushashing.BlockHash(block))
			if err != nil {
				t.Fatalf("TestBoundedMergeDepth: Failed to get block info: %v", err)
			} else if !blockInfo.Exists {
				t.Fatalf("TestBoundedMergeDepth: Failed to get block info, block doesn't exists")
			}
			return blockInfo.BlockStatus
		}

		factory := consensus.NewFactory()
		consensusBuild, teardownFunc1, err := factory.NewTestConsensus(consensusConfig, "TestBoundedMergeTestBuild")
		if err != nil {
			t.Fatalf("TestBoundedMergeDepth: Error setting up consensus: %+v", err)
		}

		consensusReal, teardownFunc2, err := factory.NewTestConsensus(consensusConfig, "TestBoundedMergeTestReal")
		if err != nil {
			t.Fatalf("TestBoundedMergeDepth: Error setting up consensus: %+v", err)
		}
		defer teardownFunc2(false)

		// Create a block on top on genesis
		block1 := buildAndInsertBlock(consensusBuild, []*externalapi.DomainHash{consensusConfig.GenesisHash})

		// Create a chain
		selectedChain := make([]*externalapi.DomainBlock, 0, finalityInterval+1)
		parent := consensushashing.BlockHash(block1)
		// Make sure this is always bigger than `blocksChain2` so it will stay the selected chain
		for i := 0; i < finalityInterval+2; i++ {
			block := buildAndInsertBlock(consensusBuild, []*externalapi.DomainHash{parent})
			selectedChain = append(selectedChain, block)
			parent = consensushashing.BlockHash(block)
		}

		// Create another chain
		blocksChain2 := make([]*externalapi.DomainBlock, 0, finalityInterval+1)
		parent = consensushashing.BlockHash(block1)
		for i := 0; i < finalityInterval+1; i++ {
			block := buildAndInsertBlock(consensusBuild, []*externalapi.DomainHash{parent})
			blocksChain2 = append(blocksChain2, block)
			parent = consensushashing.BlockHash(block)
		}

		// Teardown and assign nil to make sure we use the right DAG from here on.
		teardownFunc1(false)
		consensusBuild = nil

		// Now test against the real DAG
		// submit block1
		processBlock(consensusReal, block1, "block1")

		// submit chain1
		for i, block := range selectedChain {
			processBlock(consensusReal, block, fmt.Sprintf("selectedChain block No %d", i))
		}

		// submit chain2
		for i, block := range blocksChain2 {
			processBlock(consensusReal, block, fmt.Sprintf("blocksChain2 block No %d", i))
		}

		// submit a block pointing at tip(chain1) and on first block in chain2 directly
		mergeDepthViolatingBlockBottom, isViolatingMergeDepth := checkViolatingMergeDepth(consensusReal, []*externalapi.DomainHash{consensushashing.BlockHash(blocksChain2[0]), consensushashing.BlockHash(selectedChain[len(selectedChain)-1])})
		if !isViolatingMergeDepth {
			t.Fatalf("TestBoundedMergeDepth: Expected mergeDepthViolatingBlockBottom to violate merge depth")
		}

		// submit a block pointing at tip(chain1) and tip(chain2) should also obviously violate merge depth (this points at first block in chain2 indirectly)
		mergeDepthViolatingTop, isViolatingMergeDepth := checkViolatingMergeDepth(consensusReal, []*externalapi.DomainHash{consensushashing.BlockHash(blocksChain2[len(blocksChain2)-1]), consensushashing.BlockHash(selectedChain[len(selectedChain)-1])})
		if !isViolatingMergeDepth {
			t.Fatalf("TestBoundedMergeDepth: Expected mergeDepthViolatingTop to violate merge depth")
		}

		// the location of the parents in the slices need to be both `-X` so the `selectedChain` one will have higher blueScore (it's a chain longer by 1)
		kosherizingBlock, isViolatingMergeDepth := checkViolatingMergeDepth(consensusReal, []*externalapi.DomainHash{consensushashing.BlockHash(blocksChain2[len(blocksChain2)-3]), consensushashing.BlockHash(selectedChain[len(selectedChain)-3])})
		kosherizingBlockHash := consensushashing.BlockHash(kosherizingBlock)
		if isViolatingMergeDepth {
			t.Fatalf("TestBoundedMergeDepth: Expected blueKosherizingBlock to not violate merge depth")
		}

		stagingArea := model.NewStagingArea()
		virtualGhotDagData, err := consensusReal.GHOSTDAGDataStore().Get(consensusReal.DatabaseContext(),
			stagingArea, model.VirtualBlockHash)
		if err != nil {
			t.Fatalf("TestBoundedMergeDepth: Failed getting the ghostdag data of the virtual: %v", err)
		}
		// Make sure it's actually blue
		found := false
		for _, blue := range virtualGhotDagData.MergeSetBlues() {
			if blue.Equal(kosherizingBlockHash) {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("TestBoundedMergeDepth: Expected kosherizingBlock to be blue by the virtual")
		}

		pointAtBlueKosherizing, isViolatingMergeDepth := checkViolatingMergeDepth(consensusReal, []*externalapi.DomainHash{kosherizingBlockHash, consensushashing.BlockHash(selectedChain[len(selectedChain)-1])})
		if isViolatingMergeDepth {

			t.Fatalf("TestBoundedMergeDepth: Expected selectedTip to not violate merge depth")
		}

		virtualSelectedParent, err := consensusReal.GetVirtualSelectedParent()
		if err != nil {
			t.Fatalf("TestBoundedMergeDepth: Failed getting the virtual selected parent %v", err)
		}

		if !virtualSelectedParent.Equal(consensushashing.BlockHash(pointAtBlueKosherizing)) {
			t.Fatalf("TestBoundedMergeDepth: Expected %s to be the selectedTip but found %s instead", consensushashing.BlockHash(pointAtBlueKosherizing), virtualSelectedParent)
		}

		// Now let's make the kosherizing block red and try to merge again
		tip := consensushashing.BlockHash(selectedChain[len(selectedChain)-1])
		// we use k-1 because `kosherizingBlock` points at tip-2, so 2+k-1 = k+1 anticone.
		for i := 0; i < int(consensusConfig.K)-1; i++ {
			block := buildAndInsertBlock(consensusReal, []*externalapi.DomainHash{tip})
			tip = consensushashing.BlockHash(block)
		}

		virtualSelectedParent, err = consensusReal.GetVirtualSelectedParent()
		if err != nil {
			t.Fatalf("TestBoundedMergeDepth: Failed getting the virtual selected parent %v", err)
		}

		if !virtualSelectedParent.Equal(tip) {
			t.Fatalf("TestBoundedMergeDepth: Expected %s to be the selectedTip but found %s instead", tip, virtualSelectedParent)
		}

		virtualGhotDagData, err = consensusReal.GHOSTDAGDataStore().Get(
			consensusReal.DatabaseContext(), stagingArea, model.VirtualBlockHash)
		if err != nil {
			t.Fatalf("TestBoundedMergeDepth: Failed getting the ghostdag data of the virtual: %v", err)
		}
		// Make sure it's actually blue
		found = false
		for _, blue := range virtualGhotDagData.MergeSetBlues() {
			if blue.Equal(kosherizingBlockHash) {
				found = true
				break
			}
		}
		if found {
			t.Fatalf("expected kosherizingBlock to be red by the virtual")
		}

		pointAtRedKosherizing, isViolatingMergeDepth := checkViolatingMergeDepth(consensusReal, []*externalapi.DomainHash{kosherizingBlockHash, tip})
		if !isViolatingMergeDepth {
			t.Fatalf("TestBoundedMergeDepth: Expected selectedTipRedKosherize to violate merge depth")
		}

		// Now `pointAtBlueKosherizing` itself is actually still blue, so we can still point at that even though we can't point at kosherizing directly anymore
		transitiveBlueKosherizing, isViolatingMergeDepth :=
			checkViolatingMergeDepth(consensusReal, []*externalapi.DomainHash{consensushashing.BlockHash(pointAtBlueKosherizing), tip})
		if isViolatingMergeDepth {
			t.Fatalf("TestBoundedMergeDepth: Expected transitiveBlueKosherizing to not violate merge depth")
		}

		virtualSelectedParent, err = consensusReal.GetVirtualSelectedParent()
		if err != nil {
			t.Fatalf("TestBoundedMergeDepth: Failed getting the virtual selected parent %v", err)
		}

		if !virtualSelectedParent.Equal(consensushashing.BlockHash(transitiveBlueKosherizing)) {
			t.Fatalf("TestBoundedMergeDepth: Expected %s to be the selectedTip but found %s instead", consensushashing.BlockHash(transitiveBlueKosherizing), virtualSelectedParent)
		}

		// Lets validate the status of all the interesting blocks
		if getStatus(consensusReal, pointAtBlueKosherizing) != externalapi.StatusUTXOValid {
			t.Fatalf("TestBoundedMergeDepth: pointAtBlueKosherizing expected status '%s' but got '%s'", externalapi.StatusUTXOValid, getStatus(consensusReal, pointAtBlueKosherizing))
		}
		if getStatus(consensusReal, pointAtRedKosherizing) != externalapi.StatusInvalid {
			t.Fatalf("TestBoundedMergeDepth: pointAtRedKosherizing expected status '%s' but got '%s'", externalapi.StatusInvalid, getStatus(consensusReal, pointAtRedKosherizing))
		}
		if getStatus(consensusReal, transitiveBlueKosherizing) != externalapi.StatusUTXOValid {
			t.Fatalf("TestBoundedMergeDepth: transitiveBlueKosherizing expected status '%s' but got '%s'", externalapi.StatusUTXOValid, getStatus(consensusReal, transitiveBlueKosherizing))
		}
		if getStatus(consensusReal, mergeDepthViolatingBlockBottom) != externalapi.StatusInvalid {
			t.Fatalf("TestBoundedMergeDepth: mergeDepthViolatingBlockBottom expected status '%s' but got '%s'", externalapi.StatusInvalid, getStatus(consensusReal, mergeDepthViolatingBlockBottom))
		}
		if getStatus(consensusReal, mergeDepthViolatingTop) != externalapi.StatusInvalid {
			t.Fatalf("TestBoundedMergeDepth: mergeDepthViolatingTop expected status '%s' but got '%s'", externalapi.StatusInvalid, getStatus(consensusReal, mergeDepthViolatingTop))
		}
		if getStatus(consensusReal, kosherizingBlock) != externalapi.StatusUTXOPendingVerification {
			t.Fatalf("kosherizingBlock expected status '%s' but got '%s'", externalapi.StatusUTXOPendingVerification, getStatus(consensusReal, kosherizingBlock))
		}

		for i, b := range blocksChain2 {
			if getStatus(consensusReal, b) != externalapi.StatusUTXOPendingVerification {
				t.Fatalf("blocksChain2[%d] expected status '%s' but got '%s'", i, externalapi.StatusUTXOPendingVerification, getStatus(consensusReal, b))
			}
		}
		for i, b := range selectedChain {
			if getStatus(consensusReal, b) != externalapi.StatusUTXOValid {
				t.Fatalf("selectedChain[%d] expected status '%s' but got '%s'", i, externalapi.StatusUTXOValid, getStatus(consensusReal, b))
			}
		}
	})
}
