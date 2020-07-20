package handshake

import (
	"github.com/kaspanet/kaspad/blockdag"
	"github.com/kaspanet/kaspad/config"
	"github.com/kaspanet/kaspad/netadapter"
	"github.com/kaspanet/kaspad/netadapter/router"
	"github.com/kaspanet/kaspad/protocol/common"
	"github.com/kaspanet/kaspad/version"
	"github.com/kaspanet/kaspad/wire"
	"github.com/pkg/errors"
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
func SendVersion(cfg *config.Config, incomingRoute *router.Route, outgoingRoute *router.Route,
	netAdapter *netadapter.NetAdapter, dag *blockdag.BlockDAG) error {

	selectedTipHash := dag.SelectedTipHash()
	subnetworkID := cfg.SubnetworkID

	// Version message.
	localAddress, err := netAdapter.GetBestLocalAddress()
	if err != nil {
		panic(err)
	}
	msg := wire.NewMsgVersion(localAddress, netAdapter.ID(), selectedTipHash, subnetworkID)
	msg.AddUserAgent(userAgentName, userAgentVersion, cfg.UserAgentComments...)

	// Advertise the services flag
	msg.Services = defaultServices

	// Advertise our max supported protocol version.
	msg.ProtocolVersion = wire.ProtocolVersion

	// Advertise if inv messages for transactions are desired.
	msg.DisableRelayTx = cfg.BlocksOnly

	isOpen := outgoingRoute.Enqueue(msg)
	if !isOpen {
		return errors.WithStack(common.ErrRouteClosed)
	}

	_, isOpen, err = incomingRoute.DequeueWithTimeout(common.DefaultTimeout)
	if err != nil {
		return err
	}
	if !isOpen {
		return errors.WithStack(common.ErrRouteClosed)
	}
	return nil
}
