package consensus

import (
	"github.com/kaspanet/kaspad/domain/consensus/utils/testutils"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"testing"
)

func TestPruningDepth(t *testing.T) {
	expectedResult := map[string]uint64{
		"kaspa-mainnet": 244838,
		"kaspa-testnet": 244838,
		"kaspa-devnet":  244838,
		"kaspa-simnet":  192038,
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
