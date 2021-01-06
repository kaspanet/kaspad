package protowire

import "github.com/kaspanet/kaspad/app/appmessage"

func (x *KaspadMessage_RequestBlockLocator) toAppMessage() (appmessage.Message, error) {
	lowHash, err := x.RequestBlockLocator.LowHash.toDomain()
	if err != nil {
		return nil, err
	}

	highHash, err := x.RequestBlockLocator.HighHash.toDomain()
	if err != nil {
		return nil, err
	}

	return &appmessage.MsgRequestBlockLocator{
		LowHash:  lowHash,
		HighHash: highHash,
		Limit:    x.RequestBlockLocator.Limit,
	}, nil
}

func (x *KaspadMessage_RequestBlockLocator) fromAppMessage(msgGetBlockLocator *appmessage.MsgRequestBlockLocator) error {
	x.RequestBlockLocator = &RequestBlockLocatorMessage{
		LowHash:  domainHashToProto(msgGetBlockLocator.LowHash),
		HighHash: domainHashToProto(msgGetBlockLocator.HighHash),
		Limit:    msgGetBlockLocator.Limit,
	}
	return nil
}
