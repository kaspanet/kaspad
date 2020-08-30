package blockrelay

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kaspanet/kaspad/app/appmessage"
	peerpkg "github.com/kaspanet/kaspad/app/protocol/peer"
	"github.com/kaspanet/kaspad/domain/blockdag"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/kaspanet/kaspad/domain/txscript"
	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/util/daghash"

	"github.com/kaspanet/kaspad/infrastructure/config"
	"github.com/kaspanet/kaspad/infrastructure/db/dbaccess"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
)

type mocRelayBlockRequestsContext struct {
	dag *blockdag.BlockDAG
}

func (m *mocRelayBlockRequestsContext) DAG() *blockdag.BlockDAG {
	return m.dag
}

func newMocRelayBlockRequestsContext(dag *blockdag.BlockDAG) *mocRelayBlockRequestsContext {
	return &mocRelayBlockRequestsContext{
		dag: dag,
	}
}

type mocRelayInvsContext struct {
	dag                   *blockdag.BlockDAG
	adapter               *netadapter.NetAdapter
	sharedRequestedBlocks *SharedRequestedBlocks
}

func (m *mocRelayInvsContext) NetAdapter() *netadapter.NetAdapter {
	return m.adapter
}

func (m *mocRelayInvsContext) DAG() *blockdag.BlockDAG {
	return m.dag
}

func (m *mocRelayInvsContext) OnNewBlock(block *util.Block) error {
	return nil
}

func (m *mocRelayInvsContext) SharedRequestedBlocks() *SharedRequestedBlocks {
	return m.sharedRequestedBlocks
}

func (m *mocRelayInvsContext) StartIBDIfRequired() {
}

func (m *mocRelayInvsContext) IsInIBD() bool {
	return false
}

func (m *mocRelayInvsContext) Broadcast(message appmessage.Message) error {
	return nil
}

func newMocRelayInvsContext(dag *blockdag.BlockDAG) *mocRelayInvsContext {
	adapter, _ := netadapter.NewNetAdapter(config.DefaultConfig())

	return &mocRelayInvsContext{
		dag:                   dag,
		adapter:               adapter,
		sharedRequestedBlocks: NewSharedRequestedBlocks(),
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

func TestHandleRelayBlockRequests(t *testing.T) {
	incomingRoute := router.NewRoute()
	outgoingRoute := router.NewRoute()
	peer := peerpkg.New(nil)
	dag, cleanDag := createDag()
	defer cleanDag()

	getRelayBlocksMessage := appmessage.MsgRequestRelayBlocks{
		Hashes: []*daghash.Hash{&daghash.Hash{10}, &daghash.Hash{20}},
	}

	t.Run("Test for invalid arguments 1", func(t *testing.T) {
		context := newMocRelayBlockRequestsContext(dag)
		incomingRoute.Enqueue(&getRelayBlocksMessage)
		HandleRelayBlockRequests(context, incomingRoute, outgoingRoute, peer)
	})

	t.Run("Test for invalid arguments 2", func(t *testing.T) {
		context := newMocRelayBlockRequestsContext(dag)
		incomingRoute.Enqueue(&appmessage.MsgAddresses{})
		HandleRelayBlockRequests(context, incomingRoute, outgoingRoute, peer)
	})

	t.Run("Test for invalid arguments 3", func(t *testing.T) {
		context := newMocRelayBlockRequestsContext(nil)
		HandleRelayBlockRequests(context, incomingRoute, outgoingRoute, peer)
	})
}

func TestHandleRelayInvs(t *testing.T) {
	peer := peerpkg.New(nil)
	dag, cleanDag := createDag()
	defer cleanDag()

	msgInvRelayBlock := appmessage.MsgInvRelayBlock{
		Hash: &daghash.Hash{10},
	}

	t.Run("Test for invalid arguments 1", func(t *testing.T) {
		context := newMocRelayInvsContext(dag)
		incomingRoute := router.NewRoute()
		outgoingRoute := router.NewRoute()
		incomingRoute.Enqueue(&msgInvRelayBlock)
		incomingRoute.Enqueue(&appmessage.MsgAddresses{})
		HandleRelayInvs(context, incomingRoute, outgoingRoute, peer)
	})

	t.Run("Test for invalid arguments 2", func(t *testing.T) {
		context := newMocRelayInvsContext(dag)
		incomingRoute := router.NewRoute()
		outgoingRoute := router.NewRoute()
		incomingRoute.Enqueue(&appmessage.MsgAddresses{})
		incomingRoute.Enqueue(&appmessage.MsgAddresses{})
		HandleRelayInvs(context, incomingRoute, outgoingRoute, peer)
	})

	t.Run("Test for invalid arguments 3", func(t *testing.T) {
		context := newMocRelayInvsContext(nil)
		incomingRoute := router.NewRoute()
		outgoingRoute := router.NewRoute()
		incomingRoute.Enqueue(&msgInvRelayBlock)
		HandleRelayInvs(context, incomingRoute, outgoingRoute, peer)
	})

}
