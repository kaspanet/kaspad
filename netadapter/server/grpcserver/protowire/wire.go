package protowire

import (
	"bytes"

	"github.com/kaspanet/kaspad/wire"
)

func (x *KaspadMessage) ToWireMessage() (wire.Message, error) {
	msg, err := wire.MakeEmptyMessage(x.Command)
	if err != nil {
		return nil, err
	}

	payloadReader := bytes.NewReader(x.Payload)
	err = msg.KaspaDecode(payloadReader, wire.ProtocolVersion)
	if err != nil {
		return nil, err
	}

	return msg, nil
}

func FromWireMessage(msg wire.Message) (*KaspadMessage, error) {
	payloadWriter := &bytes.Buffer{}

	err := msg.KaspaEncode(payloadWriter, wire.ProtocolVersion)
	if err != nil {
		return nil, err
	}

	return &KaspadMessage{
		Command: msg.Command(),
		Payload: payloadWriter.Bytes(),
	}, nil
}
