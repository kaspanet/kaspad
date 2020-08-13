package blockdag

import (
	"github.com/kaspanet/kaspad/dagconfig"
	"testing"
)

func TestIsDAGCurrentMaxDiff(t *testing.T) {
	netParams := []*dagconfig.Params{
		&dagconfig.MainnetParams,
		&dagconfig.TestnetParams,
		&dagconfig.DevnetParams,
		&dagconfig.RegressionNetParams,
		&dagconfig.SimnetParams,
	}
	for _, params := range netParams {
		if params.FinalityDuration < isDAGCurrentMaxDiff*params.TargetTimePerBlock {
			t.Errorf("in %s, a DAG can be considered current even if it's below the finality point", params.Name)
		}
	}
}
