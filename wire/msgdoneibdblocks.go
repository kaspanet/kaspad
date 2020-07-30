package wire

import (
	"io"
)

// MsgDoneIBDBlocks implements the Message interface and represents a kaspa
// DoneIBDBlocks message. It is used to notify the IBD syncing peer that the
// syncer sent all the requested blocks.
//
// This message has no payload.
type MsgDoneIBDBlocks struct{}

// KaspaDecode decodes r using the kaspa protocol encoding into the receiver.
// This is part of the Message interface implementation.
func (msg *MsgDoneIBDBlocks) KaspaDecode(r io.Reader, pver uint32) error {
	return nil
}

// KaspaEncode encodes the receiver to w using the kaspa protocol encoding.
// This is part of the Message interface implementation.
func (msg *MsgDoneIBDBlocks) KaspaEncode(w io.Writer, pver uint32) error {
	return nil
}

// Command returns the protocol command string for the message. This is part
// of the Message interface implementation.
func (msg *MsgDoneIBDBlocks) Command() MessageCommand {
	return CmdDoneIBDBlocks
}

// MaxPayloadLength returns the maximum length the payload can be for the
// receiver. This is part of the Message interface implementation.
func (msg *MsgDoneIBDBlocks) MaxPayloadLength(pver uint32) uint32 {
	return 0
}

// NewMsgDoneIBDBlocks returns a new kaspa GetNextIBDBlocks message that conforms to the
// Message interface.
func NewMsgDoneIBDBlocks() *MsgDoneIBDBlocks {
	return &MsgDoneIBDBlocks{}
}
