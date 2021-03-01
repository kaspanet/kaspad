package grpcclient

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/server/grpcserver/protowire"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/encoding/protojson"
)

// PostJSON is a helper function that converts the given requestJSON
// to protobuf, sends it to the RPC server, accepts the first response
// that arrives back, and returns the response as JSON
func (c *GRPCClient) PostJSON(requestJSON string) (string, error) {
	requestBytes := []byte(requestJSON)
	var parsedRequest protowire.KaspadMessage
	err := protojson.Unmarshal(requestBytes, &parsedRequest)
	if err != nil {
		return "", errors.Wrapf(err, "error parsing the request")
	}
	response, err := c.Post(&parsedRequest)
	if err != nil {
		return "", err
	}
	responseBytes, err := protojson.MarshalOptions{EmitUnpopulated: true}.Marshal(response)
	if err != nil {
		return "", errors.Wrapf(err, "error parsing the response from the RPC server")
	}
	return string(responseBytes), nil
}

// PostAppMessage is a helper function that converts the given
// requestAppMessage to protobuf, sends it to the RPC server,
// accepts the first response that arrives back, and returns the
// response as an appMessage
func (c *GRPCClient) PostAppMessage(requestAppMessage appmessage.Message) (appmessage.Message, error) {
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

// Post is a helper function that sends the given request to the
// RPC server, accepts the first response that arrives back, and
// returns the response
func (c *GRPCClient) Post(request *protowire.KaspadMessage) (*protowire.KaspadMessage, error) {
	err := c.stream.Send(request)
	if err != nil {
		return nil, errors.Wrapf(err, "error sending the request to the RPC server")
	}
	response, err := c.stream.Recv()
	if err != nil {
		return nil, errors.Wrapf(err, "error receiving the response from the RPC server")
	}
	return response, nil
}
