package sendversion

import (
	"github.com/kaspanet/kaspad/blockdag"
	"github.com/kaspanet/kaspad/config"
	"github.com/kaspanet/kaspad/p2pserver"
	peerpkg "github.com/kaspanet/kaspad/protocol/peer"
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

func SendVersion(msgChan <-chan wire.Message, connection p2pserver.Connection, peer *peerpkg.Peer,
	dag *blockdag.BlockDAG) error {

	selectedTipHash := dag.SelectedTipHash()
	subnetworkID := config.ActiveConfig().SubnetworkID

	// Version message.
	msg := wire.NewMsgVersion(versionNonce, selectedTipHash, subnetworkID)
	err := msg.AddUserAgent(userAgentName, userAgentVersion,
		config.ActiveConfig().UserAgentComments...)
	if err != nil {
		panic(errors.Wrapf(err, "error with our own user agent"))
	}

	// Advertise the services flag
	msg.Services = defaultServices

	// Advertise our max supported protocol version.
	msg.ProtocolVersion = wire.ProtocolVersion

	// Advertise if inv messages for transactions are desired.
	msg.DisableRelayTx = config.ActiveConfig().BlocksOnly

	const stallResponseTimeout = 30 * time.Second
	select {
	case <-msgChan:
	case <-time.After(stallResponseTimeout):
		return errors.New("didn't receive a verack message")
	}
	return nil
}
