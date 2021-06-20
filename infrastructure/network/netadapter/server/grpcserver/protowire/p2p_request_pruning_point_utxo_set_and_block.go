package protowire

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/pkg/errors"
)

func (x *KaspadMessage_RequestPruningPointUTXOSetAndBlock) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_RequestPruningPointUTXOSetAndBlock is nil")
	}
	return x.RequestPruningPointUTXOSetAndBlock.toAppMessage()
}

func (x *RequestPruningPointUTXOSetAndBlockMessage) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "RequestPruningPointUTXOSetAndBlockMessage is nil")
	}
	pruningPointHash, err := x.PruningPointHash.toDomain()
	if err != nil {
		return nil, err
	}
	return &appmessage.MsgRequestPruningPointUTXOSet{PruningPointHash: pruningPointHash}, nil
}

func (x *KaspadMessage_RequestPruningPointUTXOSetAndBlock) fromAppMessage(
	msgRequestPruningPointUTXOSetAndBlock *appmessage.MsgRequestPruningPointUTXOSet) error {

	x.RequestPruningPointUTXOSetAndBlock = &RequestPruningPointUTXOSetAndBlockMessage{}
	x.RequestPruningPointUTXOSetAndBlock.PruningPointHash = domainHashToProto(msgRequestPruningPointUTXOSetAndBlock.PruningPointHash)
	return nil
}
