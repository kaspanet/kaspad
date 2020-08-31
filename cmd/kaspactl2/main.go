package main

import (
	"fmt"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/server/grpcserver/grpcclient"
	"os"
)

func main() {
	cfg, err := parseConfig()
	if err != nil {
		printErrorAndExit(fmt.Sprintf("error parsing command-line arguments: %s", err))
	}

	client, err := grpcclient.Connect(cfg.RPCServer)
	if err != nil {
		printErrorAndExit(fmt.Sprintf("error connecting to the RPC server: %s", err))
	}
	defer client.Disconnect()

	requestString := cfg.RequestJSON
	responseString, err := client.PostString(requestString)
	if err != nil {
		printErrorAndExit(fmt.Sprintf("error posting the request to the RPC server: %s", err))
	}

	fmt.Println(responseString)
}

func printErrorAndExit(message string) {
	fmt.Fprintf(os.Stderr, fmt.Sprintf("%s\n", message))
	os.Exit(1)
}
