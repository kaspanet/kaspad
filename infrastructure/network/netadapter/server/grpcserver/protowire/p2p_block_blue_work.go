package protowire

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/pkg/errors"
	"math/big"
)

func (x *KaspadMessage_BlockBlueWork) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_BlockBlueWork is nil")
	}
	return x.BlockBlueWork.toAppMessage()
}

func (x *BlockBlueWorkMessage) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "BlockBlueWorkMessage is nil")
	}

	return &appmessage.MsgBlockBlueWork{BlueWork: big.NewInt(0).SetBytes(x.BlueWork)}, nil

}

func (x *KaspadMessage_BlockBlueWork) fromAppMessage(msgBlockBlueWork *appmessage.MsgBlockBlueWork) error {
	x.BlockBlueWork = &BlockBlueWorkMessage{
		BlueWork: msgBlockBlueWork.BlueWork.Bytes(),
	}
	return nil
}
