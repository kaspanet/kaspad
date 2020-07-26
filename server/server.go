package server

import (
	"github.com/kaspanet/kaspad/util/mstime"
	"sync/atomic"

	"github.com/kaspanet/kaspad/config"
	"github.com/kaspanet/kaspad/dagconfig"
	"github.com/kaspanet/kaspad/mempool"
	"github.com/kaspanet/kaspad/mining"
	"github.com/kaspanet/kaspad/server/p2p"
	"github.com/kaspanet/kaspad/server/rpc"
	"github.com/kaspanet/kaspad/signal"
)

// Server is a wrapper for p2p server and rpc server
type Server struct {
	rpcServer   *rpc.Server
	p2pServer   *p2p.Server
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

	s.p2pServer.Start()

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

	s.p2pServer.Stop()

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
	s := &Server{}
	var err error
	notifyNewTransactions := func(txns []*mempool.TxDesc) {
		// Notify both websocket and getblocktemplate long poll clients of all
		// newly accepted transactions.
		if s.rpcServer != nil {
			s.rpcServer.NotifyNewTransactions(txns)
		}
	}
	s.p2pServer, err = p2p.NewServer(listenAddrs, dagParams, interrupt, notifyNewTransactions)
	if err != nil {
		return nil, err
	}

	cfg := config.ActiveConfig()

	// Create the mining policy and block template generator based on the
	// configuration options.
	policy := mining.Policy{
		BlockMaxMass: cfg.BlockMaxMass,
	}
	blockTemplateGenerator := mining.NewBlkTmplGenerator(&policy,
		s.p2pServer.DAGParams, s.p2pServer.TxMemPool, s.p2pServer.DAG, s.p2pServer.TimeSource, s.p2pServer.SigCache)

	if !cfg.DisableRPC {
		s.rpcServer, err = rpc.NewRPCServer(
			s.startupTime,
			s.p2pServer,
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

	return s, nil
}

// WaitForShutdown blocks until the main listener and peer handlers are stopped.
func (s *Server) WaitForShutdown() {
	s.p2pServer.WaitForShutdown()
}
