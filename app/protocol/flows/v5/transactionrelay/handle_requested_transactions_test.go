package transactionrelay_test

import (
	"testing"

	"github.com/zoomy-network/zoomyd/app/protocol/flowcontext"
	"github.com/zoomy-network/zoomyd/app/protocol/flows/v5/transactionrelay"

	"github.com/pkg/errors"
	"github.com/zoomy-network/zoomyd/app/appmessage"
	"github.com/zoomy-network/zoomyd/domain"
	"github.com/zoomy-network/zoomyd/domain/consensus"
	"github.com/zoomy-network/zoomyd/domain/consensus/model/externalapi"
	"github.com/zoomy-network/zoomyd/domain/consensus/utils/testutils"
	"github.com/zoomy-network/zoomyd/domain/miningmanager/mempool"
	"github.com/zoomy-network/zoomyd/infrastructure/config"
	"github.com/zoomy-network/zoomyd/infrastructure/logger"
	"github.com/zoomy-network/zoomyd/infrastructure/network/netadapter"
	"github.com/zoomy-network/zoomyd/infrastructure/network/netadapter/router"
	"github.com/zoomy-network/zoomyd/util/panics"
)

// TestHandleRequestedTransactionsNotFound tests the flow of  HandleRequestedTransactions
// when the requested transactions don't found in the mempool.
func TestHandleRequestedTransactionsNotFound(t *testing.T) {
	testutils.ForAllNets(t, true, func(t *testing.T, consensusConfig *consensus.Config) {
		var log = logger.RegisterSubSystem("PROT")
		var spawn = panics.GoroutineWrapperFunc(log)
		factory := consensus.NewFactory()
		tc, teardown, err := factory.NewTestConsensus(consensusConfig, "TestHandleRequestedTransactionsNotFound")
		if err != nil {
			t.Fatalf("Error setting up test Consensus: %+v", err)
		}
		defer teardown(false)

		sharedRequestedTransactions := flowcontext.NewSharedRequestedTransactions()
		adapter, err := netadapter.NewNetAdapter(config.DefaultConfig())
		if err != nil {
			t.Fatalf("Failed to create a NetAdapter: %v", err)
		}
		domainInstance, err := domain.New(consensusConfig, mempool.DefaultConfig(&consensusConfig.Params), tc.Database())
		if err != nil {
			t.Fatalf("Failed to set up a domain Instance: %v", err)
		}
		context := &mocTransactionsRelayContext{
			netAdapter:                  adapter,
			domain:                      domainInstance,
			sharedRequestedTransactions: sharedRequestedTransactions,
		}
		incomingRoute := router.NewRoute("incoming")
		outgoingRoute := router.NewRoute("outgoing")
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
			t.Fatalf("Unexpected error from incomingRoute.Enqueue: %v", err)
		}
		// The goroutine is representing the peer's actions.
		spawn("peerResponseToTheTransactionsMessages", func() {
			for i, id := range txIDs {
				msg, err := outgoingRoute.Dequeue()
				if err != nil {
					t.Fatalf("Dequeue: %s", err)
				}
				outMsg := msg.(*appmessage.MsgTransactionNotFound)
				if txIDs[i].String() != outMsg.ID.String() {
					t.Fatalf("TestHandleRelayedTransactions: expected equal txID: expected %s, but got %s", txIDs[i].String(), id.String())
				}
			}
			// Closed the incomingRoute for stop the infinity loop.
			incomingRoute.Close()
		})

		err = transactionrelay.HandleRequestedTransactions(context, incomingRoute, outgoingRoute)
		// Make sure the error is due to the closed route.
		if err == nil || !errors.Is(err, router.ErrRouteClosed) {
			t.Fatalf("Unexpected error: expected: %v, got : %v", router.ErrRouteClosed, err)
		}
	})
}
