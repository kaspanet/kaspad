package integration

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/kaspanet/kaspad/util/locks"

	"github.com/kaspanet/kaspad/wire"
)

func Test64IncomingConnections(t *testing.T) {
	const numBullies = 64 // Much more then 64 hosts creates a risk of running out of available file descriptors
	params := make([]*harnessParams, numBullies+1)
	for i := 0; i < numBullies+1; i++ {
		params[i] = &harnessParams{
			p2pAddress:              fmt.Sprintf("127.0.0.1:%d", 12345+i),
			rpcAddress:              fmt.Sprintf("127.0.0.1:%d", 22345+i),
			miningAddress:           miningAddress1,
			miningAddressPrivateKey: miningAddress1PrivateKey,
		}
	}

	appHarnesses, teardown := setupHarnesses(t, params)
	defer teardown()

	victim, bullies := appHarnesses[0], appHarnesses[1:]

	for _, bully := range bullies {
		connect(t, victim, bully)
	}

	blockAddedWG := sync.WaitGroup{}
	blockAddedWG.Add(numBullies)
	for _, bully := range bullies {
		err := bully.rpcClient.NotifyBlocks()
		if err != nil {
			t.Fatalf("Error from NotifyBlocks: %+v", err)
		}

		bully.rpcClient.onBlockAdded = func(header *wire.BlockHeader) {
			blockAddedWG.Done()
		}
	}

	_ = mineNextBlock(t, victim)

	select {
	case <-time.After(defaultTimeout):
		t.Fatalf("Timeout waiting for block added notification from the bullies")
	case <-locks.ReceiveFromChanWhenDone(func() { blockAddedWG.Wait() }):
	}
}
