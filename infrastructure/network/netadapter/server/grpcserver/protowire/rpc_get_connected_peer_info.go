package protowire

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/pkg/errors"
)

func (x *KaspadMessage_GetConnectedPeerInfoRequest) toAppMessage() (appmessage.Message, error) {
	return &appmessage.GetConnectedPeerInfoRequestMessage{}, nil
}

func (x *KaspadMessage_GetConnectedPeerInfoRequest) fromAppMessage(_ *appmessage.GetConnectedPeerInfoRequestMessage) error {
	return nil
}

func (x *KaspadMessage_GetConnectedPeerInfoResponse) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_GetConnectedPeerInfoResponse is nil")
	}
	return x.GetConnectedPeerInfoResponse.toAppMessage()
}

func (x *KaspadMessage_GetConnectedPeerInfoResponse) fromAppMessage(message *appmessage.GetConnectedPeerInfoResponseMessage) error {
	var err *RPCError
	if message.Error != nil {
		err = &RPCError{Message: message.Error.Message}
	}
	infos := make([]*GetConnectedPeerInfoMessage, len(message.Infos))
	for i, info := range message.Infos {
		infos[i] = &GetConnectedPeerInfoMessage{
			Id:                        info.ID,
			Address:                   info.Address,
			LastPingDuration:          info.LastPingDuration,
			IsOutbound:                info.IsOutbound,
			TimeOffset:                info.TimeOffset,
			UserAgent:                 info.UserAgent,
			AdvertisedProtocolVersion: info.AdvertisedProtocolVersion,
			TimeConnected:             info.TimeOffset,
			IsIbdPeer:                 info.IsIBDPeer,
		}
	}
	x.GetConnectedPeerInfoResponse = &GetConnectedPeerInfoResponseMessage{
		Infos: infos,
		Error: err,
	}
	return nil
}

func (x *GetConnectedPeerInfoResponseMessage) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "GetConnectedPeerInfoResponseMessage is nil")
	}
	rpcErr, err := x.Error.toAppMessage()
	// Error is an optional field
	if err != nil && !errors.Is(err, errorNil) {
		return nil, err
	}
	// Return verbose data only if there's no error
	if rpcErr != nil && len(x.Infos) != 0 {
		return nil, errors.New("GetConnectedPeerInfoResponseMessage contains both an error and a response")
	}
	infos := make([]*appmessage.GetConnectedPeerInfoMessage, len(x.Infos))
	for i, info := range x.Infos {
		appInfo, err := info.toAppMessage()
		if err != nil {
			return nil, err
		}
		infos[i] = appInfo
	}

	return &appmessage.GetConnectedPeerInfoResponseMessage{
		Infos: infos,
		Error: rpcErr,
	}, nil
}

func (x *GetConnectedPeerInfoMessage) toAppMessage() (*appmessage.GetConnectedPeerInfoMessage, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "GetConnectedPeerInfoMessage is nil")
	}
	return &appmessage.GetConnectedPeerInfoMessage{
		ID:                        x.Id,
		Address:                   x.Address,
		LastPingDuration:          x.LastPingDuration,
		IsOutbound:                x.IsOutbound,
		TimeOffset:                x.TimeOffset,
		UserAgent:                 x.UserAgent,
		AdvertisedProtocolVersion: x.AdvertisedProtocolVersion,
		TimeConnected:             x.TimeOffset,
		IsIBDPeer:                 x.IsIbdPeer,
	}, nil
}
