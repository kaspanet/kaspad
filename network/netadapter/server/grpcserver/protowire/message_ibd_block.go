package protowire

import "github.com/kaspanet/kaspad/network/domainmessage"

func (x *KaspadMessage_IbdBlock) toDomainMessage() (domainmessage.Message, error) {
	msgBlock, err := x.IbdBlock.toDomainMessage()
	if err != nil {
		return nil, err
	}
	return &domainmessage.MsgIBDBlock{MsgBlock: msgBlock.(*domainmessage.MsgBlock)}, nil
}

func (x *KaspadMessage_IbdBlock) fromDomainMessage(msgIBDBlock *domainmessage.MsgIBDBlock) error {
	x.IbdBlock = new(BlockMessage)
	return x.IbdBlock.fromDomainMessage(msgIBDBlock.MsgBlock)
}
