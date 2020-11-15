package wallethandlers

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/wallet/walletcontext"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
)

// HandlerFunc is an alias for the handler function type
type HandlerFunc func(context walletcontext.Context, router *router.Router, request appmessage.Message) (appmessage.Message, error)

// HandlerInContext is a structure that contains a handler function and an appropriate context for it
type HandlerInContext struct {
	Context walletcontext.Context
	Handler HandlerFunc
}

// Execute executes the handler for the provided router and request
func (hic *HandlerInContext) Execute(router *router.Router, request appmessage.Message) (appmessage.Message, error) {
	return hic.Handler(hic.Context, router, request)
}
