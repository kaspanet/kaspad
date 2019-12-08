package rpc

import (
	"fmt"
	"github.com/kaspanet/kaspad/btcjson"
	"time"
)

// handleGetPeerInfo implements the getPeerInfo command.
func handleGetPeerInfo(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	peers := s.cfg.ConnMgr.ConnectedPeers()
	syncPeerID := s.cfg.SyncMgr.SyncPeerID()
	infos := make([]*btcjson.GetPeerInfoResult, 0, len(peers))
	for _, p := range peers {
		statsSnap := p.ToPeer().StatsSnapshot()
		info := &btcjson.GetPeerInfoResult{
			ID:          statsSnap.ID,
			Addr:        statsSnap.Addr,
			Services:    fmt.Sprintf("%08d", uint64(statsSnap.Services)),
			RelayTxes:   !p.IsTxRelayDisabled(),
			LastSend:    statsSnap.LastSend.Unix(),
			LastRecv:    statsSnap.LastRecv.Unix(),
			BytesSent:   statsSnap.BytesSent,
			BytesRecv:   statsSnap.BytesRecv,
			ConnTime:    statsSnap.ConnTime.Unix(),
			PingTime:    float64(statsSnap.LastPingMicros),
			TimeOffset:  statsSnap.TimeOffset,
			Version:     statsSnap.Version,
			SubVer:      statsSnap.UserAgent,
			Inbound:     statsSnap.Inbound,
			SelectedTip: statsSnap.SelectedTip.String(),
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
