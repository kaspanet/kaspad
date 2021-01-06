package consensus_test

import (
	"github.com/kaspanet/kaspad/domain/consensus"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"os"
	"testing"
)

func TestSutton(t *testing.T) {
	tc, teardown, err := consensus.NewFactory().NewTestConsensus(&dagconfig.DevnetParams, "TestSutton")
	if err != nil {
		t.Fatalf("Error setting up consensus: %+v", err)
	}
	defer teardown(true)

	f, err := os.Open("../../testdata/dags/wide-dag-blocks--2^11-delay-factor--1-k--18.json")
	if err != nil {
		t.Fatal(f)
	}
	err = tc.MineJSON(f)
	if err != nil {
		t.Fatal(err)
	}
	// Do whatever you want with TestConsensus.
}
