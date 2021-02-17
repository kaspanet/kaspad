package blockvalidator_test

import (
	"github.com/kaspanet/kaspad/domain/consensus"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
	"github.com/kaspanet/kaspad/domain/consensus/utils/constants"
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

		beforePruningBlockBlockStatus, err := tc.BlockStatusStore().Get(tc.DatabaseContext(),
			consensushashing.BlockHash(beforePruningBlock))
		if err != nil {
			t.Fatalf("BlockStatusStore().Get: %+v", err)
		}

		// Check that the block still has header only status although it got rejected.
		if beforePruningBlockBlockStatus != externalapi.StatusHeaderOnly {
			t.Fatalf("Unexpected status %s", beforePruningBlockBlockStatus)
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

func TestIsFinalizedTransaction(t *testing.T) {
	testutils.ForAllNets(t, true, func(t *testing.T, params *dagconfig.Params) {
		params.BlockCoinbaseMaturity = 0
		factory := consensus.NewFactory()

		tc, teardown, err := factory.NewTestConsensus(params, false, "TestIsFinalizedTransaction")
		if err != nil {
			t.Fatalf("Error setting up consensus: %+v", err)
		}
		defer teardown(false)

		block1Hash, _, err := tc.AddBlock([]*externalapi.DomainHash{params.GenesisHash}, nil, nil)
		if err != nil {
			t.Fatalf("AddBlock: %+v", err)
		}

		block1, err := tc.GetBlock(block1Hash)
		if err != nil {
			t.Fatalf("Error getting block1: %+v", err)
		}

		checkForLockTimeAndSequence := func(lockTime, sequence uint64, shouldPass bool) {
			tx, err := testutils.CreateTransaction(block1.Transactions[0])
			if err != nil {
				t.Fatalf("Error creating tx: %+v", err)
			}

			tx.LockTime = lockTime
			tx.Inputs[0].Sequence = sequence

			_, _, err = tc.AddBlock([]*externalapi.DomainHash{block1Hash}, nil, []*externalapi.DomainTransaction{tx})
			if (shouldPass && err != nil) || (!shouldPass && !errors.Is(err, ruleerrors.ErrUnfinalizedTx)) {
				t.Fatalf("Unexpected error: %+v", err)
			}
		}

		// The next block blue score is 2, so we check if we see the expected
		// behaviour when the lock time blue score is higher, lower or equal
		// to it.
		checkForLockTimeAndSequence(3, 0, false)
		checkForLockTimeAndSequence(2, 0, false)
		checkForLockTimeAndSequence(1, 0, true)

		pastMedianTime, err := tc.PastMedianTimeManager().PastMedianTime(model.VirtualBlockHash)
		if err != nil {
			t.Fatalf("PastMedianTime: %+v", err)
		}
		checkForLockTimeAndSequence(uint64(pastMedianTime)+1, 0, false)
		checkForLockTimeAndSequence(uint64(pastMedianTime), 0, false)
		checkForLockTimeAndSequence(uint64(pastMedianTime)-1, 0, true)

		// We check that if the transaction is marked as finalized it'll pass for any lock time.
		checkForLockTimeAndSequence(uint64(pastMedianTime), constants.MaxTxInSequenceNum, true)
		checkForLockTimeAndSequence(2, constants.MaxTxInSequenceNum, true)
	})
}
