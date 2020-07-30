package integration

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/kaspanet/kaspad/util/locks"

	"github.com/kaspanet/kaspad/wire"

	clientpkg "github.com/kaspanet/kaspad/rpc/client"
)

func Test117IncomingConnections(t *testing.T) {
	const numBullies = 117
	params := make([]*harnessParams, numBullies+1)
	for i := 0; i < numBullies+1; i++ {
		params[i] = &harnessParams{
			p2pAddress: fmt.Sprintf("127.0.0.1:%d", 12345+i),
			rpcAddress: fmt.Sprintf("127.0.0.1:%d", 22345+i),
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

	blockTemplate, err := victim.rpcClient.GetBlockTemplate(testAddress1, "")
	if err != nil {
		t.Fatalf("Error getting block template: %+v", err)
	}

	block, err := clientpkg.ConvertGetBlockTemplateResultToBlock(blockTemplate)
	if err != nil {
		t.Fatalf("Error parsing blockTemplate: %s", err)
	}

	solveBlock(t, block)

	err = victim.rpcClient.SubmitBlock(block, nil)
	if err != nil {
		t.Fatalf("Error submitting block: %s", err)
	}

	select {
	case <-time.After(defaultTimeout):
		t.Fatalf("Timeout waiting for block added notification from the bullies")
	case <-locks.ReceiveFromChanWhenDone(func() { blockAddedWG.Wait() }):
	}
}
