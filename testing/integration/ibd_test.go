package integration

import (
	"sync"
	"testing"
	"time"

	"github.com/kaspanet/kaspad/util/locks"

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
	setOnBlockAddedHandler(t, syncee, func(header *appmessage.BlockHeader) {
		receivedBlocks++
		blockAddedWG.Done()
	})

	connect(t, syncer, syncee)

	select {
	case <-time.After(defaultTimeout):
		t.Fatalf("Timeout waiting for IBD to finish. Received %d blocks out of %d", receivedBlocks, numBlocks)
	case <-locks.ReceiveFromChanWhenDone(func() { blockAddedWG.Wait() }):
	}

	tip1, err := syncer.rpcClient.GetSelectedTip()
	if err != nil {
		t.Fatalf("Error getting tip for syncer")
	}
	tip2, err := syncee.rpcClient.GetSelectedTip()
	if err != nil {
		t.Fatalf("Error getting tip for syncee")
	}

	if tip1.Hash != tip2.Hash {
		t.Errorf("Tips of syncer: '%s' and syncee '%s' are not equal", tip1.Hash, tip2.Hash)
	}
}
