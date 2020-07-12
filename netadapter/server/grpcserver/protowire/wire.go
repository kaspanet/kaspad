package protowire

import (
	"bytes"

	"github.com/kaspanet/kaspad/wire"
)

// ToWireMessage converts a KaspadMessage to its wire.Message representation
func (x *KaspadMessage) ToWireMessage() (wire.Message, error) {
	message, err := wire.MakeEmptyMessage(x.Command)
	if err != nil {
		return nil, err
	}

	payloadReader := bytes.NewReader(x.Payload)
	err = message.KaspaDecode(payloadReader, wire.ProtocolVersion)
	if err != nil {
		return nil, err
	}

	return message, nil
}

// FromWireMessage creates a KaspadMessage from a wire.Message
func FromWireMessage(message wire.Message) (*KaspadMessage, error) {
	payloadWriter := &bytes.Buffer{}

	err := message.KaspaEncode(payloadWriter, wire.ProtocolVersion)
	if err != nil {
		return nil, err
	}

	return &KaspadMessage{
		Command: message.Command(),
		Payload: payloadWriter.Bytes(),
	}, nil
}
