package p2p

import (
	"github.com/kaspanet/kaspad/peer"
	"github.com/kaspanet/kaspad/wire"
)

func (sp *Peer) OnSelectedTip(peer *peer.Peer, msg *wire.MsgSelectedTip) {
	if msg.SelectedTip.IsEqual(peer.SelectedTip()) {
		return
	}
	peer.SetSelectedTip(msg.SelectedTip)
	sp.server.SyncManager.StartSync()
}
