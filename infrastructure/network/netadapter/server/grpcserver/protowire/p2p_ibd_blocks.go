package protowire

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/pkg/errors"
)

func (x *KaspadMessage_IbdBlocks) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_IbdBlocks is nil")
	}
	blocks, err := x.IbdBlocks.toAppMessage()
	if err != nil {
		return nil, err
	}
	return &appmessage.IBDBlocksMessage{
		Blocks: blocks,
	}, nil
}

func (x *IbdBlocksMessage) toAppMessage() ([]*appmessage.MsgBlock, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "IBDBlocksMessage is nil")
	}
	blocks := make([]*appmessage.MsgBlock, len(x.Blocks))
	for i, block := range x.Blocks {
		var err error
		blocks[i], err = block.toAppMessage()
		if err != nil {
			return nil, err
		}
	}

	return blocks, nil
}

func (x *KaspadMessage_IbdBlocks) fromAppMessage(ibdBlocksMessage *appmessage.IBDBlocksMessage) error {
	blocks := make([]*BlockMessage, len(ibdBlocksMessage.Blocks))
	for i, blockHeader := range ibdBlocksMessage.Blocks {
		blocks[i] = &BlockMessage{}
		err := blocks[i].fromAppMessage(blockHeader)
		if err != nil {
			return err
		}
	}

	x.IbdBlocks = &IbdBlocksMessage{
		Blocks: blocks,
	}
	return nil
}
