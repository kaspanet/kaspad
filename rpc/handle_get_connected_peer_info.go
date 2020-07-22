package rpc

import (
	"github.com/kaspanet/kaspad/rpc/model"
)

// handleGetConnectedPeerInfo implements the getConnectedPeerInfo command.
func handleGetConnectedPeerInfo(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	peers := s.protocolManager.Peers()
	infos := make([]*model.GetConnectedPeerInfoResult, 0, len(peers))
	for _, peer := range peers {
		info := &model.GetConnectedPeerInfoResult{
			ID:               peer.ID().String(),
			Address:          peer.Address(),
			LastPingDuration: peer.LastPingDuration().Milliseconds(),
			SelectedTipHash:  peer.SelectedTipHash().String(),
			IsSyncNode:       peer == s.protocolManager.IBDPeer(),

			// TODO(libp2p): populate the following with real values
			IsInbound:       false,
			BanScore:        0,
			TimeOffset:      0,
			UserAgent:       "",
			ProtocolVersion: 0,
			TimeConnected:   0,
		}
		infos = append(infos, info)
	}
	return infos, nil
}
