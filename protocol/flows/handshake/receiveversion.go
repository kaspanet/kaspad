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

type receiveVersionFlow struct {
	HandleHandshakeContext
	incomingRoute, outgoingRoute *router.Route
	peer                         *peerpkg.Peer
}

// ReceiveVersion waits for the peer to send a version message, sends a
// verack in response, and updates its info accordingly.
func ReceiveVersion(context HandleHandshakeContext, incomingRoute *router.Route, outgoingRoute *router.Route,
	peer *peerpkg.Peer) (*wire.NetAddress, error) {

	flow := &receiveVersionFlow{
		HandleHandshakeContext: context,
		incomingRoute:          incomingRoute,
		outgoingRoute:          outgoingRoute,
		peer:                   peer,
	}

	return flow.start()
}

func (flow *receiveVersionFlow) start() (*wire.NetAddress, error) {
	message, err := flow.incomingRoute.DequeueWithTimeout(common.DefaultTimeout)
	if err != nil {
		return nil, err
	}

	msgVersion, ok := message.(*wire.MsgVersion)
	if !ok {
		return nil, protocolerrors.New(true, "a version message must precede all others")
	}

	if !allowSelfConnections && flow.NetAdapter().ID().IsEqual(msgVersion.ID) {
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
	if !flow.DAG().Params.EnableNonNativeSubnetworks && msgVersion.SubnetworkID != nil {
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

	flow.peer.UpdateFieldsFromMsgVersion(msgVersion)
	err = flow.outgoingRoute.Enqueue(wire.NewMsgVerAck())
	if err != nil {
		return nil, err
	}
	return msgVersion.Address, nil
}
