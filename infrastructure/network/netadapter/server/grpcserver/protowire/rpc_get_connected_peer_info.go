package protowire

import (
	"github.com/kaspanet/kaspad/app/appmessage"
)

func (x *KaspadMessage_GetConnectedPeerInfoRequest) toAppMessage() (appmessage.Message, error) {
	return &appmessage.GetConnectedPeerInfoRequestMessage{}, nil
}

func (x *KaspadMessage_GetConnectedPeerInfoRequest) fromAppMessage(_ *appmessage.GetConnectedPeerInfoRequestMessage) error {
	return nil
}

func (x *KaspadMessage_GetConnectedPeerInfoResponse) toAppMessage() (appmessage.Message, error) {
	var err *appmessage.RPCError
	if x.GetConnectedPeerInfoResponse.Error != nil {
		err = &appmessage.RPCError{Message: x.GetConnectedPeerInfoResponse.Error.Message}
	}
	infos := make([]*appmessage.GetConnectedPeerInfoMessage, len(x.GetConnectedPeerInfoResponse.Infos))
	for i, info := range x.GetConnectedPeerInfoResponse.Infos {
		infos[i] = &appmessage.GetConnectedPeerInfoMessage{
			ID:                        info.Id,
			Address:                   info.Address,
			LastPingDuration:          info.LastPingDuration,
			IsOutbound:                info.IsOutbound,
			TimeOffset:                info.TimeOffset,
			UserAgent:                 info.UserAgent,
			AdvertisedProtocolVersion: info.AdvertisedProtocolVersion,
			TimeConnected:             info.TimeOffset,
			IsIBDPeer:                 info.IsIbdPeer,
		}
	}
	return &appmessage.GetConnectedPeerInfoResponseMessage{
		Infos: infos,
		Error: err,
	}, nil
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
