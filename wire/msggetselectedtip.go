package wire

import (
	"io"
)

// MsgGetSelectedTip implements the Message interface and represents a kaspa
// getseltip message. It is used to request the selected tip of another tip.
//
// This message has no payload.
type MsgGetSelectedTip struct{}

// KaspaDecode decodes r using the kaspa protocol encoding into the receiver.
// This is part of the Message interface implementation.
func (msg *MsgGetSelectedTip) KaspaDecode(r io.Reader, pver uint32) error {
	return nil
}

// KaspaEncode encodes the receiver to w using the kaspa protocol encoding.
// This is part of the Message interface implementation.
func (msg *MsgGetSelectedTip) KaspaEncode(w io.Writer, pver uint32) error {
	return nil
}

// Command returns the protocol command string for the message. This is part
// of the Message interface implementation.
func (msg *MsgGetSelectedTip) Command() string {
	return CmdGetSelectedTip
}

// MaxPayloadLength returns the maximum length the payload can be for the
// receiver. This is part of the Message interface implementation.
func (msg *MsgGetSelectedTip) MaxPayloadLength(pver uint32) uint32 {
	return 0
}

// NewMsgGetSelectedTip returns a new kaspa getseltip message that conforms to the
// Message interface.
func NewMsgGetSelectedTip() *MsgGetSelectedTip {
	return &MsgGetSelectedTip{}
}
