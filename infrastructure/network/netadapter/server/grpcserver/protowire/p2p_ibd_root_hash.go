package protowire

import "github.com/kaspanet/kaspad/app/appmessage"

func (x *KaspadMessage_IbdRootHash) toAppMessage() (appmessage.Message, error) {
	hash, err := x.IbdRootHash.Hash.toDomain()
	if err != nil {
		return nil, err
	}

	return &appmessage.MsgIBDRootHash{Hash: hash}, nil
}

func (x *KaspadMessage_IbdRootHash) fromAppMessage(msgIBDRootHash *appmessage.MsgIBDRootHash) error {
	x.IbdRootHash = &IBDRootHash{
		Hash: domainHashToProto(msgIBDRootHash.Hash),
	}
	return nil
}
