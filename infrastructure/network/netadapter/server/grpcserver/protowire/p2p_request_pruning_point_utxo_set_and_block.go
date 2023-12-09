package protowire

import (
<<<<<<< Updated upstream
	"github.com/zoomy-network/zoomyd/app/appmessage"
=======
>>>>>>> Stashed changes
	"github.com/pkg/errors"
	"github.com/zoomy-network/zoomyd/app/appmessage"
)

func (x *KaspadMessage_RequestPruningPointUTXOSet) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_RequestPruningPointUTXOSet is nil")
	}
	return x.RequestPruningPointUTXOSet.toAppMessage()
}

func (x *RequestPruningPointUTXOSetMessage) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "RequestPruningPointUTXOSetMessage is nil")
	}
	pruningPointHash, err := x.PruningPointHash.toDomain()
	if err != nil {
		return nil, err
	}
	return &appmessage.MsgRequestPruningPointUTXOSet{PruningPointHash: pruningPointHash}, nil
}

func (x *KaspadMessage_RequestPruningPointUTXOSet) fromAppMessage(
	msgRequestPruningPointUTXOSet *appmessage.MsgRequestPruningPointUTXOSet) error {

	x.RequestPruningPointUTXOSet = &RequestPruningPointUTXOSetMessage{}
	x.RequestPruningPointUTXOSet.PruningPointHash = domainHashToProto(msgRequestPruningPointUTXOSet.PruningPointHash)
	return nil
}
