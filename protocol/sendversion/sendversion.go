package sendversion

import (
	"github.com/kaspanet/kaspad/blockdag"
	"github.com/kaspanet/kaspad/config"
	"github.com/kaspanet/kaspad/netadapter"
	"github.com/kaspanet/kaspad/version"
	"github.com/kaspanet/kaspad/wire"
	"github.com/pkg/errors"
	"time"
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
func SendVersion(msgChan <-chan wire.Message, router *netadapter.Router, netAdapter *netadapter.NetAdapter,
	dag *blockdag.BlockDAG) error {

	selectedTipHash := dag.SelectedTipHash()
	subnetworkID := config.ActiveConfig().SubnetworkID

	// Version message.
	msg := wire.NewMsgVersion(netAdapter.GetBestLocalAddress(), netAdapter.ID(), selectedTipHash, subnetworkID)
	msg.AddUserAgent(userAgentName, userAgentVersion, config.ActiveConfig().UserAgentComments...)

	// Advertise the services flag
	msg.Services = defaultServices

	// Advertise our max supported protocol version.
	msg.ProtocolVersion = wire.ProtocolVersion

	// Advertise if inv messages for transactions are desired.
	msg.DisableRelayTx = config.ActiveConfig().BlocksOnly

	router.WriteOutgoingMessage(msg)
	const timeout = 30 * time.Second
	select {
	case <-msgChan:
	case <-time.After(timeout):
		return errors.New("didn't receive a verack message")
	}
	return nil
}
