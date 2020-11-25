package handshake

import (
	"sync/atomic"

	"github.com/kaspanet/kaspad/domain"

	"github.com/kaspanet/kaspad/app/protocol/common"
	"github.com/kaspanet/kaspad/app/protocol/protocolerrors"
	"github.com/kaspanet/kaspad/infrastructure/network/addressmanager"

	"github.com/kaspanet/kaspad/infrastructure/config"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter"

	"github.com/kaspanet/kaspad/app/appmessage"
	peerpkg "github.com/kaspanet/kaspad/app/protocol/peer"
	routerpkg "github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
	"github.com/pkg/errors"
)

// HandleHandshakeContext is the interface for the context needed for the HandleHandshake flow.
type HandleHandshakeContext interface {
	Config() *config.Config
	NetAdapter() *netadapter.NetAdapter
	Domain() domain.Domain
	AddressManager() *addressmanager.AddressManager
	StartIBDIfRequired() error
	AddToPeers(peer *peerpkg.Peer) error
	HandleError(err error, flowName string, isStopping *uint32, errChan chan<- error)
}

// HandleHandshake sets up the handshake protocol - It sends a version message and waits for an incoming
// version message, as well as a verack for the sent version
func HandleHandshake(context HandleHandshakeContext, netConnection *netadapter.NetConnection,
	receiveVersionRoute *routerpkg.Route, sendVersionRoute *routerpkg.Route, outgoingRoute *routerpkg.Route,
) (*peerpkg.Peer, error) {

	// For HandleHandshake to finish, we need to get from the other node
	// a version and verack messages, so we set doneCount to 2, decrease it
	// when sending and receiving the version, and close the doneChan when
	// it's 0. Then we wait for on select for a tick from doneChan or from
	// errChan.
	doneCount := int32(2)
	doneChan := make(chan struct{})

	isStopping := uint32(0)
	errChan := make(chan error)

	peer := peerpkg.New(netConnection)

	var peerAddress *appmessage.NetAddress
	spawn("HandleHandshake-ReceiveVersion", func() {
		address, err := ReceiveVersion(context, receiveVersionRoute, outgoingRoute, peer)
		if err != nil {
			handleError(err, "ReceiveVersion", &isStopping, errChan)
			return
		}
		peerAddress = address
		if atomic.AddInt32(&doneCount, -1) == 0 {
			close(doneChan)
		}
	})

	spawn("HandleHandshake-SendVersion", func() {
		err := SendVersion(context, sendVersionRoute, outgoingRoute, peer)
		if err != nil {
			handleError(err, "SendVersion", &isStopping, errChan)
			return
		}
		if atomic.AddInt32(&doneCount, -1) == 0 {
			close(doneChan)
		}
	})

	select {
	case err := <-errChan:
		if err != nil {
			return nil, err
		}
		return nil, nil
	case <-doneChan:
	}

	err := context.AddToPeers(peer)
	if err != nil {
		if errors.As(err, &common.ErrPeerWithSameIDExists) {
			return nil, protocolerrors.Wrap(false, err, "peer already exists")
		}
		return nil, err
	}

	if peerAddress != nil {
		context.AddressManager().AddAddresses(peerAddress)
	}

	err = context.StartIBDIfRequired()
	if err != nil {
		return nil, err
	}

	return peer, nil
}

// Handshake is different from other flows, since in it should forward router.ErrRouteClosed to errChan
// Therefore we implement a separate handleError for handshake
func handleError(err error, flowName string, isStopping *uint32, errChan chan error) {
	if errors.Is(err, routerpkg.ErrRouteClosed) {
		if atomic.AddUint32(isStopping, 1) == 1 {
			errChan <- err
		}
		return
	}

	if protocolErr := &(protocolerrors.ProtocolError{}); errors.As(err, &protocolErr) {
		log.Errorf("Handshake protocol error from %s: %s", flowName, err)
		if atomic.AddUint32(isStopping, 1) == 1 {
			errChan <- err
		}
		return
	}
	panic(err)
}
