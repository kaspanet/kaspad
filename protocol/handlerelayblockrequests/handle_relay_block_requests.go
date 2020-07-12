package handlerelayblockrequests

import (
	"github.com/kaspanet/kaspad/blockdag"
	"github.com/kaspanet/kaspad/netadapter"
	peerpkg "github.com/kaspanet/kaspad/protocol/peer"
	"github.com/kaspanet/kaspad/wire"
	"github.com/pkg/errors"
)

// HandleRelayBlockRequests listens to wire.MsgGetRelayBlocks messages and sends
// their corresponding blocks to the requesting peer.
func HandleRelayBlockRequests(msgChan <-chan wire.Message, peer *peerpkg.Peer, router *netadapter.Router,
	dag *blockdag.BlockDAG) error {

	for msg := range msgChan {
		getRelayBlocksMsg := msg.(*wire.MsgGetRelayBlocks)
		for _, hash := range getRelayBlocksMsg.Hashes {
			// Fetch the block from the database.
			block, err := dag.BlockByHash(hash)
			if blockdag.IsNotInDAGErr(err) {
				return errors.Errorf("block %s not found", hash)
			} else if err != nil {
				panic(errors.Wrapf(err, "unable to fetch requested block hash %s", hash))
			}
			msgBlock := block.MsgBlock()

			// If we are a full node and the peer is a partial node, we must convert
			// the block to a partial block.
			nodeSubnetworkID := dag.SubnetworkID()
			peerSubnetworkID, err := peer.SubnetworkID()
			if err != nil {
				panic(err)
			}

			isNodeFull := nodeSubnetworkID == nil
			isPeerFull := peerSubnetworkID == nil
			if isNodeFull && !isPeerFull {
				msgBlock.ConvertToPartial(peerSubnetworkID)
			}

			router.WriteOutgoingMessage(msgBlock)
		}
	}
	return nil
}
