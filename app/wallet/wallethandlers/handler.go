package wallethandlers

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/wallet/walletcontext"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
)

type HandlerFunc func(context walletcontext.Context, router *router.Router, request appmessage.Message) (appmessage.Message, error)

type HandlerInContext struct {
	Context walletcontext.Context
	Handler HandlerFunc
}

func (hic *HandlerInContext) Execute(router *router.Router, request appmessage.Message) (appmessage.Message, error) {
	return hic.Handler(hic.Context, router, request)
}
