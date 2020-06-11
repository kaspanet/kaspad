package p2p

import (
	"fmt"
	"github.com/kaspanet/kaspad/peer"
	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/wire"
	"sync/atomic"
)

// OnFeeFilter is invoked when a peer receives a feefilter kaspa message and
// is used by remote peers to request that no transactions which have a fee rate
// lower than provided value are inventoried to them. The peer will be
// disconnected if an invalid fee filter value is provided.
func (sp *Peer) OnFeeFilter(_ *peer.Peer, msg *wire.MsgFeeFilter) {
	// Check that the passed minimum fee is a valid amount.
	if msg.MinFee < 0 || msg.MinFee > util.MaxSompi {
		sp.AddBanScoreAndPushRejectMsg(wire.CmdFeeFilter, wire.RejectInvalid, nil,
			100, 0, fmt.Sprintf("sent an invalid feefilter '%s'", util.Amount(msg.MinFee)))
		return
	}

	atomic.StoreInt64(&sp.FeeFilterInt, msg.MinFee)
}
