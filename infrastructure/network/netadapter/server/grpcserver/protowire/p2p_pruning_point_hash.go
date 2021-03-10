package protowire

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/pkg/errors"
)

func (x *KaspadMessage_PruningPointHash) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_PruningPointHash is nil")
	}
	hash, err := x.PruningPointHash.Hash.toDomain()
	if err != nil {
		return nil, err
	}

	return &appmessage.MsgPruningPointHashMessage{Hash: hash}, nil
}

func (x *PruningPointHashMessage) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "PruningPointHashMessage is nil")
	}
	hash, err := x.Hash.toDomain()
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
