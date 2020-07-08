package receiveversion

import (
	"fmt"
	"github.com/kaspanet/kaspad/blockdag"
	"github.com/kaspanet/kaspad/config"
	"github.com/kaspanet/kaspad/p2pserver"
	protocolcommon "github.com/kaspanet/kaspad/protocol/common"
	peerpkg "github.com/kaspanet/kaspad/protocol/peer"
	"github.com/kaspanet/kaspad/util/random"
	"github.com/kaspanet/kaspad/wire"
	"github.com/pkg/errors"
	"sync/atomic"
)

var (
	versionNonce uint64

	// allowSelfConns is only used to allow the tests to bypass the self
	// connection detecting and disconnect logic since they intentionally
	// do so for testing purposes.
	allowSelfConns bool

	// minAcceptableProtocolVersion is the lowest protocol version that a
	// connected peer may support.
	minAcceptableProtocolVersion = wire.ProtocolVersion

	nodeCount uint32
)

func ReceiveVersion(msgChan <-chan wire.Message, connection p2pserver.Connection, peer *peerpkg.Peer,
	dag *blockdag.BlockDAG) (addr *wire.NetAddress, channelClosed bool, err error) {
	msg, isClosed := <-msgChan
	if isClosed {
		return nil, true, nil
	}

	msgVersion, ok := msg.(*wire.MsgVersion)
	if !ok {
		errStr := "a version message must precede all others"
		protocolcommon.AddBanScoreAndPushRejectMsg(connection,
			msg.Command(),
			wire.RejectNotRequested,
			nil,
			peerpkg.BanScoreNonVersionFirstMessage,
			0,
			errStr)
		return nil, false, errors.New(errStr)
	}

	if !allowSelfConns && msgVersion.Nonce == versionNonce {
		return nil, false, errors.New("disconnecting peer connected to self")
	}

	// Notify and disconnect clients that have a protocol version that is
	// too old.
	//
	// NOTE: If minAcceptableProtocolVersion is raised to be higher than
	// wire.RejectVersion, this should send a reject packet before
	// disconnecting.
	if msgVersion.ProtocolVersion < minAcceptableProtocolVersion {
		reason := fmt.Sprintf("protocol version must be %d or greater",
			minAcceptableProtocolVersion)
		protocolcommon.PushRejectMsg(connection, msgVersion.Command(), wire.RejectObsolete, reason, nil)
		return nil, false, errors.New(reason)
	}

	// Disconnect from partial nodes in networks that don't allow them
	if !dag.Params.EnableNonNativeSubnetworks && msgVersion.SubnetworkID != nil {
		return nil, false, errors.New("partial nodes are not allowed")
	}

	// Disconnect if:
	// - we are a full node and the outbound connection we've initiated is a partial node
	// - the remote node is partial and our subnetwork doesn't match their subnetwork
	localSubnetworkID := config.ActiveConfig().SubnetworkID
	isLocalNodeFull := localSubnetworkID == nil
	isRemoteNodeFull := msgVersion.SubnetworkID == nil
	if (isLocalNodeFull && !isRemoteNodeFull && !connection.IsInbound()) ||
		(!isLocalNodeFull && !isRemoteNodeFull && !msgVersion.SubnetworkID.IsEqual(localSubnetworkID)) {

		return nil, false, errors.New("incompatible subnetworks")
	}

	peer.UpdateFlagsFromVersionMsg(msgVersion, atomic.AddUint32(&nodeCount, 1))
	connection.Send(wire.NewMsgVerAck())
	return msgVersion.Address, false, nil
}

func init() {
	var err error
	versionNonce, err = random.Uint64()
	if err != nil {
		panic(errors.Wrap(err, "error initializing version nonce"))
	}
}
