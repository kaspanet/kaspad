package protowire

import "github.com/kaspanet/kaspad/domainmessage"

func (x *KaspadMessage_InvRelayBlock) toWireMessage() (domainmessage.Message, error) {
	hash, err := x.InvRelayBlock.Hash.toWire()
	if err != nil {
		return nil, err
	}

	return &domainmessage.MsgInvRelayBlock{Hash: hash}, nil
}

func (x *KaspadMessage_InvRelayBlock) fromWireMessage(msgInvRelayBlock *domainmessage.MsgInvRelayBlock) error {
	x.InvRelayBlock = &InvRelayBlockMessage{
		Hash: wireHashToProto(msgInvRelayBlock.Hash),
	}
	return nil
}
