package walletcontext

import (
	"github.com/kaspanet/kaspad/app/wallet/walletnotification"
	routerpkg "github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
)

// Context represents the Wallet handler context
type Context interface {
	Listener(router *routerpkg.Router) (*walletnotification.Listener, error)
}
