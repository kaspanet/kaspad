package addressexchange_test

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/protocol/flows/addressexchange"
	peerpkg "github.com/kaspanet/kaspad/app/protocol/peer"
	"github.com/kaspanet/kaspad/app/protocol/protocolerrors"
	"github.com/kaspanet/kaspad/domain/consensus/utils/testutils"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/kaspanet/kaspad/infrastructure/network/addressmanager"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
	"github.com/pkg/errors"
	"strings"
	"testing"
	"time"
)

type fakeReceiveAddressesContext struct{}

func (f fakeReceiveAddressesContext) AddressManager() *addressmanager.AddressManager {
	return nil
}

func checkFlowError(t *testing.T, err error, isProtocolError bool, shouldBan bool, contains string) {
	pErr := &protocolerrors.ProtocolError{}
	if errors.As(err, &pErr) != isProtocolError {
		t.Fatalf("Unexepcted error %+v", err)
	}

	if pErr.ShouldBan != shouldBan {
		t.Fatalf("Exepcted shouldBan %t but got %t", shouldBan, pErr.ShouldBan)
	}

	if !strings.Contains(err.Error(), contains) {
		t.Fatalf("Unexpected error: %+v", err)
	}
}

func TestReceiveAddressesErrors(t *testing.T) {
	testutils.ForAllNets(t, true, func(t *testing.T, params *dagconfig.Params) {
		incomingRoute := router.NewRoute()
		outgoingRoute := router.NewRoute()
		peer := peerpkg.New(nil)
		errChan := make(chan error)
		go func() {
			errChan <- addressexchange.ReceiveAddresses(fakeReceiveAddressesContext{}, incomingRoute, outgoingRoute, peer)
		}()

		_, err := outgoingRoute.DequeueWithTimeout(time.Second)
		if err != nil {
			t.Fatalf("DequeueWithTimeout: %+v", err)
		}

		// Sending addressmanager.GetAddressesMax+1 addresses should trigger a ban
		err = incomingRoute.Enqueue(appmessage.NewMsgAddresses(make([]*appmessage.NetAddress,
			addressmanager.GetAddressesMax+1)))
		if err != nil {
			t.Fatalf("Enqueue: %+v", err)
		}

		select {
		case err := <-errChan:
			checkFlowError(t, err, true, true, "address count exceeded")
		case <-time.After(time.Second):
			t.Fatalf("timed out after %s", time.Second)
		}
	})
}
