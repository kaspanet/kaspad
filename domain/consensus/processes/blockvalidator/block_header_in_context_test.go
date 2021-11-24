package blockvalidator_test

import (
	"errors"
	"math/big"
	"testing"

	"github.com/kaspanet/kaspad/domain/consensus/utils/blockheader"

	"github.com/kaspanet/kaspad/domain/consensus"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
	"github.com/kaspanet/kaspad/domain/consensus/utils/testutils"
)

func TestValidateMedianTime(t *testing.T) {
	testutils.ForAllNets(t, true, func(t *testing.T, consensusConfig *consensus.Config) {
		factory := consensus.NewFactory()
		tc, teardown, err := factory.NewTestConsensus(consensusConfig, "TestValidateMedianTime")
		if err != nil {
			t.Fatalf("Error setting up consensus: %+v", err)
		}
		defer teardown(false)

		addBlock := func(blockTime int64, parents []*externalapi.DomainHash, expectedErr error) (*externalapi.DomainBlock, *externalapi.DomainHash) {
			block, _, err := tc.BuildBlockWithParents(parents, nil, nil)
			if err != nil {
				t.Fatalf("BuildBlockWithParents: %+v", err)
			}

			newHeader := block.Header.ToMutable()
			newHeader.SetTimeInMilliseconds(blockTime)
			block.Header = newHeader.ToImmutable()
			_, err = tc.ValidateAndInsertBlock(block, true)
			if !errors.Is(err, expectedErr) {
				t.Fatalf("expected error %s but got %+v", expectedErr, err)
			}

			return block, consensushashing.BlockHash(block)
		}

		pastMedianTime := func(parents ...*externalapi.DomainHash) int64 {
			stagingArea := model.NewStagingArea()
			var tempHash externalapi.DomainHash
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

			return pastMedianTime
		}

		tip := consensusConfig.GenesisBlock
		tipHash := consensusConfig.GenesisHash

		blockTime := tip.Header.TimeInMilliseconds()

		for i := 0; i < 100; i++ {
			blockTime += 1000
			_, tipHash = addBlock(blockTime, []*externalapi.DomainHash{tipHash}, nil)
		}

		// Checks that a block is invalid if it has timestamp equals to past median time
		addBlock(pastMedianTime(tipHash), []*externalapi.DomainHash{tipHash}, ruleerrors.ErrTimeTooOld)

		// Checks that a block is valid if its timestamp is after past median time
		addBlock(pastMedianTime(tipHash)+1, []*externalapi.DomainHash{tipHash}, nil)

		// Checks that a block is invalid if its timestamp is before past median time
		addBlock(pastMedianTime(tipHash)-1, []*externalapi.DomainHash{tipHash}, ruleerrors.ErrTimeTooOld)
	})
}

func TestCheckParentsIncest(t *testing.T) {
	testutils.ForAllNets(t, true, func(t *testing.T, consensusConfig *consensus.Config) {
		factory := consensus.NewFactory()
		tc, teardown, err := factory.NewTestConsensus(consensusConfig, "TestCheckParentsIncest")
		if err != nil {
			t.Fatalf("Error setting up consensus: %+v", err)
		}
		defer teardown(false)

		a, _, err := tc.AddBlock([]*externalapi.DomainHash{consensusConfig.GenesisHash}, nil, nil)
		if err != nil {
			t.Fatalf("AddBlock: %+v", err)
		}

		b, _, err := tc.AddBlock([]*externalapi.DomainHash{a}, nil, nil)
		if err != nil {
			t.Fatalf("AddBlock: %+v", err)
		}

		c, _, err := tc.AddBlock([]*externalapi.DomainHash{consensusConfig.GenesisHash}, nil, nil)
		if err != nil {
			t.Fatalf("AddBlock: %+v", err)
		}

		directParentsRelationBlock := &externalapi.DomainBlock{
			Header: blockheader.NewImmutableBlockHeader(
				0,
				[]externalapi.BlockLevelParents{[]*externalapi.DomainHash{a, b}},
				&externalapi.DomainHash{},
				&externalapi.DomainHash{},
				&externalapi.DomainHash{},
				0,
				0,
				0,
				0,
				0,
				big.NewInt(0),
				&externalapi.DomainHash{},
			),
			Transactions: nil,
		}

		_, err = tc.ValidateAndInsertBlock(directParentsRelationBlock, true)
		if !errors.Is(err, ruleerrors.ErrInvalidParentsRelation) {
			t.Fatalf("unexpected error %+v", err)
		}

		indirectParentsRelationBlock := &externalapi.DomainBlock{
			Header: blockheader.NewImmutableBlockHeader(
				0,
				[]externalapi.BlockLevelParents{[]*externalapi.DomainHash{consensusConfig.GenesisHash, b}},
				&externalapi.DomainHash{},
				&externalapi.DomainHash{},
				&externalapi.DomainHash{},
				0,
				0,
				0,
				0,
				0,
				big.NewInt(0),
				&externalapi.DomainHash{},
			),
			Transactions: nil,
		}

		_, err = tc.ValidateAndInsertBlock(indirectParentsRelationBlock, true)
		if !errors.Is(err, ruleerrors.ErrInvalidParentsRelation) {
			t.Fatalf("unexpected error %+v", err)
		}

		// Try to add block with unrelated parents
		_, _, err = tc.AddBlock([]*externalapi.DomainHash{b, c}, nil, nil)
		if err != nil {
			t.Fatalf("AddBlock: %s", err)
		}
	})
}

func TestCheckMergeSizeLimit(t *testing.T) {
	testutils.ForAllNets(t, true, func(t *testing.T, consensusConfig *consensus.Config) {
		consensusConfig.MergeSetSizeLimit = 2 * uint64(consensusConfig.K)
		factory := consensus.NewFactory()
		tc, teardown, err := factory.NewTestConsensus(consensusConfig, "TestCheckMergeSizeLimit")
		if err != nil {
			t.Fatalf("Error setting up consensus: %+v", err)
		}
		defer teardown(false)

		chain1TipHash := consensusConfig.GenesisHash
		// We add a chain larger by one than chain2 below, to make this one the selected chain
		for i := uint64(0); i < consensusConfig.MergeSetSizeLimit+1; i++ {
			chain1TipHash, _, err = tc.AddBlock([]*externalapi.DomainHash{chain1TipHash}, nil, nil)
			if err != nil {
				t.Fatalf("AddBlock: %+v", err)
			}
		}

		chain2TipHash := consensusConfig.GenesisHash
		// We add a merge set of size exactly MergeSetSizeLimit (to violate the limit),
		// since selected parent is also counted
		for i := uint64(0); i < consensusConfig.MergeSetSizeLimit; i++ {
			chain2TipHash, _, err = tc.AddBlock([]*externalapi.DomainHash{chain2TipHash}, nil, nil)
			if err != nil {
				t.Fatalf("AddBlock: %+v", err)
			}
		}

		_, _, err = tc.AddBlock([]*externalapi.DomainHash{chain1TipHash, chain2TipHash}, nil, nil)
		if !errors.Is(err, ruleerrors.ErrViolatingMergeLimit) {
			t.Fatalf("unexpected error: %+v", err)
		}
	})
}

func TestVirtualSelectionViolatingMergeSizeLimit(t *testing.T) {
	testutils.ForAllNets(t, true, func(t *testing.T, consensusConfig *consensus.Config) {
		consensusConfig.MergeSetSizeLimit = 2 * uint64(consensusConfig.K)
		factory := consensus.NewFactory()
		tc, teardown, err := factory.NewTestConsensus(consensusConfig, "TestVirtualSelectionViolatingMergeSizeLimit")
		if err != nil {
			t.Fatalf("Error setting up consensus: %+v", err)
		}
		defer teardown(false)

		chain1TipHash := consensusConfig.GenesisHash
		// We add a chain larger than chain2 below, to make this one the selected chain
		for i := uint64(0); i < consensusConfig.MergeSetSizeLimit; i++ {
			chain1TipHash, _, err = tc.AddBlock([]*externalapi.DomainHash{chain1TipHash}, nil, nil)
			if err != nil {
				t.Fatalf("AddBlock: %+v", err)
			}
		}

		chain2TipHash := consensusConfig.GenesisHash
		// We add a merge set of size exactly MergeSetSizeLimit-1 (to still not violate the limit)
		for i := uint64(0); i < consensusConfig.MergeSetSizeLimit-1; i++ {
			chain2TipHash, _, err = tc.AddBlock([]*externalapi.DomainHash{chain2TipHash}, nil, nil)
			if err != nil {
				t.Fatalf("AddBlock: %+v", err)
			}
		}

		// We now add a single block over genesis which is expected to exceed the limit
		_, _, err = tc.AddBlock([]*externalapi.DomainHash{consensusConfig.GenesisHash}, nil, nil)
		if err != nil {
			t.Fatalf("AddBlock: %+v", err)
		}

		stagingArea := model.NewStagingArea()
		virtualSelectedParent, err := tc.GetVirtualSelectedParent()
		if err != nil {
			t.Fatalf("GetVirtualSelectedParent: %+v", err)
		}
		selectedParentAnticone, err := tc.DAGTraversalManager().AnticoneFromVirtualPOV(stagingArea, virtualSelectedParent)
		if err != nil {
			t.Fatalf("AnticoneFromVirtualPOV: %+v", err)
		}

		// Test if Virtual's mergeset is too large
		// Note: the selected parent itself is also counted in the mergeset limit
		if len(selectedParentAnticone)+1 > (int)(consensusConfig.MergeSetSizeLimit) {
			t.Fatalf("Virtual's mergset size (%d) exeeds merge set limit (%d)",
				len(selectedParentAnticone)+1, consensusConfig.MergeSetSizeLimit)
		}
	})
}
