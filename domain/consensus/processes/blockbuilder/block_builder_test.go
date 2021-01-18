package blockbuilder_test

import (
	"fmt"
	"testing"

	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"

	"github.com/kaspanet/kaspad/domain/consensus"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/testutils"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/pkg/errors"
)

func TestBuildBlockErrorCases(t *testing.T) {
	testutils.ForAllNets(t, true, func(t *testing.T, params *dagconfig.Params) {
		factory := consensus.NewFactory()

		tests := []struct {
			name              string
			coinbaseData      *externalapi.DomainCoinbaseData
			transactions      []*externalapi.DomainTransaction
			expectedErrorType error
		}{
			{
				"scriptPublicKey too long",
				&externalapi.DomainCoinbaseData{
					ScriptPublicKey: &externalapi.ScriptPublicKey{
						Script:  make([]byte, params.CoinbasePayloadScriptPublicKeyMaxLength+1),
						Version: 0,
					},
					ExtraData: nil,
				},
				nil,
				ruleerrors.ErrBadCoinbasePayloadLen,
			},
		}

		for _, test := range tests {
			func() {
				consensus, teardown, err := factory.NewTestConsensus(
					params, false, fmt.Sprintf("TestBlockBuilderErrorCases-%s", test.name))
				if err != nil {
					t.Fatalf("Error initializing consensus for %s: %+v", test.name, err)
				}
				defer teardown(false)

				_, err = consensus.BlockBuilder().BuildBlock(test.coinbaseData, test.transactions)
				if err == nil {
					t.Errorf("%s: No error from BuildBlock", test.name)
					return
				}
				if test.expectedErrorType != nil && !errors.Is(err, test.expectedErrorType) {
					t.Errorf("%s: Expected error '%s', but got '%s'", test.name, test.expectedErrorType, err)
				}
			}()
		}
	})
}
