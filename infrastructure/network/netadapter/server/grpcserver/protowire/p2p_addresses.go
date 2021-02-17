package protowire

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/pkg/errors"
)

func (x *KaspadMessage_Addresses) toAppMessage() (appmessage.Message, error) {
	protoAddresses := x.Addresses
	if len(x.Addresses.AddressList) > appmessage.MaxAddressesPerMsg {
		return nil, errors.Errorf("too many addresses for message "+
			"[count %d, max %d]", len(x.Addresses.AddressList), appmessage.MaxAddressesPerMsg)
	}

	addressList := make([]*appmessage.NetAddress, len(protoAddresses.AddressList))
	for i, address := range protoAddresses.AddressList {
		var err error
		addressList[i], err = address.toAppMessage()
		if err != nil {
			return nil, err
		}
	}
	return &appmessage.MsgAddresses{
		AddressList: addressList,
	}, nil
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
