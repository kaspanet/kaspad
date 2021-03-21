package transactionrelay_test

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/protocol/flows/transactionrelay"
	"github.com/kaspanet/kaspad/domain"
	"github.com/kaspanet/kaspad/domain/consensus"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/testutils"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/kaspanet/kaspad/infrastructure/config"
	"github.com/kaspanet/kaspad/infrastructure/logger"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
	"github.com/kaspanet/kaspad/util/panics"
	"github.com/pkg/errors"
	"testing"
)

// TestHandleRequestedTransactions verifies that the flow of  HandleRequestedTransactions
// is working as expected. The goroutine is representing the peer's actions.
func TestHandleRequestedTransactions(t *testing.T) {
	testutils.ForAllNets(t, true, func(t *testing.T, params *dagconfig.Params) {

		var log = logger.RegisterSubSystem("PROT")
		var spawn = panics.GoroutineWrapperFunc(log)
		factory := consensus.NewFactory()
		tc, teardown, err := factory.NewTestConsensus(params, false, "TestHandleRequestedTransactions")
		if err != nil {
			t.Fatalf("Error setting up test Consensus: %+v", err)
		}
		defer teardown(false)

		sharedRequestedTransactions := transactionrelay.NewSharedRequestedTransactions()
		adapter, err := netadapter.NewNetAdapter(config.DefaultConfig())
		if err != nil {
			t.Fatalf("Failed to create a NetAdapter: %v", err)
		}
		domainInstance, err := domain.New(params, tc.Database(), false)
		if err != nil {
			t.Fatalf("Failed to set up a domain Instance: %v", err)
		}
		context := &mocTransactionsRelayContext{
			netAdapter:                  adapter,
			domain:                      domainInstance,
			sharedRequestedTransactions: sharedRequestedTransactions,
		}
		incomingRoute := router.NewRoute()
		outgoingRoute := router.NewRoute()
		defer outgoingRoute.Close()

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
		msg := appmessage.NewMsgRequestTransactions(txIDs)
		err = incomingRoute.Enqueue(msg)
		if err != nil {
			t.Fatalf("Enqueue: %v", err)
		}

		spawn("peerResponseToTheTransactionsMessages", func() {
			for i, id := range txIDs {
				msg, err := outgoingRoute.Dequeue()
				if err != nil {
					t.Fatalf("Dequeue: %s", err)
				}
				outMsg := msg.(*appmessage.MsgTransactionNotFound)
				if txIDs[i].String() != outMsg.ID.String() {
					t.Fatalf("TestHandleRelayedTransactions: expected equal txID: expected to %s, but got %s", txIDs[i].String(), id.String())
				}
			}
			incomingRoute.Close()
		})

		err = transactionrelay.HandleRequestedTransactions(context, incomingRoute, outgoingRoute)
		if err == nil || !errors.Is(err, router.ErrRouteClosed) {
			t.Fatalf("Unexpected error: expected: %v, got : %v", router.ErrRouteClosed, err)
		}
	})
}
