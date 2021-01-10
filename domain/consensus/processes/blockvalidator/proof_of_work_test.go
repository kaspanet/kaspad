package blockvalidator_test

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/pow"
	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
	"github.com/kaspanet/kaspad/util"

	"math/rand"
	"testing"

	"github.com/kaspanet/kaspad/domain/consensus"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/testutils"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/pkg/errors"
)

// Set the flag "skip pow" to be false (second argument in the function)
func TestPOW(t *testing.T) {
	testutils.ForAllNets(t, false, func(t *testing.T, params *dagconfig.Params) {
		// First, it checks the validation of the genesis's POW.
		factory := consensus.NewFactory()
		tc, teardown, err := factory.NewTestConsensus(params, "TestPOW")
		if err != nil {
			t.Fatalf("Error setting up consensus: %+v", err)
		}
		defer teardown(false)
		// Builds and checks block with invalid POW.
		invalidBlockWrongPOW, _, err := tc.BuildBlockWithParents([]*externalapi.DomainHash{params.GenesisHash}, nil, nil)
		invalidBlockWrongPOW = solveBlockWithWrongPOW(invalidBlockWrongPOW)
		if err != nil {
			t.Fatal(err)
		}
		_, err = tc.ValidateAndInsertBlock(invalidBlockWrongPOW)
		if !errors.Is(err, ruleerrors.ErrInvalidPoW) {
			t.Fatalf("Expected block to be invalid with err: %v, instead found: %v", ruleerrors.ErrTimeTooOld, err)
		}
		// test on a valid block.
		validBlock, _, err := tc.BuildBlockWithParents([]*externalapi.DomainHash{params.GenesisHash}, nil, nil)
		validBlock = solveBlock(validBlock)
		if err != nil {
			t.Fatal(err)
		}
		_, err = tc.ValidateAndInsertBlock(validBlock)
		if err != nil {
			t.Fatal(err)
		}

	})
}

// solveBlockWithWrongPOW increments the given block's nonce until it gets wrong POW (for test!).
func solveBlockWithWrongPOW(block *externalapi.DomainBlock) *externalapi.DomainBlock {
	targetDifficulty := util.CompactToBig(block.Header.Bits())
	headerForMining := block.Header.ToMutable()
	initialNonce := rand.Uint64()
	for i := initialNonce; i != initialNonce-1; i++ {
		headerForMining.SetNonce(i)
		if !pow.CheckProofOfWorkWithTarget(headerForMining, targetDifficulty) {
			block.Header = headerForMining.ToImmutable()
			return block
		}
	}

	panic("Failed to solve block! cannot find a invalid POW for the test")
}

// solveBlock increments the given block's nonce until it gets valid POW.
func solveBlock(block *externalapi.DomainBlock) *externalapi.DomainBlock {
	targetDifficulty := util.CompactToBig(block.Header.Bits())
	headerForMining := block.Header.ToMutable()
	initialNonce := rand.Uint64()
	for i := initialNonce; i != initialNonce-1; i++ {
		headerForMining.SetNonce(i)
		if pow.CheckProofOfWorkWithTarget(headerForMining, targetDifficulty) {
			block.Header = headerForMining.ToImmutable()
			return block
		}
	}
	panic("Failed to solve block! cannot find a valid POW ")
}
