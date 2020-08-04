package protowire

import (
	"github.com/kaspanet/kaspad/netadapter/id"
	"github.com/kaspanet/kaspad/util/mstime"
	"github.com/kaspanet/kaspad/wire"
)

func (x *KaspadMessage_Version) toWireMessage() (wire.Message, error) {
	var address *wire.NetAddress
	if x.Version.Address != nil {
		var err error
		address, err = x.Version.Address.toWire()
		if err != nil {
			return nil, err
		}
	}

	selectedTipHash, err := x.Version.SelectedTipHash.toWire()
	if err != nil {
		return nil, err
	}

	subnetworkID, err := x.Version.SubnetworkID.toWire()
	if err != nil {
		return nil, err
	}

	err = wire.ValidateUserAgent(x.Version.UserAgent)
	if err != nil {
		return nil, err
	}

	return &wire.MsgVersion{
		ProtocolVersion: x.Version.ProtocolVersion,
		Services:        wire.ServiceFlag(x.Version.Services),
		Timestamp:       mstime.UnixMilliseconds(x.Version.Timestamp),
		Address:         address,
		ID:              id.FromBytes(x.Version.Id),
		UserAgent:       x.Version.UserAgent,
		SelectedTipHash: selectedTipHash,
		DisableRelayTx:  x.Version.DisableRelayTx,
		SubnetworkID:    subnetworkID,
	}, nil
}

func (x *KaspadMessage_Version) fromWireMessage(msgVersion *wire.MsgVersion) error {
	err := wire.ValidateUserAgent(msgVersion.UserAgent)
	if err != nil {
		return err
	}

	versionID, err := msgVersion.ID.SerializeToBytes()
	if err != nil {
		return err
	}

	var address *NetAddress
	if msgVersion.Address != nil {
		address = wireNetAddressToProto(msgVersion.Address)
	}

	x.Version = &VersionMessage{
		ProtocolVersion: msgVersion.ProtocolVersion,
		Services:        uint64(msgVersion.Services),
		Timestamp:       msgVersion.Timestamp.UnixMilliseconds(),
		Address:         address,
		Id:              versionID,
		UserAgent:       msgVersion.UserAgent,
		SelectedTipHash: wireHashToProto(msgVersion.SelectedTipHash),
		DisableRelayTx:  msgVersion.DisableRelayTx,
		SubnetworkID:    wireSubnetworkIDToProto(msgVersion.SubnetworkID),
	}
	return nil
}
