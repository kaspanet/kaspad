package server

import (
	"sync/atomic"
	"time"

	"github.com/daglabs/kaspad/config"
	"github.com/daglabs/kaspad/dagconfig"
	"github.com/daglabs/kaspad/database"
	"github.com/daglabs/kaspad/mempool"
	"github.com/daglabs/kaspad/mining"
	"github.com/daglabs/kaspad/mining/cpuminer"
	"github.com/daglabs/kaspad/server/p2p"
	"github.com/daglabs/kaspad/server/rpc"
	"github.com/daglabs/kaspad/signal"
)

// Server is a wrapper for p2p server and rpc server
type Server struct {
	rpcServer   *rpc.Server
	p2pServer   *p2p.Server
	cpuminer    *cpuminer.CPUMiner
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
	s.startupTime = time.Now().Unix()

	s.p2pServer.Start()

	// Start the CPU miner if generation is enabled.
	cfg := config.ActiveConfig()
	if cfg.Generate {
		s.cpuminer.Start()
	}

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

	// Stop the CPU miner if needed
	s.cpuminer.Stop()

	s.p2pServer.Stop()

	// Shutdown the RPC server if it's not disabled.
	if !config.ActiveConfig().DisableRPC {
		s.rpcServer.Stop()
	}

	return nil
}

// NewServer returns a new btcd server configured to listen on addr for the
// bitcoin network type specified by chainParams.  Use start to begin accepting
// connections from peers.
func NewServer(listenAddrs []string, db database.DB, dagParams *dagconfig.Params, interrupt <-chan struct{}) (*Server, error) {
	s := &Server{}
	var err error
	notifyNewTransactions := func(txns []*mempool.TxDesc) {
		// Notify both websocket and getblocktemplate long poll clients of all
		// newly accepted transactions.
		if s.rpcServer != nil {
			s.rpcServer.NotifyNewTransactions(txns)
		}
	}
	s.p2pServer, err = p2p.NewServer(listenAddrs, db, dagParams, interrupt, notifyNewTransactions)
	if err != nil {
		return nil, err
	}

	cfg := config.ActiveConfig()

	// Create the mining policy and block template generator based on the
	// configuration options.
	//
	// NOTE: The CPU miner relies on the mempool, so the mempool has to be
	// created before calling the function to create the CPU miner.
	policy := mining.Policy{
		BlockMaxMass: cfg.BlockMaxMass,
	}
	blockTemplateGenerator := mining.NewBlkTmplGenerator(&policy,
		s.p2pServer.DAGParams, s.p2pServer.TxMemPool, s.p2pServer.DAG, s.p2pServer.TimeSource, s.p2pServer.SigCache)
	s.cpuminer = cpuminer.New(&cpuminer.Config{
		DAGParams:              dagParams,
		BlockTemplateGenerator: blockTemplateGenerator,
		MiningAddrs:            cfg.MiningAddrs,
		ProcessBlock:           s.p2pServer.SyncManager.ProcessBlock,
		ConnectedCount:         s.p2pServer.ConnectedCount,
		ShouldMineOnGenesis:    s.p2pServer.ShouldMineOnGenesis,
		IsCurrent:              s.p2pServer.SyncManager.IsCurrent,
	})

	if !cfg.DisableRPC {

		s.rpcServer, err = rpc.NewRPCServer(
			s.startupTime,
			s.p2pServer,
			db,
			blockTemplateGenerator,
			s.cpuminer,
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

	return s, nil
}

// WaitForShutdown blocks until the main listener and peer handlers are stopped.
func (s *Server) WaitForShutdown() {
	s.p2pServer.WaitForShutdown()
}
