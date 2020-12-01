package testutils

import (
	"testing"

	"github.com/kaspanet/kaspad/domain/dagconfig"
)

// ForAllNets runs the passed testFunc with all available networks
// if setDifficultyToMinumum = true - will modify the net params to have minimal difficulty, like in SimNet
func ForAllNets(t *testing.T, skipPow bool, testFunc func(*testing.T, *dagconfig.Params)) {
	allParams := []dagconfig.Params{
		dagconfig.MainnetParams,
		dagconfig.TestnetParams,
		dagconfig.SimnetParams,
		dagconfig.DevnetParams,
	}

	for _, params := range allParams {
		paramsCopy := params
		paramsCopy.SkipProofOfWork = skipPow
		t.Logf("Running test for %s", params.Name)
		testFunc(t, &paramsCopy)
	}
}
