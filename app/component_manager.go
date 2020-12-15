package app

import (
	"fmt"
	"github.com/kaspanet/kaspad/domain/utxoindex"
	"sync/atomic"

	infrastructuredatabase "github.com/kaspanet/kaspad/infrastructure/db/database"

	"github.com/kaspanet/kaspad/domain"

	"github.com/kaspanet/kaspad/infrastructure/network/addressmanager"

	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/id"

	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/protocol"
	"github.com/kaspanet/kaspad/app/rpc"
	"github.com/kaspanet/kaspad/infrastructure/config"
	"github.com/kaspanet/kaspad/infrastructure/network/connmanager"
	"github.com/kaspanet/kaspad/infrastructure/network/dnsseed"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter"
	"github.com/kaspanet/kaspad/util/panics"
)

// ComponentManager is a wrapper for all the kaspad services
type ComponentManager struct {
	cfg               *config.Config
	addressManager    *addressmanager.AddressManager
	protocolManager   *protocol.Manager
	rpcManager        *rpc.Manager
	connectionManager *connmanager.ConnectionManager
	netAdapter        *netadapter.NetAdapter

	started, shutdown int32
}

// Start launches all the kaspad services.
func (a *ComponentManager) Start() {
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
func (a *ComponentManager) Stop() {
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

// NewComponentManager returns a new ComponentManager instance.
// Use Start() to begin all services within this ComponentManager
func NewComponentManager(cfg *config.Config, db infrastructuredatabase.Database, interrupt chan<- struct{}) (
	*ComponentManager, error) {

	domain, err := domain.New(cfg.ActiveNetParams, db)
	if err != nil {
		return nil, err
	}

	netAdapter, err := netadapter.NewNetAdapter(cfg)
	if err != nil {
		return nil, err
	}

	addressManager, err := addressmanager.New(addressmanager.NewConfig(cfg))
	if err != nil {
		return nil, err
	}

	var utxoIndex *utxoindex.UTXOIndex
	if cfg.UTXOIndex {
		utxoIndex = utxoindex.New(domain.Consensus(), db)
	}

	connectionManager, err := connmanager.New(cfg, netAdapter, addressManager)
	if err != nil {
		return nil, err
	}
	protocolManager, err := protocol.NewManager(cfg, domain, netAdapter, addressManager, connectionManager)
	if err != nil {
		return nil, err
	}
	rpcManager := setupRPC(cfg, domain, netAdapter, protocolManager, connectionManager, addressManager, utxoIndex, interrupt)

	return &ComponentManager{
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
	domain domain.Domain,
	netAdapter *netadapter.NetAdapter,
	protocolManager *protocol.Manager,
	connectionManager *connmanager.ConnectionManager,
	addressManager *addressmanager.AddressManager,
	utxoIndex *utxoindex.UTXOIndex,
	shutDownChan chan<- struct{},
) *rpc.Manager {

	rpcManager := rpc.NewManager(
		cfg,
		domain,
		netAdapter,
		protocolManager,
		connectionManager,
		addressManager,
		utxoIndex,
		shutDownChan,
	)
	protocolManager.SetOnBlockAddedToDAGHandler(rpcManager.NotifyBlockAddedToDAG)

	return rpcManager
}

func (a *ComponentManager) maybeSeedFromDNS() {
	if !a.cfg.DisableDNSSeed {
		dnsseed.SeedFromDNS(a.cfg.NetParams(), a.cfg.DNSSeed, appmessage.SFNodeNetwork, false, nil,
			a.cfg.Lookup, func(addresses []*appmessage.NetAddress) {
				// Kaspad uses a lookup of the dns seeder here. Since seeder returns
				// IPs of nodes and not its own IP, we can not know real IP of
				// source. So we'll take first returned address as source.
				a.addressManager.AddAddresses(addresses...)
			})

		dnsseed.SeedFromGRPC(a.cfg.NetParams(), a.cfg.GRPCSeed, appmessage.SFNodeNetwork, false, nil,
			func(addresses []*appmessage.NetAddress) {
				a.addressManager.AddAddresses(addresses...)
			})
	}
}

// P2PNodeID returns the network ID associated with this ComponentManager
func (a *ComponentManager) P2PNodeID() *id.ID {
	return a.netAdapter.ID()
}

// AddressManager returns the AddressManager associated with this ComponentManager
func (a *ComponentManager) AddressManager() *addressmanager.AddressManager {
	return a.addressManager
}
