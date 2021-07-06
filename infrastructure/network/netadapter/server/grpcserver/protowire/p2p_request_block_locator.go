package protowire

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/pkg/errors"
)

func (x *KaspadMessage_RequestBlockLocator) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_RequestBlockLocator is nil")
	}
	return x.RequestBlockLocator.toAppMessage()
}

func (x *RequestBlockLocatorMessage) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "RequestBlockLocatorMessage is nil")
	}

	var lowHash *externalapi.DomainHash
	if x.LowHash != nil {
		var err error
		lowHash, err = x.LowHash.toDomain()
		if err != nil {
			return nil, err
		}
	}

	highHash, err := x.HighHash.toDomain()
	if err != nil {
		return nil, err
	}

	return &appmessage.MsgRequestBlockLocator{
		LowHash:  lowHash,
		HighHash: highHash,
		Limit:    x.Limit,
	}, nil

}

func (x *KaspadMessage_RequestBlockLocator) fromAppMessage(msgGetBlockLocator *appmessage.MsgRequestBlockLocator) error {
	x.RequestBlockLocator = &RequestBlockLocatorMessage{
		HighHash: domainHashToProto(msgGetBlockLocator.HighHash),
		Limit:    msgGetBlockLocator.Limit,
	}

	if msgGetBlockLocator.LowHash != nil {
		x.RequestBlockLocator.LowHash = domainHashToProto(msgGetBlockLocator.LowHash)
	}

	return nil
}
