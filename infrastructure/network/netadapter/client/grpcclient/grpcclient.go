package grpcclient

import (
	"context"
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/server/grpcserver/protowire"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding/gzip"
	"google.golang.org/protobuf/encoding/protojson"
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

	rpcClient := protowire.NewRPCClient(gRPCConnection)
	stream, err := rpcClient.MessageStream(context.Background(), grpc.UseCompressor(gzip.Name))
	if err != nil {
		return nil, errors.Wrapf(err, "error getting client stream for %s", address)
	}

	return &RPCClient{stream: stream}, nil
}

func (c *RPCClient) Disconnect() error {
	return c.stream.CloseSend()
}

func (c *RPCClient) PostString(requestString string) (string, error) {
	requestBytes := []byte(requestString)
	var parsedRequest protowire.KaspadMessage
	err := protojson.Unmarshal(requestBytes, &parsedRequest)
	if err != nil {
		return "", errors.Wrapf(err, "error parsing the request")
	}
	response, err := c.Post(&parsedRequest)
	if err != nil {
		return "", err
	}
	responseBytes, err := protojson.Marshal(response)
	if err != nil {
		return "", errors.Wrapf(err, "error parsing the response from the RPC server")
	}
	return string(responseBytes), nil
}

func (c *RPCClient) PostAppMessage(requestAppMessage appmessage.Message) (appmessage.Message, error) {
	request, err := protowire.FromAppMessage(requestAppMessage)
	if err != nil {
		return nil, errors.Wrapf(err, "error converting the request")
	}
	response, err := c.Post(request)
	if err != nil {
		return nil, err
	}
	responseAppMessage, err := response.ToAppMessage()
	if err != nil {
		return nil, errors.Wrapf(err, "error converting the response")
	}
	return responseAppMessage, nil
}

func (c *RPCClient) Post(request *protowire.KaspadMessage) (*protowire.KaspadMessage, error) {
	err := c.stream.Send(request)
	if err != nil {
		return nil, errors.Wrapf(err, "error sending the request to the RPC server")
	}
	response, err := c.stream.Recv()
	if err != nil {
		return nil, errors.Wrapf(err, "error receiving the response from the RPC server")
	}
	errorResponse, isErrorResponse := response.Payload.(*protowire.KaspadMessage_RpcError)
	if isErrorResponse {
		return nil, errors.Errorf("received error from RPC: %s", errorResponse.RpcError.Message)
	}
	return response, nil
}

func (c *RPCClient) Send(request *protowire.KaspadMessage) error {
	return c.stream.Send(request)
}

func (c *RPCClient) SendAppMessage(requestAppMessage appmessage.Message) error {
	request, err := protowire.FromAppMessage(requestAppMessage)
	if err != nil {
		return errors.Wrapf(err, "error converting the request")
	}
	return c.Send(request)
}

func (c *RPCClient) Receive() (*protowire.KaspadMessage, error) {
	response, err := c.stream.Recv()
	if err != nil {
		return nil, err
	}
	errorResponse, isErrorResponse := response.Payload.(*protowire.KaspadMessage_RpcError)
	if isErrorResponse {
		return nil, errors.Errorf("received error from RPC: %s", errorResponse.RpcError.Message)
	}
	return response, nil
}

func (c *RPCClient) ReceiveAppMessage() (appmessage.Message, error) {
	response, err := c.Receive()
	if err != nil {
		return nil, errors.Wrapf(err, "error receiving the response from the RPC server")
	}
	return response.ToAppMessage()
}
