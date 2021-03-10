package protowire

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/pkg/errors"
)

func (x *KaspadMessage_RequestAddresses) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_RequestAddresses is nil")
	}
	return x.RequestAddresses.toAppMessage()
}

func (x *RequestAddressesMessage) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "RequestAddressesMessage is nil")
	}
	subnetworkID, err := x.SubnetworkId.toDomain()
	//  Full kaspa nodes set SubnetworkId==nil
	if err != nil && !errors.Is(err, errorNil) {
		return nil, err
	}

	return &appmessage.MsgRequestAddresses{
		IncludeAllSubnetworks: x.IncludeAllSubnetworks,
		SubnetworkID:          subnetworkID,
	}, nil

}

func (x *KaspadMessage_RequestAddresses) fromAppMessage(msgGetAddresses *appmessage.MsgRequestAddresses) error {
	x.RequestAddresses = &RequestAddressesMessage{
		IncludeAllSubnetworks: msgGetAddresses.IncludeAllSubnetworks,
		SubnetworkId:          domainSubnetworkIDToProto(msgGetAddresses.SubnetworkID),
	}
	return nil
}
