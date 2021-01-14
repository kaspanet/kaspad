package blockvalidator_test

import (
	"github.com/kaspanet/kaspad/domain/consensus"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
	"github.com/kaspanet/kaspad/domain/consensus/utils/testutils"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/pkg/errors"
	"testing"
)

func TestCheckBlockIsNotPruned(t *testing.T) {
	testutils.ForAllNets(t, true, func(t *testing.T, params *dagconfig.Params) {
		// This is done to reduce the pruning depth to 6 blocks
		params.FinalityDuration = 2 * params.TargetTimePerBlock
		params.K = 0

		factory := consensus.NewFactory()

		tc, teardown, err := factory.NewTestConsensus(params, false, "TestCheckBlockIsNotPruned")
		if err != nil {
			t.Fatalf("Error setting up consensus: %+v", err)
		}
		defer teardown(false)

		// Add blocks until the pruning point changes
		tipHash := params.GenesisHash
		tipHash, _, err = tc.AddBlock([]*externalapi.DomainHash{tipHash}, nil, nil)
		if err != nil {
			t.Fatalf("AddBlock: %+v", err)
		}

		beforePruningBlock, err := tc.GetBlock(tipHash)
		if err != nil {
			t.Fatalf("beforePruningBlock: %+v", err)
		}

		for {
			tipHash, _, err = tc.AddBlock([]*externalapi.DomainHash{tipHash}, nil, nil)
			if err != nil {
				t.Fatalf("AddBlock: %+v", err)
			}

			pruningPoint, err := tc.PruningPoint()
			if err != nil {
				t.Fatalf("PruningPoint: %+v", err)
			}

			if !pruningPoint.Equal(params.GenesisHash) {
				break
			}
		}

		_, err = tc.ValidateAndInsertBlock(beforePruningBlock)
		if !errors.Is(err, ruleerrors.ErrPrunedBlock) {
			t.Fatalf("Unexpected error: %+v", err)
		}
	})
}

func TestCheckParentBlockBodiesExist(t *testing.T) {
	testutils.ForAllNets(t, true, func(t *testing.T, params *dagconfig.Params) {
		// This is done to reduce the pruning depth to 6 blocks
		params.FinalityDuration = 2 * params.TargetTimePerBlock
		params.K = 0

		factory := consensus.NewFactory()

		tc, teardown, err := factory.NewTestConsensus(params, false, "TestCheckParentBlockBodiesExist")
		if err != nil {
			t.Fatalf("Error setting up consensus: %+v", err)
		}
		defer teardown(false)

		headerHash, _, err := tc.AddUTXOInvalidHeader([]*externalapi.DomainHash{params.GenesisHash})
		if err != nil {
			t.Fatalf("AddUTXOInvalidHeader: %+v", err)
		}

		_, _, err = tc.AddUTXOInvalidBlock([]*externalapi.DomainHash{headerHash})
		errMissingParents := &ruleerrors.ErrMissingParents{}
		if !errors.As(err, errMissingParents) {
			t.Fatalf("Unexpected error: %+v", err)
		}

		if !externalapi.HashesEqual(errMissingParents.MissingParentHashes, []*externalapi.DomainHash{headerHash}) {
			t.Fatalf("unexpected missing parents %s", errMissingParents.MissingParentHashes)
		}

		// Add blocks until the pruning point changes
		tipHash := params.GenesisHash
		anticonePruningBlock, err := tc.BuildUTXOInvalidBlock([]*externalapi.DomainHash{tipHash})
		if err != nil {
			t.Fatalf("BuildUTXOInvalidBlock: %+v", err)
		}

		// Add only the header of anticonePruningBlock
		_, err = tc.ValidateAndInsertBlock(&externalapi.DomainBlock{
			Header:       anticonePruningBlock.Header,
			Transactions: nil,
		})
		if err != nil {
			t.Fatalf("ValidateAndInsertBlock: %+v", err)
		}

		for {
			tipHash, _, err = tc.AddBlock([]*externalapi.DomainHash{tipHash}, nil, nil)
			if err != nil {
				t.Fatalf("AddUTXOInvalidHeader: %+v", err)
			}

			pruningPoint, err := tc.PruningPoint()
			if err != nil {
				t.Fatalf("PruningPoint: %+v", err)
			}

			if !pruningPoint.Equal(params.GenesisHash) {
				break
			}
		}

		// Add anticonePruningBlock's body and Check that it's valid to point to
		// a header only block in the past of the pruning point.
		_, err = tc.ValidateAndInsertBlock(anticonePruningBlock)
		if err != nil {
			t.Fatalf("ValidateAndInsertBlock: %+v", err)
		}
	})
}
