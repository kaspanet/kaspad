package protocolcommon

import (
	"github.com/kaspanet/kaspad/p2pserver"
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/kaspanet/kaspad/wire"
)

// AddBanScoreAndPushRejectMsg increases ban score and sends a
// reject message to the misbehaving peer.
func AddBanScoreAndPushRejectMsg(connection p2pserver.Connection, command string, code wire.RejectCode, hash *daghash.Hash, persistent, transient uint32, reason string) (isBanned bool) {
	PushRejectMsg(connection, command, code, reason, hash)
	return connection.AddBanScore(persistent, transient, reason)
}

func PushRejectMsg(connection p2pserver.Connection, command string, code wire.RejectCode, reason string, hash *daghash.Hash) {
	msg := wire.NewMsgReject(command, code, reason)
	msg.Hash = hash
	err := connection.Send(msg)
	if err != nil {
		log.Errorf("couldn't send reject message to %s", connection)
	}
}
