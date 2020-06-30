package rpc

import (
	"fmt"
	"github.com/kaspanet/kaspad/rpcmodel"
	"github.com/kaspanet/kaspad/util/mstime"
	"time"
)

// handleGetConnectedPeerInfo implements the getConnectedPeerInfo command.
func handleGetConnectedPeerInfo(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	peers := s.cfg.ConnMgr.ConnectedPeers()
	syncPeerID := s.cfg.SyncMgr.SyncPeerID()
	infos := make([]*rpcmodel.GetConnectedPeerInfoResult, 0, len(peers))
	for _, p := range peers {
		statsSnap := p.ToPeer().StatsSnapshot()
		info := &rpcmodel.GetConnectedPeerInfoResult{
			ID:          statsSnap.ID,
			Addr:        statsSnap.Addr,
			Services:    fmt.Sprintf("%08d", uint64(statsSnap.Services)),
			RelayTxes:   !p.IsTxRelayDisabled(),
			LastSend:    mstime.TimeToUnixMilli(statsSnap.LastSend),
			LastRecv:    mstime.TimeToUnixMilli(statsSnap.LastRecv),
			BytesSent:   statsSnap.BytesSent,
			BytesRecv:   statsSnap.BytesRecv,
			ConnTime:    mstime.TimeToUnixMilli(statsSnap.ConnTime),
			PingTime:    float64(statsSnap.LastPingMicros),
			TimeOffset:  statsSnap.TimeOffset,
			Version:     statsSnap.Version,
			SubVer:      statsSnap.UserAgent,
			Inbound:     statsSnap.Inbound,
			SelectedTip: statsSnap.SelectedTipHash.String(),
			BanScore:    int32(p.BanScore()),
			FeeFilter:   p.FeeFilter(),
			SyncNode:    statsSnap.ID == syncPeerID,
		}
		if p.ToPeer().LastPingNonce() != 0 {
			wait := float64(time.Since(statsSnap.LastPingTime).Nanoseconds())
			// We actually want microseconds.
			info.PingWait = wait / 1000
		}
		infos = append(infos, info)
	}
	return infos, nil
}
