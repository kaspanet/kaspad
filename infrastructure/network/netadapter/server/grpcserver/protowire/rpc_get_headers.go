package protowire

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/pkg/errors"
)

func (x *KaspadMessage_GetHeadersRequest) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_GetHeadersRequest is nil")
	}
	return x.GetHeadersRequest.toAppMessage()
}

func (x *KaspadMessage_GetHeadersRequest) fromAppMessage(message *appmessage.GetHeadersRequestMessage) error {
	x.GetHeadersRequest = &GetHeadersRequestMessage{
		StartHash:   message.StartHash,
		Limit:       message.Limit,
		IsAscending: message.IsAscending,
	}
	return nil
}

func (x *GetHeadersRequestMessage) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "GetHeadersRequestMessage is nil")
	}
	return &appmessage.GetHeadersRequestMessage{
		StartHash:   x.StartHash,
		Limit:       x.Limit,
		IsAscending: x.IsAscending,
	}, nil
}

func (x *KaspadMessage_GetHeadersResponse) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_GetHeadersResponse is nil")
	}
	return x.GetHeadersResponse.toAppMessage()
}

func (x *KaspadMessage_GetHeadersResponse) fromAppMessage(message *appmessage.GetHeadersResponseMessage) error {
	var err *RPCError
	if message.Error != nil {
		err = &RPCError{Message: message.Error.Message}
	}
	x.GetHeadersResponse = &GetHeadersResponseMessage{
		Headers: message.Headers,
		Error:   err,
	}
	return nil
}

func (x *GetHeadersResponseMessage) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "GetHeadersResponseMessage is nil")
	}
	rpcErr, err := x.Error.toAppMessage()
	// Error is an optional field
	if err != nil && !errors.Is(err, errorNil) {
		return nil, err
	}

	if rpcErr != nil && len(x.Headers) != 0 {
		return nil, errors.New("GetHeadersResponseMessage contains both an error and a response")
	}

	return &appmessage.GetHeadersResponseMessage{
		Headers: x.Headers,
		Error:   rpcErr,
	}, nil
}
