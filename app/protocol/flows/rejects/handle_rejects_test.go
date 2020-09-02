package rejects

import (
	"crypto/rand"
	"testing"

	"github.com/kaspanet/kaspad/app/protocol/protocolerrors"

	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
)

func TestHandleRejects(t *testing.T) {
	incomingRoute := router.NewRoute()
	outgoingRoute := router.NewRoute()

	t.Run("SimpleCall", func(t *testing.T) {
		err := incomingRoute.Enqueue(appmessage.NewMsgReject("test reason"))
		if err != nil {
			t.Fatalf("HandleRejects: %s", err)
		}

		err = HandleRejects(nil, incomingRoute, outgoingRoute)
		if err == nil {
			t.Fatal("HandleRejects: expected error, got nil")
		}
	})

	t.Run("CheckReturnType", func(t *testing.T) {
		err := incomingRoute.Enqueue(appmessage.NewMsgReject("test reason"))
		if err != nil {
			t.Fatalf("HandleRejects: %s", err)
		}

		err = HandleRejects(nil, incomingRoute, outgoingRoute)
		if err == nil {
			t.Fatal("HandleRejects: expected error, got nil")
		}

		if _, ok := err.(*protocolerrors.ProtocolError); !ok {
			t.Fatalf("HandleRejects: expected ProtocolError, got %T", err)
		}
	})

	t.Run("CallMultipleTimes", func(t *testing.T) {
		const callTimes = 5
		for i := 0; i < callTimes; i++ {
			err := incomingRoute.Enqueue(appmessage.NewMsgReject("test reason"))
			if err != nil {
				t.Fatalf("HandleRejects: %s", err)
			}

			err = HandleRejects(nil, incomingRoute, outgoingRoute)
			if err == nil {
				t.Fatal("HandleRejects: expected error, got nil")
			}
		}
	})

	t.Run("CallWithLongMessage", func(t *testing.T) {
		const messageLength = 512
		var buffer [messageLength]byte
		rand.Read(buffer[:])
		longRejectMessage := string(buffer[:])
		err := incomingRoute.Enqueue(appmessage.NewMsgReject(longRejectMessage))
		if err != nil {
			t.Fatalf("HandleRejects: %s", err)
		}

		// TODO: how to differentiate expected reject error from execution error? They both are same type
		err = HandleRejects(nil, incomingRoute, outgoingRoute)
		if err == nil {
			t.Fatal("HandleRejects: expected error, got nil")
		}
	})

	t.Run("CallOnClosedRoute", func(t *testing.T) {
		closedRoute := router.NewRoute()
		closedRoute.Close()
		err := HandleRejects(nil, closedRoute, outgoingRoute)
		if err == nil {
			t.Fatal("HandleRejects: expected error, got nil")
		}
	})

	t.Run("CallOnNilRoutes", func(t *testing.T) {
		err := HandleRejects(nil, nil, nil)
		if err == nil {
			t.Fatal("HandleRejects: expected err, got nil")
		}
	})

	t.Run("CallWithEnqueuedInvalidMessage", func(t *testing.T) {
		err := incomingRoute.Enqueue(appmessage.NewMsgPing(1))
		if err != nil {
			t.Fatalf("HandleRejects: %s", err)
		}

		err = HandleRejects(nil, incomingRoute, outgoingRoute)
		if err == nil {
			t.Fatal("HandleRejects: expected error, got nil")
		}
	})
}
