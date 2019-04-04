package main

import (
	"fmt"
	"log"
	"os"
	"runtime/debug"
	"sync/atomic"

	"github.com/daglabs/btcd/rpcclient"
)

var isRunning int32

func main() {
	defer handlePanic()

	cfg, err := parseConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing command-line arguments: %s", err)
		os.Exit(1)
	}

	addressList, err := getAddressList(cfg)
	if err != nil {
		panic(fmt.Errorf("Couldn't load address list: %s", err))
	}

	clients, err := connectToServers(cfg, addressList)
	if err != nil {
		panic(fmt.Errorf("Error connecting to servers: %s", err))
	}
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
		log.Printf("Stack trace: %s", debug.Stack())
	}
}
