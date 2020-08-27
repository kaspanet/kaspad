package handshake

import (
	"testing"

	peerpkg "github.com/kaspanet/kaspad/app/protocol/peer"
	"github.com/kaspanet/kaspad/domain/blockdag"
	"github.com/kaspanet/kaspad/infrastructure/config"
	"github.com/kaspanet/kaspad/infrastructure/network/addressmanager"
	"github.com/kaspanet/kaspad/infrastructure/network/connmanager"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
)

type MockHandleHandshakeContext struct {
	HandleHandshakeContext

	cfg               *config.Config
	netAdapter        *netadapter.NetAdapter
	dag               *blockdag.BlockDAG
	addressManager    *addressmanager.AddressManager
	connectionManager *connmanager.ConnectionManager
}

func Config(c *MockHandleHandshakeContext) *config.Config {
	return c.cfg
}

func NetAdapter(c *MockHandleHandshakeContext) *netadapter.NetAdapter {
	return c.netAdapter
}

func DAG(c *MockHandleHandshakeContext) *blockdag.BlockDAG {
	return c.dag
}

func AddressManager(c *MockHandleHandshakeContext) *addressmanager.AddressManager {
	return c.addressManager
}

func StartIBDIfRequired(c *MockHandleHandshakeContext) {
}

func AddToPeers(c *MockHandleHandshakeContext, peer *peerpkg.Peer) error {
	return nil
}

func TestSendVersion(t *testing.T) {
	incomingRoute := router.NewRoute()
	outgoingRoute := router.NewRoute()
	peer := peerpkg.New(nil)
	adapter, err := netadapter.NewNetAdapter(config.DefaultConfig())
	if err != nil {
		t.Fatal("netadapter.NewNetAdapter: error creating NetAdapter", err)
	}

	t.Run("Test for invalid arguments 1", func(t *testing.T) {
		err := SendVersion(nil, nil, nil, nil)
		if err == nil {
			t.Error("Expected error")
		}
	})

	t.Run("Test for invalid arguments 2", func(t *testing.T) {
		mockHandleHandshakeContext := MockHandleHandshakeContext{
			cfg:        config.DefaultConfig(),
			netAdapter: adapter,
			dag:        nil,
		}

		SendVersion(mockHandleHandshakeContext, incomingRoute, outgoingRoute, peer)
	})

	t.Run("Test for invalid arguments 3", func(t *testing.T) {
		mockHandleHandshakeContext := MockHandleHandshakeContext{
			cfg:        config.DefaultConfig(),
			netAdapter: adapter,
			dag:        nil,
		}

		err := SendVersion(mockHandleHandshakeContext, nil, nil, peer)
		if err == nil {
			t.Error("Expected error")
		}
	})

	t.Run("Test for invalid arguments 4", func(t *testing.T) {
		mockHandleHandshakeContext := MockHandleHandshakeContext{
			cfg:        nil,
			netAdapter: nil,
			dag:        nil,
		}
		SendVersion(mockHandleHandshakeContext, incomingRoute, outgoingRoute, peer)
	})
}

func TestReceiveVersion(t *testing.T) {
	incomingRoute := router.NewRoute()
	outgoingRoute := router.NewRoute()
	peer := peerpkg.New(nil)
	adapter, err := netadapter.NewNetAdapter(config.DefaultConfig())
	if err != nil {
		t.Fatal("netadapter.NewNetAdapter: error creating NetAdapter", err)
	}

	t.Run("Test for invalid arguments 1", func(t *testing.T) {
		_, err := ReceiveVersion(nil, nil, nil, nil)
		if err == nil {
			t.Error("Expected error")
		}
	})

	t.Run("Test for invalid arguments 2", func(t *testing.T) {
		mockHandleHandshakeContext := MockHandleHandshakeContext{
			cfg:        config.DefaultConfig(),
			netAdapter: adapter,
			dag:        nil,
		}

		ReceiveVersion(mockHandleHandshakeContext, incomingRoute, outgoingRoute, peer)
	})

	t.Run("Test for invalid arguments 3", func(t *testing.T) {
		mockHandleHandshakeContext := MockHandleHandshakeContext{
			cfg:        config.DefaultConfig(),
			netAdapter: adapter,
			dag:        nil,
		}

		_, err := ReceiveVersion(mockHandleHandshakeContext, nil, nil, peer)
		if err == nil {
			t.Error("Expected error")
		}
	})

	t.Run("Test for invalid arguments 4", func(t *testing.T) {
		mockHandleHandshakeContext := MockHandleHandshakeContext{
			cfg:        nil,
			netAdapter: nil,
			dag:        nil,
		}
		ReceiveVersion(mockHandleHandshakeContext, incomingRoute, outgoingRoute, peer)
	})

}

func TestHandleHandshake(t *testing.T) {
	receiveVersionRoute := router.NewRoute()
	sendVersionRoute := router.NewRoute()
	outgoingRoute := router.NewRoute()
	connection := &netadapter.NetConnection{}
	adapter, err := netadapter.NewNetAdapter(config.DefaultConfig())
	if err != nil {
		t.Fatal("netadapter.NewNetAdapter: error creating NetAdapter", err)
	}

	t.Run("Test for invalid arguments 1", func(t *testing.T) {
		_, err := HandleHandshake(nil, nil, nil, nil, nil)
		if err == nil {
			t.Error("Expected error")
		}
	})

	t.Run("Test for invalid arguments 2", func(t *testing.T) {
		mockHandleHandshakeContext := MockHandleHandshakeContext{
			cfg:        config.DefaultConfig(),
			netAdapter: adapter,
			dag:        nil,
		}

		HandleHandshake(mockHandleHandshakeContext, connection, receiveVersionRoute, sendVersionRoute, outgoingRoute)
	})

	t.Run("Test for invalid arguments 3", func(t *testing.T) {
		mockHandleHandshakeContext := MockHandleHandshakeContext{
			cfg:        config.DefaultConfig(),
			netAdapter: adapter,
			dag:        nil,
		}

		_, err := HandleHandshake(mockHandleHandshakeContext, connection, nil, nil, nil)
		if err == nil {
			t.Error("Expected error")
		}
	})

	t.Run("Test for invalid arguments 4", func(t *testing.T) {
		mockHandleHandshakeContext := MockHandleHandshakeContext{
			cfg:        nil,
			netAdapter: nil,
			dag:        nil,
		}
		HandleHandshake(mockHandleHandshakeContext, connection, receiveVersionRoute, sendVersionRoute, outgoingRoute)
	})
}
