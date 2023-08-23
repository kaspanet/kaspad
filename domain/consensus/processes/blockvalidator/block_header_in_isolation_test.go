package blockvalidator_test

import (
	"reflect"
	"runtime"
	"testing"

	"github.com/c4ei/YunSeokYeol/domain/consensus/model/testapi"

	"github.com/c4ei/YunSeokYeol/domain/consensus"
	"github.com/c4ei/YunSeokYeol/domain/consensus/model/externalapi"
	"github.com/c4ei/YunSeokYeol/domain/consensus/ruleerrors"
	"github.com/c4ei/YunSeokYeol/domain/consensus/utils/blockheader"
	"github.com/c4ei/YunSeokYeol/domain/consensus/utils/constants"
	"github.com/c4ei/YunSeokYeol/domain/consensus/utils/testutils"
	"github.com/c4ei/YunSeokYeol/util/mstime"
	"github.com/pkg/errors"
)

func TestBlockValidator_ValidateHeaderInIsolation(t *testing.T) {
	tests := []func(t *testing.T, tc testapi.TestConsensus, cfg *consensus.Config){
		CheckParentsLimit,
		CheckBlockVersion,
		CheckBlockTimestampInIsolation,
	}
	testutils.ForAllNets(t, true, func(t *testing.T, consensusConfig *consensus.Config) {
		tc, teardown, err := consensus.NewFactory().NewTestConsensus(consensusConfig, "TestBlockValidator_ValidateHeaderInIsolation")
		if err != nil {
			t.Fatalf("Error setting up consensus: %+v", err)
		}
		defer teardown(false)
		for _, test := range tests {
			testName := runtime.FuncForPC(reflect.ValueOf(test).Pointer()).Name()
			t.Run(testName, func(t *testing.T) {
				test(t, tc, consensusConfig)
			})
		}
	})
}

func CheckParentsLimit(t *testing.T, tc testapi.TestConsensus, consensusConfig *consensus.Config) {
	for i := externalapi.KType(0); i < consensusConfig.MaxBlockParents+1; i++ {
		_, _, err := tc.AddBlock([]*externalapi.DomainHash{consensusConfig.GenesisHash}, nil, nil)
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
}

func CheckBlockVersion(t *testing.T, tc testapi.TestConsensus, consensusConfig *consensus.Config) {
	block, _, err := tc.BuildBlockWithParents([]*externalapi.DomainHash{consensusConfig.GenesisHash}, nil, nil)
	if err != nil {
		t.Fatalf("BuildBlockWithParents: %+v", err)
	}

	expectedVersion := constants.BlockVersion
	block.Header = blockheader.NewImmutableBlockHeader(
		expectedVersion+1,
		block.Header.Parents(),
		block.Header.HashMerkleRoot(),
		block.Header.AcceptedIDMerkleRoot(),
		block.Header.UTXOCommitment(),
		block.Header.TimeInMilliseconds(),
		block.Header.Bits(),
		block.Header.Nonce(),
		block.Header.DAAScore(),
		block.Header.BlueScore(),
		block.Header.BlueWork(),
		block.Header.PruningPoint(),
	)

	err = tc.ValidateAndInsertBlock(block, true)
	if !errors.Is(err, ruleerrors.ErrWrongBlockVersion) {
		t.Fatalf("Unexpected error: %+v", err)
	}
}

func CheckBlockTimestampInIsolation(t *testing.T, tc testapi.TestConsensus, cfg *consensus.Config) {
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
			block.Header.Parents(),
			block.Header.HashMerkleRoot(),
			block.Header.AcceptedIDMerkleRoot(),
			block.Header.UTXOCommitment(),
			timestamp,
			block.Header.Bits(),
			block.Header.Nonce(),
			block.Header.DAAScore(),
			block.Header.BlueScore(),
			block.Header.BlueWork(),
			block.Header.PruningPoint(),
		)

		err = tc.ValidateAndInsertBlock(block, true)
		if !errors.Is(err, ruleerrors.ErrTimeTooMuchInTheFuture) {
			t.Fatalf("Unexpected error: %+v", err)
		}
	})
}
