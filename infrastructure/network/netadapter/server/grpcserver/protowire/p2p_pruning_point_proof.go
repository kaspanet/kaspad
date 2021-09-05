package protowire

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/pkg/errors"
)

func (x *KaspadMessage_PruningPointProof) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_PruningPointProof is nil")
	}

	if x.PruningPointProof == nil {
		return nil, errors.Wrapf(errorNil, "x.PruningPointProof is nil")
	}

	blockHeaders := make([][]*appmessage.MsgBlockHeader, len(x.PruningPointProof.Headers))
	for i, blockHeaderArray := range x.PruningPointProof.Headers {
		blockHeaders[i] = make([]*appmessage.MsgBlockHeader, len(blockHeaderArray.Headers))
		for j, blockHeader := range blockHeaderArray.Headers {
			var err error
			blockHeaders[i][j], err = blockHeader.toAppMessage()
			if err != nil {
				return nil, err
			}
		}
	}
	return &appmessage.MsgPruningPointProof{
		Headers: blockHeaders,
	}, nil
}

func (x *KaspadMessage_PruningPointProof) fromAppMessage(msgPruningPointProof *appmessage.MsgPruningPointProof) error {
	blockHeaders := make([]*PruningPointProofHeaderArray, len(msgPruningPointProof.Headers))
	for i, blockHeaderArray := range msgPruningPointProof.Headers {
		blockHeaders[i] = &PruningPointProofHeaderArray{Headers: make([]*BlockHeader, len(blockHeaderArray))}
		for j, blockHeader := range blockHeaderArray {
			blockHeaders[i].Headers[j] = &BlockHeader{}
			err := blockHeaders[i].Headers[j].fromAppMessage(blockHeader)
			if err != nil {
				return err
			}
		}
	}

	x.PruningPointProof = &PruningPointProofMessage{
		Headers: blockHeaders,
	}
	return nil
}
