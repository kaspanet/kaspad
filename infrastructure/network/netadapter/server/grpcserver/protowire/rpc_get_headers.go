package protowire

import (
	"github.com/kaspanet/kaspad/app/appmessage"
)

func (x *KaspadMessage_GetHeadersRequest) toAppMessage() (appmessage.Message, error) {
	return &appmessage.GetHeadersRequestMessage{
		StartHash:   x.GetHeadersRequest.StartHash,
		Limit:       x.GetHeadersRequest.Limit,
		IsAscending: x.GetHeadersRequest.IsAscending,
	}, nil
}

func (x *KaspadMessage_GetHeadersRequest) fromAppMessage(message *appmessage.GetHeadersRequestMessage) error {
	x.GetHeadersRequest = &GetHeadersRequestMessage{
		StartHash:   message.StartHash,
		Limit:       message.Limit,
		IsAscending: message.IsAscending,
	}
	return nil
}

func (x *KaspadMessage_GetHeadersResponse) toAppMessage() (appmessage.Message, error) {
	var err *appmessage.RPCError
	if x.GetHeadersResponse.Error != nil {
		err = &appmessage.RPCError{Message: x.GetHeadersResponse.Error.Message}
	}
	return &appmessage.GetHeadersResponseMessage{
		Headers: x.GetHeadersResponse.Headers,
		Error:   err,
	}, nil
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
