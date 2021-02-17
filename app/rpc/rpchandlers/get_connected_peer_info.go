package rpchandlers

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/rpc/rpccontext"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
)

// HandleGetConnectedPeerInfo handles the respectively named RPC command
func HandleGetConnectedPeerInfo(context *rpccontext.Context, _ *router.Router, _ appmessage.Message) (appmessage.Message, error) {
	peers := context.ProtocolManager.Peers()
	ibdPeer := context.ProtocolManager.IBDPeer()
	infos := make([]*appmessage.GetConnectedPeerInfoMessage, 0, len(peers))
	for _, peer := range peers {
		info := &appmessage.GetConnectedPeerInfoMessage{
			ID:                        peer.ID().String(),
			Address:                   peer.Address(),
			LastPingDuration:          peer.LastPingDuration().Milliseconds(),
			IsOutbound:                peer.IsOutbound(),
			TimeOffset:                peer.TimeOffset().Milliseconds(),
			UserAgent:                 peer.UserAgent(),
			AdvertisedProtocolVersion: peer.AdvertisedProtocolVersion(),
			TimeConnected:             peer.TimeConnected().Milliseconds(),
			IsIBDPeer:                 peer == ibdPeer,
		}
		infos = append(infos, info)
	}
	response := appmessage.NewGetConnectedPeerInfoResponseMessage(infos)
	return response, nil
}
