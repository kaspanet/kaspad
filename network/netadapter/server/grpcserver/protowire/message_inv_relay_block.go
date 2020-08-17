package protowire

import "github.com/kaspanet/kaspad/network/appmessage"

func (x *KaspadMessage_InvRelayBlock) toDomainMessage() (appmessage.Message, error) {
	hash, err := x.InvRelayBlock.Hash.toWire()
	if err != nil {
		return nil, err
	}

	return &appmessage.MsgInvRelayBlock{Hash: hash}, nil
}

func (x *KaspadMessage_InvRelayBlock) fromDomainMessage(msgInvRelayBlock *appmessage.MsgInvRelayBlock) error {
	x.InvRelayBlock = &InvRelayBlockMessage{
		Hash: wireHashToProto(msgInvRelayBlock.Hash),
	}
	return nil
}
