package grpcclient

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/server/grpcserver/protowire"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/encoding/protojson"
)

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
