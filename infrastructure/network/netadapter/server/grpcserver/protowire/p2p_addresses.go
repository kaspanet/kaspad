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

	subnetworkID, err := protoAddresses.SubnetworkID.toDomain()
	if err != nil {
		return nil, err
	}

	addressList := make([]*appmessage.NetAddress, len(protoAddresses.AddressList))
	for i, address := range protoAddresses.AddressList {
		addressList[i], err = address.toAppMessage()
		if err != nil {
			return nil, err
		}
	}
	return &appmessage.MsgAddresses{
		IncludeAllSubnetworks: protoAddresses.IncludeAllSubnetworks,
		SubnetworkID:          subnetworkID,
		AddrList:              addressList,
	}, nil
}

func (x *KaspadMessage_Addresses) fromAppMessage(msgAddresses *appmessage.MsgAddresses) error {
	if len(msgAddresses.AddrList) > appmessage.MaxAddressesPerMsg {
		return errors.Errorf("too many addresses for message "+
			"[count %d, max %d]", len(msgAddresses.AddrList), appmessage.MaxAddressesPerMsg)
	}

	addressList := make([]*NetAddress, len(msgAddresses.AddrList))
	for i, address := range msgAddresses.AddrList {
		addressList[i] = appMessageNetAddressToProto(address)
	}

	x.Addresses = &AddressesMessage{
		IncludeAllSubnetworks: msgAddresses.IncludeAllSubnetworks,
		SubnetworkID:          domainSubnetworkIDToProto(msgAddresses.SubnetworkID),
		AddressList:           addressList,
	}
	return nil
}
