package testutils

import (
	"testing"

	"github.com/kaspanet/kaspad/domain/dagconfig"
)

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

		testFunc(t, &params)
	}
}
