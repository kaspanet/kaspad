package addressexchange

import (
	"time"

	"github.com/kaspanet/kaspad/addrmgr"
	"github.com/kaspanet/kaspad/config"
	"github.com/kaspanet/kaspad/netadapter/router"
	peerpkg "github.com/kaspanet/kaspad/protocol/peer"
	"github.com/kaspanet/kaspad/protocol/protocolerrors"
	"github.com/kaspanet/kaspad/wire"
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

	msgAddresses := message.(*wire.MsgAddresses)
	if len(msgAddresses.AddrList) > addrmgr.GetAddressesMax {
		return false, protocolerrors.Errorf(true, "address count excceeded %d", addrmgr.GetAddressesMax)
	}

	if msgAddresses.IncludeAllSubnetworks {
		return false, protocolerrors.Errorf(true, "got unexpected "+
			"IncludeAllSubnetworks=true in [%s] command", msgAddresses.Command())
	}
	if !msgAddresses.SubnetworkID.IsEqual(config.ActiveConfig().SubnetworkID) && msgAddresses.SubnetworkID != nil {
		return false, protocolerrors.Errorf(false, "only full nodes and %s subnetwork IDs "+
			"are allowed in [%s] command, but got subnetwork ID %s",
			config.ActiveConfig().SubnetworkID, msgAddresses.Command(), msgAddresses.SubnetworkID)
	}

	// TODO(libp2p) Consider adding to peer known addresses set
	// TODO(libp2p) Replace with real peer IP
	fakeSourceAddress := new(wire.NetAddress)
	addressManager.AddAddresses(msgAddresses.AddrList, fakeSourceAddress, msgAddresses.SubnetworkID)
	return false, nil
}
