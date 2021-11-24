package handshake

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/protocol/common"
	peerpkg "github.com/kaspanet/kaspad/app/protocol/peer"
	"github.com/kaspanet/kaspad/app/protocol/protocolerrors"
	"github.com/kaspanet/kaspad/infrastructure/logger"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
)

var (
	// allowSelfConnections is only used to allow the tests to bypass the self
	// connection detecting and disconnect logic since they intentionally
	// do so for testing purposes.
	allowSelfConnections bool

	// minAcceptableProtocolVersion is the lowest protocol version that a
	// connected peer may support.
	minAcceptableProtocolVersion = appmessage.ProtocolVersion
)

type receiveVersionFlow struct {
	HandleHandshakeContext
	incomingRoute, outgoingRoute *router.Route
	peer                         *peerpkg.Peer
}

// ReceiveVersion waits for the peer to send a version message, sends a
// verack in response, and updates its info accordingly.
func ReceiveVersion(context HandleHandshakeContext, incomingRoute *router.Route, outgoingRoute *router.Route,
	peer *peerpkg.Peer) (*appmessage.NetAddress, error) {

	flow := &receiveVersionFlow{
		HandleHandshakeContext: context,
		incomingRoute:          incomingRoute,
		outgoingRoute:          outgoingRoute,
		peer:                   peer,
	}

	return flow.start()
}

func (flow *receiveVersionFlow) start() (*appmessage.NetAddress, error) {
	onEnd := logger.LogAndMeasureExecutionTime(log, "receiveVersionFlow.start")
	defer onEnd()

	log.Debugf("Starting receiveVersionFlow with %s", flow.peer.Address())

	message, err := flow.incomingRoute.DequeueWithTimeout(common.DefaultTimeout)
	if err != nil {
		return nil, err
	}

	log.Debugf("Got version message")

	msgVersion, ok := message.(*appmessage.MsgVersion)
	if !ok {
		return nil, protocolerrors.New(true, "a version message must precede all others")
	}

	if !allowSelfConnections && flow.NetAdapter().ID().IsEqual(msgVersion.ID) {
		return nil, protocolerrors.New(false, "connected to self")
	}

	// Disconnect and ban peers from a different network
	if msgVersion.Network != flow.Config().ActiveNetParams.Name {
		return nil, protocolerrors.Errorf(true, "wrong network")
	}

	// Notify and disconnect clients that have a protocol version that is
	// too old.
	//
	// NOTE: If minAcceptableProtocolVersion is raised to be higher than
	// appmessage.RejectVersion, this should send a reject packet before
	// disconnecting.
	if msgVersion.ProtocolVersion < minAcceptableProtocolVersion {
		return nil, protocolerrors.Errorf(false, "protocol version must be %d or greater",
			minAcceptableProtocolVersion)
	}

	// Disconnect from partial nodes in networks that don't allow them
	if !flow.Config().ActiveNetParams.EnableNonNativeSubnetworks && msgVersion.SubnetworkID != nil {
		return nil, protocolerrors.New(true, "partial nodes are not allowed")
	}

	// Disconnect if:
	// - we are a full node and the outbound connection we've initiated is a partial node
	// - the remote node is partial and our subnetwork doesn't match their subnetwork
	localSubnetworkID := flow.Config().SubnetworkID
	isLocalNodeFull := localSubnetworkID == nil
	isRemoteNodeFull := msgVersion.SubnetworkID == nil
	isOutbound := flow.peer.Connection().IsOutbound()
	if (isLocalNodeFull && !isRemoteNodeFull && isOutbound) ||
		(!isLocalNodeFull && !isRemoteNodeFull && !msgVersion.SubnetworkID.Equal(localSubnetworkID)) {

		return nil, protocolerrors.New(false, "incompatible subnetworks")
	}

	flow.peer.UpdateFieldsFromMsgVersion(msgVersion)
	err = flow.outgoingRoute.Enqueue(appmessage.NewMsgVerAck())
	if err != nil {
		return nil, err
	}

	flow.peer.Connection().SetID(msgVersion.ID)
	if msgVersion.Banner != "" {
		log.Infof("Got peer’s banner: %s\n", msgVersion.Banner)
	}

	return msgVersion.Address, nil
}
