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
		dag:                         nil,
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
	incomingRoute := router.NewRoute()
	outgoingRoute := router.NewRoute()
	dag, cleanDag := createDag()
	defer cleanDag()

	t.Run("Test for invalid arguments 1", func(t *testing.T) {
		context := newMocTransactionsRelayContext(dag)
		incomingRoute.Enqueue(&appmessage.MsgAddresses{})
		HandleRelayedTransactions(context, incomingRoute, outgoingRoute)
	})

	t.Run("Test for invalid arguments 2", func(t *testing.T) {
		context := newMocTransactionsRelayContext(dag)
		incomingRoute.Enqueue(&appmessage.MsgInvTransaction{})
		incomingRoute.Enqueue(&appmessage.MsgAddresses{})
		HandleRelayedTransactions(context, incomingRoute, outgoingRoute)
	})

	t.Run("Test for invalid arguments 3", func(t *testing.T) {
		context := newMocTransactionsRelayContext(dag)
		incomingRoute.Close()
		HandleRelayedTransactions(context, incomingRoute, outgoingRoute)
	})
}

func TestHandleRequestedTransactions(t *testing.T) {
	incomingRoute := router.NewRoute()
	outgoingRoute := router.NewRoute()
	dag, cleanDag := createDag()
	defer cleanDag()

	t.Run("Test for invalid arguments 1", func(t *testing.T) {
		context := newMocTransactionsRelayContext(dag)
		incomingRoute.Enqueue(&appmessage.MsgAddresses{})
		HandleRequestedTransactions(context, incomingRoute, outgoingRoute)
	})

	t.Run("Test for invalid arguments 2", func(t *testing.T) {
		context := newMocTransactionsRelayContext(dag)
		incomingRoute.Enqueue(&appmessage.MsgRequestTransactions{})
		incomingRoute.Enqueue(&appmessage.MsgAddresses{})
		HandleRequestedTransactions(context, incomingRoute, outgoingRoute)
	})

	t.Run("Test for invalid arguments 3", func(t *testing.T) {
		context := newMocTransactionsRelayContext(nil)
		incomingRoute.Enqueue(&appmessage.MsgRequestTransactions{})
		incomingRoute.Enqueue(&appmessage.MsgAddresses{})
		HandleRequestedTransactions(context, incomingRoute, outgoingRoute)
	})
}
