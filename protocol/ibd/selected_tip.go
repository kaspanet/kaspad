package ibd

import (
	"github.com/kaspanet/kaspad/blockdag"
	"github.com/kaspanet/kaspad/netadapter/router"
	peerpkg "github.com/kaspanet/kaspad/protocol/peer"
)

func requestSelectedTipsIfRequired(dag *blockdag.BlockDAG) error {
	if recentlyReceivedBlock(dag) {
		return nil
	}
	return requestSelectedTips(dag)
}

func recentlyReceivedBlock(dag *blockdag.BlockDAG) bool {
	return false
}

func requestSelectedTips(dag *blockdag.BlockDAG) error {
	return nil
}

func RequestSelectedTip(incomingRoute *router.Route, outgoingRoute *router.Route,
	peer *peerpkg.Peer) error {
	return nil
}

func HandleGetSelectedTip(incomingRoute *router.Route, outgoingRoute *router.Route,
	peer *peerpkg.Peer) error {
	return nil
}
