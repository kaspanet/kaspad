package protowire

import (
<<<<<<< Updated upstream
	"github.com/zoomy-network/zoomyd/app/appmessage"
=======
>>>>>>> Stashed changes
	"github.com/pkg/errors"
	"github.com/zoomy-network/zoomyd/app/appmessage"
)

func (x *KaspadMessage_RequestPruningPointAndItsAnticone) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_RequestPruningPointAndItsAnticone is nil")
	}
	return &appmessage.MsgRequestPruningPointAndItsAnticone{}, nil
}

func (x *KaspadMessage_RequestPruningPointAndItsAnticone) fromAppMessage(_ *appmessage.MsgRequestPruningPointAndItsAnticone) error {
	return nil
}
