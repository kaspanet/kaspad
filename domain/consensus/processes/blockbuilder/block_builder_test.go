package blockbuilder_test

import (
	"fmt"
	"testing"

	"github.com/kaspanet/kaspad/domain/consensus"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/model/testapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/testutils"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/pkg/errors"
)

func TestBlockBuilderErrorCases(t *testing.T) {
	testutils.ForAllNets(t, true, func(t *testing.T, params *dagconfig.Params) {
		factory := consensus.NewFactory()

		tests := []struct {
			name          string
			preparation   func(testConsensus testapi.TestConsensus) error
			coinbaseData  *externalapi.DomainCoinbaseData
			transactions  []*externalapi.DomainTransaction
			expectedError error
		}{}

		for _, test := range tests {
			func() {
				consensus, teardown, err := factory.NewTestConsensus(
					params, false, fmt.Sprintf("TestBlockBuilderErrorCases-%s", test.name))
				if err != nil {
					t.Fatalf("Error initializing consensus for %s: %+v", test.name, err)
				}
				defer teardown(false)

				if test.preparation != nil {
					err := test.preparation(consensus)
					if err != nil {
						t.Errorf("%s: Error during preparation: %+v", test.name, err)
						return
					}
				}

				_, err = consensus.BuildBlock(test.coinbaseData, test.transactions)
				if err == nil {
					t.Errorf("%s: No error from BuildBlock")
					return
				}
				if !errors.Is(test.expectedError, err) {
					t.Errorf("%s: Expected error '%s', but got '%s'", test.name, test.expectedError, err)
				}
			}()
		}
	})
}
