package protowire

import (
	"github.com/c4ei/yunseokyeol/app/appmessage"
	"github.com/c4ei/yunseokyeol/domain/consensus/model/externalapi"
	"github.com/pkg/errors"
)

func (x *C4exdMessage_BlockLocator) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "C4exdMessage_BlockLocator is nil")
	}
	hashes, err := x.BlockLocator.toAppMessage()
	if err != nil {
		return nil, err
	}
	return &appmessage.MsgBlockLocator{BlockLocatorHashes: hashes}, nil
}

func (x *BlockLocatorMessage) toAppMessage() ([]*externalapi.DomainHash, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "BlockLocatorMessage is nil")
	}
	if len(x.Hashes) > appmessage.MaxBlockLocatorsPerMsg {
		return nil, errors.Errorf("too many block locator hashes for message "+
			"[count %d, max %d]", len(x.Hashes), appmessage.MaxBlockLocatorsPerMsg)
	}
	return protoHashesToDomain(x.Hashes)
}

func (x *C4exdMessage_BlockLocator) fromAppMessage(msgBlockLocator *appmessage.MsgBlockLocator) error {
	if len(msgBlockLocator.BlockLocatorHashes) > appmessage.MaxBlockLocatorsPerMsg {
		return errors.Errorf("too many block locator hashes for message "+
			"[count %d, max %d]", len(msgBlockLocator.BlockLocatorHashes), appmessage.MaxBlockLocatorsPerMsg)
	}
	x.BlockLocator = &BlockLocatorMessage{
		Hashes: domainHashesToProto(msgBlockLocator.BlockLocatorHashes),
	}
	return nil
}
