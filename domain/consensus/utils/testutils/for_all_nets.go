package testutils

import (
	"testing"

	"github.com/kaspanet/kaspad/domain/dagconfig"
)

// ForAllNets runs the passed testFunc with all available networks
// if setDifficultyToMinumum = true - will modify the net params to have minimal difficulty, like in SimNet
func ForAllNets(t *testing.T, setDifficultyToMinimum bool, testFunc func(*testing.T, *dagconfig.Params)) {
	allParams := []dagconfig.Params{
		dagconfig.MainnetParams,
		dagconfig.TestnetParams,
		dagconfig.SimnetParams,
		dagconfig.DevnetParams,
	}

	for _, params := range allParams {
		if setDifficultyToMinimum {
			params.DisableDifficultyAdjustment = dagconfig.SimnetParams.DisableDifficultyAdjustment
			params.TargetTimePerBlock = dagconfig.SimnetParams.TargetTimePerBlock
		}

		t.Run(params.Name, func(t *testing.T) { testFunc(t, &params) })
	}
}
