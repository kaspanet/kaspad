package main

import (
	"context"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/server/grpcserver/protowire"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding/gzip"
	"time"
)

type client struct {
	stream protowire.RPC_MessageStreamClient
}

func connectToServer(cfg *configFlags) (*client, error) {
	log.Infof("Dialing to %s", cfg.RPCServer)

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

	log.Infof("Connected to %s", cfg.RPCServer)

	return &client{stream: stream}, nil
}

func (c *client) disconnect() error {
	return c.stream.CloseSend()
}
