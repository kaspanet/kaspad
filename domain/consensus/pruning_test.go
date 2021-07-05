package consensus_test

import (
	"testing"

	"github.com/kaspanet/kaspad/domain/consensus"
	"github.com/kaspanet/kaspad/domain/consensus/utils/testutils"
	"github.com/kaspanet/kaspad/domain/dagconfig"
)

func TestPruningDepth(t *testing.T) {
	expectedResult := map[string]uint64{
		dagconfig.MainnetParams.Name: 185798,
		dagconfig.TestnetParams.Name: 185798,
		dagconfig.DevnetParams.Name:  185798,
		dagconfig.SimnetParams.Name:  132998,
	}
	testutils.ForAllNets(t, true, func(t *testing.T, consensusConfig *consensus.Config) {
		expected, found := expectedResult[consensusConfig.Name]
		if !found {
			t.Fatalf("TestPruningDepth: expectedResult doesn't contain '%s'", consensusConfig.Name)
		}
		if consensusConfig.PruningDepth() != expected {
			t.Errorf("pruningDepth in %s is expected to be %d but got %d", consensusConfig.Name, expected, consensusConfig.PruningDepth())
		}
	})
}
