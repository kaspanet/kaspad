package protowire

import (
	"github.com/kaspanet/kaspad/domainmessage"
	"github.com/pkg/errors"
)

func (x *KaspadMessage_Addresses) toDomainMessage() (domainmessage.Message, error) {
	protoAddresses := x.Addresses
	if len(x.Addresses.AddressList) > domainmessage.MaxAddressesPerMsg {
		return nil, errors.Errorf("too many addresses for message "+
			"[count %d, max %d]", len(x.Addresses.AddressList), domainmessage.MaxAddressesPerMsg)
	}

	subnetworkID, err := protoAddresses.SubnetworkID.toWire()
	if err != nil {
		return nil, err
	}

	addressList := make([]*domainmessage.NetAddress, len(protoAddresses.AddressList))
	for i, address := range protoAddresses.AddressList {
		addressList[i], err = address.toWire()
		if err != nil {
			return nil, err
		}
	}
	return &domainmessage.MsgAddresses{
		IncludeAllSubnetworks: protoAddresses.IncludeAllSubnetworks,
		SubnetworkID:          subnetworkID,
		AddrList:              addressList,
	}, nil
}

func (x *KaspadMessage_Addresses) fromDomainMessage(msgAddresses *domainmessage.MsgAddresses) error {
	if len(msgAddresses.AddrList) > domainmessage.MaxAddressesPerMsg {
		return errors.Errorf("too many addresses for message "+
			"[count %d, max %d]", len(msgAddresses.AddrList), domainmessage.MaxAddressesPerMsg)
	}

	addressList := make([]*NetAddress, len(msgAddresses.AddrList))
	for i, address := range msgAddresses.AddrList {
		addressList[i] = wireNetAddressToProto(address)
	}

	x.Addresses = &AddressesMessage{
		IncludeAllSubnetworks: msgAddresses.IncludeAllSubnetworks,
		SubnetworkID:          wireSubnetworkIDToProto(msgAddresses.SubnetworkID),
		AddressList:           addressList,
	}
	return nil
}
