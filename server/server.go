package server

import (
	"sync/atomic"

	"github.com/kaspanet/kaspad/blockdag"
	"github.com/kaspanet/kaspad/blockdag/indexers"
	"github.com/kaspanet/kaspad/txscript"
	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/util/mstime"

	"github.com/kaspanet/kaspad/config"
	"github.com/kaspanet/kaspad/dagconfig"
	"github.com/kaspanet/kaspad/mempool"
	"github.com/kaspanet/kaspad/mining"
	"github.com/kaspanet/kaspad/server/rpc"
	"github.com/kaspanet/kaspad/signal"
)

// Server is a wrapper for p2p server and rpc server
type Server struct {
	rpcServer   *rpc.Server
	SigCache    *txscript.SigCache
	DAG         *blockdag.BlockDAG
	Mempool     *mempool.TxPool
	startupTime int64

	started, shutdown int32
}

// Start begins accepting connections from peers.
func (s *Server) Start() {
	// Already started?
	if atomic.AddInt32(&s.started, 1) != 1 {
		return
	}

	log.Trace("Starting server")

	// Server startup time. Used for the uptime command for uptime calculation.
	s.startupTime = mstime.Now().UnixMilliseconds()

	cfg := config.ActiveConfig()

	if !cfg.DisableRPC {
		s.rpcServer.Start()
	}
}

// Stop gracefully shuts down the server by stopping and disconnecting all
// peers and the main listener.
func (s *Server) Stop() error {
	// Make sure this only happens once.
	if atomic.AddInt32(&s.shutdown, 1) != 1 {
		log.Infof("Server is already in the process of shutting down")
		return nil
	}

	log.Warnf("Server shutting down")

	// Shutdown the RPC server if it's not disabled.
	if !config.ActiveConfig().DisableRPC {
		s.rpcServer.Stop()
	}

	return nil
}

// NewServer returns a new kaspad server configured to listen on addr for the
// kaspa network type specified by dagParams. Use start to begin accepting
// connections from peers.
func NewServer(listenAddrs []string, dagParams *dagconfig.Params, interrupt <-chan struct{}) (*Server, error) {
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

	mempoolConfig := setupMempool(dag, sigCache)
	s := &Server{
		SigCache: sigCache,
		DAG:      dag,
		Mempool:  mempool.New(&mempoolConfig),
	}

	server, err := setupRPC(s, dag, sigCache, acceptanceIndex)
	if err != nil {
		return server, err
	}

	return s, nil
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
	var indexManager blockdag.IndexManager
	if len(indexes) > 0 {
		indexManager = indexers.NewManager(indexes)
	}
	return indexManager, acceptanceIndex
}

func setupMempool(dag *blockdag.BlockDAG, sigCache *txscript.SigCache) mempool.Config {
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
	return mempoolConfig
}

func setupRPC(s *Server, dag *blockdag.BlockDAG, sigCache *txscript.SigCache, acceptanceIndex *indexers.AcceptanceIndex) (*Server, error) {
	cfg := config.ActiveConfig()
	if !cfg.DisableRPC {
		policy := mining.Policy{
			BlockMaxMass: cfg.BlockMaxMass,
		}
		blockTemplateGenerator := mining.NewBlkTmplGenerator(&policy,
			s.Mempool, dag, sigCache)

		var err error
		s.rpcServer, err = rpc.NewRPCServer(
			s.startupTime,
			dag,
			s.Mempool,
			acceptanceIndex,
			blockTemplateGenerator,
		)
		if err != nil {
			return nil, err
		}

		// Signal process shutdown when the RPC server requests it.
		spawn(func() {
			<-s.rpcServer.RequestedProcessShutdown()
			signal.ShutdownRequestChannel <- struct{}{}
		})
	}
	return nil, nil
}

// WaitForShutdown blocks until the main listener and peer handlers are stopped.
func (s *Server) WaitForShutdown() {
	// TODO(libp2p)
	//	s.p2pServer.WaitForShutdown()
}
