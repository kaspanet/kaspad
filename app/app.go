package app

import (
	"fmt"
	"sync/atomic"

	"github.com/kaspanet/kaspad/infrastructure/network/addressmanager"

	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/id"

	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/protocol"
	"github.com/kaspanet/kaspad/app/rpc"
	"github.com/kaspanet/kaspad/domain/blockdag"
	"github.com/kaspanet/kaspad/domain/blockdag/indexers"
	"github.com/kaspanet/kaspad/domain/mempool"
	"github.com/kaspanet/kaspad/domain/mining"
	"github.com/kaspanet/kaspad/domain/txscript"
	"github.com/kaspanet/kaspad/infrastructure/config"
	"github.com/kaspanet/kaspad/infrastructure/db/dbaccess"
	"github.com/kaspanet/kaspad/infrastructure/network/connmanager"
	"github.com/kaspanet/kaspad/infrastructure/network/dnsseed"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter"
	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/util/panics"
)

// App is a wrapper for all the kaspad services
type App struct {
	cfg               *config.Config
	addressManager    *addressmanager.AddressManager
	protocolManager   *protocol.Manager
	rpcManager        *rpc.Manager
	connectionManager *connmanager.ConnectionManager
	netAdapter        *netadapter.NetAdapter

	started, shutdown int32
}

// Start launches all the kaspad services.
func (a *App) Start() {
	// Already started?
	if atomic.AddInt32(&a.started, 1) != 1 {
		return
	}

	log.Trace("Starting kaspad")

	err := a.netAdapter.Start()
	if err != nil {
		panics.Exit(log, fmt.Sprintf("Error starting the net adapter: %+v", err))
	}

	a.maybeSeedFromDNS()

	a.connectionManager.Start()
}

// Stop gracefully shuts down all the kaspad services.
func (a *App) Stop() {
	// Make sure this only happens once.
	if atomic.AddInt32(&a.shutdown, 1) != 1 {
		log.Infof("Kaspad is already in the process of shutting down")
		return
	}

	log.Warnf("Kaspad shutting down")

	a.connectionManager.Stop()

	err := a.netAdapter.Stop()
	if err != nil {
		log.Errorf("Error stopping the net adapter: %+v", err)
	}

	return
}

// New returns a new App instance configured to listen on addr for the
// kaspa network type specified by dagParams. Use start to begin accepting
// connections from peers.
func New(cfg *config.Config, databaseContext *dbaccess.DatabaseContext, interrupt <-chan struct{}) (*App, error) {
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
	addressManager, err := addressmanager.New(addressmanager.NewConfig(cfg))
	if err != nil {
		return nil, err
	}
	connectionManager, err := connmanager.New(cfg, netAdapter, addressManager)
	if err != nil {
		return nil, err
	}
	protocolManager, err := protocol.NewManager(cfg, dag, netAdapter, addressManager, txMempool, connectionManager)
	if err != nil {
		return nil, err
	}
	rpcManager := setupRPC(cfg, txMempool, dag, sigCache, netAdapter, protocolManager, connectionManager, addressManager, acceptanceIndex)

	return &App{
		cfg:               cfg,
		protocolManager:   protocolManager,
		rpcManager:        rpcManager,
		connectionManager: connectionManager,
		netAdapter:        netAdapter,
		addressManager:    addressManager,
	}, nil
}

func setupRPC(
	cfg *config.Config,
	txMempool *mempool.TxPool,
	dag *blockdag.BlockDAG,
	sigCache *txscript.SigCache,
	netAdapter *netadapter.NetAdapter,
	protocolManager *protocol.Manager,
	connectionManager *connmanager.ConnectionManager,
	addressManager *addressmanager.AddressManager,
	acceptanceIndex *indexers.AcceptanceIndex) *rpc.Manager {

	blockTemplateGenerator := mining.NewBlkTmplGenerator(&mining.Policy{BlockMaxMass: cfg.BlockMaxMass}, txMempool, dag, sigCache)
	rpcManager := rpc.NewManager(cfg, netAdapter, dag, protocolManager, connectionManager, blockTemplateGenerator, txMempool, addressManager, acceptanceIndex)
	protocolManager.SetOnBlockAddedToDAGHandler(rpcManager.NotifyBlockAddedToDAG)
	protocolManager.SetOnTransactionAddedToMempoolHandler(rpcManager.NotifyTransactionAddedToMempool)
	dag.Subscribe(func(notification *blockdag.Notification) {
		err := handleBlockDAGNotifications(notification, acceptanceIndex, rpcManager)
		if err != nil {
			panic(err)
		}
	})
	return rpcManager
}

func handleBlockDAGNotifications(notification *blockdag.Notification,
	acceptanceIndex *indexers.AcceptanceIndex, rpcManager *rpc.Manager) error {

	if notification.Type == blockdag.NTChainChanged && acceptanceIndex != nil {
		chainChangedNotificationData := notification.Data.(*blockdag.ChainChangedNotificationData)
		err := rpcManager.NotifyChainChanged(chainChangedNotificationData.RemovedChainBlockHashes,
			chainChangedNotificationData.AddedChainBlockHashes)
		if err != nil {
			return err
		}
	}
	return nil
}

func (a *App) maybeSeedFromDNS() {
	if !a.cfg.DisableDNSSeed {
		dnsseed.SeedFromDNS(a.cfg.NetParams(), a.cfg.DNSSeed, appmessage.SFNodeNetwork, false, nil,
			a.cfg.Lookup, func(addresses []*appmessage.NetAddress) {
				// Kaspad uses a lookup of the dns seeder here. Since seeder returns
				// IPs of nodes and not its own IP, we can not know real IP of
				// source. So we'll take first returned address as source.
				a.addressManager.AddAddresses(addresses...)
			})
	}

	if a.cfg.GRPCSeed != "" {
		dnsseed.SeedFromGRPC(a.cfg.NetParams(), a.cfg.GRPCSeed, appmessage.SFNodeNetwork, false, nil,
			func(addresses []*appmessage.NetAddress) {
				a.addressManager.AddAddresses(addresses, addresses[0], nil)
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
		acceptanceIndex = indexers.NewAcceptanceIndex()
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
			return dag.CalcSequenceLockNoLock(tx, utxoSet)
		},
		SigCache: sigCache,
		DAG:      dag,
	}

	return mempool.New(&mempoolConfig)
}

// P2PNodeID returns the network ID associated with this App
func (a *App) P2PNodeID() *id.ID {
	return a.netAdapter.ID()
}

// AddressManager returns the AddressManager associated with this App
func (a *App) AddressManager() *addressmanager.AddressManager {
	return a.addressManager
}
