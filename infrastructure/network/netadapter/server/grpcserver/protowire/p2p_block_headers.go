package protowire

import "github.com/kaspanet/kaspad/app/appmessage"

func (x *KaspadMessage_BlockHeaders) toAppMessage() (appmessage.Message, error) {
	blockHeaders := make([]*appmessage.MsgBlockHeader, len(x.BlockHeaders.BlockHeaders))
	for i, blockHeader := range x.BlockHeaders.BlockHeaders {
		var err error
		blockHeaders[i], err = blockHeader.toAppMessage()
		if err != nil {
			return nil, err
		}
	}

	return &appmessage.BlockHeadersMessage{
		BlockHeaders: blockHeaders,
	}, nil
}

func (x *KaspadMessage_BlockHeaders) fromAppMessage(blockHeadersMessage *appmessage.BlockHeadersMessage) error {
	blockHeaders := make([]*BlockHeaderMessage, len(blockHeadersMessage.BlockHeaders))
	for i, blockHeader := range blockHeadersMessage.BlockHeaders {
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
