package wire

import (
	"io"
)

// MsgGetNextIBDBlocks implements the Message interface and represents a kaspa
// GetNextIBDBlocks message. It is used to notify the IBD syncer peer to send
// more blocks.
//
// This message has no payload.
type MsgGetNextIBDBlocks struct{}

// KaspaDecode decodes r using the kaspa protocol encoding into the receiver.
// This is part of the Message interface implementation.
func (msg *MsgGetNextIBDBlocks) KaspaDecode(r io.Reader, pver uint32) error {
	return nil
}

// KaspaEncode encodes the receiver to w using the kaspa protocol encoding.
// This is part of the Message interface implementation.
func (msg *MsgGetNextIBDBlocks) KaspaEncode(w io.Writer, pver uint32) error {
	return nil
}

// Command returns the protocol command string for the message. This is part
// of the Message interface implementation.
func (msg *MsgGetNextIBDBlocks) Command() MessageCommand {
	return CmdGetNextIBDBlocks
}

// MaxPayloadLength returns the maximum length the payload can be for the
// receiver. This is part of the Message interface implementation.
func (msg *MsgGetNextIBDBlocks) MaxPayloadLength(pver uint32) uint32 {
	return 0
}

// NewMsgGetNextIBDBlocks returns a new kaspa GetNextIBDBlocks message that conforms to the
// Message interface.
func NewMsgGetNextIBDBlocks() *MsgGetNextIBDBlocks {
	return &MsgGetNextIBDBlocks{}
}
