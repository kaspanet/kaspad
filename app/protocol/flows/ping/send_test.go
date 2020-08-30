package ping

import (
	"testing"
	"time"

	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/protocol/common"
	peerpkg "github.com/kaspanet/kaspad/app/protocol/peer"

	"github.com/kaspanet/kaspad/infrastructure/network/netadapter"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
)

// routerInitializerForTest returns new RouterInitializer which simply sets
// new incoming route for router.
func routerInitializerForTest(t *testing.T) func(*router.Router, *netadapter.NetConnection) {
	return func(router *router.Router, connection *netadapter.NetConnection) {
		_, err := router.AddIncomingRoute([]appmessage.MessageCommand{appmessage.CmdPing})
		if err != nil {
			t.Fatalf("SendPings: %s", err)
		}
	}
}

func TestSendPings(t *testing.T) {
	incomingRoute := router.NewRoute()
	outgoingRoute := router.NewRoute()

	go func() {
		err := ReceivePings(nil, outgoingRoute, incomingRoute)
		if err != nil {
			t.Fatalf("SendPings: %s", err)
		}
	}()

	go func() {
		err := SendPings(nil, incomingRoute, outgoingRoute, &peerpkg.Peer{})
		if err != nil {
			t.Fatalf("SendPings: %s", err)
		}
	}()

	// waitTime based on pingInterval in SendPings function + default timeout needed to receive answer +
	// some extra time. Change it when pingInterval become customizable
	waitTime := 2*time.Minute + common.DefaultTimeout + 2*time.Second
	<-time.After(waitTime)
}
