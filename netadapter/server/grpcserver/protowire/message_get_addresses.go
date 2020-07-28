package protowire

import (
	"github.com/kaspanet/kaspad/util/subnetworkid"
	"github.com/kaspanet/kaspad/wire"
)

func (x *KaspadMessage_GetAddresses_) toWireMessage() (*wire.MsgGetAddresses, error) {
	protoGetAddresses := x.GetAddresses_
	subnetworkID, err := subnetworkid.New(protoGetAddresses.SubnetworkID.Bytes)
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
		SubnetworkID: &SubnetworkID{
			Bytes: msgGetAddresses.SubnetworkID.CloneBytes(),
		},
	}
	return nil
}
