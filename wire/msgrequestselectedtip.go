package wire

import (
	"io"
)

// MsgRequestSelectedTip implements the Message interface and represents a kaspa
// RequestSelectedTip message. It is used to request the selected tip of another peer.
//
// This message has no payload.
type MsgRequestSelectedTip struct{}

// KaspaDecode decodes r using the kaspa protocol encoding into the receiver.
// This is part of the Message interface implementation.
func (msg *MsgRequestSelectedTip) KaspaDecode(r io.Reader, pver uint32) error {
	return nil
}

// KaspaEncode encodes the receiver to w using the kaspa protocol encoding.
// This is part of the Message interface implementation.
func (msg *MsgRequestSelectedTip) KaspaEncode(w io.Writer, pver uint32) error {
	return nil
}

// Command returns the protocol command string for the message. This is part
// of the Message interface implementation.
func (msg *MsgRequestSelectedTip) Command() MessageCommand {
	return CmdRequestSelectedTip
}

// MaxPayloadLength returns the maximum length the payload can be for the
// receiver. This is part of the Message interface implementation.
func (msg *MsgRequestSelectedTip) MaxPayloadLength(pver uint32) uint32 {
	return 0
}

// NewMsgGetSelectedTip returns a new kaspa RequestSelectedTip message that conforms to the
// Message interface.
func NewMsgGetSelectedTip() *MsgRequestSelectedTip {
	return &MsgRequestSelectedTip{}
}
