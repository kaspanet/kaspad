package protowire

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/pkg/errors"
)

func (x *KaspadMessage_RequestIBDBlocks) toAppMessage() (appmessage.Message, error) {
	if len(x.RequestIBDBlocks.Hashes) > appmessage.MaxRequestIBDBlocksHashes {
		return nil, errors.Errorf("too many hashes for message "+
			"[count %d, max %d]", len(x.RequestIBDBlocks.Hashes), appmessage.MaxRequestIBDBlocksHashes)
	}
	hashes, err := protoHashesToDomain(x.RequestIBDBlocks.Hashes)
	if err != nil {
		return nil, err
	}
	return &appmessage.MsgRequestIBDBlocks{Hashes: hashes}, nil
}

func (x *KaspadMessage_RequestIBDBlocks) fromAppMessage(msgRequestIBDBlocks *appmessage.MsgRequestIBDBlocks) error {
	if len(msgRequestIBDBlocks.Hashes) > appmessage.MaxRequestIBDBlocksHashes {
		return errors.Errorf("too many hashes for message "+
			"[count %d, max %d]", len(msgRequestIBDBlocks.Hashes), appmessage.MaxRequestIBDBlocksHashes)
	}

	x.RequestIBDBlocks = &RequestIBDBlocksMessage{
		Hashes: domainHashesToProto(msgRequestIBDBlocks.Hashes),
	}

	return nil
}
