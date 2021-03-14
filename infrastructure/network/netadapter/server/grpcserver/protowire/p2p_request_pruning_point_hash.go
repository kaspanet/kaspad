package protowire

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/pkg/errors"
)

func (x *KaspadMessage_RequestPruningPointHash) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_RequestPruningPointHash is nil")
	}
	return &appmessage.MsgRequestPruningPointHashMessage{}, nil
}

func (x *KaspadMessage_RequestPruningPointHash) fromAppMessage(_ *appmessage.MsgRequestPruningPointHashMessage) error {
	return nil
}
