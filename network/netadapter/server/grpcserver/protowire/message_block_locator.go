package protowire

import (
	"github.com/kaspanet/kaspad/network/appmessage"
	"github.com/pkg/errors"
)

func (x *KaspadMessage_BlockLocator) toDomainMessage() (appmessage.Message, error) {
	if len(x.BlockLocator.Hashes) > appmessage.MaxBlockLocatorsPerMsg {
		return nil, errors.Errorf("too many block locator hashes for message "+
			"[count %d, max %d]", len(x.BlockLocator.Hashes), appmessage.MaxBlockLocatorsPerMsg)
	}
	hashes, err := protoHashesToWire(x.BlockLocator.Hashes)
	if err != nil {
		return nil, err
	}
	return &appmessage.MsgBlockLocator{BlockLocatorHashes: hashes}, nil
}

func (x *KaspadMessage_BlockLocator) fromDomainMessage(msgBlockLocator *appmessage.MsgBlockLocator) error {
	if len(msgBlockLocator.BlockLocatorHashes) > appmessage.MaxBlockLocatorsPerMsg {
		return errors.Errorf("too many block locator hashes for message "+
			"[count %d, max %d]", len(msgBlockLocator.BlockLocatorHashes), appmessage.MaxBlockLocatorsPerMsg)
	}
	x.BlockLocator = &BlockLocatorMessage{
		Hashes: wireHashesToProto(msgBlockLocator.BlockLocatorHashes),
	}
	return nil
}
