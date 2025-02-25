package protowire

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/pkg/errors"
)

func (x *KaspadMessage_GetPruningWindowRootsRequest) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_GetPruningWindowRootsRequest is nil")
	}
	return &appmessage.GetPeerAddressesRequestMessage{}, nil
}

func (x *KaspadMessage_GetPruningWindowRootsRequest) fromAppMessage(_ *appmessage.GetPruningWindowRootsRequestMessage) error {
	return nil
}

func (x *KaspadMessage_GetPruningWindowRootsResponse) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_GetPruningWindowRootsResponse is nil")
	}
	return x.GetPruningWindowRootsResponse.toAppMessage()
}

func (x *KaspadMessage_GetPruningWindowRootsResponse) fromAppMessage(message *appmessage.GetPruningWindowRootsResponseMessage) error {
	var err *RPCError
	if message.Error != nil {
		err = &RPCError{Message: message.Error.Message}
	}

	roots := make([]*PruningWindowRoot, len(message.Roots))
	for i, root := range message.Roots {
		roots[i] = &PruningWindowRoot{
			Root:    root.Root,
			PpIndex: root.PPIndex,
		}
	}

	x.GetPruningWindowRootsResponse = &GetPruningWindowRootsResponseMessage{
		Roots: roots,
		Error: err,
	}

	return nil
}

func (x *GetPruningWindowRootsResponseMessage) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "GetPeerAddressesResponseMessage is nil")
	}
	rpcErr, err := x.Error.toAppMessage()
	// Error is an optional field
	if err != nil && !errors.Is(err, errorNil) {
		return nil, err
	}

	roots := make([]*appmessage.PruningWindowRoot, len(x.Roots))
	for i, root := range x.Roots {
		roots[i] = &appmessage.PruningWindowRoot{
			Root:    root.Root,
			PPIndex: root.PpIndex,
		}
	}

	return &appmessage.GetPruningWindowRootsResponseMessage{
		Roots: roots,
		Error: rpcErr,
	}, nil
}
