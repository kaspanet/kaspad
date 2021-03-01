package rpccontext

import (
	"github.com/kaspanet/kaspad/domain/consensus"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/model/testapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/testutils"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/kaspanet/kaspad/domain/miningmanager"
	"github.com/kaspanet/kaspad/infrastructure/config"
	"testing"
)

type fakeDomain struct{ testapi.TestConsensus }

func (d fakeDomain) MiningManager() miningmanager.MiningManager { return nil }
func (d fakeDomain) Consensus() externalapi.Consensus           { return d }

func TestContext_BuildBlockVerboseDataRoundTrip(t *testing.T) {
	testutils.ForAllNets(t, true, func(t *testing.T, params *dagconfig.Params) {
		consensus, teardown, err := consensus.NewFactory().NewTestConsensus(params, false, "TestContext_BuildBlockVerboseDataRoundTrip")
		if err != nil {
			t.Fatalf("Error setting up consensus: %+v", err)
		}
		defer teardown(false)

		config := config.Config{Flags: &config.Flags{NetworkFlags: config.NetworkFlags{ActiveNetParams: params}}}
		ctx := NewContext(&config, fakeDomain{consensus}, nil, nil, nil, nil, nil, nil)
		genesisVerbose, err := ctx.BuildBlockVerboseData(params.GenesisBlock.Header, nil, true)
		if err != nil {
			t.Fatalf("Failed getting verbose block data for genesis: %v", err)
		}

		genesisConverted, err := ConvertBlockVerboseDataToDomainBlock(genesisVerbose)
		if err != nil {
			t.Fatalf("Failed converting verbose block data back to block: %v", err)
		}

		if !genesisConverted.Equal(params.GenesisBlock) {
			t.Fatal("Expected the converted block to be equal to the original genesis block")
		}

		genesisNoTXVerbose, err := ctx.BuildBlockVerboseData(params.GenesisBlock.Header, nil, false)
		if err != nil {
			t.Fatalf("Failed getting verbose block data for genesis: %v", err)
		}

		genesisNoTXConverted, err := ConvertBlockVerboseDataToDomainBlock(genesisNoTXVerbose)
		if err != nil {
			t.Fatalf("Failed converting verbose block data back to block: %v", err)
		}

		if !genesisNoTXConverted.Header.Equal(params.GenesisBlock.Header) {
			t.Fatal("Expected the converted block header to be equal to the original genesis block header")
		}

		if len(genesisNoTXConverted.Transactions) > 0 {
			t.Fatalf("Expected no transactions in the converted verbose block with includeTransactionVerboseData=false")
		}
	})
}
