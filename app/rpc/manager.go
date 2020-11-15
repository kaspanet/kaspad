package rpc

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/protocol"
	"github.com/kaspanet/kaspad/app/rpc/rpccontext"
	"github.com/kaspanet/kaspad/domain"
	"github.com/kaspanet/kaspad/domain/addressindex"
	"github.com/kaspanet/kaspad/infrastructure/config"
	"github.com/kaspanet/kaspad/infrastructure/network/addressmanager"
	"github.com/kaspanet/kaspad/infrastructure/network/connmanager"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
)

// Handler is the interface for the RPC handlers
type Handler interface {
	Execute(router *router.Router, request appmessage.Message) (appmessage.Message, error)
}

// Notifier is the interface for the RPC notifier
type Notifier interface {
	AddListener(router *router.Router)
	RemoveListener(router *router.Router)
}

type handlerInContext struct {
	context *rpccontext.Context
	handler handlerFunc
}

func (hic *handlerInContext) Execute(router *router.Router, request appmessage.Message) (appmessage.Message, error) {
	return hic.handler(hic.context, router, request)
}

// Manager is an RPC manager
type Manager struct {
	context  *rpccontext.Context
	handlers map[appmessage.MessageCommand]Handler
	notifier Notifier
}

// NewManager creates a new RPC Manager
func NewManager(
	cfg *config.Config,
	domain domain.Domain,
	netAdapter *netadapter.NetAdapter,
	protocolManager *protocol.Manager,
	connectionManager *connmanager.ConnectionManager,
	addressManager *addressmanager.AddressManager,
	utxoAddressIndex *addressindex.Index,
	shutDownChan chan<- struct{}) *Manager {

	manager := Manager{
		context: rpccontext.NewContext(
			cfg,
			domain,
			netAdapter,
			protocolManager,
			connectionManager,
			addressManager,
			utxoAddressIndex,
			shutDownChan,
		),
		handlers: make(map[appmessage.MessageCommand]Handler),
	}

	netAdapter.SetRPCRouterInitializer(manager.routerInitializer)

	for command, handler := range rpcHandlers {
		handlerInContext := handlerInContext{
			context: manager.context,
			handler: handler,
		}
		manager.RegisterHandler(command, &handlerInContext)
	}

	return &manager
}
