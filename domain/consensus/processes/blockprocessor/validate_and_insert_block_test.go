package blockprocessor_test

import (
	"strings"
	"testing"

	"github.com/kaspanet/kaspad/domain/consensus"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
	"github.com/kaspanet/kaspad/domain/consensus/utils/blockheader"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
	"github.com/kaspanet/kaspad/domain/consensus/utils/constants"
	"github.com/kaspanet/kaspad/domain/consensus/utils/merkle"
	"github.com/kaspanet/kaspad/domain/consensus/utils/testutils"
	"github.com/pkg/errors"
)

func TestBlockStatus(t *testing.T) {
	testutils.ForAllNets(t, true, func(t *testing.T, consensusConfig *consensus.Config) {
		factory := consensus.NewFactory()
		tc, teardown, err := factory.NewTestConsensus(consensusConfig, "TestBlockStatus")
		if err != nil {
			t.Fatalf("Error setting up consensus: %+v", err)
		}
		defer teardown(false)

		checkStatus := func(hash *externalapi.DomainHash, expectedStatus externalapi.BlockStatus) {
			blockStatus, err := tc.BlockStatusStore().Get(tc.DatabaseContext(), model.NewStagingArea(), hash)
			if err != nil {
				t.Fatalf("BlockStatusStore().Get: %+v", err)
			}

			if blockStatus != expectedStatus {
				t.Fatalf("Expected to have status %s but got %s", expectedStatus, blockStatus)
			}
		}

		tipHash := consensusConfig.GenesisHash
		for i := 0; i < 2; i++ {
			tipHash, _, err = tc.AddBlock([]*externalapi.DomainHash{tipHash}, nil, nil)
			if err != nil {
				t.Fatalf("AddBlock: %+v", err)
			}

			checkStatus(tipHash, externalapi.StatusUTXOValid)
		}

		headerHash, _, err := tc.AddUTXOInvalidHeader([]*externalapi.DomainHash{tipHash})
		if err != nil {
			t.Fatalf("AddUTXOInvalidHeader: %+v", err)
		}

		checkStatus(headerHash, externalapi.StatusHeaderOnly)

		nonChainBlockHash, _, err := tc.AddBlock([]*externalapi.DomainHash{consensusConfig.GenesisHash}, nil, nil)
		if err != nil {
			t.Fatalf("AddBlock: %+v", err)
		}

		checkStatus(nonChainBlockHash, externalapi.StatusUTXOPendingVerification)

		disqualifiedBlock, _, err := tc.BuildBlockWithParents([]*externalapi.DomainHash{tipHash}, nil, nil)
		if err != nil {
			t.Fatalf("AddBlock: %+v", err)
		}
		disqualifiedBlock.Header = blockheader.NewImmutableBlockHeader(
			disqualifiedBlock.Header.Version(),
			disqualifiedBlock.Header.ParentHashes(),
			disqualifiedBlock.Header.HashMerkleRoot(),
			externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{}), // This should disqualify the block
			disqualifiedBlock.Header.UTXOCommitment(),
			disqualifiedBlock.Header.TimeInMilliseconds(),
			disqualifiedBlock.Header.Bits(),
			disqualifiedBlock.Header.Nonce(),
			disqualifiedBlock.Header.DAAScore(),
			disqualifiedBlock.Header.BlueWork(),
			disqualifiedBlock.Header.FinalityPoint(),
		)

		_, err = tc.ValidateAndInsertBlock(disqualifiedBlock, true)
		if err != nil {
			t.Fatalf("ValidateAndInsertBlock: %+v", err)
		}

		checkStatus(consensushashing.BlockHash(disqualifiedBlock), externalapi.StatusDisqualifiedFromChain)

		invalidBlock, _, err := tc.BuildBlockWithParents([]*externalapi.DomainHash{tipHash}, nil, nil)
		if err != nil {
			t.Fatalf("AddBlock: %+v", err)
		}
		invalidBlock.Transactions[0].Version = constants.MaxTransactionVersion + 1 // This should invalidate the block
		invalidBlock.Header = blockheader.NewImmutableBlockHeader(
			disqualifiedBlock.Header.Version(),
			disqualifiedBlock.Header.ParentHashes(),
			merkle.CalculateHashMerkleRoot(invalidBlock.Transactions),
			disqualifiedBlock.Header.AcceptedIDMerkleRoot(),
			disqualifiedBlock.Header.UTXOCommitment(),
			disqualifiedBlock.Header.TimeInMilliseconds(),
			disqualifiedBlock.Header.Bits(),
			disqualifiedBlock.Header.Nonce(),
			disqualifiedBlock.Header.DAAScore(),
			disqualifiedBlock.Header.BlueWork(),
			disqualifiedBlock.Header.FinalityPoint(),
		)

		_, err = tc.ValidateAndInsertBlock(invalidBlock, true)
		if err == nil {
			t.Fatalf("block is expected to be invalid")
		}
		if !errors.As(err, &ruleerrors.RuleError{}) {
			t.Fatalf("ValidateAndInsertBlock: %+v", err)
		}

		checkStatus(consensushashing.BlockHash(invalidBlock), externalapi.StatusInvalid)
	})
}

// TestValidateAndInsertErrors tests the error cases on "validateBlock" function.
func TestValidateAndInsertErrors(t *testing.T) {
	// All the tests below tests the error cases in "validateAndInsertBlock" function.
	// Each test is covering the error cases in a sub-function in "validateAndInsertBlock" function.
	// Currently, implemented only for some of the errors.

	testutils.ForAllNets(t, true, func(t *testing.T, consensusConfig *consensus.Config) {
		factory := consensus.NewFactory()
		tc, teardown, err := factory.NewTestConsensus(consensusConfig, "TestBlockStatus")
		if err != nil {
			t.Fatalf("Error setting up consensus: %+v", err)
		}
		defer teardown(false)
		tipHash, emptyCoinbase, tx1 := initData(consensusConfig)

		// Tests all the error case on the function: "checkBlockStatus"(sub-function in function validateBlock)
		blockWithStatusInvalid, _, err := tc.BuildBlockWithParents([]*externalapi.DomainHash{consensusConfig.GenesisHash},
			&emptyCoinbase, []*externalapi.DomainTransaction{tx1, tx1})
		if err != nil {
			t.Fatalf("AddBlock: %+v", err)
		}
		_, err = tc.ValidateAndInsertBlock(blockWithStatusInvalid, true)
		if err == nil {
			t.Fatalf("Test ValidateAndInsertBlock: Expected an error, because the block is invalid.")
		}
		_, err = tc.ValidateAndInsertBlock(blockWithStatusInvalid, true)
		if err == nil || !errors.Is(err, ruleerrors.ErrKnownInvalid) {
			t.Fatalf("Expected block to be invalid with err: %v, instead found: %v", ruleerrors.ErrKnownInvalid, err)
		}
		if !strings.Contains(err.Error(), "is a known invalid block") {
			t.Fatalf("Test ValidateAndInsertBlock: Expected an diff error, got: %+v.", err)
		}

		block, _, err := tc.BuildBlockWithParents([]*externalapi.DomainHash{tipHash}, nil, nil)
		if err != nil {
			t.Fatalf("AddBlock: %+v", err)
		}
		_, err = tc.ValidateAndInsertBlock(block, true)
		if err != nil {
			t.Fatalf("ValidateAndInsertBlock: %+v", err)
		}
		// resend the same block.
		_, err = tc.ValidateAndInsertBlock(block, true)
		if err == nil || !errors.Is(err, ruleerrors.ErrDuplicateBlock) {
			t.Fatalf("Expected block to be invalid with err: %v, instead found: %v", ruleerrors.ErrDuplicateBlock, err)
		}
		if !strings.Contains(err.Error(), " already exists") {
			t.Fatalf("Test ValidateAndInsertBlock: Expected an diff error, got: %+v.", err)
		}

		onlyHeader, err := tc.BuildBlock(&emptyCoinbase, []*externalapi.DomainTransaction{})
		if err != nil {
			t.Fatalf("AddBlock: %+v", err)
		}
		onlyHeader.Transactions = []*externalapi.DomainTransaction{}
		_, err = tc.ValidateAndInsertBlock(onlyHeader, true)
		if err != nil {
			t.Fatalf("AddBlock: %+v", err)
		}
		// resend the same header.
		_, err = tc.ValidateAndInsertBlock(onlyHeader, true)
		if err == nil || !errors.Is(err, ruleerrors.ErrDuplicateBlock) {
			t.Fatalf("Expected block to be invalid with err: %v, instead found: %v", ruleerrors.ErrDuplicateBlock, err)
		}
		if !strings.Contains(err.Error(), "header already exists") {
			t.Fatalf("Test ValidateAndInsertBlock: Expected an diff error, got: %+v.", err)
		}

	})
}

func initData(consensusConfig *consensus.Config) (*externalapi.DomainHash, externalapi.DomainCoinbaseData, *externalapi.DomainTransaction) {
	return consensusConfig.GenesisHash,
		externalapi.DomainCoinbaseData{
			ScriptPublicKey: &externalapi.ScriptPublicKey{
				Script:  nil,
				Version: 0,
			},
		},

		&externalapi.DomainTransaction{
			Version: 0,
			Inputs:  []*externalapi.DomainTransactionInput{},
			Outputs: []*externalapi.DomainTransactionOutput{{uint64(0xFFFF),
				&externalapi.ScriptPublicKey{Script: []byte{1, 2}, Version: 0}}, {uint64(0xFFFF),
				&externalapi.ScriptPublicKey{Script: []byte{1, 3}, Version: 0}}},
			LockTime:     1,
			SubnetworkID: externalapi.DomainSubnetworkID{0x01},
			Gas:          1,
			Payload:      []byte{0x01},
			Fee:          0,
			Mass:         1,
			ID: externalapi.NewDomainTransactionIDFromByteArray(&[externalapi.DomainHashSize]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02}),
		}
}
