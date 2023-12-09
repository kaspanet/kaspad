package main

import (
	"fmt"
	"os"
	"time"

	"github.com/zoomy-network/zoomyd/version"

	"github.com/zoomy-network/zoomyd/infrastructure/network/netadapter/server/grpcserver/protowire"

	"github.com/pkg/errors"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/zoomy-network/zoomyd/infrastructure/network/rpcclient/grpcclient"
)

func main() {
	cfg, err := parseConfig()
	if err != nil {
		printErrorAndExit(fmt.Sprintf("error parsing command-line arguments: %s", err))
	}
	if cfg.ListCommands {
		printAllCommands()
		return
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

	if !cfg.AllowConnectionToDifferentVersions {
		zoomydMessage, err := client.Post(&protowire.ZoomydMessage{Payload: &protowire.ZoomydMessage_GetInfoRequest{GetInfoRequest: &protowire.GetInfoRequestMessage{}}})
		if err != nil {
			printErrorAndExit(fmt.Sprintf("Cannot post GetInfo message: %s", err))
		}

		localVersion := version.Version()
		remoteVersion := zoomydMessage.GetGetInfoResponse().ServerVersion

		if localVersion != remoteVersion {
			printErrorAndExit(fmt.Sprintf("Server version mismatch, expect: %s, got: %s", localVersion, remoteVersion))
		}
	}

	responseChan := make(chan string)

	if cfg.RequestJSON != "" {
		go postJSON(cfg, client, responseChan)
	} else {
		go postCommand(cfg, client, responseChan)
	}

	timeout := time.Duration(cfg.Timeout) * time.Second
	select {
	case responseString := <-responseChan:
		prettyResponseString := prettifyResponse(responseString)
		fmt.Println(prettyResponseString)
	case <-time.After(timeout):
		printErrorAndExit(fmt.Sprintf("timeout of %s has been exceeded", timeout))
	}
}

func printAllCommands() {
	requestDescs := commandDescriptions()
	for _, requestDesc := range requestDescs {
		fmt.Printf("\t%s\n", requestDesc.help())
	}
}

func postCommand(cfg *configFlags, client *grpcclient.GRPCClient, responseChan chan string) {
	message, err := parseCommand(cfg.CommandAndParameters, commandDescriptions())
	if err != nil {
		printErrorAndExit(fmt.Sprintf("error parsing command: %s", err))
	}

	response, err := client.Post(message)
	if err != nil {
		printErrorAndExit(fmt.Sprintf("error posting the request to the RPC server: %s", err))
	}
	responseBytes, err := protojson.MarshalOptions{EmitUnpopulated: true}.Marshal(response)
	if err != nil {
		printErrorAndExit(errors.Wrapf(err, "error parsing the response from the RPC server").Error())
	}

	responseChan <- string(responseBytes)
}

func postJSON(cfg *configFlags, client *grpcclient.GRPCClient, doneChan chan string) {
	responseString, err := client.PostJSON(cfg.RequestJSON)
	if err != nil {
		printErrorAndExit(fmt.Sprintf("error posting the request to the RPC server: %s", err))
	}
	doneChan <- responseString
}

func prettifyResponse(response string) string {
	zoomydMessage := &protowire.ZoomydMessage{}
	err := protojson.Unmarshal([]byte(response), zoomydMessage)
	if err != nil {
		printErrorAndExit(fmt.Sprintf("error parsing the response from the RPC server: %s", err))
	}

	marshalOptions := &protojson.MarshalOptions{}
	marshalOptions.Indent = "    "
	marshalOptions.EmitUnpopulated = true
	return marshalOptions.Format(zoomydMessage)
}

func printErrorAndExit(message string) {
	fmt.Fprintf(os.Stderr, fmt.Sprintf("%s\n", message))
	os.Exit(1)
}
