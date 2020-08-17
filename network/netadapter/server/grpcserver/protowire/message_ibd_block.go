package protowire

import "github.com/kaspanet/kaspad/network/appmessage"

func (x *KaspadMessage_IbdBlock) toDomainMessage() (appmessage.Message, error) {
	msgBlock, err := x.IbdBlock.toDomainMessage()
	if err != nil {
		return nil, err
	}
	return &appmessage.MsgIBDBlock{MsgBlock: msgBlock.(*appmessage.MsgBlock)}, nil
}

func (x *KaspadMessage_IbdBlock) fromDomainMessage(msgIBDBlock *appmessage.MsgIBDBlock) error {
	x.IbdBlock = new(BlockMessage)
	return x.IbdBlock.fromDomainMessage(msgIBDBlock.MsgBlock)
}
