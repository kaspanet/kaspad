package kaspad

import (
	"fmt"
	"sync/atomic"

	"github.com/kaspanet/kaspad/addrmgr"
	"github.com/kaspanet/kaspad/blockdag"
	"github.com/kaspanet/kaspad/blockdag/indexers"
	"github.com/kaspanet/kaspad/config"
	"github.com/kaspanet/kaspad/connmanager"
	"github.com/kaspanet/kaspad/dbaccess"
	"github.com/kaspanet/kaspad/dnsseed"
	"github.com/kaspanet/kaspad/mempool"
	"github.com/kaspanet/kaspad/mining"
	"github.com/kaspanet/kaspad/netadapter"
	"github.com/kaspanet/kaspad/protocol"
	"github.com/kaspanet/kaspad/rpc"
	"github.com/kaspanet/kaspad/signal"
	"github.com/kaspanet/kaspad/txscript"
	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/util/panics"
	"github.com/kaspanet/kaspad/wire"
)

// Kaspad is a wrapper for all the kaspad services
type Kaspad struct {
	cfg               *config.Config
	rpcServer         *rpc.Server
	addressManager    *addrmgr.AddrManager
	protocolManager   *protocol.Manager
	connectionManager *connmanager.ConnectionManager
	NetAdapter        *netadapter.NetAdapter

	started, shutdown int32
}

// Kaspad start launches all the kaspad services.
func (k *Kaspad) Start() {
	// Already started?
	if atomic.AddInt32(&k.started, 1) != 1 {
		return
	}

	log.Trace("Starting kaspad")

	err := k.protocolManager.Start()
	if err != nil {
		panics.Exit(log, fmt.Sprintf("Error starting the p2p protocol: %+v", err))
	}

	k.maybeSeedFromDNS()

	k.connectionManager.Start()

	if !k.cfg.DisableRPC {
		k.rpcServer.Start()
	}
}

// stop gracefully shuts down all the kaspad services.
func (k *Kaspad) Stop() error {
	// Make sure this only happens once.
	if atomic.AddInt32(&k.shutdown, 1) != 1 {
		log.Infof("Kaspad is already in the process of shutting down")
		return nil
	}

	log.Warnf("Kaspad shutting down")

	k.connectionManager.Stop()

	err := k.protocolManager.Stop()
	if err != nil {
		log.Errorf("Error stopping the p2p protocol: %+v", err)
	}

	// Shutdown the RPC server if it's not disabled.
	if !k.cfg.DisableRPC {
		err := k.rpcServer.Stop()
		if err != nil {
			log.Errorf("Error stopping rpcServer: %+v", err)
		}
	}

	return nil
}

// newKaspad returns a new kaspad instance configured to listen on addr for the
// kaspa network type specified by dagParams. Use start to begin accepting
// connections from peers.
func New(cfg *config.Config, databaseContext *dbaccess.DatabaseContext, interrupt <-chan struct{}) (*Kaspad, error) {
	indexManager, acceptanceIndex := setupIndexes(cfg)

	sigCache := txscript.NewSigCache(cfg.SigCacheMaxSize)

	// Create a new block DAG instance with the appropriate configuration.
	dag, err := setupDAG(cfg, databaseContext, interrupt, sigCache, indexManager)
	if err != nil {
		return nil, err
	}

	txMempool := setupMempool(cfg, dag, sigCache)

	netAdapter, err := netadapter.NewNetAdapter(cfg)
	if err != nil {
		return nil, err
	}
	addressManager := addrmgr.New(cfg, databaseContext)

	protocolManager, err := protocol.NewManager(cfg, dag, netAdapter, addressManager, txMempool)
	if err != nil {
		return nil, err
	}

	connectionManager, err := connmanager.New(cfg, netAdapter, addressManager)
	if err != nil {
		return nil, err
	}
	rpcServer, err := setupRPC(
		cfg, dag, txMempool, sigCache, acceptanceIndex, connectionManager, addressManager, protocolManager)
	if err != nil {
		return nil, err
	}

	return &Kaspad{
		cfg:               cfg,
		rpcServer:         rpcServer,
		protocolManager:   protocolManager,
		connectionManager: connectionManager,
		NetAdapter:        netAdapter,
	}, nil
}

func (k *Kaspad) maybeSeedFromDNS() {
	if !k.cfg.DisableDNSSeed {
		dnsseed.SeedFromDNS(k.cfg.NetParams(), k.cfg.DNSSeed, wire.SFNodeNetwork, false, nil,
			k.cfg.Lookup, func(addresses []*wire.NetAddress) {
				// Kaspad uses a lookup of the dns seeder here. Since seeder returns
				// IPs of nodes and not its own IP, we can not know real IP of
				// source. So we'll take first returned address as source.
				k.addressManager.AddAddresses(addresses, addresses[0], nil)
			})
	}
}
func setupDAG(cfg *config.Config, databaseContext *dbaccess.DatabaseContext, interrupt <-chan struct{},
	sigCache *txscript.SigCache, indexManager blockdag.IndexManager) (*blockdag.BlockDAG, error) {

	dag, err := blockdag.New(&blockdag.Config{
		Interrupt:       interrupt,
		DatabaseContext: databaseContext,
		DAGParams:       cfg.NetParams(),
		TimeSource:      blockdag.NewTimeSource(),
		SigCache:        sigCache,
		IndexManager:    indexManager,
		SubnetworkID:    cfg.SubnetworkID,
	})
	return dag, err
}

func setupIndexes(cfg *config.Config) (blockdag.IndexManager, *indexers.AcceptanceIndex) {
	// Create indexes if needed.
	var indexes []indexers.Indexer
	var acceptanceIndex *indexers.AcceptanceIndex
	if cfg.AcceptanceIndex {
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

func setupMempool(cfg *config.Config, dag *blockdag.BlockDAG, sigCache *txscript.SigCache) *mempool.TxPool {
	mempoolConfig := mempool.Config{
		Policy: mempool.Policy{
			AcceptNonStd:    cfg.RelayNonStd,
			MaxOrphanTxs:    cfg.MaxOrphanTxs,
			MaxOrphanTxSize: config.DefaultMaxOrphanTxSize,
			MinRelayTxFee:   cfg.MinRelayTxFee,
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

func setupRPC(cfg *config.Config,
	dag *blockdag.BlockDAG,
	txMempool *mempool.TxPool,
	sigCache *txscript.SigCache,
	acceptanceIndex *indexers.AcceptanceIndex,
	connectionManager *connmanager.ConnectionManager,
	addressManager *addrmgr.AddrManager,
	protocolManager *protocol.Manager) (*rpc.Server, error) {

	if !cfg.DisableRPC {
		policy := mining.Policy{
			BlockMaxMass: cfg.BlockMaxMass,
		}
		blockTemplateGenerator := mining.NewBlkTmplGenerator(&policy, txMempool, dag, sigCache)

		rpcServer, err := rpc.NewRPCServer(cfg, dag, txMempool, acceptanceIndex, blockTemplateGenerator,
			connectionManager, addressManager, protocolManager)
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
func (k *Kaspad) WaitForShutdown() {
	// TODO(libp2p)
	//	k.p2pServer.WaitForShutdown()
}
