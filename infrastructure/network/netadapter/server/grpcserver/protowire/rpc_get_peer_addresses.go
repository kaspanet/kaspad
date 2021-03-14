package protowire

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/pkg/errors"
)

func (x *KaspadMessage_GetPeerAddressesRequest) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_GetPeerAddressesRequest is nil")
	}
	return &appmessage.GetPeerAddressesRequestMessage{}, nil
}

func (x *KaspadMessage_GetPeerAddressesRequest) fromAppMessage(_ *appmessage.GetPeerAddressesRequestMessage) error {
	return nil
}

func (x *KaspadMessage_GetPeerAddressesResponse) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_GetPeerAddressesResponse is nil")
	}
	return x.GetPeerAddressesResponse.toAppMessage()
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

func (x *GetPeerAddressesResponseMessage) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "GetPeerAddressesResponseMessage is nil")
	}
	rpcErr, err := x.Error.toAppMessage()
	// Error is an optional field
	if err != nil && !errors.Is(err, errorNil) {
		return nil, err
	}

	addresses := make([]*appmessage.GetPeerAddressesKnownAddressMessage, len(x.Addresses))
	for i, address := range x.Addresses {
		appAddress, err := address.toAppMessage()
		if err != nil {
			return nil, err
		}
		addresses[i] = appAddress
	}
	bannedAddresses := make([]*appmessage.GetPeerAddressesKnownAddressMessage, len(x.BannedAddresses))
	for i, address := range x.BannedAddresses {
		bannedAddress, err := address.toAppMessage()
		if err != nil {
			return nil, err
		}
		bannedAddresses[i] = bannedAddress
	}

	if rpcErr != nil && (len(addresses) != 0 || len(bannedAddresses) != 0) {
		return nil, errors.New("GetPeerAddressesResponseMessage contains both an error and a response")
	}
	return &appmessage.GetPeerAddressesResponseMessage{
		Addresses:       addresses,
		BannedAddresses: bannedAddresses,
		Error:           rpcErr,
	}, nil
}

func (x *GetPeerAddressesKnownAddressMessage) toAppMessage() (*appmessage.GetPeerAddressesKnownAddressMessage, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "GetPeerAddressesKnownAddressMessage is nil")
	}
	return &appmessage.GetPeerAddressesKnownAddressMessage{Addr: x.Addr}, nil
}
