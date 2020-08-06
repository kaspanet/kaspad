package protowire

import (
	"github.com/kaspanet/kaspad/domainmessage"
	"github.com/pkg/errors"
)

func (x *KaspadMessage_BlockLocator) toDomainMessage() (domainmessage.Message, error) {
	if len(x.BlockLocator.Hashes) > domainmessage.MaxBlockLocatorsPerMsg {
		return nil, errors.Errorf("too many block locator hashes for message "+
			"[count %d, max %d]", len(x.BlockLocator.Hashes), domainmessage.MaxBlockLocatorsPerMsg)
	}
	hashes, err := protoHashesToWire(x.BlockLocator.Hashes)
	if err != nil {
		return nil, err
	}
	return &domainmessage.MsgBlockLocator{BlockLocatorHashes: hashes}, nil
}

func (x *KaspadMessage_BlockLocator) fromDomainMessage(msgBlockLocator *domainmessage.MsgBlockLocator) error {
	if len(msgBlockLocator.BlockLocatorHashes) > domainmessage.MaxBlockLocatorsPerMsg {
		return errors.Errorf("too many block locator hashes for message "+
			"[count %d, max %d]", len(msgBlockLocator.BlockLocatorHashes), domainmessage.MaxBlockLocatorsPerMsg)
	}
	x.BlockLocator = &BlockLocatorMessage{
		Hashes: wireHashesToProto(msgBlockLocator.BlockLocatorHashes),
	}
	return nil
}
