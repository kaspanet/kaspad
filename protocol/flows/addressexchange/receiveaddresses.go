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

type ReceiveAddressesContext interface {
	Config() *config.Config
	AddressManager() *addrmgr.AddrManager
}

// ReceiveAddresses asks a peer for more addresses if needed.
func ReceiveAddresses(context ReceiveAddressesContext, incomingRoute *router.Route, outgoingRoute *router.Route,
	peer *peerpkg.Peer) error {

	if !context.AddressManager().NeedMoreAddresses() {
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
	if !msgAddresses.SubnetworkID.IsEqual(context.Config().SubnetworkID) && msgAddresses.SubnetworkID != nil {
		return protocolerrors.Errorf(false, "only full nodes and %s subnetwork IDs "+
			"are allowed in [%s] command, but got subnetwork ID %s",
			context.Config().SubnetworkID, msgAddresses.Command(), msgAddresses.SubnetworkID)
	}

	// TODO(libp2p) Consider adding to peer known addresses set
	// TODO(libp2p) Replace with real peer IP
	fakeSourceAddress := new(wire.NetAddress)
	context.AddressManager().AddAddresses(msgAddresses.AddrList, fakeSourceAddress, msgAddresses.SubnetworkID)
	return nil
}
