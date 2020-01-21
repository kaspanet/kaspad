package p2p

import (
	"github.com/kaspanet/kaspad/wire"
)

// OnGetSelectedTip is invoked when a peer receives a getSelectedTip kaspa
// message.
func (sp *Peer) OnGetSelectedTip() {
	sp.QueueMessage(wire.NewMsgSelectedTip(sp.selectedTip()), nil)
}
