package ping

import (
	"testing"

	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
)

func TestReceivePings(t *testing.T) {
	incomingRoute := router.NewRoute()
	outgoingRoute := router.NewRoute()

	err := incomingRoute.Enqueue(appmessage.NewMsgPing(1))
	if err != nil {
		t.Fatalf("ReceivePings: %s", err)
	}

	go func() {
		err := ReceivePings(nil, incomingRoute, outgoingRoute)
		if err != nil {
			t.Fatalf("ReceivePings: %s", err)
		}
	}()

	msg, err := outgoingRoute.Dequeue()
	if err != nil {
		t.Fatalf("ReceivePings: %s", err)
	}

	if _, ok := msg.(*appmessage.MsgPong); !ok {
		t.Fatalf("ReceivePings: expected *appmessage.MsgPong, got %T", msg)
	}
}
