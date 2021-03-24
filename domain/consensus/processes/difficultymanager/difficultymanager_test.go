package difficultymanager_test

import (
	"testing"
	"time"

	"github.com/kaspanet/kaspad/util/difficulty"

	"github.com/kaspanet/kaspad/util/mstime"

	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"

	"github.com/kaspanet/kaspad/domain/consensus"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/testutils"
	"github.com/kaspanet/kaspad/domain/dagconfig"
)

func TestDifficulty(t *testing.T) {
	testutils.ForAllNets(t, true, func(t *testing.T, params *dagconfig.Params) {
		if params.DisableDifficultyAdjustment {
			return
		}
		// This test generates 3066 blocks above genesis with at least 1 second between each block, amounting to
		// a bit less then an hour of timestamps.
		// To prevent rejected blocks due to timestamps in the future, the following safeguard makes sure
		// the genesis block is at least 1 hour in the past.
		if params.GenesisBlock.Header.TimeInMilliseconds() > mstime.ToMSTime(time.Now().Add(-time.Hour)).UnixMilliseconds() {
			t.Fatalf("TestDifficulty requires the GenesisBlock to be at least 1 hour old to pass")
		}

		params.K = 1
		params.DifficultyAdjustmentWindowSize = 264

		factory := consensus.NewFactory()
		tc, teardown, err := factory.NewTestConsensus(params, false, "TestDifficulty")
		if err != nil {
			t.Fatalf("Error setting up consensus: %+v", err)
		}
		defer teardown(false)

		stagingArea := model.NewStagingArea()

		addBlock := func(blockTime int64, parents ...*externalapi.DomainHash) (*externalapi.DomainBlock, *externalapi.DomainHash) {
			bluestParent, err := tc.GHOSTDAGManager().ChooseSelectedParent(stagingArea, parents...)
			if err != nil {
				t.Fatalf("ChooseSelectedParent: %+v", err)
			}

			if blockTime == 0 {
				header, err := tc.BlockHeaderStore().BlockHeader(tc.DatabaseContext(), stagingArea, bluestParent)
				if err != nil {
					t.Fatalf("BlockHeader: %+v", err)
				}

				blockTime = header.TimeInMilliseconds() + params.TargetTimePerBlock.Milliseconds()
			}

			block, _, err := tc.BuildBlockWithParents(parents, nil, nil)
			if err != nil {
				t.Fatalf("BuildBlockWithParents: %+v", err)
			}

			newHeader := block.Header.ToMutable()
			newHeader.SetTimeInMilliseconds(blockTime)
			block.Header = newHeader.ToImmutable()
			_, err = tc.ValidateAndInsertBlock(block)
			if err != nil {
				t.Fatalf("ValidateAndInsertBlock: %+v", err)
			}

			return block, consensushashing.BlockHash(block)
		}

		minimumTime := func(parents ...*externalapi.DomainHash) int64 {
			var tempHash externalapi.DomainHash
			stagingArea := model.NewStagingArea()
			tc.BlockRelationStore().StageBlockRelation(stagingArea, &tempHash, &model.BlockRelations{
				Parents:  parents,
				Children: nil,
			})

			err = tc.GHOSTDAGManager().GHOSTDAG(stagingArea, &tempHash)
			if err != nil {
				t.Fatalf("GHOSTDAG: %+v", err)
			}

			pastMedianTime, err := tc.PastMedianTimeManager().PastMedianTime(stagingArea, &tempHash)
			if err != nil {
				t.Fatalf("PastMedianTime: %+v", err)
			}

			return pastMedianTime + 1
		}

		addBlockWithMinimumTime := func(parents ...*externalapi.DomainHash) (*externalapi.DomainBlock, *externalapi.DomainHash) {
			minTime := minimumTime(parents...)
			return addBlock(minTime, parents...)
		}

		tipHash := params.GenesisHash
		tip := params.GenesisBlock
		for i := 0; i < params.DifficultyAdjustmentWindowSize; i++ {
			tip, tipHash = addBlock(0, tipHash)
			if tip.Header.Bits() != params.GenesisBlock.Header.Bits() {
				t.Fatalf("As long as the block blue score is less then the difficulty adjustment " +
					"window size, the difficulty should be the same as genesis'")
			}
		}
		for i := 0; i < params.DifficultyAdjustmentWindowSize+100; i++ {
			tip, tipHash = addBlock(0, tipHash)
			if tip.Header.Bits() != params.GenesisBlock.Header.Bits() {
				t.Fatalf("As long as the block rate remains the same, the difficulty shouldn't change")
			}
		}

		blockInThePast, tipHash := addBlockWithMinimumTime(tipHash)
		if blockInThePast.Header.Bits() != tip.Header.Bits() {
			t.Fatalf("The difficulty should only change when blockInThePast is in the past of a block")
		}
		tip = blockInThePast

		tip, tipHash = addBlock(0, tipHash)
		if compareBits(tip.Header.Bits(), blockInThePast.Header.Bits()) >= 0 {
			t.Fatalf("tip.bits should be smaller than blockInThePast.bits because blockInThePast increased the " +
				"block rate, so the difficulty should increase as well")
		}

		var expectedBits uint32
		switch params.Name {
		case dagconfig.TestnetParams.Name, dagconfig.DevnetParams.Name:
			expectedBits = uint32(0x1e7f83df)
		case dagconfig.MainnetParams.Name:
			expectedBits = uint32(0x207f83df)
		}

		if tip.Header.Bits() != expectedBits {
			t.Errorf("tip.bits was expected to be %x but got %x", expectedBits, tip.Header.Bits())
		}

		// Increase block rate to increase difficulty
		for i := 0; i < params.DifficultyAdjustmentWindowSize; i++ {
			tip, tipHash = addBlockWithMinimumTime(tipHash)
			tipGHOSTDAGData, err := tc.GHOSTDAGDataStore().Get(tc.DatabaseContext(), stagingArea, tipHash)
			if err != nil {
				t.Fatalf("GHOSTDAGDataStore: %+v", err)
			}

			selectedParentHeader, err :=
				tc.BlockHeaderStore().BlockHeader(tc.DatabaseContext(), stagingArea, tipGHOSTDAGData.SelectedParent())
			if err != nil {
				t.Fatalf("BlockHeader: %+v", err)
			}

			if compareBits(tip.Header.Bits(), selectedParentHeader.Bits()) > 0 {
				t.Fatalf("Because we're increasing the block rate, the difficulty can't decrease")
			}
		}

		// Add blocks until difficulty stabilizes
		lastBits := tip.Header.Bits()
		sameBitsCount := 0
		for sameBitsCount < params.DifficultyAdjustmentWindowSize+1 {
			tip, tipHash = addBlock(0, tipHash)
			if tip.Header.Bits() == lastBits {
				sameBitsCount++
			} else {
				lastBits = tip.Header.Bits()
				sameBitsCount = 0
			}
		}

		slowBlockTime := tip.Header.TimeInMilliseconds() + params.TargetTimePerBlock.Milliseconds() + 1000
		slowBlock, tipHash := addBlock(slowBlockTime, tipHash)
		if slowBlock.Header.Bits() != tip.Header.Bits() {
			t.Fatalf("The difficulty should only change when slowBlock is in the past of a block")
		}

		tip = slowBlock

		tip, tipHash = addBlock(0, tipHash)
		if compareBits(tip.Header.Bits(), slowBlock.Header.Bits()) <= 0 {
			t.Fatalf("tip.bits should be smaller than slowBlock.bits because slowBlock decreased the block" +
				" rate, so the difficulty should decrease as well")
		}

		// Here we create two chains: a chain of blue blocks, and a chain of red blocks with
		// very low timestamps. Because the red blocks should be part of the difficulty
		// window, their low timestamps should lower the difficulty, and we check it by
		// comparing the bits of two blocks with the same blue score, one with the red
		// blocks in its past and one without.
		splitBlockHash := tipHash
		blueTipHash := splitBlockHash
		for i := 0; i < params.DifficultyAdjustmentWindowSize; i++ {
			_, blueTipHash = addBlock(0, blueTipHash)
		}

		redChainTipHash := splitBlockHash
		const redChainLength = 10
		for i := 0; i < redChainLength; i++ {
			_, redChainTipHash = addBlockWithMinimumTime(redChainTipHash)
		}
		tipWithRedPast, _ := addBlock(0, redChainTipHash, blueTipHash)
		tipWithoutRedPast, _ := addBlock(0, blueTipHash)
		if tipWithRedPast.Header.Bits() <= tipWithoutRedPast.Header.Bits() {
			t.Fatalf("tipWithRedPast.bits should be greater than tipWithoutRedPast.bits because the red blocks" +
				" blocks have very low timestamp and should lower the difficulty")
		}

		// We repeat the test, but now we make the blue chain longer in order to filter
		// out the red blocks from the window, and check that the red blocks don't
		// affect the difficulty.
		blueTipHash = splitBlockHash
		for i := 0; i < params.DifficultyAdjustmentWindowSize+redChainLength+1; i++ {
			_, blueTipHash = addBlock(0, blueTipHash)
		}

		redChainTipHash = splitBlockHash
		for i := 0; i < redChainLength; i++ {
			_, redChainTipHash = addBlockWithMinimumTime(redChainTipHash)
		}
		tipWithRedPast, _ = addBlock(0, redChainTipHash, blueTipHash)
		tipWithoutRedPast, _ = addBlock(0, blueTipHash)
		if tipWithRedPast.Header.Bits() != tipWithoutRedPast.Header.Bits() {
			t.Fatalf("tipWithoutRedPast.bits should be the same as tipWithRedPast.bits because the red blocks" +
				" are not part of the difficulty window")
		}
	})
}

func TestDAAScore(t *testing.T) {
	testutils.ForAllNets(t, true, func(t *testing.T, params *dagconfig.Params) {
		params.DifficultyAdjustmentWindowSize = 264

		stagingArea := model.NewStagingArea()

		factory := consensus.NewFactory()
		tc, teardown, err := factory.NewTestConsensus(params, false, "TestDifficulty")
		if err != nil {
			t.Fatalf("Error setting up consensus: %+v", err)
		}
		defer teardown(false)

		// We create a small DAG in order to skip from block with blue score of 1 directly to 3
		split1Hash, _, err := tc.AddBlock([]*externalapi.DomainHash{params.GenesisHash}, nil, nil)
		if err != nil {
			t.Fatalf("AddBlock: %+v", err)
		}
		block, _, err := tc.AddBlock([]*externalapi.DomainHash{params.GenesisHash}, nil, nil)
		if err != nil {
			t.Fatalf("AddBlock: %+v", err)
		}

		blockBlueScore3, _, err := tc.AddBlock([]*externalapi.DomainHash{split1Hash, block}, nil, nil)
		if err != nil {
			t.Fatalf("AddBlock: %+v", err)
		}

		tipHash := blockBlueScore3
		blockBlueScore3DAAScore, err := tc.DAABlocksStore().DAAScore(tc.DatabaseContext(), stagingArea, tipHash)
		if err != nil {
			t.Fatalf("DAAScore: %+v", err)
		}

		blockBlueScore3ExpectedDAAScore := uint64(2)
		if blockBlueScore3DAAScore != blockBlueScore3ExpectedDAAScore {
			t.Fatalf("DAA score is expected to be %d but got %d", blockBlueScore3ExpectedDAAScore, blockBlueScore3ExpectedDAAScore)
		}
		tipDAAScore := blockBlueScore3ExpectedDAAScore

		for i := uint64(0); i < 10; i++ {
			tipHash, _, err = tc.AddBlock([]*externalapi.DomainHash{tipHash}, nil, nil)
			if err != nil {
				t.Fatalf("AddBlock: %+v", err)
			}
			tipDAAScore, err = tc.DAABlocksStore().DAAScore(tc.DatabaseContext(), stagingArea, tipHash)
			if err != nil {
				t.Fatalf("DAAScore: %+v", err)
			}

			expectedDAAScore := blockBlueScore3ExpectedDAAScore + i + 1
			if tipDAAScore != expectedDAAScore {
				t.Fatalf("DAA score is expected to be %d but got %d", expectedDAAScore, tipDAAScore)
			}
		}

		split2Hash := tipHash
		split2DAAScore := tipDAAScore
		for i := uint64(0); i < uint64(params.DifficultyAdjustmentWindowSize)-1; i++ {
			tipHash, _, err = tc.AddBlock([]*externalapi.DomainHash{tipHash}, nil, nil)
			if err != nil {
				t.Fatalf("AddBlock: %+v", err)
			}
			tipDAAScore, err = tc.DAABlocksStore().DAAScore(tc.DatabaseContext(), stagingArea, tipHash)
			if err != nil {
				t.Fatalf("DAAScore: %+v", err)
			}

			expectedDAAScore := split2DAAScore + i + 1
			if tipDAAScore != expectedDAAScore {
				t.Fatalf("DAA score is expected to be %d but got %d", expectedDAAScore, split2DAAScore)
			}
		}

		// This block should have blue score of 2 so it shouldn't be added to the DAA window of a merging block
		blockAboveSplit1, _, err := tc.AddBlock([]*externalapi.DomainHash{split1Hash}, nil, nil)
		if err != nil {
			t.Fatalf("AddBlock: %+v", err)
		}

		// This block is in the anticone of params.DifficultyAdjustmentWindowSize-1 blocks, so it must be part
		// of the DAA window of a merging block
		blockAboveSplit2, _, err := tc.AddBlock([]*externalapi.DomainHash{split2Hash}, nil, nil)
		if err != nil {
			t.Fatalf("AddBlock: %+v", err)
		}

		currentSelectedTipDAAScore := tipDAAScore
		currentSelectedTip := tipHash
		tipHash, _, err = tc.AddBlock([]*externalapi.DomainHash{blockAboveSplit1, blockAboveSplit2, tipHash}, nil, nil)
		if err != nil {
			t.Fatalf("AddBlock: %+v", err)
		}

		tipDAAScore, err = tc.DAABlocksStore().DAAScore(tc.DatabaseContext(), stagingArea, tipHash)
		if err != nil {
			t.Fatalf("DAAScore: %+v", err)
		}

		// The DAA score should be increased only by 2, because 1 of the 3 merged blocks
		// is not in the DAA window
		expectedDAAScore := currentSelectedTipDAAScore + 2
		if tipDAAScore != expectedDAAScore {
			t.Fatalf("DAA score is expected to be %d but got %d", expectedDAAScore, tipDAAScore)
		}

		tipDAAAddedBlocks, err := tc.DAABlocksStore().DAAAddedBlocks(tc.DatabaseContext(), stagingArea, tipHash)
		if err != nil {
			t.Fatalf("DAAScore: %+v", err)
		}

		// blockAboveSplit2 should be excluded from the DAA added blocks because it's not in the tip's
		// DAA window.
		expectedDAABlocks := []*externalapi.DomainHash{blockAboveSplit2, currentSelectedTip}
		if !externalapi.HashesEqual(tipDAAAddedBlocks, expectedDAABlocks) {
			t.Fatalf("DAA added blocks are expected to be %s but got %s", expectedDAABlocks, tipDAAAddedBlocks)
		}
	})
}

func compareBits(a uint32, b uint32) int {
	aTarget := difficulty.CompactToBig(a)
	bTarget := difficulty.CompactToBig(b)
	return aTarget.Cmp(bTarget)
}
