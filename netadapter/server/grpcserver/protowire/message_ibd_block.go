package protowire

import "github.com/kaspanet/kaspad/domainmessage"

func (x *KaspadMessage_IbdBlock) toWireMessage() (domainmessage.Message, error) {
	msgBlock, err := x.IbdBlock.toWireMessage()
	if err != nil {
		return nil, err
	}
	return &domainmessage.MsgIBDBlock{MsgBlock: msgBlock.(*domainmessage.MsgBlock)}, nil
}

func (x *KaspadMessage_IbdBlock) fromWireMessage(msgIBDBlock *domainmessage.MsgIBDBlock) error {
	x.IbdBlock = new(BlockMessage)
	return x.IbdBlock.fromWireMessage(msgIBDBlock.MsgBlock)
}
