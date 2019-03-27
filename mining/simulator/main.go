package main

import (
	"fmt"
	"log"
	"os/user"
	"path"
	"runtime/debug"
	"sync/atomic"

	"github.com/daglabs/btcd/rpcclient"
)

var isRunning int32

func main() {
	defer handlePanic()

	err := initPaths()
	if err != nil {
		panic(fmt.Errorf("Error initializing paths: %s", err))
	}

	addressList, err := getAddressList()
	if err != nil {
		panic(fmt.Errorf("Couldn't load address list: %s", err))
	}

	clients, err := connectToServers(addressList)
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

func initPaths() error {
	usr, err := user.Current()
	if err != nil {
		return fmt.Errorf("Error getting current user: %s", err)
	}

	basePath := ".btcd/mining_simulator"

	certificatePath = path.Join(usr.HomeDir, basePath, "rpc.cert")
	addressListPath = path.Join(usr.HomeDir, basePath, "addresses")

	return nil
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
