package protowire

import "github.com/kaspanet/kaspad/app/appmessage"

func (x *KaspadMessage_PruningPointHash) toAppMessage() (appmessage.Message, error) {
	hash, err := x.PruningPointHash.Hash.toDomain()
	if err != nil {
		return nil, err
	}

	return &appmessage.MsgPruningPointHashMessage{Hash: hash}, nil
}

func (x *KaspadMessage_PruningPointHash) fromAppMessage(msgPruningPointHash *appmessage.MsgPruningPointHashMessage) error {
	x.PruningPointHash = &PruningPointHashMessage{
		Hash: domainHashToProto(msgPruningPointHash.Hash),
	}
	return nil
}
