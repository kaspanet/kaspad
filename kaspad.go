package main

import (
	"fmt"
	"sync/atomic"

	"github.com/kaspanet/kaspad/dnsseed"
	"github.com/kaspanet/kaspad/wire"

	"github.com/kaspanet/kaspad/connmanager"

	"github.com/kaspanet/kaspad/addrmgr"
	"github.com/kaspanet/kaspad/server/serverutils"

	"github.com/kaspanet/kaspad/netadapter"

	"github.com/kaspanet/kaspad/util/panics"

	"github.com/kaspanet/kaspad/blockdag"
	"github.com/kaspanet/kaspad/blockdag/indexers"
	"github.com/kaspanet/kaspad/config"
	"github.com/kaspanet/kaspad/mempool"
	"github.com/kaspanet/kaspad/mining"
	"github.com/kaspanet/kaspad/protocol"
	"github.com/kaspanet/kaspad/server/rpc"
	"github.com/kaspanet/kaspad/signal"
	"github.com/kaspanet/kaspad/txscript"
	"github.com/kaspanet/kaspad/util"
)

// kaspad is a wrapper for all the kaspad services
type kaspad struct {
	rpcServer         *rpc.Server
	addressManager    *addrmgr.AddrManager
	networkAdapter    *netadapter.NetAdapter
	connectionManager *connmanager.ConnectionManager

	started, shutdown int32
}

// start launches all the kaspad services.
func (s *kaspad) start() {
	// Already started?
	if atomic.AddInt32(&s.started, 1) != 1 {
		return
	}

	log.Trace("Starting kaspad")

	cfg := config.ActiveConfig()

	err := s.networkAdapter.Start()
	if err != nil {
		panics.Exit(log, fmt.Sprintf("Error starting the p2p protocol: %+v", err))
	}

	maybeSeedFromDNS(cfg, s.addressManager)

	s.connectionManager.Start()

	if !cfg.DisableRPC {
		s.rpcServer.Start()
	}
}

func maybeSeedFromDNS(cfg *config.Config, addressManager *addrmgr.AddrManager) {
	if !cfg.DisableDNSSeed {
		dnsseed.SeedFromDNS(cfg.NetParams(), wire.SFNodeNetwork, false, nil,
			config.ActiveConfig().Lookup, func(addresses []*wire.NetAddress) {
				// Kaspad uses a lookup of the dns seeder here. Since seeder returns
				// IPs of nodes and not its own IP, we can not know real IP of
				// source. So we'll take first returned address as source.
				addressManager.AddAddresses(addresses, addresses[0], nil)
			})
	}
}

// stop gracefully shuts down all the kaspad services.
func (s *kaspad) stop() error {
	// Make sure this only happens once.
	if atomic.AddInt32(&s.shutdown, 1) != 1 {
		log.Infof("Kaspad is already in the process of shutting down")
		return nil
	}

	log.Warnf("Kaspad shutting down")

	s.connectionManager.Stop()

	err := s.networkAdapter.Stop()
	if err != nil {
		log.Errorf("Error stopping the p2p protocol: %+v", err)
	}

	// Shutdown the RPC server if it's not disabled.
	if !config.ActiveConfig().DisableRPC {
		err := s.rpcServer.Stop()
		if err != nil {
			log.Errorf("Error stopping rpcServer: %+v", err)
		}
	}

	return nil
}

// newKaspad returns a new kaspad instance configured to listen on addr for the
// kaspa network type specified by dagParams. Use start to begin accepting
// connections from peers.
func newKaspad(interrupt <-chan struct{}) (*kaspad, error) {
	cfg := config.ActiveConfig()

	indexManager, acceptanceIndex := setupIndexes()

	sigCache := txscript.NewSigCache(cfg.SigCacheMaxSize)

	// Create a new block DAG instance with the appropriate configuration.
	dag, err := setupDAG(interrupt, sigCache, indexManager)
	if err != nil {
		return nil, err
	}

	txMempool := setupMempool(dag, sigCache)

	netAdapter, err := netadapter.NewNetAdapter(cfg.Listeners)
	if err != nil {
		return nil, err
	}
	addressManager := addrmgr.New(serverutils.KaspadLookup, config.ActiveConfig().SubnetworkID)

	protocol.Init(netAdapter, addressManager, dag)

	connectionManager, err := connmanager.New(netAdapter, addressManager)
	if err != nil {
		return nil, err
	}

	rpcServer, err := setupRPC(dag, txMempool, sigCache, acceptanceIndex)
	if err != nil {
		return nil, err
	}

	return &kaspad{
		rpcServer:         rpcServer,
		networkAdapter:    netAdapter,
		connectionManager: connectionManager,
	}, nil
}

func setupDAG(interrupt <-chan struct{}, sigCache *txscript.SigCache, indexManager blockdag.IndexManager) (*blockdag.BlockDAG, error) {
	dag, err := blockdag.New(&blockdag.Config{
		Interrupt:    interrupt,
		DAGParams:    config.ActiveConfig().NetParams(),
		TimeSource:   blockdag.NewTimeSource(),
		SigCache:     sigCache,
		IndexManager: indexManager,
		SubnetworkID: config.ActiveConfig().SubnetworkID,
	})
	return dag, err
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
		spawn("setupRPC-handleShutdownRequest", func() {
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
