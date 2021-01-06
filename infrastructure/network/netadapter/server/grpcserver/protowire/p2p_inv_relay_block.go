package protowire

import "github.com/kaspanet/kaspad/app/appmessage"

func (x *KaspadMessage_InvRelayBlock) toAppMessage() (appmessage.Message, error) {
	hash, err := x.InvRelayBlock.Hash.toDomain()
	if err != nil {
		return nil, err
	}

	return &appmessage.MsgInvRelayBlock{Hash: hash}, nil
}

func (x *KaspadMessage_InvRelayBlock) fromAppMessage(msgInvRelayBlock *appmessage.MsgInvRelayBlock) error {
	x.InvRelayBlock = &InvRelayBlockMessage{
		Hash: domainHashToProto(msgInvRelayBlock.Hash),
	}
	return nil
}
