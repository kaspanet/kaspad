package transactionrelay_test

import (
	"github.com/kaspanet/kaspad/app/protocol/flows/transactionrelay"
	"github.com/kaspanet/kaspad/domain"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/infrastructure/db/database/ldb"
	"io/ioutil"
	"os"
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

func (m *mocTransactionsRelayContext) Broadcast(message appmessage.Message) error {
	return nil
}

func (m *mocTransactionsRelayContext) OnTransactionAddedToMempool() {
}

func newMocTransactionsRelayContext(testName string) (mock *mocTransactionsRelayContext, teardown func(), err error) {
	sharedRequestedTransactions := transactionrelay.NewSharedRequestedTransactions()
	adapter, _ := netadapter.NewNetAdapter(config.DefaultConfig())
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
	domainInstance, err := domain.New(&params, db)
	if err != nil {
		teardown()
		return nil, nil, err
	}

	return &mocTransactionsRelayContext{
		netAdapter:                  adapter,
		domain:                      domainInstance,
		sharedRequestedTransactions: sharedRequestedTransactions,
	}, teardown, nil
}

func TestHandleRelayedTransactions(t *testing.T) {
	context, teardownFunc, err := newMocTransactionsRelayContext(t.Name())
	if err != nil {
		t.Fatalf("Failed to setup new MocTransactionsRelayContext instance: %v", err)
	}
	defer teardownFunc()

	t.Run("Simple call test", func(t *testing.T) {
		incomingRoute := router.NewRoute()
		outgoingRoute := router.NewRoute()

		incomingRoute.Enqueue(&appmessage.MsgInvTransaction{})
		incomingRoute.Enqueue(&appmessage.MsgAddresses{})
		transactionrelay.HandleRelayedTransactions(context, incomingRoute, outgoingRoute)
	})

	t.Run("Test on wrong message type", func(t *testing.T) {
		incomingRoute := router.NewRoute()
		outgoingRoute := router.NewRoute()

		incomingRoute.Enqueue(&appmessage.MsgAddresses{})
		transactionrelay.HandleRelayedTransactions(context, incomingRoute, outgoingRoute)
	})

	t.Run("Test on closed route", func(t *testing.T) {
		incomingRoute := router.NewRoute()
		outgoingRoute := router.NewRoute()

		incomingRoute.Close()
		transactionrelay.HandleRelayedTransactions(context, incomingRoute, outgoingRoute)
	})

	t.Run("Test Outgoing Message", func(t *testing.T) {
		incomingRoute := router.NewRoute()
		outgoingRoute := router.NewRoute()

		txID1 := &externalapi.DomainTransactionID{1, 2, 3}
		txID2 := &externalapi.DomainTransactionID{3, 4, 5}
		txIDs := []*externalapi.DomainTransactionID{txID1, txID2}
		msg := appmessage.NewMsgInvTransaction(txIDs)

		go func() {
			defer func() {
				incomingRoute.Close()
				outgoingRoute.Close()
			}()

			msg, err := outgoingRoute.Dequeue()
			if err != nil {
				t.Fatalf("TestHandleRelayedTransactions: %s", err)
			}

			inv := msg.(*appmessage.MsgRequestTransactions)

			if len(txIDs) != len(inv.IDs) {
				t.Fatalf("TestHandleRelayedTransactions: expected %d got %d", len(txIDs), len(inv.IDs))
			}

			for i, id := range inv.IDs {
				if txIDs[i].String() != id.String() {
					t.Fatalf("TestHandleRelayedTransactions: expected equal txID  %s != %s", txIDs[i].String(), id.String())
				}

				incomingRoute.Enqueue(appmessage.NewMsgTransactionNotFound(id))
			}
		}()

		incomingRoute.Enqueue(msg)
		transactionrelay.HandleRelayedTransactions(context, incomingRoute, outgoingRoute)
	})
}

func TestHandleRequestedTransactions(t *testing.T) {
	context, teardownFunc, err := newMocTransactionsRelayContext(t.Name())
	if err != nil {
		t.Fatalf("Failed to setup new MocTransactionsRelayContext instance: %v", err)
	}
	defer teardownFunc()

	t.Run("Simple call test", func(t *testing.T) {
		incomingRoute := router.NewRoute()
		outgoingRoute := router.NewRoute()

		incomingRoute.Enqueue(&appmessage.MsgRequestTransactions{})
		incomingRoute.Enqueue(&appmessage.MsgRequestTransactions{})
		incomingRoute.Close()
		transactionrelay.HandleRequestedTransactions(context, incomingRoute, outgoingRoute)
	})

	t.Run("Test on wrong message type", func(t *testing.T) {
		incomingRoute := router.NewRoute()
		outgoingRoute := router.NewRoute()

		incomingRoute.Enqueue(&appmessage.MsgAddresses{})
		transactionrelay.HandleRequestedTransactions(context, incomingRoute, outgoingRoute)
	})

	t.Run("Test on closed route", func(t *testing.T) {
		incomingRoute := router.NewRoute()
		outgoingRoute := router.NewRoute()

		incomingRoute.Close()
		outgoingRoute.Close()
		transactionrelay.HandleRequestedTransactions(context, incomingRoute, outgoingRoute)
	})

	t.Run("Test Outgoing Message", func(t *testing.T) {
		incomingRoute := router.NewRoute()
		outgoingRoute := router.NewRoute()

		txID1 := &externalapi.DomainTransactionID{1, 2, 3}
		txID2 := &externalapi.DomainTransactionID{3, 4, 5}
		txIDs := []*externalapi.DomainTransactionID{txID1, txID2}
		msg := appmessage.NewMsgRequestTransactions(txIDs)

		go func() {
			defer func() {
				incomingRoute.Close()
				outgoingRoute.Close()
			}()

			for i, id := range txIDs {
				msg, err := outgoingRoute.Dequeue()
				if err != nil {
					t.Fatalf("TestHandleRequestedTransactions: %s", err)
				}

				outMsg := msg.(*appmessage.MsgTransactionNotFound)

				if txIDs[i].String() != outMsg.ID.String() {
					t.Fatalf("TestHandleRequestedTransactions: expected equal txID  %s != %s", txIDs[i].String(), id.String())
				}
			}
		}()

		incomingRoute.Enqueue(msg)
		transactionrelay.HandleRequestedTransactions(context, incomingRoute, outgoingRoute)
	})
}
