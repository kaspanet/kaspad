package protowire

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/pkg/errors"
)

func (x *KaspadMessage_BlockLocator) toAppMessage() (appmessage.Message, error) {
	if len(x.BlockLocator.Hashes) > appmessage.MaxBlockLocatorsPerMsg {
		return nil, errors.Errorf("too many block locator hashes for message "+
			"[count %d, max %d]", len(x.BlockLocator.Hashes), appmessage.MaxBlockLocatorsPerMsg)
	}
	hashes, err := protoHashesToDomain(x.BlockLocator.Hashes)
	if err != nil {
		return nil, err
	}
	return &appmessage.MsgBlockLocator{BlockLocatorHashes: hashes}, nil
}

func (x *KaspadMessage_BlockLocator) fromAppMessage(msgBlockLocator *appmessage.MsgBlockLocator) error {
	if len(msgBlockLocator.BlockLocatorHashes) > appmessage.MaxBlockLocatorsPerMsg {
		return errors.Errorf("too many block locator hashes for message "+
			"[count %d, max %d]", len(msgBlockLocator.BlockLocatorHashes), appmessage.MaxBlockLocatorsPerMsg)
	}
	x.BlockLocator = &BlockLocatorMessage{
		Hashes: domainHashesToProto(msgBlockLocator.BlockLocatorHashes),
	}
	return nil
}
