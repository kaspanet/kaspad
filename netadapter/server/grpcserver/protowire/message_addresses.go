package protowire

import (
	"github.com/kaspanet/kaspad/wire"
	"github.com/pkg/errors"
)

func (x *KaspadMessage_Addresses) toWireMessage() (*wire.MsgAddresses, error) {
	protoAddresses := x.Addresses
	if len(x.Addresses.AddressList) > wire.MaxAddressesPerMsg {
		return nil, errors.Errorf("too many addresses for message "+
			"[count %d, max %d]", len(x.Addresses.AddressList), wire.MaxAddressesPerMsg)
	}

	subnetworkID, err := protoAddresses.SubnetworkID.toWire()
	if err != nil {
		return nil, err
	}

	addressList := make([]*wire.NetAddress, len(protoAddresses.AddressList))
	for i, address := range protoAddresses.AddressList {
		addressList[i], err = address.toWire()
		if err != nil {
			return nil, err
		}
	}
	return &wire.MsgAddresses{
		IncludeAllSubnetworks: protoAddresses.IncludeAllSubnetworks,
		SubnetworkID:          subnetworkID,
		AddrList:              addressList,
	}, nil
}

func (x *KaspadMessage_Addresses) fromWireMessage(msgAddresses *wire.MsgAddresses) error {
	if len(msgAddresses.AddrList) > wire.MaxAddressesPerMsg {
		return errors.Errorf("too many addresses for message "+
			"[count %d, max %d]", len(msgAddresses.AddrList), wire.MaxAddressesPerMsg)
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
