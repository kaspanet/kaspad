package blockrelay

import (
	"github.com/kaspanet/kaspad/blockdag"
	"github.com/kaspanet/kaspad/netadapter/router"
	peerpkg "github.com/kaspanet/kaspad/protocol/peer"
	"github.com/kaspanet/kaspad/protocol/protocolerrors"
	"github.com/kaspanet/kaspad/wire"
	"github.com/pkg/errors"
)

// HandleRelayBlockRequests listens to wire.MsgGetRelayBlocks messages and sends
// their corresponding blocks to the requesting peer.
func HandleRelayBlockRequests(incomingRoute *router.Route, outgoingRoute *router.Route,
	peer *peerpkg.Peer, dag *blockdag.BlockDAG) error {

	for {
		message, isOpen := incomingRoute.Dequeue()
		if !isOpen {
			return nil
		}
		getRelayBlocksMessage := message.(*wire.MsgGetRelayBlocks)
		for _, hash := range getRelayBlocksMessage.Hashes {
			// Fetch the block from the database.
			block, err := dag.BlockByHash(hash)
			if blockdag.IsNotInDAGErr(err) {
				return protocolerrors.Errorf(true, "block %s not found", hash)
			} else if err != nil {
				panic(errors.Wrapf(err, "unable to fetch requested block hash %s", hash))
			}
			msgBlock := block.MsgBlock()

			// If we are a full node and the peer is a partial node, we must convert
			// the block to a partial block.
			nodeSubnetworkID := dag.SubnetworkID()
			peerSubnetworkID := peer.SubnetworkID()

			isNodeFull := nodeSubnetworkID == nil
			isPeerFull := peerSubnetworkID == nil
			if isNodeFull && !isPeerFull {
				msgBlock.ConvertToPartial(peerSubnetworkID)
			}

			isOpen = outgoingRoute.Enqueue(msgBlock)
			if !isOpen {
				return nil
			}
		}
	}
}
