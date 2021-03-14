package protowire

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/pkg/errors"
)

func (x *KaspadMessage_Addresses) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrap(errorNil, "KaspadMessage_Addresses is nil")
	}
	addressList, err := x.Addresses.toAppMessage()
	if err != nil {
		return nil, err
	}
	return &appmessage.MsgAddresses{
		AddressList: addressList,
	}, nil
}

func (x *AddressesMessage) toAppMessage() ([]*appmessage.NetAddress, error) {
	if x == nil {
		return nil, errors.Wrap(errorNil, "AddressesMessage is nil")
	}

	if len(x.AddressList) > appmessage.MaxAddressesPerMsg {
		return nil, errors.Errorf("too many addresses for message "+
			"[count %d, max %d]", len(x.AddressList), appmessage.MaxAddressesPerMsg)
	}
	addressList := make([]*appmessage.NetAddress, len(x.AddressList))
	for i, address := range x.AddressList {
		var err error
		addressList[i], err = address.toAppMessage()
		if err != nil {
			return nil, err
		}
	}
	return addressList, nil
}

func (x *KaspadMessage_Addresses) fromAppMessage(msgAddresses *appmessage.MsgAddresses) error {
	if len(msgAddresses.AddressList) > appmessage.MaxAddressesPerMsg {
		return errors.Errorf("too many addresses for message "+
			"[count %d, max %d]", len(msgAddresses.AddressList), appmessage.MaxAddressesPerMsg)
	}

	addressList := make([]*NetAddress, len(msgAddresses.AddressList))
	for i, address := range msgAddresses.AddressList {
		addressList[i] = appMessageNetAddressToProto(address)
	}

	x.Addresses = &AddressesMessage{
		AddressList: addressList,
	}
	return nil
}
