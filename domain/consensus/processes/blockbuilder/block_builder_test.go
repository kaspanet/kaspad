package blockbuilder_test

import (
	"github.com/pkg/errors"
	"testing"

	"github.com/kaspanet/kaspad/domain/consensus"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
	"github.com/kaspanet/kaspad/domain/consensus/utils/constants"
	"github.com/kaspanet/kaspad/domain/consensus/utils/subnetworks"
	"github.com/kaspanet/kaspad/domain/consensus/utils/testutils"
)

func TestBuildBlockErrorCases(t *testing.T) {
	testutils.ForAllNets(t, true, func(t *testing.T, consensusConfig *consensus.Config) {
		factory := consensus.NewFactory()
		testConsensus, teardown, err := factory.NewTestConsensus(consensusConfig, "TestBlockBuilderErrorCases")
		if err != nil {
			t.Fatalf("Error initializing consensus for: %+v", err)
		}
		defer teardown(false)

		type testData struct {
			name         string
			coinbaseData *externalapi.DomainCoinbaseData
			transactions []*externalapi.DomainTransaction
			testFunc     func(test testData, err error) error
		}

		tests := []testData{
			{
				"scriptPublicKey too long",
				&externalapi.DomainCoinbaseData{
					ScriptPublicKey: &externalapi.ScriptPublicKey{
						Script:  make([]byte, consensusConfig.CoinbasePayloadScriptPublicKeyMaxLength+1),
						Version: 0,
					},
					ExtraData: nil,
				},
				nil,
				func(_ testData, err error) error {
					if !errors.Is(err, ruleerrors.ErrBadCoinbasePayloadLen) {
						return errors.Errorf("Unexpected error: %+v", err)
					}
					return nil
				},
			},
			{
				"missing UTXO transactions",
				&externalapi.DomainCoinbaseData{
					ScriptPublicKey: &externalapi.ScriptPublicKey{
						Script:  nil,
						Version: 0,
					},
					ExtraData: nil,
				},
				[]*externalapi.DomainTransaction{
					{
						Version: constants.MaxTransactionVersion,
						Inputs: []*externalapi.DomainTransactionInput{
							{
								PreviousOutpoint: externalapi.DomainOutpoint{
									TransactionID: externalapi.DomainTransactionID{}, Index: 0},
							},
						},
						Outputs:      nil,
						LockTime:     0,
						SubnetworkID: subnetworks.SubnetworkIDNative,
						Gas:          0,
						Payload:      []byte{0},
					},
					{
						Version: constants.MaxTransactionVersion,
						Inputs: []*externalapi.DomainTransactionInput{
							{
								PreviousOutpoint: externalapi.DomainOutpoint{
									TransactionID: externalapi.DomainTransactionID{}, Index: 0},
							},
						},
						Outputs:      nil,
						LockTime:     0,
						SubnetworkID: subnetworks.SubnetworkIDNative,
						Gas:          0,
						Payload:      []byte{1},
					},
				},

				func(test testData, err error) error {
					errInvalidTransactionsInNewBlock := ruleerrors.ErrInvalidTransactionsInNewBlock{}
					if !errors.As(err, &errInvalidTransactionsInNewBlock) {
						return errors.Errorf("Unexpected error: %+v", err)
					}

					if len(errInvalidTransactionsInNewBlock.InvalidTransactions) != len(test.transactions) {
						return errors.Errorf("Expected %d transaction but got %d",
							len(test.transactions), len(errInvalidTransactionsInNewBlock.InvalidTransactions))
					}

					for i, invalidTx := range errInvalidTransactionsInNewBlock.InvalidTransactions {
						if !invalidTx.Transaction.Equal(test.transactions[i]) {
							return errors.Errorf("Expected transaction %d to be equal to its corresponding "+
								"test transaction", i)
						}

						if !errors.As(invalidTx.Error, &ruleerrors.ErrMissingTxOut{}) {
							return errors.Errorf("Unexpected error for transaction %d: %+v", i, invalidTx.Error)
						}
					}
					return nil
				},
			},
		}

		for _, test := range tests {
			_, err = testConsensus.BlockBuilder().BuildBlock(test.coinbaseData, test.transactions)
			if err == nil {
				t.Errorf("%s: No error from BuildBlock", test.name)
				return
			}

			err := test.testFunc(test, err)
			if err != nil {
				t.Errorf("%s: %s", test.name, err)
				return
			}
		}
	})
}
