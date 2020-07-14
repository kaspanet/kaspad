package receiveversion

import (
	"github.com/kaspanet/kaspad/blockdag"
	"github.com/kaspanet/kaspad/netadapter"
	"github.com/kaspanet/kaspad/netadapter/router"
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
func ReceiveVersion(incomingRoute *router.Route, outgoingRoute *router.Route, netAdapter *netadapter.NetAdapter,
	peer *peerpkg.Peer, dag *blockdag.BlockDAG) (addr *wire.NetAddress, routeClosed bool, err error) {

	message, isOpen := incomingRoute.Dequeue()
	if !isOpen {
		return nil, true, nil
	}

	msgVersion, ok := message.(*wire.MsgVersion)
	if !ok {
		return nil, false, protocolerrors.New(true, "a version message must precede all others")
	}

	if !allowSelfConnections && netAdapter.ID().IsEqual(msgVersion.ID) {
		return nil, false, protocolerrors.New(true, "connected to self")
	}

	// Notify and disconnect clients that have a protocol version that is
	// too old.
	//
	// NOTE: If minAcceptableProtocolVersion is raised to be higher than
	// wire.RejectVersion, this should send a reject packet before
	// disconnecting.
	if msgVersion.ProtocolVersion < minAcceptableProtocolVersion {
		//TODO(libp2p) create error type for disconnect but don't ban
		return nil, false, protocolerrors.Errorf(false, "protocol version must be %d or greater",
			minAcceptableProtocolVersion)
	}

	// Disconnect from partial nodes in networks that don't allow them
	if !dag.Params.EnableNonNativeSubnetworks && msgVersion.SubnetworkID != nil {
		return nil, false, protocolerrors.New(true, "partial nodes are not allowed")
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
	isOpen = outgoingRoute.Enqueue(wire.NewMsgVerAck())
	if !isOpen {
		return nil, true, nil
	}
	// TODO(libp2p) Register peer ID
	return msgVersion.Address, false, nil
}
