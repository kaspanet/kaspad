package sendversion

import (
	"github.com/kaspanet/kaspad/blockdag"
	"github.com/kaspanet/kaspad/config"
	"github.com/kaspanet/kaspad/netadapter"
	"github.com/kaspanet/kaspad/netadapter/router"
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
func SendVersion(incomingRoute *router.Route, outgoingRoute *router.Route, netAdapter *netadapter.NetAdapter,
	dag *blockdag.BlockDAG) (routeClosed bool, err error) {

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

	isOpen := outgoingRoute.Enqueue(msg)
	if !isOpen {
		return true, nil
	}

	closeChan := make(chan struct{})
	gotVerack := make(chan struct{})
	spawn(func() {
		_, isOpen := incomingRoute.Dequeue()
		if !isOpen {
			closeChan <- struct{}{}
			return
		}
		close(gotVerack)
	})

	const timeout = 30 * time.Second
	select {
	case <-gotVerack:
	case <-closeChan:
		return true, nil
	case <-time.After(timeout):
		return false, errors.New("didn't receive a verack message")
	}
	return false, nil
}
