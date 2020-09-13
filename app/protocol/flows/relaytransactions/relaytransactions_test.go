package relaytransactions

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/domain/blockdag"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/kaspanet/kaspad/domain/mempool"
	"github.com/kaspanet/kaspad/domain/txscript"
	"github.com/kaspanet/kaspad/infrastructure/config"
	"github.com/kaspanet/kaspad/infrastructure/db/dbaccess"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
	"github.com/kaspanet/kaspad/util/daghash"
)

type mocTransactionsRelayContext struct {
	netAdapter                  *netadapter.NetAdapter
	dag                         *blockdag.BlockDAG
	txPool                      *mempool.TxPool
	sharedRequestedTransactions *SharedRequestedTransactions
}

func (m *mocTransactionsRelayContext) NetAdapter() *netadapter.NetAdapter {
	return m.netAdapter
}

func (m *mocTransactionsRelayContext) DAG() *blockdag.BlockDAG {
	return m.dag
}

func (m *mocTransactionsRelayContext) SharedRequestedTransactions() *SharedRequestedTransactions {
	return m.sharedRequestedTransactions
}

func (m *mocTransactionsRelayContext) TxPool() *mempool.TxPool {
	return m.txPool
}

func (m *mocTransactionsRelayContext) Broadcast(message appmessage.Message) error {
	return nil
}

func newMocTransactionsRelayContext(dag *blockdag.BlockDAG) *mocTransactionsRelayContext {
	sharedRequestedTransactions := NewSharedRequestedTransactions()
	adapter, _ := netadapter.NewNetAdapter(config.DefaultConfig())
	mempoolConfig := createMempoolConfig(dag)
	txPool := mempool.New(mempoolConfig)

	return &mocTransactionsRelayContext{
		netAdapter:                  adapter,
		dag:                         dag,
		txPool:                      txPool,
		sharedRequestedTransactions: sharedRequestedTransactions,
	}
}

func createDag() (*blockdag.BlockDAG, func()) {
	tempDir := os.TempDir()
	dbPath := filepath.Join(tempDir, "TestRelaytransactions")
	_ = os.RemoveAll(dbPath)

	databaseContext, err := dbaccess.New(dbPath)
	if err != nil {
		return nil, nil
	}

	cfg := &blockdag.Config{
		DAGParams:  &dagconfig.SimnetParams,
		TimeSource: blockdag.NewTimeSource(),
		SigCache:   txscript.NewSigCache(1000),
	}

	cfg.DatabaseContext = databaseContext
	dag, err := blockdag.New(cfg)

	return dag, func() {
		databaseContext.Close()
		os.RemoveAll(dbPath)
	}
}

func createMempoolConfig(dag *blockdag.BlockDAG) *mempool.Config {
	cfg := config.DefaultConfig()
	mempoolConfig := &mempool.Config{
		Policy: mempool.Policy{
			AcceptNonStd:    cfg.RelayNonStd,
			MaxOrphanTxs:    cfg.MaxOrphanTxs,
			MaxOrphanTxSize: config.DefaultMaxOrphanTxSize,
			MinRelayTxFee:   cfg.MinRelayTxFee,
			MaxTxVersion:    1,
		},
		DAG: dag,
	}

	return mempoolConfig
}

func TestHandleRelayedTransactions(t *testing.T) {
	params := dagconfig.SimnetParams
	dagConfig := blockdag.Config{DAGParams: &params}
	dag, teardownFunc, err := blockdag.DAGSetup("TestHandleRelayedTransactions", true, dagConfig)

	if err != nil {
		t.Fatalf("Failed to setup DAG instance: %v", err)
	}
	defer teardownFunc()

	t.Run("Simple call test", func(t *testing.T) {
		incomingRoute := router.NewRoute()
		outgoingRoute := router.NewRoute()
		context := newMocTransactionsRelayContext(dag)
		incomingRoute.Enqueue(&appmessage.MsgInvTransaction{})
		incomingRoute.Enqueue(&appmessage.MsgAddresses{})
		HandleRelayedTransactions(context, incomingRoute, outgoingRoute)
	})

	t.Run("Test on wrong message type", func(t *testing.T) {
		incomingRoute := router.NewRoute()
		outgoingRoute := router.NewRoute()
		context := newMocTransactionsRelayContext(dag)
		incomingRoute.Enqueue(&appmessage.MsgAddresses{})
		HandleRelayedTransactions(context, incomingRoute, outgoingRoute)
	})

	t.Run("Test on closed route", func(t *testing.T) {
		incomingRoute := router.NewRoute()
		outgoingRoute := router.NewRoute()
		context := newMocTransactionsRelayContext(dag)
		incomingRoute.Close()
		HandleRelayedTransactions(context, incomingRoute, outgoingRoute)
	})

	t.Run("Test Outgoing Message", func(t *testing.T) {
		incomingRoute := router.NewRoute()
		outgoingRoute := router.NewRoute()
		context := newMocTransactionsRelayContext(dag)

		txID1 := &daghash.TxID{1, 2, 3}
		txID2 := &daghash.TxID{3, 4, 5}
		txIDs := []*daghash.TxID{txID1, txID2}
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
		HandleRelayedTransactions(context, incomingRoute, outgoingRoute)
	})
}

func TestHandleRequestedTransactions(t *testing.T) {
	params := dagconfig.SimnetParams
	dagConfig := blockdag.Config{DAGParams: &params}
	dag, teardownFunc, err := blockdag.DAGSetup("TestHandleRequestedTransactions", true, dagConfig)

	if err != nil {
		t.Fatalf("Failed to setup DAG instance: %v", err)
	}
	defer teardownFunc()

	t.Run("Simple call test", func(t *testing.T) {
		incomingRoute := router.NewRoute()
		outgoingRoute := router.NewRoute()
		context := newMocTransactionsRelayContext(dag)
		incomingRoute.Enqueue(&appmessage.MsgRequestTransactions{})
		incomingRoute.Enqueue(&appmessage.MsgRequestTransactions{})
		incomingRoute.Close()
		HandleRequestedTransactions(context, incomingRoute, outgoingRoute)
	})

	t.Run("Test on wrong message type", func(t *testing.T) {
		incomingRoute := router.NewRoute()
		outgoingRoute := router.NewRoute()
		context := newMocTransactionsRelayContext(dag)
		incomingRoute.Enqueue(&appmessage.MsgAddresses{})
		HandleRequestedTransactions(context, incomingRoute, outgoingRoute)
	})

	t.Run("Test on closed route", func(t *testing.T) {
		incomingRoute := router.NewRoute()
		outgoingRoute := router.NewRoute()
		context := newMocTransactionsRelayContext(dag)
		incomingRoute.Close()
		outgoingRoute.Close()
		HandleRequestedTransactions(context, incomingRoute, outgoingRoute)
	})

	t.Run("Test Outgoing Message", func(t *testing.T) {
		incomingRoute := router.NewRoute()
		outgoingRoute := router.NewRoute()
		context := newMocTransactionsRelayContext(dag)

		txID1 := &daghash.TxID{1, 2, 3}
		txID2 := &daghash.TxID{3, 4, 5}
		txIDs := []*daghash.TxID{txID1, txID2}
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
		HandleRequestedTransactions(context, incomingRoute, outgoingRoute)
	})
}
