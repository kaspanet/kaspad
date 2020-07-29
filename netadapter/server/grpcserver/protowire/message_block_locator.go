package protowire

import (
	"github.com/kaspanet/kaspad/wire"
	"github.com/pkg/errors"
)

func (x *KaspadMessage_BlockLocator) toWireMessage() (wire.Message, error) {
	if len(x.BlockLocator.Hashes) > wire.MaxBlockLocatorsPerMsg {
		return nil, errors.Errorf("too many block locator hashes for message "+
			"[count %d, max %d]", len(x.BlockLocator.Hashes), wire.MaxBlockLocatorsPerMsg)
	}
	hashes, err := protoHashesToWire(x.BlockLocator.Hashes)
	if err != nil {
		return nil, err
	}
	return &wire.MsgBlockLocator{BlockLocatorHashes: hashes}, nil
}

func (x *KaspadMessage_BlockLocator) fromWireMessage(msgBlockLocator *wire.MsgBlockLocator) error {
	if len(msgBlockLocator.BlockLocatorHashes) > wire.MaxBlockLocatorsPerMsg {
		return errors.Errorf("too many block locator hashes for message "+
			"[count %d, max %d]", len(msgBlockLocator.BlockLocatorHashes), wire.MaxBlockLocatorsPerMsg)
	}
	x.BlockLocator = &BlockLocatorMessage{
		Hashes: wireHashesToProto(msgBlockLocator.BlockLocatorHashes),
	}
	return nil
}
