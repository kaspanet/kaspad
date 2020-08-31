package grpcclient

import (
	"context"
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/server/grpcserver/protowire"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding/gzip"
	"time"
)

type RPCClient struct {
	stream protowire.RPC_MessageStreamClient
}

func Connect(address string) (*RPCClient, error) {
	const dialTimeout = 30 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), dialTimeout)
	defer cancel()

	gRPCConnection, err := grpc.DialContext(ctx, address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return nil, errors.Wrapf(err, "error connecting to %s", address)
	}

	grpcClient := protowire.NewRPCClient(gRPCConnection)
	stream, err := grpcClient.MessageStream(context.Background(), grpc.UseCompressor(gzip.Name))
	if err != nil {
		return nil, errors.Wrapf(err, "error getting client stream for %s", address)
	}
	return &RPCClient{stream: stream}, nil
}

func (c *RPCClient) Disconnect() error {
	return c.stream.CloseSend()
}

func (c *RPCClient) AttachRouter(router *router.Router) {
	spawn("RPCClient.AttachRouter-sendLoop", func() {
		for {
			message, err := router.OutgoingRoute().Dequeue()
			if err != nil {
				c.handleError(err)
				return
			}
			err = c.send(message)
			if err != nil {
				c.handleError(err)
				return
			}
		}
	})
	spawn("RPCClient.AttachRouter-receiveLoop", func() {
		for {
			message, err := c.receive()
			if err != nil {
				c.handleError(err)
				return
			}
			err = router.EnqueueIncomingMessage(message)
			if err != nil {
				c.handleError(err)
				return
			}
		}
	})
}

func (c *RPCClient) send(requestAppMessage appmessage.Message) error {
	request, err := protowire.FromAppMessage(requestAppMessage)
	if err != nil {
		return errors.Wrapf(err, "error converting the request")
	}
	return c.stream.Send(request)
}

func (c *RPCClient) receive() (appmessage.Message, error) {
	response, err := c.stream.Recv()
	if err != nil {
		return nil, err
	}
	return response.ToAppMessage()
}

func (c *RPCClient) handleError(err error) {
	panic(err)
}
