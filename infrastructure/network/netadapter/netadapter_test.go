package netadapter

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/kaspanet/kaspad/app/appmessage"

	"github.com/kaspanet/kaspad/infrastructure/config"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
)

// routerInitializerForTest returns new RouterInitializer which simply sets
// new incoming route for router and stores this route in map for further usage in tests
func routerInitializerForTest(t *testing.T, routes *sync.Map,
	routeName string, wg *sync.WaitGroup) func(*router.Router, *NetConnection) {
	return func(router *router.Router, connection *NetConnection) {
		route, err := router.AddIncomingRoute(routeName, []appmessage.MessageCommand{appmessage.CmdPing})
		if err != nil {
			t.Fatalf("TestNetAdapter: AddIncomingRoute failed: %+v", err)
		}
		routes.Store(routeName, route)
		wg.Done()
	}
}

func TestNetAdapter(t *testing.T) {
	const (
		timeout = time.Second * 5
		nonce   = uint64(1)

		host  = "127.0.0.1"
		portA = 3000
		portB = 3001
		portC = 3002
	)

	addressA := fmt.Sprintf("%s:%d", host, portA)
	addressB := fmt.Sprintf("%s:%d", host, portB)
	addressC := fmt.Sprintf("%s:%d", host, portC)

	cfgA, cfgB, cfgC := config.DefaultConfig(), config.DefaultConfig(), config.DefaultConfig()
	cfgA.Listeners = []string{addressA}
	cfgB.Listeners = []string{addressB}
	cfgC.Listeners = []string{addressC}

	routes := &sync.Map{}
	wg := &sync.WaitGroup{}
	wg.Add(2)

	adapterA, err := NewNetAdapter(cfgA)
	if err != nil {
		t.Fatalf("TestNetAdapter: NetAdapter instantiation failed: %+v", err)
	}

	adapterA.SetP2PRouterInitializer(func(router *router.Router, connection *NetConnection) {})
	adapterA.SetRPCRouterInitializer(func(router *router.Router, connection *NetConnection) {})
	err = adapterA.Start()
	if err != nil {
		t.Fatalf("TestNetAdapter: Start() failed: %+v", err)
	}

	adapterB, err := NewNetAdapter(cfgB)
	if err != nil {
		t.Fatalf("TestNetAdapter: NetAdapter instantiation failed: %+v", err)
	}

	initializer := routerInitializerForTest(t, routes, "B", wg)
	adapterB.SetP2PRouterInitializer(initializer)
	adapterB.SetRPCRouterInitializer(func(router *router.Router, connection *NetConnection) {})
	err = adapterB.Start()
	if err != nil {
		t.Fatalf("TestNetAdapter: Start() failed: %+v", err)
	}

	adapterC, err := NewNetAdapter(cfgC)
	if err != nil {
		t.Fatalf("TestNetAdapter: NetAdapter instantiation failed: %+v", err)
	}

	initializer = routerInitializerForTest(t, routes, "C", wg)
	adapterC.SetP2PRouterInitializer(initializer)
	adapterC.SetRPCRouterInitializer(func(router *router.Router, connection *NetConnection) {})
	err = adapterC.Start()
	if err != nil {
		t.Fatalf("TestNetAdapter: Start() failed: %+v", err)
	}

	err = adapterA.P2PConnect(addressB)
	if err != nil {
		t.Fatalf("TestNetAdapter: connection to %s failed: %+v", addressB, err)
	}

	err = adapterA.P2PConnect(addressC)
	if err != nil {
		t.Fatalf("TestNetAdapter: connection to %s failed: %+v", addressC, err)
	}

	// Ensure adapter has two connections
	if count := adapterA.P2PConnectionCount(); count != 2 {
		t.Fatalf("TestNetAdapter: expected 2 connections, got - %d", count)
	}

	// Ensure all connected peers have received broadcasted message
	connections := adapterA.P2PConnections()
	err = adapterA.P2PBroadcast(connections, appmessage.NewMsgPing(1))
	if err != nil {
		t.Fatalf("TestNetAdapter: broadcast failed: %+v", err)
	}

	// wait for routes to be added to map, then they can be used to receive broadcasted message
	wg.Wait()

	r, ok := routes.Load("B")
	if !ok {
		t.Fatal("TestNetAdapter: route loading failed")
	}

	msg, err := r.(*router.Route).DequeueWithTimeout(timeout)
	if err != nil {
		t.Fatalf("TestNetAdapter: dequeuing message failed: %+v", err)
	}

	if command := msg.Command(); command != appmessage.CmdPing {
		t.Fatalf("TestNetAdapter: expected '%s' message to be received but got '%s'",
			appmessage.ProtocolMessageCommandToString[appmessage.CmdPing],
			appmessage.ProtocolMessageCommandToString[command])
	}

	if number := msg.MessageNumber(); number != nonce {
		t.Fatalf("TestNetAdapter: expected '%d' message number but got %d", nonce, number)
	}

	r, ok = routes.Load("C")
	if !ok {
		t.Fatal("TestNetAdapter: route loading failed")
	}

	msg, err = r.(*router.Route).DequeueWithTimeout(timeout)
	if err != nil {
		t.Fatalf("TestNetAdapter: dequeuing message failed: %+v", err)
	}

	if command := msg.Command(); command != appmessage.CmdPing {
		t.Fatalf("TestNetAdapter: expected '%s' message to be received but got '%s'",
			appmessage.ProtocolMessageCommandToString[appmessage.CmdPing],
			appmessage.ProtocolMessageCommandToString[command])
	}

	if number := msg.MessageNumber(); number != nonce {
		t.Fatalf("TestNetAdapter: expected '%d' message number but got %d", nonce, number)
	}

	err = adapterA.Stop()
	if err != nil {
		t.Fatalf("TestNetAdapter: stopping adapter failed: %+v", err)
	}

	// Ensure adapter can't be stopped multiple times
	err = adapterA.Stop()
	if err == nil {
		t.Fatalf("TestNetAdapter: error expected at attempt to stop adapter second time, but got nothing")
	}
}
