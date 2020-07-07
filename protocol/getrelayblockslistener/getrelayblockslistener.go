package getrelayblockslistener

import (
	"fmt"
	"github.com/kaspanet/kaspad/blockdag"
	"github.com/kaspanet/kaspad/p2pserver"
	"github.com/kaspanet/kaspad/peer"
	"github.com/kaspanet/kaspad/wire"
)

// StartGetRelayBlocksListener listens to wire.MsgGetRelayBlocks messages and sends
// their corresponding blocks to the requesting peer.
func StartGetRelayBlocksListener(msgChan <-chan wire.Message, connection p2pserver.Connection,
	dag *blockdag.BlockDAG) error {

	for msg := range msgChan {
		getRelayBlocksMsg := msg.(*wire.MsgGetRelayBlocks)

		length := len(getRelayBlocksMsg.Hashes)
		// A decaying ban score increase is applied to prevent exhausting resources
		// with unusually large inventory queries.
		// Requesting more than the maximum inventory vector length within a short
		// period of time yields a score above the default ban threshold. Sustained
		// bursts of small requests are not penalized as that would potentially ban
		// peers performing IBD.
		// This incremental score decays each minute to half of its value.
		isBanned := connection.AddBanScore(0, uint32(length)*99/wire.MsgGetRelayBlocksHashes, "getrelblks")
		if isBanned {
			return nil
		}

		for _, hash := range getRelayBlocksMsg.Hashes {
			// Fetch the block from the database.
			block, err := dag.BlockByHash(hash)
			if blockdag.IsNotInDAGErr(err) {
				isBanned := connection.AddBanScore(peer.BanScoreRequestNonExistingBlock, 0, fmt.Sprintf("block %s not found", hash))
				if isBanned {
					return nil
				}
				continue
			} else if err != nil {
				log.Tracef("Unable to fetch requested block hash %s: %s",
					hash, err)
				return err
			}
			msgBlock := block.MsgBlock()

			// TODO(libp2p)
			//// If we are a full node and the peer is a partial node, we must convert
			//// the block to a partial block.
			//nodeSubnetworkID := s.DAG.SubnetworkID()
			//peerSubnetworkID := sp.Peer.SubnetworkID()
			//isNodeFull := nodeSubnetworkID == nil
			//isPeerFull := peerSubnetworkID == nil
			//if isNodeFull && !isPeerFull {
			//	msgBlock.ConvertToPartial(peerSubnetworkID)
			//}

			err = connection.Send(msgBlock)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
