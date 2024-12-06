package protowire

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/pkg/errors"
)

func (x *KaspadMessage_GetInfoRequest) toAppMessage() (appmessage.Message, error) {
	return &appmessage.GetInfoRequestMessage{}, nil
}

func (x *KaspadMessage_GetInfoRequest) fromAppMessage(_ *appmessage.GetInfoRequestMessage) error {
	x.GetInfoRequest = &GetInfoRequestMessage{}
	return nil
}

func (x *KaspadMessage_GetInfoResponse) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_GetInfoResponse is nil")
	}
	return x.GetInfoResponse.toAppMessage()
}

func (x *KaspadMessage_GetInfoResponse) fromAppMessage(message *appmessage.GetInfoResponseMessage) error {
	var err *RPCError
	if message.Error != nil {
		err = &RPCError{Message: message.Error.Message}
	}
	x.GetInfoResponse = &GetInfoResponseMessage{
		P2PId:         message.P2PID,
		ServerVersion: message.ServerVersion,
		MempoolSize:   message.MempoolSize,

		IsUtxoIndexed: message.IsUtxoIndexed,
		IsSynced:      message.IsSynced,

		MaxRPCClients:          message.MaxRPCClients,
		NumberOfRPCConnections: message.NumberOfRPCConnections,
		MaxP2PClients:          message.MaxP2PClients,
		NumberOfP2PConnections: message.NumberOfP2PConnections,

		BanDurationInMilliseconds: message.BanDurationInMilliseconds,

		UptimeInMilliseconds: message.UptimeInMilliseconds,

		Error: err,
	}
	return nil
}

func (x *GetInfoResponseMessage) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "GetInfoResponseMessage is nil")
	}
	rpcErr, err := x.Error.toAppMessage()
	// Error is an optional field
	if err != nil && !errors.Is(err, errorNil) {
		return nil, err
	}

	if rpcErr != nil && len(x.P2PId) != 0 {
		return nil, errors.New("GetInfoResponseMessage contains both an error and a response")
	}

	return &appmessage.GetInfoResponseMessage{
		P2PID:         x.P2PId,
		MempoolSize:   x.MempoolSize,
		ServerVersion: x.ServerVersion,

		IsUtxoIndexed: x.IsUtxoIndexed,
		IsSynced:      x.IsSynced,

		MaxRPCClients:          x.MaxRPCClients,
		NumberOfRPCConnections: x.NumberOfRPCConnections,

		MaxP2PClients:          x.MaxP2PClients,
		NumberOfP2PConnections: x.NumberOfP2PConnections,

		BanDurationInMilliseconds: x.BanDurationInMilliseconds,
		UptimeInMilliseconds:      x.UptimeInMilliseconds,

		Error: rpcErr,
	}, nil
}
