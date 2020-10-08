package protowire

import "github.com/kaspanet/kaspad/app/appmessage"

func (x *KaspadMessage_GetPeerAddressesRequest) toAppMessage() (appmessage.Message, error) {
	return &appmessage.GetPeerAddressesRequestMessage{}, nil
}

func (x *KaspadMessage_GetPeerAddressesRequest) fromAppMessage(_ *appmessage.GetPeerAddressesRequestMessage) error {
	return nil
}

func (x *KaspadMessage_GetPeerAddressesResponse) toAppMessage() (appmessage.Message, error) {
	var err *appmessage.RPCError
	if x.GetPeerAddressesResponse.Error != nil {
		err = &appmessage.RPCError{Message: x.GetPeerAddressesResponse.Error.Message}
	}
	addresses := make([]*appmessage.GetPeerAddressesKnownAddressMessage, len(x.GetPeerAddressesResponse.Addresses))
	for i, address := range x.GetPeerAddressesResponse.Addresses {
		addresses[i] = &appmessage.GetPeerAddressesKnownAddressMessage{Addr: address.Addr}
	}
	bannedAddresses := make([]*appmessage.GetPeerAddressesKnownAddressMessage, len(x.GetPeerAddressesResponse.BannedAddresses))
	for i, address := range x.GetPeerAddressesResponse.BannedAddresses {
		bannedAddresses[i] = &appmessage.GetPeerAddressesKnownAddressMessage{Addr: address.Addr}
	}
	return &appmessage.GetPeerAddressesResponseMessage{
		Addresses:       addresses,
		BannedAddresses: bannedAddresses,
		Error:           err,
	}, nil
}

func (x *KaspadMessage_GetPeerAddressesResponse) fromAppMessage(message *appmessage.GetPeerAddressesResponseMessage) error {
	var err *RPCError
	if message.Error != nil {
		err = &RPCError{Message: message.Error.Message}
	}
	addresses := make([]*GetPeerAddressesKnownAddressMessage, len(message.Addresses))
	for i, address := range message.Addresses {
		addresses[i] = &GetPeerAddressesKnownAddressMessage{Addr: address.Addr}
	}
	bannedAddresses := make([]*GetPeerAddressesKnownAddressMessage, len(message.BannedAddresses))
	for i, address := range message.BannedAddresses {
		bannedAddresses[i] = &GetPeerAddressesKnownAddressMessage{Addr: address.Addr}
	}
	x.GetPeerAddressesResponse = &GetPeerAddressesResponseMessage{
		Addresses:       addresses,
		BannedAddresses: bannedAddresses,
		Error:           err,
	}
	return nil
}
