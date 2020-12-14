package addressexchange

import (
	"fmt"
	"testing"

	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/protocol/flowcontext"
	peerpkg "github.com/kaspanet/kaspad/app/protocol/peer"
	"github.com/kaspanet/kaspad/infrastructure/config"
	"github.com/kaspanet/kaspad/infrastructure/network/addressmanager"
	"github.com/kaspanet/kaspad/infrastructure/network/connmanager"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
)

func TestReceiveAddresses(t *testing.T) {
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

	testDomain, teardown, err := setupTestDomain(t.Name())
	if err != nil {
		t.Fatalf("SendAddresses: %s", err)
	}
	defer teardown()

	defaultCfg.ActiveNetParams.AcceptUnroutable = true
	addressManager, err := addressmanager.New(addressmanager.NewConfig(defaultCfg))
	if err != nil {
		t.Fatalf("SendAddresses: %s", err)
	}

	addresses := generateAddressesForTest(5)
	addressManager.AddAddresses(addresses...)

	netAdapterA, err := netadapter.NewNetAdapter(cfgA)
	if err != nil {
		t.Fatalf("SendAddresses: %s", err)
	}

	netAdapterA.SetP2PRouterInitializer(func(router *router.Router, connection *netadapter.NetConnection) {})
	netAdapterA.SetRPCRouterInitializer(func(router *router.Router, connection *netadapter.NetConnection) {})
	err = netAdapterA.Start()
	if err != nil {
		t.Fatalf("SendAddresses: %s", err)
	}

	// netAdapterB is needed to have a connection with netAdapterA
	netAdapterB, err := netadapter.NewNetAdapter(cfgB)
	if err != nil {
		t.Fatalf("SendAddresses: %s", err)
	}

	netAdapterB.SetP2PRouterInitializer(func(router *router.Router, connection *netadapter.NetConnection) {})
	netAdapterB.SetRPCRouterInitializer(func(router *router.Router, connection *netadapter.NetConnection) {})
	err = netAdapterB.Start()
	if err != nil {
		t.Fatalf("SendAddresses: %s", err)
	}

	err = netAdapterA.P2PConnect(addressB)
	if err != nil {
		t.Fatalf("SendAddresses: %s", err)
	}

	connManager, err := connmanager.New(defaultCfg, netAdapterA, addressManager)
	if err != nil {
		t.Fatalf("SendAddresses: %s", err)
	}

	peer := peerpkg.New(netAdapterA.P2PConnections()[0])

	ctx := flowcontext.New(defaultCfg, testDomain, addressManager, netAdapterA, connManager)
	incomingRoute := router.NewRoute()
	outgoingRoute := router.NewRoute()
	err = incomingRoute.Enqueue(appmessage.NewMsgAddresses(nil))
	if err != nil {
		t.Fatalf("ReceiveAddresses: %s", err)
	}

	err = ReceiveAddresses(ctx, incomingRoute, outgoingRoute, peer)
	if err != nil {
		t.Fatalf("ReceiveAddresses: %s", err)
	}
}
