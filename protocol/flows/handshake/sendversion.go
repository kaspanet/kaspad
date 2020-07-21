package handshake

import (
	"github.com/kaspanet/kaspad/netadapter/router"
	"github.com/kaspanet/kaspad/protocol/common"
	"github.com/kaspanet/kaspad/version"
	"github.com/kaspanet/kaspad/wire"
)

var (
	// userAgentName is the user agent name and is used to help identify
	// ourselves to other kaspa peers.
	userAgentName = "kaspad"

	// userAgentVersion is the user agent version and is used to help
	// identify ourselves to other kaspa peers.
	userAgentVersion = version.Version()

	// defaultServices describes the default services that are supported by
	// the server.
	defaultServices = wire.SFNodeNetwork | wire.SFNodeBloom | wire.SFNodeCF

	// defaultRequiredServices describes the default services that are
	// required to be supported by outbound peers.
	defaultRequiredServices = wire.SFNodeNetwork
)

// SendVersion sends a version to a peer and waits for verack.
func SendVersion(context Context, incomingRoute *router.Route, outgoingRoute *router.Route) error {

	selectedTipHash := context.DAG().SelectedTipHash()
	subnetworkID := context.Config().SubnetworkID

	// Version message.
	localAddress, err := context.NetAdapter().GetBestLocalAddress()
	if err != nil {
		panic(err)
	}
	msg := wire.NewMsgVersion(localAddress, context.NetAdapter().ID(), selectedTipHash, subnetworkID)
	msg.AddUserAgent(userAgentName, userAgentVersion, context.Config().UserAgentComments...)

	// Advertise the services flag
	msg.Services = defaultServices

	// Advertise our max supported protocol version.
	msg.ProtocolVersion = wire.ProtocolVersion

	// Advertise if inv messages for transactions are desired.
	msg.DisableRelayTx = context.Config().BlocksOnly

	err = outgoingRoute.Enqueue(msg)
	if err != nil {
		return err
	}

	// Wait for verack
	_, err = incomingRoute.DequeueWithTimeout(common.DefaultTimeout)
	if err != nil {
		return err
	}
	return nil
}
