package blockvalidator_test

import (
	"errors"
	"github.com/kaspanet/kaspad/domain/consensus/utils/blockheader"
	"testing"

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
		tc, teardown, err := factory.NewTestConsensus(params, "TestValidateMedianTime")
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
			var tempHash externalapi.DomainHash
			tc.BlockRelationStore().StageBlockRelation(&tempHash, &model.BlockRelations{
				Parents:  parents,
				Children: nil,
			})
			defer tc.BlockRelationStore().Discard()

			err = tc.GHOSTDAGManager().GHOSTDAG(&tempHash)
			if err != nil {
				t.Fatalf("GHOSTDAG: %+v", err)
			}
			defer tc.GHOSTDAGDataStore().Discard()

			pastMedianTime, err := tc.PastMedianTimeManager().PastMedianTime(&tempHash)
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
		tc, teardown, err := factory.NewTestConsensus(params, "TestCheckParentsIncest")
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
