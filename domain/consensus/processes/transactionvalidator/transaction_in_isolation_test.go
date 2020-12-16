package transactionvalidator_test

import (
	"testing"

	"github.com/kaspanet/kaspad/domain/consensus"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
	"github.com/kaspanet/kaspad/domain/consensus/utils/constants"
	"github.com/kaspanet/kaspad/domain/consensus/utils/hashes"
	"github.com/kaspanet/kaspad/domain/consensus/utils/subnetworks"
	"github.com/kaspanet/kaspad/domain/consensus/utils/testutils"
	"github.com/kaspanet/kaspad/domain/consensus/utils/transactionhelper"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/kaspanet/kaspad/util"
	"github.com/pkg/errors"
)

type txSubnetworkData struct {
	subnetworkID externalapi.DomainSubnetworkID
	gas          uint64
	payload      []byte
}

func TestValidateTransactionInIsolation(t *testing.T) {
	testutils.ForAllNets(t, true, func(t *testing.T, params *dagconfig.Params) {
		factory := consensus.NewFactory()
		tc, teardown, err := factory.NewTestConsensus(params, "TestValidateTransactionInIsolation")
		if err != nil {
			t.Fatalf("Error setting up consensus: %+v", err)
		}
		defer teardown(false)

		tests := []struct {
			name                   string
			numInputs              uint32
			numOutputs             uint32
			outputValue            uint64
			nodeSubnetworkID       externalapi.DomainSubnetworkID
			txSubnetworkData       *txSubnetworkData
			extraModificationsFunc func(*externalapi.DomainTransaction)
			expectedErr            error
		}{
			{"good one", 1, 1, 1, subnetworks.SubnetworkIDNative, nil, nil, nil},
			{"no inputs", 0, 1, 1, subnetworks.SubnetworkIDNative, nil, nil, ruleerrors.ErrNoTxInputs},
			{"no outputs", 1, 0, 1, subnetworks.SubnetworkIDNative, nil, nil, nil},
			{"too much sompi in one output", 1, 1, util.MaxSompi + 1,
				subnetworks.SubnetworkIDNative,
				nil,
				nil,
				ruleerrors.ErrBadTxOutValue},
			{"too much sompi in total outputs", 1, 2, util.MaxSompi - 1,
				subnetworks.SubnetworkIDNative,
				nil,
				nil,
				ruleerrors.ErrBadTxOutValue},
			{"duplicate inputs", 2, 1, 1,
				subnetworks.SubnetworkIDNative,
				nil,
				func(tx *externalapi.DomainTransaction) { tx.Inputs[1].PreviousOutpoint.Index = 0 },
				ruleerrors.ErrDuplicateTxInputs},
			{"1 input coinbase",
				1,
				1,
				1,
				subnetworks.SubnetworkIDNative,
				&txSubnetworkData{subnetworks.SubnetworkIDCoinbase, 0, nil},
				nil,
				nil},
			{"no inputs coinbase",
				0,
				1,
				1,
				subnetworks.SubnetworkIDNative,
				&txSubnetworkData{subnetworks.SubnetworkIDCoinbase, 0, nil},
				nil,
				nil},
			{"too long payload coinbase",
				1,
				1,
				1,
				subnetworks.SubnetworkIDNative,
				&txSubnetworkData{subnetworks.SubnetworkIDCoinbase, 0, make([]byte, params.MaxCoinbasePayloadLength+1)},
				nil,
				ruleerrors.ErrBadCoinbasePayloadLen},
			{"non-zero gas in Kaspa", 1, 1, 0,
				subnetworks.SubnetworkIDNative,
				nil,
				func(tx *externalapi.DomainTransaction) {
					tx.Gas = 1
				},
				ruleerrors.ErrInvalidGas},
			{"non-zero gas in subnetwork registry", 1, 1, 0,
				subnetworks.SubnetworkIDRegistry,
				&txSubnetworkData{subnetworks.SubnetworkIDRegistry, 1, []byte{}},
				nil,
				ruleerrors.ErrInvalidGas},
			{"non-zero payload in Kaspa", 1, 1, 0,
				subnetworks.SubnetworkIDNative,
				nil,
				func(tx *externalapi.DomainTransaction) {
					tx.Payload = []byte{1}
				},
				ruleerrors.ErrInvalidPayload},
			{"invalid payload hash", 1, 1, 0,
				externalapi.DomainSubnetworkID{123},
				&txSubnetworkData{externalapi.DomainSubnetworkID{123}, 0, []byte{1}},
				func(tx *externalapi.DomainTransaction) {
					tx.PayloadHash = externalapi.DomainHash{}
				},
				ruleerrors.ErrInvalidPayloadHash},
			{"invalid payload hash in native subnetwork", 1, 1, 0,
				subnetworks.SubnetworkIDNative,
				nil,
				func(tx *externalapi.DomainTransaction) {
					tx.PayloadHash = *hashes.HashData(tx.Payload)
				},
				ruleerrors.ErrInvalidPayloadHash},
		}

		for _, test := range tests {
			tx := createTxForTest(test.numInputs, test.numOutputs, test.outputValue, test.txSubnetworkData)

			if test.extraModificationsFunc != nil {
				test.extraModificationsFunc(tx)
			}

			err := tc.TransactionValidator().ValidateTransactionInIsolation(tx)
			if !errors.Is(err, test.expectedErr) {
				t.Errorf("TestValidateTransactionInIsolation: '%s': unexpected error %+v", test.name, err)
			}
		}
	})
}

func createTxForTest(numInputs uint32, numOutputs uint32, outputValue uint64, subnetworkData *txSubnetworkData) *externalapi.DomainTransaction {
	txIns := []*externalapi.DomainTransactionInput{}
	txOuts := []*externalapi.DomainTransactionOutput{}

	for i := uint32(0); i < numInputs; i++ {
		txIns = append(txIns, &externalapi.DomainTransactionInput{
			PreviousOutpoint: externalapi.DomainOutpoint{
				TransactionID: externalapi.DomainTransactionID{},
				Index:         i,
			},
			SignatureScript: []byte{},
			Sequence:        constants.MaxTxInSequenceNum,
		})
	}

	for i := uint32(0); i < numOutputs; i++ {
		txOuts = append(txOuts, &externalapi.DomainTransactionOutput{
			ScriptPublicKey: []byte{},
			Value:           outputValue,
		})
	}

	if subnetworkData != nil {
		return transactionhelper.NewSubnetworkTransaction(constants.TransactionVersion, txIns, txOuts, &subnetworkData.subnetworkID, subnetworkData.gas, subnetworkData.payload)
	}

	return transactionhelper.NewNativeTransaction(constants.TransactionVersion, txIns, txOuts)
}
