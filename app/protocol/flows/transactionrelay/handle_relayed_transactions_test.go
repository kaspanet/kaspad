package transactionrelay_test

import (
	"errors"
	"github.com/kaspanet/kaspad/app/protocol/flows/transactionrelay"
	"github.com/kaspanet/kaspad/app/protocol/protocolerrors"
	"github.com/kaspanet/kaspad/domain"
	"github.com/kaspanet/kaspad/domain/consensus"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/testutils"
	"github.com/kaspanet/kaspad/infrastructure/logger"
	"github.com/kaspanet/kaspad/util/panics"
	"strings"
	"testing"

	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/kaspanet/kaspad/infrastructure/config"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
)

type mocTransactionsRelayContext struct {
	netAdapter                  *netadapter.NetAdapter
	domain                      domain.Domain
	sharedRequestedTransactions *transactionrelay.SharedRequestedTransactions
}

func (m *mocTransactionsRelayContext) NetAdapter() *netadapter.NetAdapter {
	return m.netAdapter
}

func (m *mocTransactionsRelayContext) Domain() domain.Domain {
	return m.domain
}

func (m *mocTransactionsRelayContext) SharedRequestedTransactions() *transactionrelay.SharedRequestedTransactions {
	return m.sharedRequestedTransactions
}

func (m *mocTransactionsRelayContext) Broadcast(appmessage.Message) error {
	return nil
}

func (m *mocTransactionsRelayContext) OnTransactionAddedToMempool() {
}

// TestHandleRelayedTransactionsNotFound tests the flow of  HandleRelayedTransactions when the peer doesn't
// have the requested transactions in the mempool.
func TestHandleRelayedTransactionsNotFound(t *testing.T) {
	testutils.ForAllNets(t, true, func(t *testing.T, params *dagconfig.Params) {

		var log = logger.RegisterSubSystem("PROT")
		var spawn = panics.GoroutineWrapperFunc(log)
		factory := consensus.NewFactory()
		tc, teardown, err := factory.NewTestConsensus(params, false, "TestHandleRelayedTransactionsNotFound")
		if err != nil {
			t.Fatalf("Error setting up test consensus: %+v", err)
		}
		defer teardown(false)

		sharedRequestedTransactions := transactionrelay.NewSharedRequestedTransactions()
		adapter, err := netadapter.NewNetAdapter(config.DefaultConfig())
		if err != nil {
			t.Fatalf("Failed to create a NetAdapter: %v", err)
		}
		domainInstance, err := domain.New(params, tc.Database(), false)
		if err != nil {
			t.Fatalf("Failed to set up a domain instance: %v", err)
		}
		context := &mocTransactionsRelayContext{
			netAdapter:                  adapter,
			domain:                      domainInstance,
			sharedRequestedTransactions: sharedRequestedTransactions,
		}
		incomingRoute := router.NewRoute()
		defer incomingRoute.Close()
		peerIncomingRoute := router.NewRoute()
		defer peerIncomingRoute.Close()

		txID1 := externalapi.NewDomainTransactionIDFromByteArray(&[externalapi.DomainHashSize]byte{
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01})
		txID2 := externalapi.NewDomainTransactionIDFromByteArray(&[externalapi.DomainHashSize]byte{
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02})
		txIDs := []*externalapi.DomainTransactionID{txID1, txID2}
		invMessage := appmessage.NewMsgInvTransaction(txIDs)
		err = incomingRoute.Enqueue(invMessage)
		if err != nil {
			t.Fatalf("Unexpected error from incomingRoute.Enqueue: %v", err)
		}
		// The goroutine is representing the peer's actions.
		spawn("peerResponseToTheTransactionsRequest", func() {
			msg, err := peerIncomingRoute.Dequeue()
			if err != nil {
				t.Fatalf("Dequeue: %v", err)
			}
			inv := msg.(*appmessage.MsgRequestTransactions)

			if len(txIDs) != len(inv.IDs) {
				t.Fatalf("TestHandleRelayedTransactions: expected %d transactions ID, but got %d", len(txIDs), len(inv.IDs))
			}

			for i, id := range inv.IDs {
				if txIDs[i].String() != id.String() {
					t.Fatalf("TestHandleRelayedTransactions: expected equal txID: expected %s, but got %s", txIDs[i].String(), id.String())
				}
				err = incomingRoute.Enqueue(appmessage.NewMsgTransactionNotFound(txIDs[i]))
				if err != nil {
					t.Fatalf("Unexpected error from incomingRoute.Enqueue: %v", err)
				}
			}
			// Insert an unexpected message type to stop the infinity loop.
			err = incomingRoute.Enqueue(&appmessage.MsgAddresses{})
			if err != nil {
				t.Fatalf("Unexpected error from incomingRoute.Enqueue: %v", err)
			}
		})

		err = transactionrelay.HandleRelayedTransactions(context, incomingRoute, peerIncomingRoute)
		// Since we inserted an unexpected message type to stop the infinity loop,
		// we expect the error will be infected from this specific message and also the
		// error will count as a protocol message.
		if protocolErr := (protocolerrors.ProtocolError{}); err == nil || !errors.As(err, &protocolErr) {
			t.Fatalf("Expected to protocol error")
		} else {
			if !protocolErr.ShouldBan {
				t.Fatalf("Exepcted shouldBan true, but got false.")
			}
			if !strings.Contains(err.Error(), "unexpected Addresses [code 3] message in the block relay flow while expecting an inv message") {
				t.Fatalf("Unexpected error: expected: an error due to existence of an Addresses message "+
					"in the block relay flow, but got: %v", protocolErr.Cause)
			}
		}
	})
}

// TestOnClosedIncomingRoute verifies that an appropriate error message will be returned when
// trying to dequeue a message from a closed route.
func TestOnClosedIncomingRoute(t *testing.T) {
	testutils.ForAllNets(t, true, func(t *testing.T, params *dagconfig.Params) {

		factory := consensus.NewFactory()
		tc, teardown, err := factory.NewTestConsensus(params, false, "TestOnClosedIncomingRoute")
		if err != nil {
			t.Fatalf("Error setting up test consensus: %+v", err)
		}
		defer teardown(false)

		sharedRequestedTransactions := transactionrelay.NewSharedRequestedTransactions()
		adapter, err := netadapter.NewNetAdapter(config.DefaultConfig())
		if err != nil {
			t.Fatalf("Failed to creat a NetAdapter : %v", err)
		}
		domainInstance, err := domain.New(params, tc.Database(), false)
		if err != nil {
			t.Fatalf("Failed to set up a domain instance: %v", err)
		}
		context := &mocTransactionsRelayContext{
			netAdapter:                  adapter,
			domain:                      domainInstance,
			sharedRequestedTransactions: sharedRequestedTransactions,
		}
		incomingRoute := router.NewRoute()
		outgoingRoute := router.NewRoute()
		defer outgoingRoute.Close()

		txID := externalapi.NewDomainTransactionIDFromByteArray(&[externalapi.DomainHashSize]byte{
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01})
		txIDs := []*externalapi.DomainTransactionID{txID}

		err = incomingRoute.Enqueue(&appmessage.MsgInvTransaction{TxIDs: txIDs})
		if err != nil {
			t.Fatalf("Unexpected error from incomingRoute.Enqueue: %v", err)
		}
		incomingRoute.Close()
		err = transactionrelay.HandleRelayedTransactions(context, incomingRoute, outgoingRoute)
		if err == nil || !errors.Is(err, router.ErrRouteClosed) {
			t.Fatalf("Unexpected error: expected: %v, got : %v", router.ErrRouteClosed, err)
		}
	})
}
