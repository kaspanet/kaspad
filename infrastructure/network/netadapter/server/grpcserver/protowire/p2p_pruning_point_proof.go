package protowire

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/pkg/errors"
	"math/big"
)

func (x *KaspadMessage_PruningPointProof) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_PruningPointProof is nil")
	}

	if x.PruningPointProof == nil {
		return nil, errors.Wrapf(errorNil, "x.PruningPointProof is nil")
	}

	blockHeaders := make([]*appmessage.MsgBlockHeader, len(x.PruningPointProof.Headers))
	for i, blockHeader := range x.PruningPointProof.Headers {
		var err error
		blockHeaders[i], err = blockHeader.toAppMessage()
		if err != nil {
			return nil, err
		}
	}
	return &appmessage.MsgPruningPointProof{
		Headers:              blockHeaders,
		PruningPointBlueWork: big.NewInt(0).SetBytes(x.PruningPointProof.PruningPointBlueWork),
	}, nil
}

func (x *KaspadMessage_PruningPointProof) fromAppMessage(msgPruningPointProof *appmessage.MsgPruningPointProof) error {
	blockHeaders := make([]*BlockHeader, len(msgPruningPointProof.Headers))
	for i, blockHeader := range msgPruningPointProof.Headers {
		blockHeaders[i] = &BlockHeader{}
		err := blockHeaders[i].fromAppMessage(blockHeader)
		if err != nil {
			return err
		}
	}

	x.PruningPointProof = &PruningPointProofMessage{
		Headers:              blockHeaders,
		PruningPointBlueWork: msgPruningPointProof.PruningPointBlueWork.Bytes(),
	}
	return nil
}
