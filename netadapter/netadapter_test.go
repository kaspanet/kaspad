package netadapter

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/kaspanet/kaspad/domainmessage"

	"github.com/kaspanet/kaspad/config"
	"github.com/kaspanet/kaspad/netadapter/router"
)

// getRouterInitializer returns new RouterInitializer which simply sets
// new incoming route for router and stores this route in map for further usage in tests
func getRouterInitializer(routes *sync.Map, routeName string,
	t *testing.T, wg *sync.WaitGroup) func(*router.Router, *NetConnection) {
	return func(router *router.Router, connection *NetConnection) {
		route, err := router.AddIncomingRoute([]domainmessage.MessageCommand{domainmessage.CmdPing})
		if err != nil {
			t.Fatalf("TestNetAdapter: AddIncomingRoute failed: %+v", err)
		}
		routes.Store(routeName, route)
		wg.Done()
	}
}

func TestNetAdapter(t *testing.T) {
	timeout := time.Second * 5
	nonce := uint64(1)

	host := "127.0.0.1"
	portA, portB, portC := 3000, 3001, 3002
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

	adapterA.SetRouterInitializer(func(router *router.Router, connection *NetConnection) {})
	err = adapterA.Start()
	if err != nil {
		t.Fatalf("TestNetAdapter: Start() failed: %+v", err)
	}

	adapterB, err := NewNetAdapter(cfgB)
	if err != nil {
		t.Fatalf("TestNetAdapter: NetAdapter instantiation failed: %+v", err)
	}

	initializer := getRouterInitializer(routes, "B", t, wg)
	adapterB.SetRouterInitializer(initializer)
	err = adapterB.Start()
	if err != nil {
		t.Fatalf("TestNetAdapter: Start() failed: %+v", err)
	}

	adapterC, err := NewNetAdapter(cfgC)
	if err != nil {
		t.Fatalf("TestNetAdapter: NetAdapter instantiation failed: %+v", err)
	}

	initializer = getRouterInitializer(routes, "C", t, wg)
	adapterC.SetRouterInitializer(initializer)
	err = adapterC.Start()
	if err != nil {
		t.Fatalf("TestNetAdapter: Start() failed: %+v", err)
	}

	err = adapterA.Connect(addressB)
	if err != nil {
		t.Fatalf("TestNetAdapter: connection to %s failed: %+v", addressB, err)
	}

	err = adapterA.Connect(addressC)
	if err != nil {
		t.Fatalf("TestNetAdapter: connection to %s failed: %+v", addressC, err)
	}

	// Ensure adapter has two connections
	if count := adapterA.ConnectionCount(); count != 2 {
		t.Fatalf("TestNetAdapter: expected 2 connections, got - %d", count)
	}

	// Ensure all connected peers have received broadcasted message
	connections := adapterA.Connections()
	err = adapterA.Broadcast(connections, domainmessage.NewMsgPing(1))
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

	if command := msg.Command(); command != domainmessage.CmdPing {
		t.Fatalf("TestNetAdapter: expected '%s' message to be received but got '%s'",
			domainmessage.MessageCommandToString[domainmessage.CmdPing],
			domainmessage.MessageCommandToString[command])
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

	if command := msg.Command(); command != domainmessage.CmdPing {
		t.Fatalf("TestNetAdapter: expected '%s' message to be received but got '%s'",
			domainmessage.MessageCommandToString[domainmessage.CmdPing],
			domainmessage.MessageCommandToString[command])
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
