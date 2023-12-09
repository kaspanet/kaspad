package protowire

import (
<<<<<<< Updated upstream
	"github.com/zoomy-network/zoomyd/app/appmessage"
=======
>>>>>>> Stashed changes
	"github.com/pkg/errors"
	"github.com/zoomy-network/zoomyd/app/appmessage"
)

func (x *KaspadMessage_DonePruningPointUtxoSetChunks) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_DonePruningPointUtxoSetChunks is nil")
	}
	return &appmessage.MsgDonePruningPointUTXOSetChunks{}, nil
}

func (x *KaspadMessage_DonePruningPointUtxoSetChunks) fromAppMessage(_ *appmessage.MsgDonePruningPointUTXOSetChunks) error {
	x.DonePruningPointUtxoSetChunks = &DonePruningPointUtxoSetChunksMessage{}
	return nil
}
