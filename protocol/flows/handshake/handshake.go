package handshake

import (
	"sync"

	"github.com/kaspanet/kaspad/addressmanager"
	"github.com/kaspanet/kaspad/protocol/common"
	"github.com/kaspanet/kaspad/protocol/protocolerrors"

	"github.com/kaspanet/kaspad/blockdag"
	"github.com/kaspanet/kaspad/config"
	"github.com/kaspanet/kaspad/netadapter"

	routerpkg "github.com/kaspanet/kaspad/netadapter/router"
	peerpkg "github.com/kaspanet/kaspad/protocol/peer"
	"github.com/kaspanet/kaspad/util/locks"
	"github.com/kaspanet/kaspad/wire"
	"github.com/pkg/errors"
)

// HandleHandshakeContext is the interface for the context needed for the HandleHandshake flow.
type HandleHandshakeContext interface {
	Config() *config.Config
	NetAdapter() *netadapter.NetAdapter
	DAG() *blockdag.BlockDAG
	AddressManager() *addressmanager.AddressManager
	StartIBDIfRequired()
	AddToPeers(peer *peerpkg.Peer) error
	HandleError(err error, flowName string, isStopping *uint32, errChan chan<- error)
}

// HandleHandshake sets up the handshake protocol - It sends a version message and waits for an incoming
// version message, as well as a verack for the sent version
func HandleHandshake(context HandleHandshakeContext, router *routerpkg.Router, netConnection *netadapter.NetConnection,
) (peer *peerpkg.Peer, err error) {

	receiveVersionRoute, err := router.AddIncomingRoute([]wire.MessageCommand{wire.CmdVersion})
	if err != nil {
		return nil, err
	}

	sendVersionRoute, err := router.AddIncomingRoute([]wire.MessageCommand{wire.CmdVerAck})
	if err != nil {
		return nil, err
	}

	// For HandleHandshake to finish, we need to get from the other node
	// a version and verack messages, so we increase the wait group by 2
	// and block HandleHandshake with wg.Wait().
	wg := sync.WaitGroup{}
	wg.Add(2)

	isStopping := uint32(0)
	errChan := make(chan error)

	peer = peerpkg.New(netConnection)

	var peerAddress *wire.NetAddress
	spawn("HandleHandshake-ReceiveVersion", func() {
		defer wg.Done()
		address, err := ReceiveVersion(context, receiveVersionRoute, router.OutgoingRoute(), peer)
		if err != nil {
			context.HandleError(err, "SendVersion", &isStopping, errChan)
			return
		}
		peerAddress = address
	})

	spawn("HandleHandshake-SendVersion", func() {
		defer wg.Done()
		err := SendVersion(context, sendVersionRoute, router.OutgoingRoute())
		if err != nil {
			context.HandleError(err, "SendVersion", &isStopping, errChan)
			return
		}
	})

	select {
	case err := <-errChan:
		if err != nil {
			return nil, err
		}
		return nil, nil
	case <-locks.ReceiveFromChanWhenDone(func() { wg.Wait() }):
	}

	err = context.AddToPeers(peer)
	if err != nil {
		if errors.As(err, &common.ErrPeerWithSameIDExists) {
			return nil, protocolerrors.Wrap(false, err, "peer already exists")
		}
		return nil, err
	}

	if peerAddress != nil {
		subnetworkID := peer.SubnetworkID()
		context.AddressManager().AddAddress(peerAddress, peerAddress, subnetworkID)
		context.AddressManager().Good(peerAddress, subnetworkID)
	}

	context.StartIBDIfRequired()

	err = router.RemoveRoute([]wire.MessageCommand{wire.CmdVersion, wire.CmdVerAck})
	if err != nil {
		return nil, err
	}

	return peer, nil
}
