package main

import (
	"github.com/pkg/errors"
	"github.com/zoomy-network/zoomyd/infrastructure/network/rpcclient/grpcclient"
)

func sendCommands(rpcClient *grpcclient.GRPCClient, commandsChan <-chan string) error {
	for command := range commandsChan {
		log.Infof("Sending command %s", command)
		response, err := rpcClient.PostJSON(command)
		if err != nil {
			return errors.Wrap(err, "error sending message")
		}

		log.Infof("-> Got response: %s", response)
	}

	return nil
}
