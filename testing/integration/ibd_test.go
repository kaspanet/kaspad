package integration

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
	"github.com/kaspanet/kaspad/domain/consensus/utils/mining"
	"math/rand"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/kaspanet/kaspad/domain/dagconfig"

	"github.com/kaspanet/kaspad/app/appmessage"
)

func TestIBD(t *testing.T) {
	const numBlocks = 100

	syncer, syncee, _, teardown := standardSetup(t)
	defer teardown()

	for i := 0; i < numBlocks; i++ {
		mineNextBlock(t, syncer)
	}

	blockAddedWG := sync.WaitGroup{}
	blockAddedWG.Add(numBlocks)
	receivedBlocks := 0
	disableOnBlockAddedHandler := false
	setOnBlockAddedHandler(t, syncee, func(_ *appmessage.BlockAddedNotificationMessage) {
		if disableOnBlockAddedHandler {
			return
		}
		receivedBlocks++
		blockAddedWG.Done()
	})

	// We expect this to trigger IBD
	connect(t, syncer, syncee)

	select {
	case <-time.After(defaultTimeout):
		t.Fatalf("Timeout waiting for IBD to finish. Received %d blocks out of %d", receivedBlocks, numBlocks)
	case <-ReceiveFromChanWhenDone(func() { blockAddedWG.Wait() }):
	}

	disableOnBlockAddedHandler = true
	// This should trigger resolving the syncee virtual
	mineNextBlock(t, syncer)
	time.Sleep(time.Second)

	tip1Hash, err := syncer.rpcClient.GetSelectedTipHash()
	if err != nil {
		t.Fatalf("Error getting tip for syncer")
	}

	tip2Hash, err := syncee.rpcClient.GetSelectedTipHash()
	if err != nil {
		t.Fatalf("Error getting tip for syncee")
	}

	if tip1Hash.SelectedTipHash != tip2Hash.SelectedTipHash {
		t.Errorf("Tips of syncer: '%s' and syncee '%s' are not equal", tip1Hash.SelectedTipHash, tip2Hash.SelectedTipHash)
	}
}

// TestIBDWithPruning checks the IBD from a node with
// already pruned blocks.
func TestIBDWithPruning(t *testing.T) {
	testSync := func(syncer, syncee *appHarness) {
		utxoSetOverriden := make(chan struct{})
		err := syncee.rpcClient.RegisterPruningPointUTXOSetNotifications(func() {
			close(utxoSetOverriden)
		})

		if err != nil {
			t.Fatalf("RegisterPruningPointUTXOSetNotifications: %+v", err)
		}

		// We expect this to trigger IBD
		connect(t, syncer, syncee)

		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

		start := time.Now()
		for range ticker.C {
			if time.Since(start) > defaultTimeout {
				t.Fatalf("Timeout waiting for IBD to finish.")
			}

			syncerInfo, err := syncer.rpcClient.GetBlockDAGInfo()
			if err != nil {
				t.Fatalf("Error getting tip for syncer")
			}
			synceeInfo, err := syncee.rpcClient.GetBlockDAGInfo()
			if err != nil {
				t.Fatalf("Error getting tip for syncee")
			}

			if reflect.DeepEqual(syncerInfo.TipHashes, synceeInfo.TipHashes) {
				break
			}
		}

		const timeout = 10 * time.Second
		select {
		case <-utxoSetOverriden:
		case <-time.After(timeout):
			t.Fatalf("expected pruning point UTXO set override notification, but it didn't get one after %s", timeout)
		}

		// Checking that the syncee can generate block templates before resolving the virtual
		_, err = syncee.rpcClient.GetBlockTemplate(syncee.miningAddress)
		if err != nil {
			t.Fatalf("Error getting block template: %+v", err)
		}

		// This should trigger resolving the syncee virtual
		syncerTip := mineNextBlockWithMockTimestamps(t, syncer, rand.New(rand.NewSource(time.Now().UnixNano())))
		time.Sleep(time.Second)
		synceeSelectedTip, err := syncee.rpcClient.GetSelectedTipHash()
		if err != nil {
			t.Fatalf("Error getting tip for syncee")
		}

		if synceeSelectedTip.SelectedTipHash != consensushashing.BlockHash(syncerTip).String() {
			t.Fatalf("Unexpected selected tip: expected %s but got %s", consensushashing.BlockHash(syncerTip).String(), synceeSelectedTip.SelectedTipHash)
		}
	}

	const numBlocks = 100

	overrideDAGParams := dagconfig.SimnetParams

	// Increase the target time per block so that we could mine
	// blocks with timestamps that are spaced far enough apart
	// to avoid failing the timestamp threshold validation of
	// ibd-with-headers-proof
	overrideDAGParams.TargetTimePerBlock = time.Minute

	// This is done to make a pruning depth of 6 blocks
	overrideDAGParams.FinalityDuration = 2 * overrideDAGParams.TargetTimePerBlock
	overrideDAGParams.K = 0
	overrideDAGParams.PruningProofM = 20

	expectedPruningDepth := uint64(6)
	if overrideDAGParams.PruningDepth() != expectedPruningDepth {
		t.Fatalf("Unexpected pruning depth: expected %d but got %d", expectedPruningDepth, overrideDAGParams.PruningDepth())
	}

	harnesses, teardown := setupHarnesses(t, []*harnessParams{
		{
			p2pAddress:              p2pAddress1,
			rpcAddress:              rpcAddress1,
			miningAddress:           miningAddress1,
			miningAddressPrivateKey: miningAddress1PrivateKey,
			overrideDAGParams:       &overrideDAGParams,
		},
		{
			p2pAddress:              p2pAddress2,
			rpcAddress:              rpcAddress2,
			miningAddress:           miningAddress2,
			miningAddressPrivateKey: miningAddress2PrivateKey,
			overrideDAGParams:       &overrideDAGParams,
			utxoIndex:               true,
		},
		{
			p2pAddress:              p2pAddress3,
			rpcAddress:              rpcAddress3,
			miningAddress:           miningAddress3,
			miningAddressPrivateKey: miningAddress3PrivateKey,
			overrideDAGParams:       &overrideDAGParams,
			utxoIndex:               true,
		},
	})
	defer teardown()

	syncer, syncee1, syncee2 := harnesses[0], harnesses[1], harnesses[2]

	// Let syncee1 have two blocks that the syncer
	// doesn't have to test a situation where
	// the block locator will need more than one
	// iteration to find the highest shared chain
	// block.
	const synceeOnlyBlocks = 2
	rd := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < synceeOnlyBlocks; i++ {
		mineNextBlockWithMockTimestamps(t, syncee1, rd)
	}

	for i := 0; i < numBlocks-1; i++ {
		mineNextBlockWithMockTimestamps(t, syncer, rd)
	}

	testSync(syncer, syncee1)

	// Test a situation where a node with pruned headers syncs another fresh node.
	testSync(syncee1, syncee2)
}

var currentMockTimestamp int64 = 0

// mineNextBlockWithMockTimestamps mines blocks with large timestamp differences
// between every two blocks. This is done to avoid the timestamp threshold validation
// of ibd-with-headers-proof
func mineNextBlockWithMockTimestamps(t *testing.T, harness *appHarness, rd *rand.Rand) *externalapi.DomainBlock {
	blockTemplate, err := harness.rpcClient.GetBlockTemplate(harness.miningAddress)
	if err != nil {
		t.Fatalf("Error getting block template: %+v", err)
	}

	block, err := appmessage.RPCBlockToDomainBlock(blockTemplate.Block)
	if err != nil {
		t.Fatalf("Error converting block: %s", err)
	}

	if currentMockTimestamp == 0 {
		currentMockTimestamp = block.Header.TimeInMilliseconds()
	} else {
		currentMockTimestamp += 10_000
	}
	mutableHeader := block.Header.ToMutable()
	mutableHeader.SetTimeInMilliseconds(currentMockTimestamp)
	block.Header = mutableHeader.ToImmutable()

	mining.SolveBlock(block, rd)

	_, err = harness.rpcClient.SubmitBlockAlsoIfNonDAA(block)
	if err != nil {
		t.Fatalf("Error submitting block: %s", err)
	}

	return block
}
