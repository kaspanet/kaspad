package blockdag

import (
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"testing"
)

func TestPruningDepth(t *testing.T) {
	tests := []struct {
		params        *dagconfig.Params
		expectedDepth uint64
	}{
		{
			params:        &dagconfig.MainnetParams,
			expectedDepth: 244838,
		},
		{
			params:        &dagconfig.TestnetParams,
			expectedDepth: 244838,
		},
		{
			params:        &dagconfig.DevnetParams,
			expectedDepth: 244838,
		},
		{
			params:        &dagconfig.RegressionNetParams,
			expectedDepth: 244838,
		},
		{
			params:        &dagconfig.SimnetParams,
			expectedDepth: 192038,
		},
	}
	for _, test := range tests {
		func() {
			dag, teardownFunc, err := DAGSetup("TestFinalityInterval", true, Config{
				DAGParams: test.params,
			})
			if err != nil {
				t.Fatalf("Failed to setup dag instance for %s: %v", test.params.Name, err)
			}
			defer teardownFunc()

			if dag.pruningDepth() != test.expectedDepth {
				t.Errorf("pruningDepth in %s is expected to be %d but got %d", test.params.Name, test.expectedDepth, dag.pruningDepth())
			}
		}()
	}
}
