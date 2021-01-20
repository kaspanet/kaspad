package protowire

import "github.com/kaspanet/kaspad/app/appmessage"

func (x *KaspadMessage_RequestPruningPointUTXOSetAndBlock) toAppMessage() (appmessage.Message, error) {
	pruningPointHash, err := x.RequestPruningPointUTXOSetAndBlock.PruningPointHash.toDomain()
	if err != nil {
		return nil, err
	}
	return &appmessage.MsgRequestPruningPointUTXOSetAndBlock{PruningPointHash: pruningPointHash}, nil
}

func (x *KaspadMessage_RequestPruningPointUTXOSetAndBlock) fromAppMessage(
	msgRequestPruningPointUTXOSetAndBlock *appmessage.MsgRequestPruningPointUTXOSetAndBlock) error {

	x.RequestPruningPointUTXOSetAndBlock = &RequestPruningPointUTXOSetAndBlockMessage{}
	x.RequestPruningPointUTXOSetAndBlock.PruningPointHash = domainHashToProto(msgRequestPruningPointUTXOSetAndBlock.PruningPointHash)
	return nil
}
