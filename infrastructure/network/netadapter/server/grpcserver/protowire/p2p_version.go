package protowire

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/id"
	"github.com/kaspanet/kaspad/util/mstime"
)

func (x *KaspadMessage_Version) toAppMessage() (appmessage.Message, error) {
	// Address is optional for non-listening nodes
	var address *appmessage.NetAddress
	if x.Version.Address != nil {
		var err error
		address, err = x.Version.Address.toAppMessage()
		if err != nil {
			return nil, err
		}
	}

	selectedTipHash, err := x.Version.SelectedTipHash.toDomain()
	if err != nil {
		return nil, err
	}

	subnetworkID, err := x.Version.SubnetworkId.toDomain()
	if err != nil {
		return nil, err
	}

	err = appmessage.ValidateUserAgent(x.Version.UserAgent)
	if err != nil {
		return nil, err
	}

	return &appmessage.MsgVersion{
		ProtocolVersion: x.Version.ProtocolVersion,
		Network:         x.Version.Network,
		Services:        appmessage.ServiceFlag(x.Version.Services),
		Timestamp:       mstime.UnixMilliseconds(x.Version.Timestamp),
		Address:         address,
		ID:              id.FromBytes(x.Version.Id),
		UserAgent:       x.Version.UserAgent,
		SelectedTipHash: selectedTipHash,
		DisableRelayTx:  x.Version.DisableRelayTx,
		SubnetworkID:    subnetworkID,
	}, nil
}

func (x *KaspadMessage_Version) fromAppMessage(msgVersion *appmessage.MsgVersion) error {
	err := appmessage.ValidateUserAgent(msgVersion.UserAgent)
	if err != nil {
		return err
	}

	versionID, err := msgVersion.ID.SerializeToBytes()
	if err != nil {
		return err
	}

	// Address is optional for non-listening nodes
	var address *NetAddress
	if msgVersion.Address != nil {
		address = appMessageNetAddressToProto(msgVersion.Address)
	}

	x.Version = &VersionMessage{
		ProtocolVersion: msgVersion.ProtocolVersion,
		Network:         msgVersion.Network,
		Services:        uint64(msgVersion.Services),
		Timestamp:       msgVersion.Timestamp.UnixMilliseconds(),
		Address:         address,
		Id:              versionID,
		UserAgent:       msgVersion.UserAgent,
		SelectedTipHash: domainHashToProto(msgVersion.SelectedTipHash),
		DisableRelayTx:  msgVersion.DisableRelayTx,
		SubnetworkId:    domainSubnetworkIDToProto(msgVersion.SubnetworkID),
	}
	return nil
}
