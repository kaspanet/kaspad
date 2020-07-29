package protowire

import (
	"github.com/kaspanet/kaspad/wire"
	"github.com/pkg/errors"
)

func (x *KaspadMessage_BlockLocator_) toWireMessage() (wire.Message, error) {
	if len(x.BlockLocator_.Hashes) > wire.MaxBlockLocatorsPerMsg {
		return nil, errors.Errorf("too many block locator hashes for message "+
			"[count %d, max %d]", len(x.BlockLocator_.Hashes), wire.MaxBlockLocatorsPerMsg)
	}
	hashes, err := protoHashesToWire(x.BlockLocator_.Hashes)
	if err != nil {
		return nil, err
	}
	return &wire.MsgBlockLocator{BlockLocatorHashes: hashes}, nil
}

func (x *KaspadMessage_BlockLocator_) fromWireMessage(msgBlockLocator *wire.MsgBlockLocator) error {
	if len(msgBlockLocator.BlockLocatorHashes) > wire.MaxBlockLocatorsPerMsg {
		return errors.Errorf("too many block locator hashes for message "+
			"[count %d, max %d]", len(msgBlockLocator.BlockLocatorHashes), wire.MaxBlockLocatorsPerMsg)
	}
	x.BlockLocator_ = &BlockLocatorMessage{
		Hashes: wireHashesToProto(msgBlockLocator.BlockLocatorHashes),
	}
	return nil
}
