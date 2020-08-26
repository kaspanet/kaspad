package main

import (
	"context"
	"fmt"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/server/grpcserver/protowire"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding/gzip"
	"google.golang.org/protobuf/encoding/protojson"
	"time"
)

type client struct {
	stream protowire.RPC_MessageStreamClient
}

func connectToServer(cfg *configFlags) (*client, error) {
	const dialTimeout = 30 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), dialTimeout)
	defer cancel()

	gRPCConnection, err := grpc.DialContext(ctx, cfg.RPCServer, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return nil, errors.Wrapf(err, "error connecting to %s", cfg.RPCServer)
	}

	rpcClient := protowire.NewRPCClient(gRPCConnection)
	stream, err := rpcClient.MessageStream(context.Background(), grpc.UseCompressor(gzip.Name))
	if err != nil {
		return nil, errors.Wrapf(err, "error getting client stream for %s", cfg.RPCServer)
	}

	return &client{stream: stream}, nil
}

func (c *client) disconnect() error {
	return c.stream.CloseSend()
}

func (c *client) post(requestString string) string {
	requestBytes := []byte(requestString)
	var parsedRequest protowire.KaspadMessage
	err := protojson.Unmarshal(requestBytes, &parsedRequest)
	if err != nil {
		printErrorAndExit(fmt.Sprintf("error parsing the request: %s", err))
	}
	err = c.stream.Send(&parsedRequest)
	if err != nil {
		printErrorAndExit(fmt.Sprintf("error sending the request to the RPC server: %s", err))
	}
	response, err := c.stream.Recv()
	if err != nil {
		printErrorAndExit(fmt.Sprintf("error receiving the response from the RPC server: %s", err))
	}
	responseBytes, err := protojson.Marshal(response)
	if err != nil {
		printErrorAndExit(fmt.Sprintf("error parsing the response from the RPC server: %s", err))
	}
	return string(responseBytes)
}
