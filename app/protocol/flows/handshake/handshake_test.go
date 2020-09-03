package handshake

import (
	"testing"

	"github.com/kaspanet/kaspad/app/appmessage"
	peerpkg "github.com/kaspanet/kaspad/app/protocol/peer"
	"github.com/kaspanet/kaspad/domain/blockdag"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/kaspanet/kaspad/infrastructure/config"
	"github.com/kaspanet/kaspad/infrastructure/network/addressmanager"
	"github.com/kaspanet/kaspad/infrastructure/network/connmanager"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
	routerpkg "github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
)

type mockHandleHandshakeContext struct {
	cfg               *config.Config
	netAdapter        *netadapter.NetAdapter
	dag               *blockdag.BlockDAG
	addressManager    *addressmanager.AddressManager
	connectionManager *connmanager.ConnectionManager
}

func (m *mockHandleHandshakeContext) Config() *config.Config {
	return m.cfg
}

func (m *mockHandleHandshakeContext) NetAdapter() *netadapter.NetAdapter {
	return m.netAdapter
}

func (m *mockHandleHandshakeContext) DAG() *blockdag.BlockDAG {
	return m.dag
}

func (m *mockHandleHandshakeContext) AddressManager() *addressmanager.AddressManager {
	return m.addressManager
}

func (m *mockHandleHandshakeContext) StartIBDIfRequired() {
}

func (m *mockHandleHandshakeContext) AddToPeers(peer *peerpkg.Peer) error {
	return nil
}

func (m *mockHandleHandshakeContext) HandleError(err error, flowName string, isStopping *uint32, errChan chan<- error) {
}

func TestSendVersion(t *testing.T) {
	peer := peerpkg.New(&netadapter.NetConnection{})
	params := dagconfig.SimnetParams
	adapter, err := netadapter.NewNetAdapter(config.DefaultConfig())
	if err != nil {
		t.Fatalf("Failed to setup NewNetAdapter instance: %v", err)

	}

	dagConfig := blockdag.Config{DAGParams: &params}
	dag, teardownFunc, err := blockdag.DAGSetup("TestSendVersion", true, dagConfig)

	if err != nil {
		t.Fatalf("Failed to setup DAG instance: %v", err)
	}
	defer teardownFunc()
	addressManager, err := addressmanager.New(config.DefaultConfig(), dagConfig.DatabaseContext)
	if err != nil {
		t.Fatalf("Failed to setup AddressManager instance: %v", err)
	}

	context := &mockHandleHandshakeContext{
		cfg:            config.DefaultConfig(),
		netAdapter:     adapter,
		dag:            dag,
		addressManager: addressManager,
	}

	t.Run("Simple call test", func(t *testing.T) {
		incomingRoute := router.NewRoute()
		outgoingRoute := router.NewRoute()

		incomingRoute.Enqueue(&appmessage.MsgVersion{})
		SendVersion(context, incomingRoute, outgoingRoute, peer)
	})

	t.Run("Test on wrong message type", func(t *testing.T) {
		incomingRoute := router.NewRoute()
		outgoingRoute := router.NewRoute()

		incomingRoute.Enqueue(&appmessage.MsgInvTransaction{})
		SendVersion(context, incomingRoute, outgoingRoute, peer)
	})

	t.Run("Test on closed route", func(t *testing.T) {
		incomingRoute := router.NewRoute()
		outgoingRoute := router.NewRoute()

		incomingRoute.Enqueue(&appmessage.MsgVersion{})
		incomingRoute.Close()
		err := SendVersion(context, incomingRoute, outgoingRoute, peer)

		if err.Error() != routerpkg.ErrRouteClosed.Error() {
			t.Fatalf("HandleRelayBlockRequests: expected ErrRouteClosed, got %s", err)
		}
	})

	t.Run("Check Outgoing Message", func(t *testing.T) {
		incomingRoute := router.NewRoute()
		outgoingRoute := router.NewRoute()

		go func() {
			_, err := outgoingRoute.Dequeue()
			if err != nil {
				t.Fatalf("TestSendVersion: %s", err)
			}

			incomingRoute.Close()
			outgoingRoute.Close()
		}()

		SendVersion(context, incomingRoute, outgoingRoute, peer)
	})
}

func TestReceiveVersion(t *testing.T) {
	peer := peerpkg.New(&netadapter.NetConnection{})
	params := dagconfig.SimnetParams
	adapter, err := netadapter.NewNetAdapter(config.DefaultConfig())
	if err != nil {
		t.Fatalf("Failed to setup NewNetAdapter instance: %v", err)

	}

	dagConfig := blockdag.Config{DAGParams: &params}
	dag, teardownFunc, err := blockdag.DAGSetup("TestReceiveVersion", true, dagConfig)

	if err != nil {
		t.Fatalf("Failed to setup DAG instance: %v", err)
	}
	defer teardownFunc()
	addressManager, err := addressmanager.New(config.DefaultConfig(), dagConfig.DatabaseContext)
	if err != nil {
		t.Fatalf("Failed to setup AddressManager instance: %v", err)
	}

	context := &mockHandleHandshakeContext{
		cfg:            config.DefaultConfig(),
		netAdapter:     adapter,
		dag:            dag,
		addressManager: addressManager,
	}

	t.Run("Simple call test", func(t *testing.T) {
		incomingRoute := router.NewRoute()
		outgoingRoute := router.NewRoute()
		incomingRoute.Enqueue(&appmessage.MsgVersion{})
		ReceiveVersion(context, incomingRoute, outgoingRoute, peer)
	})

	t.Run("Test on wrong message type", func(t *testing.T) {
		incomingRoute := router.NewRoute()
		outgoingRoute := router.NewRoute()
		incomingRoute.Enqueue(&appmessage.MsgInvTransaction{})
		ReceiveVersion(context, incomingRoute, outgoingRoute, peer)
	})

	t.Run("Test on closed route", func(t *testing.T) {
		incomingRoute := router.NewRoute()
		outgoingRoute := router.NewRoute()
		incomingRoute.Close()
		outgoingRoute.Close()
		ReceiveVersion(context, incomingRoute, outgoingRoute, peer)
	})

	t.Run("Check Outgoing Message", func(t *testing.T) {
		incomingRoute := router.NewRoute()
		outgoingRoute := router.NewRoute()
		go func() {
			_, err := outgoingRoute.Dequeue()
			if err != nil {
				t.Fatalf("TestReceiveVersion: %s", err)
			}

			incomingRoute.Close()
			outgoingRoute.Close()
		}()

		incomingRoute.Enqueue(&appmessage.MsgVersion{})
		ReceiveVersion(context, incomingRoute, outgoingRoute, peer)
	})
}

func TestHandleHandshake(t *testing.T) {
	outgoingRoute := router.NewRoute()
	params := dagconfig.SimnetParams
	connection := &netadapter.NetConnection{}
	adapter, err := netadapter.NewNetAdapter(config.DefaultConfig())
	if err != nil {
		t.Fatal("netadapter.NewNetAdapter: error creating NetAdapter", err)
	}

	dagConfig := blockdag.Config{DAGParams: &params}
	dag, teardownFunc, err := blockdag.DAGSetup("TestReceiveVersion", true, dagConfig)

	if err != nil {
		t.Fatalf("Failed to setup DAG instance: %v", err)
	}
	defer teardownFunc()
	addressManager, err := addressmanager.New(config.DefaultConfig(), dagConfig.DatabaseContext)
	if err != nil {
		t.Fatalf("Failed to setup AddressManager instance: %v", err)
	}

	context := &mockHandleHandshakeContext{
		cfg:            config.DefaultConfig(),
		netAdapter:     adapter,
		dag:            dag,
		addressManager: addressManager,
	}

	t.Run("Simple call test", func(t *testing.T) {
		receiveVersionRoute := router.NewRoute()
		sendVersionRoute := router.NewRoute()
		HandleHandshake(context, connection, receiveVersionRoute, sendVersionRoute, outgoingRoute)
	})

	t.Run("Test on closed route 1", func(t *testing.T) {
		receiveVersionRoute := router.NewRoute()
		sendVersionRoute := router.NewRoute()
		sendVersionRoute.Close()
		HandleHandshake(context, connection, receiveVersionRoute, sendVersionRoute, outgoingRoute)
	})

	t.Run("Test on closed route 2", func(t *testing.T) {
		receiveVersionRoute := router.NewRoute()
		sendVersionRoute := router.NewRoute()
		receiveVersionRoute.Close()
		HandleHandshake(context, connection, receiveVersionRoute, sendVersionRoute, outgoingRoute)
	})

	t.Run("Test HandleHandshake", func(t *testing.T) {
		receiveVersionRoute := router.NewRoute()
		sendVersionRoute := router.NewRoute()
		_, err := HandleHandshake(context, connection, receiveVersionRoute, sendVersionRoute, outgoingRoute)

		if err != nil {
			t.Fatalf("TestHandleHandshake: %s", err)
		}
	})
}
