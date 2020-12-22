package handshake

import (
	"errors"
	"fmt"
	"github.com/kaspanet/kaspad/app/appmessage"
	peerpkg "github.com/kaspanet/kaspad/app/protocol/peer"
	"github.com/kaspanet/kaspad/domain"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/kaspanet/kaspad/infrastructure/config"
	"github.com/kaspanet/kaspad/infrastructure/db/database/ldb"
	"github.com/kaspanet/kaspad/infrastructure/network/addressmanager"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
	"io/ioutil"
	"os"
	"sync"
	"testing"
)

type mockHandleHandshakeContext struct {
	cfg            *config.Config
	netAdapter     *netadapter.NetAdapter
	domain         domain.Domain
	addressManager *addressmanager.AddressManager
	peers          []*peerpkg.Peer
}

func (m *mockHandleHandshakeContext) Config() *config.Config {
	return m.cfg
}

func (m *mockHandleHandshakeContext) NetAdapter() *netadapter.NetAdapter {
	return m.netAdapter
}

func (m *mockHandleHandshakeContext) Domain() domain.Domain {
	return m.domain
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

func newMockHandleHandshakeContext(adapter *netadapter.NetAdapter, testName string) (moc *mockHandleHandshakeContext, teardown func(), err error) {
	addressManager, err := addressmanager.New(addressmanager.NewConfig(config.DefaultConfig()))
	if err != nil {
		return nil, nil, err
	}

	domainInstance, teardown, err := setupTestDomain(testName)
	if err != nil {
		return nil, nil, err
	}

	return &mockHandleHandshakeContext{
		domain:         domainInstance,
		cfg:            config.DefaultConfig(),
		netAdapter:     adapter,
		addressManager: addressManager,
	}, teardown, nil
}

func setupTestDomain(testName string) (domainInstance domain.Domain, teardown func(), err error) {
	dataDir, err := ioutil.TempDir("", testName)
	if err != nil {
		return nil, nil, err
	}
	db, err := ldb.NewLevelDB(dataDir)
	if err != nil {
		return nil, nil, err
	}
	teardown = func() {
		db.Close()
		os.RemoveAll(dataDir)
	}

	params := dagconfig.SimnetParams
	domainInstance, err = domain.New(&params, db)
	if err != nil {
		teardown()
		return nil, nil, err
	}

	return domainInstance, teardown, nil
}

func setupP2PConnection() (*netadapter.NetAdapter, *netadapter.NetAdapter, *netadapter.NetConnection, error) {
	const (
		host  = "127.0.0.1"
		portA = 3000
		portB = 3001
	)

	addressA := fmt.Sprintf("%s:%d", host, portA)
	addressB := fmt.Sprintf("%s:%d", host, portB)

	cfgA, cfgB := config.DefaultConfig(), config.DefaultConfig()
	cfgA.Listeners = []string{addressA}
	cfgB.Listeners = []string{addressB}

	adapterA, err := netadapter.NewNetAdapter(cfgA)
	if err != nil {
		return nil, nil, nil, err
	}
	adapterA.SetP2PRouterInitializer(func(router *router.Router, connection *netadapter.NetConnection) {})
	adapterA.SetRPCRouterInitializer(func(router *router.Router, connection *netadapter.NetConnection) {})

	err = adapterA.Start()
	if err != nil {
		return nil, nil, nil, err
	}

	adapterB, err := netadapter.NewNetAdapter(cfgB)
	if err != nil {
		return nil, nil, nil, err
	}
	adapterB.SetP2PRouterInitializer(func(router *router.Router, connection *netadapter.NetConnection) {})
	adapterB.SetRPCRouterInitializer(func(router *router.Router, connection *netadapter.NetConnection) {})
	err = adapterB.Start()
	if err != nil {
		return nil, nil, nil, err
	}

	err = adapterA.P2PConnect(addressB)
	if err != nil {
		return nil, nil, nil, err
	}

	connections := adapterA.P2PConnections()
	if len(connections) == 0 {
		return nil, nil, nil, errors.New("adapterA.P2PConnections: No available connections")
	}

	return adapterA, adapterB, connections[0], nil
}

func TestSendVersion(t *testing.T) {
	adapterA, adapterB, connection, err := setupP2PConnection()
	if err != nil {
		t.Fatalf("Failed to setup p2p connection: %v", err)
	}
	defer func() {
		go adapterB.Stop()
		go adapterA.Stop()
	}()

	peer := peerpkg.New(connection)
	context, teardown, err := newMockHandleHandshakeContext(adapterA, t.Name())
	if err != nil {
		t.Fatalf("Failed to setup new MocRelayBlockRequestsContext instance: %v", err)
	}
	defer teardown()

	t.Run("Simple call test", func(t *testing.T) {
		incomingRoute := router.NewRoute()
		outgoingRoute := router.NewRoute()

		incomingRoute.Enqueue(&appmessage.MsgVerAck{})
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

		incomingRoute.Enqueue(&appmessage.MsgVerAck{})
		incomingRoute.Close()
		outgoingRoute.Close()
		err := SendVersion(context, incomingRoute, outgoingRoute, peer)

		if err.Error() != router.ErrRouteClosed.Error() {
			t.Fatalf("HandleRelayBlockRequests: expected ErrRouteClosed, got %s", err)
		}
	})

	t.Run("Check Outgoing Message", func(t *testing.T) {
		incomingRoute := router.NewRoute()
		outgoingRoute := router.NewRoute()

		go func() {
			message, err := outgoingRoute.Dequeue()
			if err != nil {
				t.Fatalf("TestSendVersion: %s", err)
			}

			messageVersion := message.(*appmessage.MsgVersion)

			if !messageVersion.ID.IsEqual(adapterA.ID()) {
				t.Fatalf("messageVersion: unexcpected result")
			}
			if messageVersion.ID.IsEqual(adapterB.ID()) {
				t.Fatalf("messageVersion: unexcpected result")
			}

			incomingRoute.Enqueue(&appmessage.MsgVerAck{})
			incomingRoute.Close()
			outgoingRoute.Close()
		}()

		SendVersion(context, incomingRoute, outgoingRoute, peer)
	})
}

func TestReceiveVersion(t *testing.T) {
	adapterA, adapterB, connection, err := setupP2PConnection()
	if err != nil {
		t.Fatalf("Failed to setup p2p connection: %v", err)
	}

	defer func() {
		go adapterB.Stop()
		go adapterA.Stop()
	}()

	peer := peerpkg.New(connection)
	context, teardown, err := newMockHandleHandshakeContext(adapterA, t.Name())
	if err != nil {
		t.Fatalf("Failed to setup new MocRelayBlockRequestsContext instance: %v", err)
	}
	defer teardown()

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
		wg := &sync.WaitGroup{}
		wg.Add(2)

		go func() {
			incomingRoute.Enqueue(&appmessage.MsgVersion{
				ID:              adapterB.ID(),
				Network:         context.Config().ActiveNetParams.Name,
				ProtocolVersion: appmessage.ProtocolVersion,
			})

			message, err := outgoingRoute.Dequeue()
			if err != nil {
				t.Fatalf("TestSendVersion: %s", err)
			}

			_, ok := message.(*appmessage.MsgVerAck)
			if !ok {
				t.Fatalf("Unexpected message type")
			}

			wg.Done()
		}()

		go func() {
			ReceiveVersion(context, incomingRoute, outgoingRoute, peer)
			wg.Done()
		}()

		wg.Wait()
	})
}

func TestHandleHandshake(t *testing.T) {
	adapterA, adapterB, connection, err := setupP2PConnection()
	if err != nil {
		t.Fatalf("Failed to setup p2p connection: %v", err)
	}

	defer func() {
		go adapterB.Stop()
		go adapterA.Stop()
	}()

	context, teardown, err := newMockHandleHandshakeContext(adapterA, t.Name())
	if err != nil {
		t.Fatalf("Failed to setup new MocRelayBlockRequestsContext instance: %v", err)
	}
	defer teardown()

	contextB, teardown, err := newMockHandleHandshakeContext(adapterB, t.Name()+"receive")
	if err != nil {
		t.Fatalf("Failed to setup new MocRelayBlockRequestsContext instance: %v", err)
	}
	defer teardown()

	peer := peerpkg.New(connection)

	t.Run("Simple call test", func(t *testing.T) {
		receiveVersionRoute := router.NewRoute()
		sendVersionRoute := router.NewRoute()
		outgoingRoute := router.NewRoute()
		outgoingRoute.Close()
		HandleHandshake(context, connection, receiveVersionRoute, sendVersionRoute, outgoingRoute)
	})

	t.Run("Test on closed route 1", func(t *testing.T) {
		receiveVersionRoute := router.NewRoute()
		sendVersionRoute := router.NewRoute()
		outgoingRoute := router.NewRoute()
		sendVersionRoute.Close()
		HandleHandshake(context, connection, receiveVersionRoute, sendVersionRoute, outgoingRoute)
	})

	t.Run("Test on closed route 2", func(t *testing.T) {
		receiveVersionRoute := router.NewRoute()
		sendVersionRoute := router.NewRoute()
		outgoingRoute := router.NewRoute()
		receiveVersionRoute.Close()
		HandleHandshake(context, connection, receiveVersionRoute, sendVersionRoute, outgoingRoute)
	})

	t.Run("Test HandleHandshake", func(t *testing.T) {
		receiveVersionRoute := router.NewRoute()
		sendVersionRoute := router.NewRoute()
		outgoingRoute := router.NewRoute()

		go func() {
			_, err = ReceiveVersion(contextB, outgoingRoute, sendVersionRoute, peer)
			if err != nil {
				t.Fatalf("ReceiveVersion: %v", err)
			}

			err := SendVersion(contextB, outgoingRoute, receiveVersionRoute, peer)
			if err != nil {
				t.Fatalf("SendVersion: %v", err)
			}
		}()

		_, err = HandleHandshake(context, connection, receiveVersionRoute, sendVersionRoute, outgoingRoute)

		if err != nil {
			t.Fatalf("TestHandleHandshake: %s", err)
		}
	})
}
