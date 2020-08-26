package rpc

import (
	"github.com/kaspanet/kaspad/domain/blockdag"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter"
)

type context struct {
	netAdapter *netadapter.NetAdapter
	dag        *blockdag.BlockDAG
}

func newContext(netAdapter *netadapter.NetAdapter, dag *blockdag.BlockDAG) *context {
	return &context{
		netAdapter: netAdapter,
		dag:        dag,
	}
}
