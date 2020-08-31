package grpcclient

import (
	"context"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/server/grpcserver/protowire"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding/gzip"
	"time"
)

type RPCClient struct {
	stream protowire.RPC_MessageStreamClient
	router *clientRouter
}

func Connect(address string) (*RPCClient, error) {
	const dialTimeout = 30 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), dialTimeout)
	defer cancel()

	gRPCConnection, err := grpc.DialContext(ctx, address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return nil, errors.Wrapf(err, "error connecting to %s", address)
	}

	rpcClient := &RPCClient{}
	router, err := newClientRouter(rpcClient)
	if err != nil {
		return nil, errors.Wrapf(err, "error creating a router")
	}
	rpcClient.router = router
	router.start()

	grpcClient := protowire.NewRPCClient(gRPCConnection)
	stream, err := grpcClient.MessageStream(context.Background(), grpc.UseCompressor(gzip.Name))
	if err != nil {
		return nil, errors.Wrapf(err, "error getting client stream for %s", address)
	}
	rpcClient.stream = stream

	return rpcClient, nil
}

func (c *RPCClient) Disconnect() error {
	c.router.close()
	return c.stream.CloseSend()
}
