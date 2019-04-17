package main

import (
	"fmt"
	"github.com/daglabs/btcd/signal"
	"log"
	"os"
	"runtime/debug"
)

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

	go func() {
		err = mineLoop(clients)
		if err != nil {
			panic(fmt.Errorf("Error in main loop: %s", err))
		}
	}()

	interrupt := signal.InterruptListener()
	<-interrupt
}

func disconnect(clients []*simulatorClient) {
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
