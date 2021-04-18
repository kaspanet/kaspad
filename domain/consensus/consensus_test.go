package consensus_test

import (
	"testing"

	"github.com/kaspanet/kaspad/domain/consensus"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
	"github.com/kaspanet/kaspad/domain/consensus/utils/testutils"
	"github.com/pkg/errors"
)

func TestConsensus_GetBlockInfo(t *testing.T) {
	testutils.ForAllNets(t, true, func(t *testing.T, consensusConfig *consensus.Config) {
		factory := consensus.NewFactory()
		consensus, teardown, err := factory.NewTestConsensus(consensusConfig, "TestConsensus_GetBlockInfo")
		if err != nil {
			t.Fatalf("Error setting up consensus: %+v", err)
		}
		defer teardown(false)

		invalidBlock, _, err := consensus.BuildBlockWithParents([]*externalapi.DomainHash{consensusConfig.GenesisHash}, nil, nil)
		if err != nil {
			t.Fatal(err)
		}
		newHeader := invalidBlock.Header.ToMutable()
		newHeader.SetTimeInMilliseconds(0)
		invalidBlock.Header = newHeader.ToImmutable()
		_, err = consensus.ValidateAndInsertBlock(invalidBlock)
		if !errors.Is(err, ruleerrors.ErrTimeTooOld) {
			t.Fatalf("Expected block to be invalid with err: %v, instead found: %v", ruleerrors.ErrTimeTooOld, err)
		}

		info, err := consensus.GetBlockInfo(consensushashing.BlockHash(invalidBlock))
		if err != nil {
			t.Fatalf("Failed to get block info: %v", err)
		}

		if !info.Exists {
			t.Fatal("The block is missing")
		}
		if info.BlockStatus != externalapi.StatusInvalid {
			t.Fatalf("Expected block status: %s, instead got: %s", externalapi.StatusInvalid, info.BlockStatus)
		}

		emptyCoinbase := externalapi.DomainCoinbaseData{
			ScriptPublicKey: &externalapi.ScriptPublicKey{
				Script:  nil,
				Version: 0,
			},
		}
		validBlock, err := consensus.BuildBlock(&emptyCoinbase, nil)
		if err != nil {
			t.Fatalf("consensus.BuildBlock with an empty coinbase shouldn't fail: %v", err)
		}

		_, err = consensus.ValidateAndInsertBlock(validBlock)
		if err != nil {
			t.Fatalf("consensus.ValidateAndInsertBlock with a block straight from consensus.BuildBlock should not fail: %v", err)
		}

		info, err = consensus.GetBlockInfo(consensushashing.BlockHash(validBlock))
		if err != nil {
			t.Fatalf("Failed to get block info: %v", err)
		}

		if !info.Exists {
			t.Fatal("The block is missing")
		}
		if info.BlockStatus != externalapi.StatusUTXOValid {
			t.Fatalf("Expected block status: %s, instead got: %s", externalapi.StatusUTXOValid, info.BlockStatus)
		}

	})
}
