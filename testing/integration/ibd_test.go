package integration

import (
	"sync"
	"testing"
	"time"

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
	setOnBlockAddedHandler(t, syncee, func(_ *appmessage.BlockAddedNotificationMessage) {
		receivedBlocks++
		blockAddedWG.Done()
	})

	connect(t, syncer, syncee)

	select {
	case <-time.After(defaultTimeout):
		t.Fatalf("Timeout waiting for IBD to finish. Received %d blocks out of %d", receivedBlocks, numBlocks)
	case <-ReceiveFromChanWhenDone(func() { blockAddedWG.Wait() }):
	}

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
