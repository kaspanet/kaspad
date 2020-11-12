package integration

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/kaspanet/kaspad/util/locks"

	"github.com/kaspanet/kaspad/app/appmessage"
)

func Test64IncomingConnections(t *testing.T) {
	// Much more than 64 hosts creates a risk of running out of available file descriptors for leveldb
	const numBullies = 64
	harnessesParams := make([]*harnessParams, numBullies+1)
	for i := 0; i < numBullies+1; i++ {
		harnessesParams[i] = &harnessParams{
			p2pAddress:              fmt.Sprintf("127.0.0.1:%d", 12345+i),
			rpcAddress:              fmt.Sprintf("127.0.0.1:%d", 22345+i),
			miningAddress:           miningAddress1,
			miningAddressPrivateKey: miningAddress1PrivateKey,
		}
	}

	appHarnesses, teardown := setupHarnesses(t, harnessesParams)
	defer teardown()

	victim, bullies := appHarnesses[0], appHarnesses[1:]

	for _, bully := range bullies {
		connect(t, victim, bully)
	}

	blockAddedWG := sync.WaitGroup{}
	blockAddedWG.Add(numBullies)
	for _, bully := range bullies {
		blockAdded := false
		onBlockAdded := func(_ *appmessage.BlockAddedNotificationMessage) {
			if blockAdded {
				t.Fatalf("Single bully reported block added twice")
			}
			blockAdded = true
			blockAddedWG.Done()
		}

		err := bully.rpcClient.RegisterForBlockAddedNotifications(onBlockAdded)
		if err != nil {
			t.Fatalf("Error from RegisterForBlockAddedNotifications: %+v", err)
		}
	}

	_ = mineNextBlock(t, victim)

	select {
	case <-time.After(defaultTimeout):
		t.Fatalf("Timeout waiting for block added notification from the bullies")
	case <-locks.ReceiveFromChanWhenDone(func() { blockAddedWG.Wait() }):
	}
}
