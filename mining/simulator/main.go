package main

import (
	"fmt"
	"log"
	"sync/atomic"

	"github.com/daglabs/btcd/rpcclient"
)

var isRunning int32

func main() {
	defer handlePanic()

	addressList, err := getAddressList()
	if err != nil {
		panic(fmt.Errorf("Couldn't load address list: %s", err))
	}

	clients, err := connectToServers(addressList)
	defer disconnect(clients)

	atomic.StoreInt32(&isRunning, 1)

	err = mineLoop(clients)
	if err != nil {
		panic(fmt.Errorf("Error in main loop: %s", err))
	}
}

func disconnect(clients []*rpcclient.Client) {
	for _, client := range clients {
		client.Disconnect()
	}
}

func handlePanic() {
	err := recover()
	if err != nil {
		log.Printf("Fatal error: %s", err)
	}
}
