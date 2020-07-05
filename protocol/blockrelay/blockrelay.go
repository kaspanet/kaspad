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
	"sync"
)

type SharedRequestedBlocks struct {
	blocks map[daghash.Hash]struct{}
	sync.Mutex
}

func (s *SharedRequestedBlocks) delete(hash *daghash.Hash) {
	s.Lock()
	defer s.Unlock()
	delete(s.blocks, *hash)
}

func (s *SharedRequestedBlocks) addIfExists(hash *daghash.Hash) (exists bool) {
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
	dag *blockdag.BlockDAG, requestedBlocks *SharedRequestedBlocks) error {
	invsQueue := make([]*wire.MsgInvRelayBlock, 0)
	for {
		shouldStop, err := func() (shouldStop bool, err error) {
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

			exists := requestedBlocks.addIfExists(inv.Hash)
			if exists {
				return false, nil
			}
			defer requestedBlocks.delete(inv.Hash)

			getRelayBlockMsg := wire.NewMsgGetRelayBlock(inv.Hash)
			err = connection.Send(getRelayBlockMsg)
			if err != nil {
				return false, err
			}

			msg, shouldStop := readNonInvMsg(msgChan, &invsQueue)
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
			}

			block := util.NewBlock(msgBlock)
			if !block.Hash().IsEqual(inv.Hash) {
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
			requestedBlocks.delete(inv.Hash)
			return false, nil
		}()
		if err != nil {
			return err
		}
		if shouldStop {
			return nil
		}
	}
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

func readNonInvMsg(msgChan <-chan wire.Message,
	invsQueue *[]*wire.MsgInvRelayBlock) (msg wire.Message, shouldStop bool) {

	for {
		msg, isClosed := <-msgChan
		if isClosed {
			return nil, true
		}

		inv, ok := msg.(*wire.MsgInvRelayBlock)
		if !ok {
			return msg, false
		}

		*invsQueue = append(*invsQueue, inv)
	}
}
