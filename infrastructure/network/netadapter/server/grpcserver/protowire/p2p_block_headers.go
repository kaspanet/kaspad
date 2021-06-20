package protowire

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/pkg/errors"
)

func (x *KaspadMessage_BlockHeaders) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_BlockHeaders is nil")
	}
	blockHeaders, err := x.BlockHeaders.toAppMessage()
	if err != nil {
		return nil, err
	}
	return &appmessage.IBDBlocksMessage{
		Blocks: blockHeaders,
	}, nil
}

func (x *BlockHeadersMessage) toAppMessage() ([]*appmessage.MsgBlockHeader, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "IBDBlocksMessage is nil")
	}
	blockHeaders := make([]*appmessage.MsgBlockHeader, len(x.BlockHeaders))
	for i, blockHeader := range x.BlockHeaders {
		var err error
		blockHeaders[i], err = blockHeader.toAppMessage()
		if err != nil {
			return nil, err
		}
	}

	return blockHeaders, nil
}

func (x *KaspadMessage_BlockHeaders) fromAppMessage(blockHeadersMessage *appmessage.IBDBlocksMessage) error {
	blockHeaders := make([]*BlockHeaderMessage, len(blockHeadersMessage.Blocks))
	for i, blockHeader := range blockHeadersMessage.Blocks {
		blockHeaders[i] = &BlockHeaderMessage{}
		err := blockHeaders[i].fromAppMessage(blockHeader)
		if err != nil {
			return err
		}
	}

	x.BlockHeaders = &BlockHeadersMessage{
		BlockHeaders: blockHeaders,
	}
	return nil
}
