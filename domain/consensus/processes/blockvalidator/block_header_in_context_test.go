package blockvalidator_test

import (
	"errors"
	"testing"

	"github.com/kaspanet/kaspad/domain/consensus/utils/blockheader"

	"github.com/kaspanet/kaspad/domain/consensus"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
	"github.com/kaspanet/kaspad/domain/consensus/utils/testutils"
	"github.com/kaspanet/kaspad/domain/dagconfig"
)

func TestValidateMedianTime(t *testing.T) {
	testutils.ForAllNets(t, true, func(t *testing.T, params *dagconfig.Params) {
		factory := consensus.NewFactory()
		tc, teardown, err := factory.NewTestConsensus(params, false, "TestValidateMedianTime")
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
			_, err = tc.ValidateAndInsertBlock(block)
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

		tip := params.GenesisBlock
		tipHash := params.GenesisHash

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
	testutils.ForAllNets(t, true, func(t *testing.T, params *dagconfig.Params) {
		factory := consensus.NewFactory()
		tc, teardown, err := factory.NewTestConsensus(params, false, "TestCheckParentsIncest")
		if err != nil {
			t.Fatalf("Error setting up consensus: %+v", err)
		}
		defer teardown(false)

		a, _, err := tc.AddBlock([]*externalapi.DomainHash{params.GenesisHash}, nil, nil)
		if err != nil {
			t.Fatalf("AddBlock: %+v", err)
		}

		b, _, err := tc.AddBlock([]*externalapi.DomainHash{a}, nil, nil)
		if err != nil {
			t.Fatalf("AddBlock: %+v", err)
		}

		c, _, err := tc.AddBlock([]*externalapi.DomainHash{params.GenesisHash}, nil, nil)
		if err != nil {
			t.Fatalf("AddBlock: %+v", err)
		}

		directParentsRelationBlock := &externalapi.DomainBlock{
			Header: blockheader.NewImmutableBlockHeader(
				0,
				[]*externalapi.DomainHash{a, b},
				&externalapi.DomainHash{},
				&externalapi.DomainHash{},
				&externalapi.DomainHash{},
				0,
				0,
				0,
			),
			Transactions: nil,
		}

		_, err = tc.ValidateAndInsertBlock(directParentsRelationBlock)
		if !errors.Is(err, ruleerrors.ErrInvalidParentsRelation) {
			t.Fatalf("unexpected error %+v", err)
		}

		indirectParentsRelationBlock := &externalapi.DomainBlock{
			Header: blockheader.NewImmutableBlockHeader(
				0,
				[]*externalapi.DomainHash{params.GenesisHash, b},
				&externalapi.DomainHash{},
				&externalapi.DomainHash{},
				&externalapi.DomainHash{},
				0,
				0,
				0,
			),
			Transactions: nil,
		}

		_, err = tc.ValidateAndInsertBlock(indirectParentsRelationBlock)
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
	testutils.ForAllNets(t, true, func(t *testing.T, params *dagconfig.Params) {
		params.MergeSetSizeLimit = 2 * uint64(params.K)
		factory := consensus.NewFactory()
		tc, teardown, err := factory.NewTestConsensus(params, false, "TestCheckParentsIncest")
		if err != nil {
			t.Fatalf("Error setting up consensus: %+v", err)
		}
		defer teardown(false)

		chain1TipHash := params.GenesisHash
		for i := uint64(0); i < params.MergeSetSizeLimit+2; i++ {
			chain1TipHash, _, err = tc.AddBlock([]*externalapi.DomainHash{chain1TipHash}, nil, nil)
			if err != nil {
				t.Fatalf("AddBlock: %+v", err)
			}
		}

		chain2TipHash := params.GenesisHash
		for i := uint64(0); i < params.MergeSetSizeLimit+1; i++ {
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
