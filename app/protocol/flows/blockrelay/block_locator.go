package blockrelay

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/protocol/common"
	"github.com/kaspanet/kaspad/app/protocol/protocolerrors"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

func (flow *handleRelayInvsFlow) sendGetBlockLocator(lowHash *externalapi.DomainHash,
	highHash *externalapi.DomainHash, limit *int) error {

	rawLimit := uint32(0)
	if limit != nil {
		rawLimit = uint32(*limit)
	}

	msgGetBlockLocator := appmessage.NewMsgRequestBlockLocator(lowHash, highHash, rawLimit)
	return flow.outgoingRoute.Enqueue(msgGetBlockLocator)
}

func (flow *handleRelayInvsFlow) receiveBlockLocator() (blockLocatorHashes []*externalapi.DomainHash, err error) {
	message, err := flow.dequeueIncomingMessageAndSkipInvs(common.DefaultTimeout)
	if err != nil {
		return nil, err
	}
	msgBlockLocator, ok := message.(*appmessage.MsgBlockLocator)
	if !ok {
		return nil,
			protocolerrors.Errorf(true, "received unexpected message type. "+
				"expected: %s, got: %s", appmessage.CmdBlockLocator, message.Command())
	}
	return msgBlockLocator.BlockLocatorHashes, nil
}
