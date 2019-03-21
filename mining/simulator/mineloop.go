package main

import (
	"fmt"
	"log"
	"math/rand"
	"sync/atomic"
	"time"

	"github.com/daglabs/btcd/rpcclient"
)

var r = rand.New(rand.NewSource(time.Now().UnixNano()))

func mineLoop(clients []*rpcclient.Client) error {
	clientsCount := int64(len(clients))

	for atomic.LoadInt32(&isRunning) == 1 {
		currentClient := clients[r.Int63n(clientsCount)]

		template, err := currentClient.GetBlockTemplate()
		if err != nil {
			return fmt.Errorf("Error getting block template: %s", err)
		}

		log.Printf("Got template: %+v", template)
	}

	return nil
}
