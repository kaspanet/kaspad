package protowire

import "github.com/kaspanet/kaspad/app/appmessage"

func (x *KaspadMessage_RequestIBDRootUTXOSetAndBlock) toAppMessage() (appmessage.Message, error) {
	ibdRoot, err := x.RequestIBDRootUTXOSetAndBlock.IbdRoot.toDomain()
	if err != nil {
		return nil, err
	}

	return &appmessage.MsgRequestIBDRootUTXOSetAndBlock{IBDRoot: ibdRoot}, nil
}

func (x *KaspadMessage_RequestIBDRootUTXOSetAndBlock) fromAppMessage(
	msgRequestIBDRootUTXOSetAndBlock *appmessage.MsgRequestIBDRootUTXOSetAndBlock) error {
	x.RequestIBDRootUTXOSetAndBlock = &RequestIBDRootUTXOSetAndBlockMessage{}
	x.RequestIBDRootUTXOSetAndBlock.IbdRoot = domainHashToProto(msgRequestIBDRootUTXOSetAndBlock.IBDRoot)
	return nil
}
