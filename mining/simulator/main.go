package main

import (
	"fmt"
	"log"
	"os/user"
	"path"
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

func init() {
	usr, err := user.Current()
	if err != nil {
		panic(fmt.Errorf("Error getting current user: %s", err))
	}
	certificatePath = path.Join(usr.HomeDir, ".btcd/simulator/rpc.cert")
	addressListPath = path.Join(usr.HomeDir, ".btcd/simulator/addresses")
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
