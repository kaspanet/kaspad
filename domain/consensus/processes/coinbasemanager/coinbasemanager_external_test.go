package coinbasemanager_test

import (
	"github.com/kaspanet/kaspad/domain/consensus"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/testutils"
	"testing"
)

func TestExtractCoinbaseDataBlueScoreAndSubsidy(t *testing.T) {
	testutils.ForAllNets(t, true, func(t *testing.T, consensusConfig *consensus.Config) {
		factory := consensus.NewFactory()
		tc, teardown, err := factory.NewTestConsensus(consensusConfig, "TestBlockStatus")
		if err != nil {
			t.Fatalf("Error setting up consensus: %+v", err)
		}
		defer teardown(false)

		tests := []struct {
			name                   string
			scriptPublicKeyVersion uint16
		}{
			{
				name:                   "below 255",
				scriptPublicKeyVersion: 100,
			},
			{
				name:                   "above 255",
				scriptPublicKeyVersion: 300,
			},
		}

		for _, test := range tests {
			coinbaseTx, _, err := tc.CoinbaseManager().ExpectedCoinbaseTransaction(model.NewStagingArea(), model.VirtualBlockHash, &externalapi.DomainCoinbaseData{
				ScriptPublicKey: &externalapi.ScriptPublicKey{
					Script:  nil,
					Version: test.scriptPublicKeyVersion,
				},
				ExtraData: nil,
			})
			if err != nil {
				t.Fatal(err)
			}

			_, cbData, _, err := tc.CoinbaseManager().ExtractCoinbaseDataBlueScoreAndSubsidy(coinbaseTx)
			if err != nil {
				t.Fatal(err)
			}

			if cbData.ScriptPublicKey.Version != test.scriptPublicKeyVersion {
				t.Fatalf("test %s post HF expected %d but got %d", test.name, test.scriptPublicKeyVersion, cbData.ScriptPublicKey.Version)
			}
		}

	})
}
