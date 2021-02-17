package consensus

import (
	"testing"

	"github.com/kaspanet/kaspad/domain/consensus/utils/testutils"
	"github.com/kaspanet/kaspad/domain/dagconfig"
)

func TestPruningDepth(t *testing.T) {
	expectedResult := map[string]uint64{
		dagconfig.MainnetParams.Name: 244838,
		dagconfig.TestnetParams.Name: 244838,
		dagconfig.DevnetParams.Name:  244838,
		dagconfig.SimnetParams.Name:  192038,
	}
	testutils.ForAllNets(t, true, func(t *testing.T, params *dagconfig.Params) {
		expected, found := expectedResult[params.Name]
		if !found {
			t.Fatalf("TestPruningDepth: expectedResult doesn't contain '%s'", params.Name)
		}
		if params.PruningDepth() != expected {
			t.Errorf("pruningDepth in %s is expected to be %d but got %d", params.Name, expected, params.PruningDepth())
		}
	})
}
