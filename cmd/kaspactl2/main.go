package main

import (
	"fmt"
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/server/grpcserver/protowire"
	"github.com/kaspanet/kaspad/util/panics"
	"github.com/pkg/errors"
	"os"
)

func main() {
	defer panics.HandlePanic(log, "MAIN", nil)

	cfg, err := parseConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing command-line arguments: %s\n", err)
		os.Exit(1)
	}

	client, err := connectToServer(cfg)
	if err != nil {
		panic(errors.Wrap(err, "error connecting to the RPC server"))
	}
	defer client.disconnect()

	getCurrentNetworkRequest := appmessage.GetCurrentNetworkRequestMessage{}
	rawRequest, err := protowire.FromAppMessage(&getCurrentNetworkRequest)
	if err != nil {
		panic(err)
	}
	err = client.stream.Send(rawRequest)
	if err != nil {
		panic(err)
	}
	rawResponse, err := client.stream.Recv()
	if err != nil {
		panic(err)
	}
	response, err := rawResponse.ToAppMessage()
	if err != nil {
		panic(err)
	}
	getCurrentNetworkResponse := response.(*appmessage.GetCurrentNetworkResponseMessage)

	log.Infof("Done! %s", getCurrentNetworkResponse.CurrentNetwork)
}
