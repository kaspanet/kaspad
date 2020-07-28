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
	case *KaspadMessage_Transaction:
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
		payload.fromWireMessage(message)
		return payload, nil
	case *wire.MsgBlock:
		payload := new(KaspadMessage_Block)
		payload.fromWireMessage(message)
		return payload, nil
	default:
		return nil, errors.Errorf("unknown message type %T", message)
	}
}
