package addressexchange

import (
	"github.com/kaspanet/kaspad/addressmanager"
	"github.com/kaspanet/kaspad/config"
	"github.com/kaspanet/kaspad/domainmessage"
	"github.com/kaspanet/kaspad/netadapter/router"
	"github.com/kaspanet/kaspad/protocol/common"
	peerpkg "github.com/kaspanet/kaspad/protocol/peer"
	"github.com/kaspanet/kaspad/protocol/protocolerrors"
)

// ReceiveAddressesContext is the interface for the context needed for the ReceiveAddresses flow.
type ReceiveAddressesContext interface {
	Config() *config.Config
	AddressManager() *addressmanager.AddressManager
}

// ReceiveAddresses asks a peer for more addresses if needed.
func ReceiveAddresses(context ReceiveAddressesContext, incomingRoute *router.Route, outgoingRoute *router.Route,
	peer *peerpkg.Peer) error {

	if !context.AddressManager().NeedMoreAddresses() {
		return nil
	}

	subnetworkID := peer.SubnetworkID()
	msgGetAddresses := domainmessage.NewMsgRequestAddresses(false, subnetworkID)
	err := outgoingRoute.Enqueue(msgGetAddresses)
	if err != nil {
		return err
	}

	message, err := incomingRoute.DequeueWithTimeout(common.DefaultTimeout)
	if err != nil {
		return err
	}

	msgAddresses := message.(*domainmessage.MsgAddresses)
	if len(msgAddresses.AddrList) > addressmanager.GetAddressesMax {
		return protocolerrors.Errorf(true, "address count excceeded %d", addressmanager.GetAddressesMax)
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

	sourceAddress := peer.Connection().NetAddress()
	context.AddressManager().AddAddresses(msgAddresses.AddrList, sourceAddress, msgAddresses.SubnetworkID)
	return nil
}
