package blockdag

import (
	"github.com/kaspanet/kaspad/dagconfig"
	"github.com/kaspanet/kaspad/util"
	"testing"
	"time"
)

func TestCheckBlockDelayed(t *testing.T) {
	// Create a new database and dag instance to run tests against.
	dag, teardownFunc, err := DAGSetup("TestCheckBlockDelayed", true, Config{
		DAGParams: &dagconfig.SimnetParams,
	})
	if err != nil {
		t.Errorf("Failed to setup dag instance: %v", err)
		return
	}
	defer teardownFunc()

	blockInTheFuture := Block100000
	expectedDelay := 10 * time.Second
	deviationTolerance := time.Duration(dag.TimestampDeviationTolerance) * dag.Params.TargetTimePerBlock
	blockInTheFuture.Header.Timestamp = dag.Now().Add(deviationTolerance + expectedDelay)
	delay, isDelayed := dag.checkBlockDelayed(util.NewBlock(&blockInTheFuture))
	if !isDelayed {
		t.Errorf("TestCheckBlockDelayed: block unexpectedly not delayed")
	}
	if delay != expectedDelay {
		t.Errorf("TestCheckBlockDelayed: expected %s delay but got %s", expectedDelay, delay)
	}
}
