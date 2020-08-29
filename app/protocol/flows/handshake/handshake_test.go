package handshake

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kaspanet/kaspad/app/appmessage"
	peerpkg "github.com/kaspanet/kaspad/app/protocol/peer"
	"github.com/kaspanet/kaspad/domain/blockdag"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/kaspanet/kaspad/domain/txscript"
	"github.com/kaspanet/kaspad/infrastructure/config"
	"github.com/kaspanet/kaspad/infrastructure/db/dbaccess"
	"github.com/kaspanet/kaspad/infrastructure/network/addressmanager"
	"github.com/kaspanet/kaspad/infrastructure/network/connmanager"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
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

func createDag() (*blockdag.BlockDAG, func()) {
	tempDir := os.TempDir()
	dbPath := filepath.Join(tempDir, "TestHandshake")
	_ = os.RemoveAll(dbPath)

	databaseContext, err := dbaccess.New(dbPath)
	if err != nil {
		return nil, nil
	}

	cfg := &blockdag.Config{
		DAGParams:  &dagconfig.SimnetParams,
		TimeSource: blockdag.NewTimeSource(),
		SigCache:   txscript.NewSigCache(1000),
	}

	cfg.DatabaseContext = databaseContext
	dag, err := blockdag.New(cfg)

	return dag, func() {
		databaseContext.Close()
		os.RemoveAll(dbPath)
	}
}

func TestSendVersion(t *testing.T) {
	incomingRoute := router.NewRoute()
	outgoingRoute := router.NewRoute()
	peer := peerpkg.New(&netadapter.NetConnection{})
	dag, cleanDag := createDag()
	defer cleanDag()
	adapter, err := netadapter.NewNetAdapter(config.DefaultConfig())
	if err != nil {
		t.Fatal("netadapter.NewNetAdapter: error creating NetAdapter", err)
	}

	t.Run("Test for invalid arguments 1", func(t *testing.T) {
		context := &mockHandleHandshakeContext{
			cfg:        config.DefaultConfig(),
			netAdapter: adapter,
			dag:        dag,
		}

		incomingRoute.Enqueue(&appmessage.MsgVersion{})
		SendVersion(context, incomingRoute, outgoingRoute, peer)
	})

	t.Run("Test for invalid arguments 2", func(t *testing.T) {
		context := &mockHandleHandshakeContext{
			cfg:        config.DefaultConfig(),
			netAdapter: adapter,
			dag:        dag,
		}

		incomingRoute.Enqueue(&appmessage.MsgVersion{})
		SendVersion(context, incomingRoute, outgoingRoute, peer)
	})

	t.Run("Test for invalid arguments 3", func(t *testing.T) {
		context := &mockHandleHandshakeContext{
			cfg:        config.DefaultConfig(),
			netAdapter: adapter,
			dag:        dag,
		}

		incomingRoute.Enqueue(&appmessage.MsgVersion{})
		SendVersion(context, incomingRoute, outgoingRoute, peer)
	})
}

func TestReceiveVersion(t *testing.T) {
	incomingRoute := router.NewRoute()
	outgoingRoute := router.NewRoute()
	peer := peerpkg.New(&netadapter.NetConnection{})
	dag, cleanDag := createDag()
	defer cleanDag()
	adapter, err := netadapter.NewNetAdapter(config.DefaultConfig())
	if err != nil {
		t.Fatal("netadapter.NewNetAdapter: error creating NetAdapter", err)
	}

	t.Run("Test for invalid arguments 1", func(t *testing.T) {
		context := &mockHandleHandshakeContext{
			cfg:        config.DefaultConfig(),
			netAdapter: adapter,
			dag:        dag,
		}

		incomingRoute.Enqueue(&appmessage.MsgVersion{})
		ReceiveVersion(context, incomingRoute, outgoingRoute, peer)
	})

	t.Run("Test for invalid arguments 2", func(t *testing.T) {
		context := &mockHandleHandshakeContext{
			cfg:        config.DefaultConfig(),
			netAdapter: adapter,
			dag:        dag,
		}

		incomingRoute.Enqueue(&appmessage.MsgInvTransaction{})
		ReceiveVersion(context, incomingRoute, outgoingRoute, peer)
	})

	t.Run("Test for invalid arguments 3", func(t *testing.T) {
		context := &mockHandleHandshakeContext{
			cfg:        config.DefaultConfig(),
			netAdapter: adapter,
			dag:        dag,
		}
		incomingRoute.Enqueue(&appmessage.MsgVersion{})
		incomingRoute.Close()
		ReceiveVersion(context, incomingRoute, outgoingRoute, peer)
	})
}

func TestHandleHandshake(t *testing.T) {
	outgoingRoute := router.NewRoute()
	connection := &netadapter.NetConnection{}
	dag, cleanDag := createDag()
	defer cleanDag()
	adapter, err := netadapter.NewNetAdapter(config.DefaultConfig())
	if err != nil {
		t.Fatal("netadapter.NewNetAdapter: error creating NetAdapter", err)
	}

	t.Run("Test for invalid arguments 1", func(t *testing.T) {
		context := &mockHandleHandshakeContext{
			cfg:        config.DefaultConfig(),
			netAdapter: adapter,
			dag:        dag,
		}

		receiveVersionRoute := router.NewRoute()
		sendVersionRoute := router.NewRoute()
		HandleHandshake(context, connection, receiveVersionRoute, sendVersionRoute, outgoingRoute)
	})

	t.Run("Test for invalid arguments 2", func(t *testing.T) {
		context := &mockHandleHandshakeContext{
			cfg:        config.DefaultConfig(),
			netAdapter: adapter,
			dag:        dag,
		}

		receiveVersionRoute := router.NewRoute()
		sendVersionRoute := router.NewRoute()
		sendVersionRoute.Close()
		HandleHandshake(context, connection, receiveVersionRoute, sendVersionRoute, outgoingRoute)
	})

	t.Run("Test for invalid arguments 3", func(t *testing.T) {
		context := &mockHandleHandshakeContext{
			cfg:        config.DefaultConfig(),
			netAdapter: adapter,
			dag:        dag,
		}
		receiveVersionRoute := router.NewRoute()
		sendVersionRoute := router.NewRoute()
		receiveVersionRoute.Close()
		HandleHandshake(context, connection, receiveVersionRoute, sendVersionRoute, outgoingRoute)
	})
}
