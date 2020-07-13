package receiveversion

import (
	"github.com/kaspanet/kaspad/blockdag"
	"github.com/kaspanet/kaspad/netadapter/router"
	peerpkg "github.com/kaspanet/kaspad/protocol/peer"
	"github.com/kaspanet/kaspad/wire"
	"github.com/pkg/errors"
)

var (
	// allowSelfConns is only used to allow the tests to bypass the self
	// connection detecting and disconnect logic since they intentionally
	// do so for testing purposes.
	allowSelfConns bool

	// minAcceptableProtocolVersion is the lowest protocol version that a
	// connected peer may support.
	minAcceptableProtocolVersion = wire.ProtocolVersion
)

// ReceiveVersion waits for the peer to send a version message, sends a
// verack in response, and updates its info accordingly.
func ReceiveVersion(incomingRoute *router.Route, outgoingRoute *router.Route, peer *peerpkg.Peer,
	dag *blockdag.BlockDAG) (addr *wire.NetAddress, routeClosed bool, err error) {

	msg, isClosed := incomingRoute.Dequeue()
	if isClosed {
		return nil, true, nil
	}

	msgVersion, ok := msg.(*wire.MsgVersion)
	if !ok {
		return nil, false, errors.New("a version message must precede all others")
	}

	if !allowSelfConns && peerpkg.IDExists(msgVersion.ID) {
		//TODO(libp2p) create error type for disconnect but don't ban
		return nil, false, errors.Errorf("already connected to peer with ID %s", msgVersion.ID)
	}

	// Notify and disconnect clients that have a protocol version that is
	// too old.
	//
	// NOTE: If minAcceptableProtocolVersion is raised to be higher than
	// wire.RejectVersion, this should send a reject packet before
	// disconnecting.
	if msgVersion.ProtocolVersion < minAcceptableProtocolVersion {
		//TODO(libp2p) create error type for disconnect but don't ban
		return nil, false, errors.Errorf("protocol version must be %d or greater",
			minAcceptableProtocolVersion)
	}

	// Disconnect from partial nodes in networks that don't allow them
	if !dag.Params.EnableNonNativeSubnetworks && msgVersion.SubnetworkID != nil {
		return nil, false, errors.New("partial nodes are not allowed")
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
	isOpen := outgoingRoute.Enqueue(wire.NewMsgVerAck())
	if !isOpen {
		return nil, true, nil
	}
	// TODO(libp2p) Register peer ID
	return msgVersion.Address, false, nil
}
