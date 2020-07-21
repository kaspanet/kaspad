package handshake

import (
	"github.com/kaspanet/kaspad/netadapter/router"
	"github.com/kaspanet/kaspad/protocol/common"
	peerpkg "github.com/kaspanet/kaspad/protocol/peer"
	"github.com/kaspanet/kaspad/protocol/protocolerrors"
	"github.com/kaspanet/kaspad/wire"
)

var (
	// allowSelfConnections is only used to allow the tests to bypass the self
	// connection detecting and disconnect logic since they intentionally
	// do so for testing purposes.
	allowSelfConnections bool

	// minAcceptableProtocolVersion is the lowest protocol version that a
	// connected peer may support.
	minAcceptableProtocolVersion = wire.ProtocolVersion
)

// ReceiveVersion waits for the peer to send a version message, sends a
// verack in response, and updates its info accordingly.
func ReceiveVersion(context Context, incomingRoute *router.Route, outgoingRoute *router.Route,
	peer *peerpkg.Peer) (*wire.NetAddress, error) {

	message, err := incomingRoute.DequeueWithTimeout(common.DefaultTimeout)
	if err != nil {
		return nil, err
	}

	msgVersion, ok := message.(*wire.MsgVersion)
	if !ok {
		return nil, protocolerrors.New(true, "a version message must precede all others")
	}

	if !allowSelfConnections && context.NetAdapter().ID().IsEqual(msgVersion.ID) {
		return nil, protocolerrors.New(true, "connected to self")
	}

	// Notify and disconnect clients that have a protocol version that is
	// too old.
	//
	// NOTE: If minAcceptableProtocolVersion is raised to be higher than
	// wire.RejectVersion, this should send a reject packet before
	// disconnecting.
	if msgVersion.ProtocolVersion < minAcceptableProtocolVersion {
		//TODO(libp2p) create error type for disconnect but don't ban
		return nil, protocolerrors.Errorf(false, "protocol version must be %d or greater",
			minAcceptableProtocolVersion)
	}

	// Disconnect from partial nodes in networks that don't allow them
	if !context.DAG().Params.EnableNonNativeSubnetworks && msgVersion.SubnetworkID != nil {
		return nil, protocolerrors.New(true, "partial nodes are not allowed")
	}

	// TODO(libp2p)
	//// Disconnect if:
	//// - we are a full node and the outbound connection we've initiated is a partial node
	//// - the remote node is partial and our subnetwork doesn't match their subnetwork
	//localSubnetworkID := config.ActiveConfig().SubnetworkID
	//isLocalNodeFull := localSubnetworkID == nil
	//isRemoteNodeFull := msgVersion.SubnetworkID == nil
	//if (isLocalNodeFull && !isRemoteNodeFull && !connection.IsInbound()) ||
	//	(!isLocalNodeFull && !isRemoteNodeFull && !msgVersion.SubnetworkID.IsEqual(localSubnetworkID)) {
	//
	//	return nil, false, errors.New("incompatible subnetworks")
	//}

	peer.UpdateFieldsFromMsgVersion(msgVersion)
	err = outgoingRoute.Enqueue(wire.NewMsgVerAck())
	if err != nil {
		return nil, err
	}
	// TODO(libp2p) Register peer ID
	return msgVersion.Address, nil
}
