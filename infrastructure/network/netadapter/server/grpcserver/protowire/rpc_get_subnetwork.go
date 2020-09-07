package protowire

import "github.com/kaspanet/kaspad/app/appmessage"

func (x *KaspadMessage_GetSubnetworkRequest) toAppMessage() (appmessage.Message, error) {
	return &appmessage.GetSubnetworkRequestMessage{
		SubnetworkID: x.GetSubnetworkRequest.SubnetworkId,
	}, nil
}

func (x *KaspadMessage_GetSubnetworkRequest) fromAppMessage(message *appmessage.GetSubnetworkRequestMessage) error {
	x.GetSubnetworkRequest = &GetSubnetworkRequestMessage{
		SubnetworkId: message.SubnetworkID,
	}
	return nil
}

func (x *KaspadMessage_GetSubnetworkResponse) toAppMessage() (appmessage.Message, error) {
	var err *appmessage.RPCError
	if x.GetSubnetworkResponse.Error != nil {
		err = &appmessage.RPCError{Message: x.GetSubnetworkResponse.Error.Message}
	}
	return &appmessage.GetSubnetworkResponseMessage{
		GasLimit: x.GetSubnetworkResponse.GasLimit,
		Error:    err,
	}, nil
}

func (x *KaspadMessage_GetSubnetworkResponse) fromAppMessage(message *appmessage.GetSubnetworkResponseMessage) error {
	var err *RPCError
	if message.Error != nil {
		err = &RPCError{Message: message.Error.Message}
	}
	x.GetSubnetworkResponse = &GetSubnetworkResponseMessage{
		GasLimit: message.GasLimit,
		Error:    err,
	}
	return nil
}
