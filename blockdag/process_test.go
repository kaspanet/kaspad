package blockdag

import (
	"bou.ke/monkey"
	"github.com/daglabs/btcd/dagconfig"
	"github.com/daglabs/btcd/util"
	"testing"
	"time"
)

func TestProcessBlock(t *testing.T) {
	dag, teardownFunc, err := DAGSetup("TestProcessBlock", Config{
		DAGParams: &dagconfig.SimNetParams,
	})
	if err != nil {
		t.Errorf("Failed to setup dag instance: %v", err)
		return
	}
	defer teardownFunc()

	// Check that BFAfterDelay skip checkBlockSanity
	called := false
	guard := monkey.Patch((*BlockDAG).checkBlockSanity, func(_ *BlockDAG, _ *util.Block, _ BehaviorFlags) (time.Duration, error) {
		called = true
		return 0, nil
	})
	defer guard.Unpatch()

	isOrphan, delay, err := dag.ProcessBlock(util.NewBlock(&Block100000), BFNoPoWCheck)
	if err != nil {
		t.Errorf("ProcessBlock: %s", err)
	}
	if delay != 0 {
		t.Errorf("ProcessBlock: unexpected returned %s delay", delay)
	}
	if !isOrphan {
		t.Errorf("ProcessBlock: unexpected returned non orphan block")
	}
	if !called {
		t.Errorf("ProcessBlock: expected checkBlockSanity to be called")
	}

	Block100000Copy := Block100000
	// Change nonce to change block hash
	Block100000Copy.Header.Nonce++
	called = false
	isOrphan, delay, err = dag.ProcessBlock(util.NewBlock(&Block100000Copy), BFAfterDelay|BFNoPoWCheck)
	if err != nil {
		t.Errorf("ProcessBlock: %s", err)
	}
	if delay != 0 {
		t.Errorf("ProcessBlock: unexpected returned %s delay", delay)
	}
	if !isOrphan {
		t.Errorf("ProcessBlock: unexpected returned non orphan block")
	}
	if called {
		t.Errorf("ProcessBlock: Didn't expected checkBlockSanity to be called")
	}
}
