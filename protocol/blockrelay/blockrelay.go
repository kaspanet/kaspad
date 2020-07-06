package blockrelay

import (
	"fmt"
	"github.com/kaspanet/kaspad/blockdag"
	"github.com/kaspanet/kaspad/p2pserver"
	"github.com/kaspanet/kaspad/peer"
	"github.com/kaspanet/kaspad/protocol"
	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/kaspanet/kaspad/wire"
	"github.com/pkg/errors"
	"sync"
	"time"
)

var requestedBlocks = &sharedRequestedBlocks{
	blocks: make(map[daghash.Hash]struct{}),
}

type sharedRequestedBlocks struct {
	blocks map[daghash.Hash]struct{}
	sync.Mutex
}

func (s *sharedRequestedBlocks) delete(hash *daghash.Hash) {
	s.Lock()
	defer s.Unlock()
	delete(s.blocks, *hash)
}

func (s *sharedRequestedBlocks) addIfExists(hash *daghash.Hash) (exists bool) {
	s.Lock()
	defer s.Unlock()
	_, ok := s.blocks[*hash]
	if ok {
		return true
	}
	s.blocks[*hash] = struct{}{}
	return false
}

func StartBlockRelay(msgChan <-chan wire.Message, server p2pserver.Server, connection p2pserver.Connection,
	dag *blockdag.BlockDAG) error {

	invsQueue := make([]*wire.MsgInvRelayBlock, 0)
	for {
		shouldStop, err := handleInv(msgChan, server, connection, dag, invsQueue)
		if err != nil {
			return err
		}
		if shouldStop {
			return nil
		}
	}
}

func handleInv(msgChan <-chan wire.Message, server p2pserver.Server, connection p2pserver.Connection,
	dag *blockdag.BlockDAG, invsQueue []*wire.MsgInvRelayBlock) (shouldStop bool, err error) {

	inv, shouldStop := readInv(connection, msgChan, &invsQueue)
	if shouldStop {
		return true, nil
	}

	if dag.IsKnownBlock(inv.Hash) {
		if dag.IsKnownInvalid(inv.Hash) {
			protocol.AddBanScoreAndPushRejectMsg(connection, inv.Command(), wire.RejectInvalid, inv.Hash,
				peer.BanScoreInvalidInvBlock, 0, fmt.Sprintf("sent inv of invalid block %s",
					inv.Hash))
		}
		return false, nil
	}

	requestQueue := []*daghash.Hash{inv.Hash}
	requestQueueSet := map[daghash.Hash]struct{}{
		*inv.Hash: {},
	}
	pendingBlocks := map[daghash.Hash]struct{}{}

	// In case the function closes earlier than expected, we wanna make sure requestedBlocks is
	// clean from any pending blocks.
	defer deleteFromRequestedBlocks(pendingBlocks)
	for len(requestQueue) > 0 {
		shouldStop, err := requestBlocks(connection, server, msgChan, dag, &invsQueue, &requestQueue, requestQueueSet)
		if err != nil {
			return false, err
		}
		if shouldStop {
			return true, nil
		}
	}
	return false, nil
}

func readInv(connection p2pserver.Connection, msgChan <-chan wire.Message,
	invsQueue *[]*wire.MsgInvRelayBlock) (inv *wire.MsgInvRelayBlock, shouldStop bool) {

	if len(*invsQueue) > 0 {
		inv, *invsQueue = (*invsQueue)[0], (*invsQueue)[1:]
		return inv, false
	}

	for {
		msg, isClosed := <-msgChan
		if isClosed {
			return nil, true
		}

		inv, ok := msg.(*wire.MsgInvRelayBlock)
		if ok {
			return inv, false
		}

		isBanned := protocol.AddBanScoreAndPushRejectMsg(connection,
			msg.Command(),
			wire.RejectNotRequested,
			nil,
			peer.BanScoreUnrequestedMessage,
			0,
			fmt.Sprintf("unrequested %s message in the block relay flow", msg.Command()))

		if isBanned {
			return nil, true
		}
	}
}

func readNonInvMsg(connection p2pserver.Connection, msgChan <-chan wire.Message,
	invsQueue *[]*wire.MsgInvRelayBlock) (msg wire.Message, shouldStop bool, err error) {

	for {
		const stallResponseTimeout = 30 * time.Second
		select {
		case <-time.After(stallResponseTimeout):
			reason := fmt.Sprintf("stalled for %s", stallResponseTimeout)
			connection.AddBanScore(peer.BanScoreStallTimeout, 0, reason)
			return nil, false, errors.New(reason)
		case msg, isClosed := <-msgChan:
			if isClosed {
				return nil, true, nil
			}

			inv, ok := msg.(*wire.MsgInvRelayBlock)
			if !ok {
				return msg, false, nil
			}

			*invsQueue = append(*invsQueue, inv)
		}
	}
}

func deleteFromRequestedBlocks(blockHashes map[daghash.Hash]struct{}) {
	for hash := range blockHashes {
		hash := hash
		requestedBlocks.delete(&hash)
	}
}

func requestBlocks(connection p2pserver.Connection, server p2pserver.Server, msgChan <-chan wire.Message,
	dag *blockdag.BlockDAG, invsQueue *[]*wire.MsgInvRelayBlock, requestQueue *[]*daghash.Hash,
	requestQueueSet map[daghash.Hash]struct{}) (shouldStop bool, err error) {

	var hashesToRequest []*daghash.Hash
	if len(*requestQueue) > wire.MsgGetRelayBlocksHashes {
		hashesToRequest, *requestQueue = (*requestQueue)[:wire.MsgGetRelayBlocksHashes],
			(*requestQueue)[wire.MsgGetRelayBlocksHashes:]
	} else {
		hashesToRequest, *requestQueue = *requestQueue, nil
	}

	pendingBlocks := map[daghash.Hash]struct{}{}
	for _, hash := range hashesToRequest {
		delete(requestQueueSet, *hash)
		pendingBlocks[*hash] = struct{}{}
		exists := requestedBlocks.addIfExists(hash)
		if exists {
			return false, nil
		}
	}

	getRelayBlockMsg := wire.NewMsgGetRelayBlocks(hashesToRequest)
	err = connection.Send(getRelayBlockMsg)
	if err != nil {
		return false, err
	}

	for len(pendingBlocks) > 0 {
		msg, shouldStop, err := readNonInvMsg(connection, msgChan, invsQueue)
		if err != nil {
			return false, err
		}
		if shouldStop {
			return true, nil
		}

		msgBlock, ok := msg.(*wire.MsgBlock)
		if !ok {
			isBanned := protocol.AddBanScoreAndPushRejectMsg(connection,
				msg.Command(),
				wire.RejectNotRequested,
				nil,
				peer.BanScoreUnrequestedMessage,
				0,
				fmt.Sprintf("unrequested %s message in the block relay flow", msg.Command()))
			if isBanned {
				return true, nil
			}
			continue
		}

		block := util.NewBlock(msgBlock)
		blockHash := block.Hash()
		if _, ok := pendingBlocks[*blockHash]; !ok {
			isBanned := protocol.AddBanScoreAndPushRejectMsg(connection,
				msg.Command(),
				wire.RejectNotRequested,
				nil,
				peer.BanScoreUnrequestedBlock,
				0,
				fmt.Sprintf("got unrequested block %s", block.Hash()))
			if isBanned {
				return true, nil
			}
		}
		delete(pendingBlocks, *blockHash)
		requestedBlocks.delete(blockHash)

		shouldStop, processedBlockHashes, err := processBlock(connection, dag, requestQueue, requestQueueSet, block)
		if err != nil {
			return false, err
		}
		if shouldStop {
			return true, nil
		}
		if len(processedBlockHashes) > 0 {
			// TODO(libp2p) need to implement relay inv with multiple hashes so we can broadcast everything together.
		}
	}
	return false, nil
}

func processBlock(connection p2pserver.Connection, dag *blockdag.BlockDAG, requestQueue *[]*daghash.Hash,
	requestQueueSet map[daghash.Hash]struct{},
	block *util.Block) (shouldStop bool, processedBlocks []*daghash.Hash, err error) {

	blockHash := block.Hash()
	isOrphan, isDelayed, err := dag.ProcessBlock(block, blockdag.BFNone)
	if err != nil {
		// When the error is a rule error, it means the block was simply
		// rejected as opposed to something actually going wrong, so log
		// it as such. Otherwise, something really did go wrong, so log
		// it as an actual error.
		if !errors.As(err, &blockdag.RuleError{}) {
			panic(errors.Wrapf(err, "failed to process block %s",
				blockHash))
		}
		log.Infof("Rejected block %s from %s: %s", blockHash,
			connection, err)

		isBanned := protocol.AddBanScoreAndPushRejectMsg(connection, wire.CmdBlock, wire.RejectInvalid, blockHash,
			peer.BanScoreInvalidBlock, 0, fmt.Sprintf("got invalid block: %s", err))
		// Whether the peer will be banned or not, syncing from a node that doesn't follow
		// the netsync protocol is undesired.
		// TODO(libp2p): sm.RemoveFromSyncCandidates(peer)
		if isBanned {
			return true, nil, nil
		}
		return false, nil, nil
	}

	if isOrphan {
		blueScore, err := block.BlueScore()
		if err != nil {
			log.Errorf("Received an orphan block %s with malformed blue score from %s. Disconnecting...",
				blockHash, connection)
			isBanned := protocol.AddBanScoreAndPushRejectMsg(connection, wire.CmdBlock,
				wire.RejectInvalid, blockHash, peer.BanScoreMalformedBlueScoreInOrphan, 0,
				fmt.Sprintf("Received an orphan block %s with malformed blue score", blockHash))
			if isBanned {
				return true, nil, nil
			}
			return false, nil, nil
		}

		const maxOrphanBlueScoreDiff = 10000
		selectedTipBlueScore := dag.SelectedTipBlueScore()
		if blueScore > selectedTipBlueScore+maxOrphanBlueScoreDiff {
			log.Infof("Orphan block %s has blue score %d and the selected tip blue score is "+
				"%d. Ignoring orphans with a blue score difference from the selected tip greater than %d",
				blockHash, blueScore, selectedTipBlueScore, maxOrphanBlueScoreDiff)
			return false, nil, nil
		}

		// Request the parents for the orphan block from the peer that sent it.
		missingAncestors := dag.GetOrphanMissingAncestorHashes(blockHash)
		for _, missingAncestor := range missingAncestors {
			if _, ok := requestQueueSet[*missingAncestor]; !ok {
				*requestQueue = append(*requestQueue, missingAncestor)
				requestQueueSet[*missingAncestor] = struct{}{}
			}
		}
		return false, []*daghash.Hash{blockHash}, nil
	}
	//TODO(libp2p)
	//// When the block is not an orphan, log information about it and
	//// update the DAG state.
	// sm.restartSyncIfNeeded()
	//blockBlueScore, err := dag.BlueScoreByBlockHash(blockHash)
	//if err != nil {
	//	log.Errorf("Failed to get blue score for block %s: %s", blockHash, err)
	//}
	//sm.progressLogger.LogBlockBlueScore(bmsg.block, blockBlueScore)
	//
	//// Clear the rejected transactions.
	//sm.rejectedTxns = make(map[daghash.TxID]struct{})
	return false, nil, nil
}
