package rpchandlers

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/rpc/rpccontext"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
	"github.com/kaspanet/kaspad/version"
)

// HandleGetInfo handles the respectively named RPC command
func HandleGetInfo(context *rpccontext.Context, _ *router.Router, _ appmessage.Message) (appmessage.Message, error) {
	isNearlySynced, err := context.Domain.Consensus().IsNearlySynced()
	if err != nil {
		return nil, err
	}

	response := appmessage.NewGetInfoResponseMessage(
		context.NetAdapter.ID().String(),
		uint64(context.Domain.MiningManager().TransactionCount()),
		version.Version(),
		context.Config.UTXOIndex,
		context.ProtocolManager.Context().HasPeers() && isNearlySynced,
		int64(context.Config.RPCMaxClients),
		int64(context.ConnectionManager.ConnectionCount()),
		int64(context.Config.MaxInboundPeers),
		int64(context.NetAdapter.P2PConnectionCount()),
		int64(context.Config.BanDuration.Seconds()),
		
	)

	return response, nil
}
