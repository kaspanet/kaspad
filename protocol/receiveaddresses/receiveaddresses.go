package receiveaddresses

import (
	"github.com/kaspanet/kaspad/addrmgr"
	"github.com/kaspanet/kaspad/config"
	"github.com/kaspanet/kaspad/netadapter/router"
	peerpkg "github.com/kaspanet/kaspad/protocol/peer"
	"github.com/kaspanet/kaspad/protocol/protocolerrors"
	"github.com/kaspanet/kaspad/wire"
	"time"
)

const timeout = 30 * time.Second

// ReceiveAddresses asks a peer for more addresses if needed.
func ReceiveAddresses(incomingRoute *router.Route, outgoingRoute *router.Route,
	peer *peerpkg.Peer, addressManager *addrmgr.AddrManager) (routeClosed bool, err error) {

	subnetworkID, err := peer.SubnetworkID()
	if err != nil {
		panic(err)
	}

	msgGetAddresses := wire.NewMsgGetAddresses(false, subnetworkID)
	isOpen, err := outgoingRoute.EnqueueWithTimeout(msgGetAddresses, timeout)
	if err != nil {
		return false, err
	}
	if !isOpen {
		return true, nil
	}

	if addressManager.NeedMoreAddresses() {
		return false, nil
	}

	message, isOpen, err := incomingRoute.DequeueWithTimeout(timeout)
	if err != nil {
		return false, err
	}
	if !isOpen {
		return true, nil
	}

	msgAddr := message.(*wire.MsgAddr)
	if len(msgAddr.AddrList) > addrmgr.GetAddrMax {
		return false, protocolerrors.Errorf(true, "address count excceeded %d", addrmgr.GetAddrMax)
	}

	if msgAddr.IncludeAllSubnetworks {
		return false, protocolerrors.Errorf(true, "got unexpected "+
			"IncludeAllSubnetworks=true in [%s] command", msgAddr.Command())
	}
	if !msgAddr.SubnetworkID.IsEqual(config.ActiveConfig().SubnetworkID) && msgAddr.SubnetworkID != nil {
		return false, protocolerrors.Errorf(false, "only full nodes and %s subnetwork IDs "+
			"are allowed in [%s] command, but got subnetwork ID %s",
			config.ActiveConfig().SubnetworkID, msgAddr.Command(), msgAddr.SubnetworkID)
	}

	// TODO(libp2p) Consider adding to peer known addresses set
	// TODO(libp2p) Replace with real peer IP
	fakeSrcAddr := new(wire.NetAddress)
	addressManager.AddAddresses(msgAddr.AddrList, fakeSrcAddr, msgAddr.SubnetworkID)
	return false, nil
}
