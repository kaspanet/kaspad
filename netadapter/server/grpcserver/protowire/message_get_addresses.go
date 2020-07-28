package protowire

import (
	"github.com/kaspanet/kaspad/wire"
)

func (x *KaspadMessage_GetAddresses_) toWireMessage() (*wire.MsgGetAddresses, error) {
	protoGetAddresses := x.GetAddresses_
	subnetworkID, err := protoGetAddresses.SubnetworkID.toWire()
	if err != nil {
		return nil, err
	}

	return &wire.MsgGetAddresses{
		IncludeAllSubnetworks: protoGetAddresses.IncludeAllSubnetworks,
		SubnetworkID:          subnetworkID,
	}, nil
}

func (x *KaspadMessage_GetAddresses_) fromWireMessage(msgGetAddresses *wire.MsgGetAddresses) error {
	x.GetAddresses_ = &GetAddressesMessage{
		IncludeAllSubnetworks: msgGetAddresses.IncludeAllSubnetworks,
		SubnetworkID:          wireSubnetworkIDToProto(msgGetAddresses.SubnetworkID),
	}
	return nil
}
