package p2p

import (
	"github.com/kaspanet/kaspad/wire"
)

func (sp *Peer) OnGetSelectedTip() {
	sp.QueueMessage(wire.NewMsgSelectedTip(sp.selectedTip()), nil)
}
