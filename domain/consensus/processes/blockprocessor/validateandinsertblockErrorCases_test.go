package blockprocessor_test

import (
	"github.com/kaspanet/kaspad/domain/consensus"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
	"github.com/kaspanet/kaspad/domain/consensus/utils/testutils"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/pkg/errors"
	"strings"
	"testing"
)

// All the tests below are tests the error cases in "validateAndInsertBlock" function.
// Each test is coverage the error cases in sub-function in "validateAndInsertBlock" function.
// Currently, implement only for some of the errors.

// TestErrorCasesOnValidateBlock tests the error cases on "validateBlock" function.
func TestErrorCasesOnValidateBlock(t *testing.T) {
	testutils.ForAllNets(t, true, func(t *testing.T, params *dagconfig.Params) {
		factory := consensus.NewFactory()
		tc, teardown, err := factory.NewTestConsensus(params, false, "TestBlockStatus")
		if err != nil {
			t.Fatalf("Error setting up consensus: %+v", err)
		}
		defer teardown(false)
		tipHash, emptyCoinbase, tx1 := initData(params)

		// Tests all the error case on the function: "checkBlockStatus"(sub-function in function validateBlock)
		blockWithStatusInvalid, err := tc.BuildBlock(&emptyCoinbase, []*externalapi.DomainTransaction{tx1, tx1})
		if err != nil {
			t.Fatalf("AddBlock: %+v", err)
		}
		_, err = tc.ValidateAndInsertBlock(blockWithStatusInvalid)
		if err == nil {
			t.Fatalf("Test ValidateAndInsertBlock: Expected an error, because the block is invalid.")
		}
		_, err = tc.ValidateAndInsertBlock(blockWithStatusInvalid)
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
		_, err = tc.ValidateAndInsertBlock(block)
		if err != nil {
			t.Fatalf("ValidateAndInsertBlock: %+v", err)
		}
		// resend the same block.
		_, err = tc.ValidateAndInsertBlock(block)
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
		_, err = tc.ValidateAndInsertBlock(onlyHeader)
		if err != nil {
			t.Fatalf("AddBlock: %+v", err)
		}
		// resend the same header.
		_, err = tc.ValidateAndInsertBlock(onlyHeader)
		if err == nil || !errors.Is(err, ruleerrors.ErrDuplicateBlock) {
			t.Fatalf("Expected block to be invalid with err: %v, instead found: %v", ruleerrors.ErrDuplicateBlock, err)
		}
		if !strings.Contains(err.Error(), "header already exists") {
			t.Fatalf("Test ValidateAndInsertBlock: Expected an diff error, got: %+v.", err)
		}

	})
}

func initData(params *dagconfig.Params) (*externalapi.DomainHash, externalapi.DomainCoinbaseData, *externalapi.DomainTransaction) {
	return params.GenesisHash,
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
			PayloadHash: *externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}),
			Payload: []byte{0x01},
			Fee:     0,
			Mass:    1,
			ID: externalapi.NewDomainTransactionIDFromByteArray(&[externalapi.DomainHashSize]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02}),
		}
}
