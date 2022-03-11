package protowire

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/pkg/errors"
)

func (x *KaspadMessage_RequestPastDiff) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_RequestPastDiff is nil")
	}
	return x.RequestPastDiff.toAppMessage()
}

func (x *RequestPastDiffMessage) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "RequestPastDiffMessage is nil")
	}
	hasHash, err := x.HasHash.toDomain()
	if err != nil {
		return nil, err
	}

	requestedHash, err := x.RequestedHash.toDomain()
	if err != nil {
		return nil, err
	}

	return &appmessage.MsgRequestPastDiff{
		HasHash:       hasHash,
		RequestedHash: requestedHash,
	}, nil

}

func (x *KaspadMessage_RequestPastDiff) fromAppMessage(msgRequestPastDiff *appmessage.MsgRequestPastDiff) error {
	x.RequestPastDiff = &RequestPastDiffMessage{
		HasHash:       domainHashToProto(msgRequestPastDiff.HasHash),
		RequestedHash: domainHashToProto(msgRequestPastDiff.RequestedHash),
	}
	return nil
}
