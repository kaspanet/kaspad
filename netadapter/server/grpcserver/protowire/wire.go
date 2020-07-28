package protowire

import (
	"github.com/kaspanet/kaspad/wire"
	"github.com/pkg/errors"
)

// ToWireMessage converts a KaspadMessage to its wire.Message representation
func (x *KaspadMessage) ToWireMessage() (wire.Message, error) {
	switch payload := x.Payload.(type) {
	case *KaspadMessage_Addresses:
		return payload.toWireMessage()
	case *KaspadMessage_Block:
		return payload.toWireMessage()
	case *KaspadMessage_BlockLocator:
		return payload.toWireMessage()
	case *KaspadMessage_GetAddresses_:
		return payload.toWireMessage()
	case *KaspadMessage_GetBlocks:
		return payload.toWireMessage()
	case *KaspadMessage_GetRelayBlocks:
		return payload.toWireMessage()
	case *KaspadMessage_GetSelectedTip:
		return payload.toWireMessage()
	case *KaspadMessage_GetTransactions:
		return payload.toWireMessage()
	case *KaspadMessage_IbdBlock:
		return payload.toWireMessage()
	case *KaspadMessage_InvRelayBlock:
		return payload.toWireMessage()
	case *KaspadMessage_InvTransactions:
		return payload.toWireMessage()
	case *KaspadMessage_Ping:
		return payload.toWireMessage()
	case *KaspadMessage_Pong:
		return payload.toWireMessage()
	case *KaspadMessage_SelectedTip_:
		return payload.toWireMessage()
	case *KaspadMessage_Transaction:
		return payload.toWireMessage()
	case *KaspadMessage_Verack:
		return payload.toWireMessage()
	case *KaspadMessage_Version:
		return payload.toWireMessage()
	default:
		return nil, errors.Errorf("unknown payload type %T", payload)
	}
}

// FromWireMessage creates a KaspadMessage from a wire.Message
func FromWireMessage(message wire.Message) (*KaspadMessage, error) {
	payload, err := toPayload(message)
	if err != nil {
		return nil, err
	}
	return &KaspadMessage{
		Payload: payload,
	}, nil
}

func toPayload(message wire.Message) (isKaspadMessage_Payload, error) {
	switch message := message.(type) {
	case *wire.MsgAddresses:
		payload := new(KaspadMessage_Addresses)
		err := payload.fromWireMessage(message)
		if err != nil {
			return nil, err
		}
		return payload, nil
	case *wire.MsgBlock:
		payload := new(KaspadMessage_Block)
		err := payload.fromWireMessage(message)
		if err != nil {
			return nil, err
		}
		return payload, nil
	case *wire.MsgBlockLocator:
		payload := new(KaspadMessage_BlockLocator)
		err := payload.fromWireMessage(message)
		if err != nil {
			return nil, err
		}
		return payload, nil
	case *wire.MsgGetAddresses:
		payload := new(KaspadMessage_GetAddresses_)
		err := payload.fromWireMessage(message)
		if err != nil {
			return nil, err
		}
		return payload, nil
	case *wire.MsgGetBlocks:
		payload := new(KaspadMessage_GetBlocks)
		err := payload.fromWireMessage(message)
		if err != nil {
			return nil, err
		}
		return payload, nil
	case *wire.MsgGetRelayBlocks:
		payload := new(KaspadMessage_GetRelayBlocks)
		err := payload.fromWireMessage(message)
		if err != nil {
			return nil, err
		}
		return payload, nil
	case *wire.MsgGetSelectedTip:
		payload := new(KaspadMessage_GetSelectedTip)
		err := payload.fromWireMessage(message)
		if err != nil {
			return nil, err
		}
		return payload, nil
	case *wire.MsgGetTransactions:
		payload := new(KaspadMessage_GetTransactions)
		err := payload.fromWireMessage(message)
		if err != nil {
			return nil, err
		}
		return payload, nil
	case *wire.MsgIBDBlock:
		payload := new(KaspadMessage_IbdBlock)
		err := payload.fromWireMessage(message)
		if err != nil {
			return nil, err
		}
		return payload, nil
	case *wire.MsgInvRelayBlock:
		payload := new(KaspadMessage_InvRelayBlock)
		err := payload.fromWireMessage(message)
		if err != nil {
			return nil, err
		}
		return payload, nil
	case *wire.MsgInvTransaction:
		payload := new(KaspadMessage_InvTransactions)
		err := payload.fromWireMessage(message)
		if err != nil {
			return nil, err
		}
		return payload, nil
	case *wire.MsgPing:
		payload := new(KaspadMessage_Ping)
		err := payload.fromWireMessage(message)
		if err != nil {
			return nil, err
		}
		return payload, nil
	case *wire.MsgPong:
		payload := new(KaspadMessage_Pong)
		err := payload.fromWireMessage(message)
		if err != nil {
			return nil, err
		}
		return payload, nil
	case *wire.MsgSelectedTip:
		payload := new(KaspadMessage_SelectedTip_)
		err := payload.fromWireMessage(message)
		if err != nil {
			return nil, err
		}
		return payload, nil
	case *wire.MsgTx:
		payload := new(KaspadMessage_Transaction)
		err := payload.fromWireMessage(message)
		if err != nil {
			return nil, err
		}
		return payload, nil
	case *wire.MsgVerAck:
		payload := new(KaspadMessage_Verack)
		err := payload.fromWireMessage(message)
		if err != nil {
			return nil, err
		}
		return payload, nil
	case *wire.MsgVersion:
		payload := new(KaspadMessage_Version)
		err := payload.fromWireMessage(message)
		if err != nil {
			return nil, err
		}
		return payload, nil
	default:
		return nil, errors.Errorf("unknown message type %T", message)
	}
}
