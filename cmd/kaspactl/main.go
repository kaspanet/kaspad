package main

import (
	"fmt"
	"github.com/kaspanet/kaspad/infrastructure/network/rpcclient/grpcclient"
	"os"
	"time"
)

func main() {
	cfg, err := parseConfig()
	if err != nil {
		printErrorAndExit(fmt.Sprintf("error parsing command-line arguments: %s", err))
	}

	rpcAddress, err := cfg.NetParams().NormalizeRPCServerAddress(cfg.RPCServer)
	if err != nil {
		printErrorAndExit(fmt.Sprintf("error parsing RPC server address: %s", err))
	}
	client, err := grpcclient.Connect(rpcAddress)
	if err != nil {
		printErrorAndExit(fmt.Sprintf("error connecting to the RPC server: %s", err))
	}
	defer client.Disconnect()

	var responseString string
	done := make(chan struct{})

	go func() {
		requestString := cfg.RequestJSON
		var err error
		responseString, err = client.PostJSON(requestString)
		if err != nil {
			printErrorAndExit(fmt.Sprintf("error posting the request to the RPC server: %s", err))
		}
		done <- struct{}{}
	}()

	timeout := time.Duration(cfg.Timeout) * time.Second
	select {
	case <-done:
		fmt.Println(responseString)
	case <-time.After(timeout):
		printErrorAndExit(fmt.Sprintf("timeout of %s has been exceeded", timeout))
	}
}

func printErrorAndExit(message string) {
	fmt.Fprintf(os.Stderr, fmt.Sprintf("%s\n", message))
	os.Exit(1)
}
