package rpccontext

import (
	"github.com/kaspanet/kaspad/domain/blockdag"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter"
)

type Context struct {
	NetAdapter *netadapter.NetAdapter
	DAG        *blockdag.BlockDAG
}

func NewContext(netAdapter *netadapter.NetAdapter, dag *blockdag.BlockDAG) *Context {
	return &Context{
		NetAdapter: netAdapter,
		DAG:        dag,
	}
}
