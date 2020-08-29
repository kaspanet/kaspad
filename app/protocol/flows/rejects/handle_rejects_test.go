package rejects

import (
	"testing"

	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/protocol/protocolerrors"

	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
)

func TestHandleRejects(t *testing.T) {
	incomingRoute := router.NewRoute()
	outgoingRoute := router.NewRoute()
	err := incomingRoute.Enqueue(appmessage.NewMsgReject("test reason"))
	if err != nil {
		t.Fatalf("HandleRejects: %s", err)
	}

	err = HandleRejects(nil, incomingRoute, outgoingRoute)
	if err == nil {
		t.Fatal("HandleRejects: expected ProtocolError, got nil")
	}

	if _, ok := err.(*protocolerrors.ProtocolError); !ok {
		t.Fatalf("HandleRejects: %s", err)
	}
}
