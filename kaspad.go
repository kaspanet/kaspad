package main

import (
	"sync/atomic"

	"github.com/kaspanet/kaspad/blockdag"
	"github.com/kaspanet/kaspad/blockdag/indexers"
	"github.com/kaspanet/kaspad/config"
	"github.com/kaspanet/kaspad/dagconfig"
	"github.com/kaspanet/kaspad/mempool"
	"github.com/kaspanet/kaspad/mining"
	"github.com/kaspanet/kaspad/server/rpc"
	"github.com/kaspanet/kaspad/signal"
	"github.com/kaspanet/kaspad/txscript"
	"github.com/kaspanet/kaspad/util"
)

// Kaspad is a wrapper for all the kaspad services
type kaspad struct {
	rpcServer *rpc.Server

	started, shutdown int32
}

// Start launches all the kaspad services.
func (s *kaspad) start() {
	// Already started?
	if atomic.AddInt32(&s.started, 1) != 1 {
		return
	}

	log.Trace("Starting kaspad")

	cfg := config.ActiveConfig()

	if !cfg.DisableRPC {
		s.rpcServer.Start()
	}
}

// Stop gracefully shuts down all the kaspad services.
func (s *kaspad) stop() error {
	// Make sure this only happens once.
	if atomic.AddInt32(&s.shutdown, 1) != 1 {
		log.Infof("Kaspad is already in the process of shutting down")
		return nil
	}

	log.Warnf("Kaspad shutting down")

	// Shutdown the RPC server if it's not disabled.
	if !config.ActiveConfig().DisableRPC {
		s.rpcServer.Stop()
	}

	return nil
}

// newKaspad returns a new kaspad instance configured to listen on addr for the
// kaspa network type specified by dagParams. Use start to begin accepting
// connections from peers.
func newKaspad(listenAddrs []string, dagParams *dagconfig.Params, interrupt <-chan struct{}) (*kaspad, error) {
	indexManager, acceptanceIndex := setupIndexes()

	sigCache := txscript.NewSigCache(config.ActiveConfig().SigCacheMaxSize)

	// Create a new block DAG instance with the appropriate configuration.
	dag, err := blockdag.New(&blockdag.Config{
		Interrupt:    interrupt,
		DAGParams:    dagParams,
		TimeSource:   blockdag.NewTimeSource(),
		SigCache:     sigCache,
		IndexManager: indexManager,
		SubnetworkID: config.ActiveConfig().SubnetworkID,
	})
	if err != nil {
		return nil, err
	}

	txMempool := setupMempool(dag, sigCache)

	rpcServer, err := setupRPC(dag, txMempool, sigCache, acceptanceIndex)
	if err != nil {
		return nil, err
	}

	return &kaspad{
		rpcServer: rpcServer,
	}, nil
}

func setupIndexes() (blockdag.IndexManager, *indexers.AcceptanceIndex) {
	// Create indexes if needed.
	var indexes []indexers.Indexer
	var acceptanceIndex *indexers.AcceptanceIndex
	if config.ActiveConfig().AcceptanceIndex {
		log.Info("acceptance index is enabled")
		indexes = append(indexes, acceptanceIndex)
	}

	// Create an index manager if any of the optional indexes are enabled.
	if len(indexes) < 0 {
		return nil, nil
	}
	indexManager := indexers.NewManager(indexes)
	return indexManager, acceptanceIndex
}

func setupMempool(dag *blockdag.BlockDAG, sigCache *txscript.SigCache) *mempool.TxPool {
	mempoolConfig := mempool.Config{
		Policy: mempool.Policy{
			AcceptNonStd:    config.ActiveConfig().RelayNonStd,
			MaxOrphanTxs:    config.ActiveConfig().MaxOrphanTxs,
			MaxOrphanTxSize: config.DefaultMaxOrphanTxSize,
			MinRelayTxFee:   config.ActiveConfig().MinRelayTxFee,
			MaxTxVersion:    1,
		},
		CalcSequenceLockNoLock: func(tx *util.Tx, utxoSet blockdag.UTXOSet) (*blockdag.SequenceLock, error) {
			return dag.CalcSequenceLockNoLock(tx, utxoSet, true)
		},
		IsDeploymentActive: dag.IsDeploymentActive,
		SigCache:           sigCache,
		DAG:                dag,
	}

	return mempool.New(&mempoolConfig)
}

func setupRPC(dag *blockdag.BlockDAG, txMempool *mempool.TxPool, sigCache *txscript.SigCache,
	acceptanceIndex *indexers.AcceptanceIndex) (*rpc.Server, error) {
	cfg := config.ActiveConfig()
	if !cfg.DisableRPC {
		policy := mining.Policy{
			BlockMaxMass: cfg.BlockMaxMass,
		}
		blockTemplateGenerator := mining.NewBlkTmplGenerator(&policy, txMempool, dag, sigCache)

		rpcServer, err := rpc.NewRPCServer(dag, txMempool, acceptanceIndex, blockTemplateGenerator)
		if err != nil {
			return nil, err
		}

		// Signal process shutdown when the RPC server requests it.
		spawn(func() {
			<-rpcServer.RequestedProcessShutdown()
			signal.ShutdownRequestChannel <- struct{}{}
		})

		return rpcServer, nil
	}
	return nil, nil
}

// WaitForShutdown blocks until the main listener and peer handlers are stopped.
func (s *kaspad) WaitForShutdown() {
	// TODO(libp2p)
	//	s.p2pServer.WaitForShutdown()
}
