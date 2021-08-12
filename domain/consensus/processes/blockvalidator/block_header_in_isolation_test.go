package blockvalidator_test

import (
	"testing"

	"github.com/kaspanet/kaspad/domain/consensus"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
	"github.com/kaspanet/kaspad/domain/consensus/utils/blockheader"
	"github.com/kaspanet/kaspad/domain/consensus/utils/constants"
	"github.com/kaspanet/kaspad/domain/consensus/utils/testutils"
	"github.com/kaspanet/kaspad/util/mstime"
	"github.com/pkg/errors"
)

func TestCheckParentsLimit(t *testing.T) {
	testutils.ForAllNets(t, true, func(t *testing.T, consensusConfig *consensus.Config) {
		factory := consensus.NewFactory()

		tc, teardown, err := factory.NewTestConsensus(consensusConfig, "TestCheckParentsLimit")
		if err != nil {
			t.Fatalf("Error setting up consensus: %+v", err)
		}
		defer teardown(false)

		for i := externalapi.KType(0); i < consensusConfig.MaxBlockParents+1; i++ {
			_, _, err = tc.AddBlock([]*externalapi.DomainHash{consensusConfig.GenesisHash}, nil, nil)
			if err != nil {
				t.Fatalf("AddBlock: %+v", err)
			}
		}

		tips, err := tc.Tips()
		if err != nil {
			t.Fatalf("Tips: %+v", err)
		}

		_, _, err = tc.AddBlock(tips, nil, nil)
		if !errors.Is(err, ruleerrors.ErrTooManyParents) {
			t.Fatalf("Unexpected error: %+v", err)
		}
	})
}

func TestCheckBlockVersion(t *testing.T) {
	testutils.ForAllNets(t, true, func(t *testing.T, consensusConfig *consensus.Config) {
		factory := consensus.NewFactory()

		tc, teardown, err := factory.NewTestConsensus(consensusConfig, "TestCheckBlockVersion")
		if err != nil {
			t.Fatalf("Error setting up consensus: %+v", err)
		}
		defer teardown(false)

		block, _, err := tc.BuildBlockWithParents([]*externalapi.DomainHash{consensusConfig.GenesisHash}, nil, nil)
		if err != nil {
			t.Fatalf("BuildBlockWithParents: %+v", err)
		}

		block.Header = blockheader.NewImmutableBlockHeader(
			constants.MaxBlockVersion+1,
			block.Header.ParentHashes(),
			block.Header.HashMerkleRoot(),
			block.Header.AcceptedIDMerkleRoot(),
			block.Header.UTXOCommitment(),
			block.Header.TimeInMilliseconds(),
			block.Header.Bits(),
			block.Header.Nonce(),
			block.Header.DAAScore(),
			block.Header.BlueWork(),
			block.Header.FinalityPoint(),
		)

		_, err = tc.ValidateAndInsertBlock(block, true)
		if !errors.Is(err, ruleerrors.ErrBlockVersionIsUnknown) {
			t.Fatalf("Unexpected error: %+v", err)
		}
	})
}

func TestCheckBlockTimestampInIsolation(t *testing.T) {
	testutils.ForAllNets(t, true, func(t *testing.T, consensusConfig *consensus.Config) {
		factory := consensus.NewFactory()

		tc, teardown, err := factory.NewTestConsensus(consensusConfig, "TestCheckBlockTimestampInIsolation")
		if err != nil {
			t.Fatalf("Error setting up consensus: %+v", err)
		}
		defer teardown(false)

		block, _, err := tc.BuildBlockWithParents([]*externalapi.DomainHash{consensusConfig.GenesisHash}, nil, nil)
		if err != nil {
			t.Fatalf("BuildBlockWithParents: %+v", err)
		}

		// Give 10 seconds slack to take care of the test duration
		timestamp := mstime.Now().UnixMilliseconds() +
			int64(consensusConfig.TimestampDeviationTolerance)*consensusConfig.TargetTimePerBlock.Milliseconds() + 10_000

		block.Header = blockheader.NewImmutableBlockHeader(
			block.Header.Version(),
			block.Header.ParentHashes(),
			block.Header.HashMerkleRoot(),
			block.Header.AcceptedIDMerkleRoot(),
			block.Header.UTXOCommitment(),
			timestamp,
			block.Header.Bits(),
			block.Header.Nonce(),
			block.Header.DAAScore(),
			block.Header.BlueWork(),
			block.Header.FinalityPoint(),
		)

		_, err = tc.ValidateAndInsertBlock(block, true)
		if !errors.Is(err, ruleerrors.ErrTimeTooMuchInTheFuture) {
			t.Fatalf("Unexpected error: %+v", err)
		}
	})
}
