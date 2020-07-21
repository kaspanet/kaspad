package addressexchange

import (
	"github.com/kaspanet/kaspad/addrmgr"
	"github.com/kaspanet/kaspad/config"
	"github.com/kaspanet/kaspad/netadapter/router"
	"github.com/kaspanet/kaspad/protocol/common"
	peerpkg "github.com/kaspanet/kaspad/protocol/peer"
	"github.com/kaspanet/kaspad/protocol/protocolerrors"
	"github.com/kaspanet/kaspad/wire"
)

// ReceiveAddresses asks a peer for more addresses if needed.
func ReceiveAddresses(incomingRoute *router.Route, outgoingRoute *router.Route, cfg *config.Config, peer *peerpkg.Peer,
	addressManager *addrmgr.AddrManager) error {

	if !addressManager.NeedMoreAddresses() {
		return nil
	}

	subnetworkID := peer.SubnetworkID()
	msgGetAddresses := wire.NewMsgGetAddresses(false, subnetworkID)
	err := outgoingRoute.Enqueue(msgGetAddresses)
	if err != nil {
		return err
	}

	message, err := incomingRoute.DequeueWithTimeout(common.DefaultTimeout)
	if err != nil {
		return err
	}

	msgAddresses := message.(*wire.MsgAddresses)
	if len(msgAddresses.AddrList) > addrmgr.GetAddressesMax {
		return protocolerrors.Errorf(true, "address count excceeded %d", addrmgr.GetAddressesMax)
	}

	if msgAddresses.IncludeAllSubnetworks {
		return protocolerrors.Errorf(true, "got unexpected "+
			"IncludeAllSubnetworks=true in [%s] command", msgAddresses.Command())
	}
	if !msgAddresses.SubnetworkID.IsEqual(cfg.SubnetworkID) && msgAddresses.SubnetworkID != nil {
		return protocolerrors.Errorf(false, "only full nodes and %s subnetwork IDs "+
			"are allowed in [%s] command, but got subnetwork ID %s",
			cfg.SubnetworkID, msgAddresses.Command(), msgAddresses.SubnetworkID)
	}

	// TODO(libp2p) Consider adding to peer known addresses set
	// TODO(libp2p) Replace with real peer IP
	fakeSourceAddress := new(wire.NetAddress)
	addressManager.AddAddresses(msgAddresses.AddrList, fakeSourceAddress, msgAddresses.SubnetworkID)
	return nil
}
