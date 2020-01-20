package wire

import (
	"io"

	"github.com/kaspanet/kaspad/util/daghash"
)

// MsgGetBlockLocator implements the Message interface and represents a kaspa
// getlocator message. It is used to request a block locator between start and stop hash.
// The locator is returned via a locator message (MsgBlockLocator).
//
// This message has no payload.
type MsgGetBlockLocator struct {
	HighHash *daghash.Hash
	LowHash  *daghash.Hash
}

// KaspaDecode decodes r using the kaspa protocol encoding into the receiver.
// This is part of the Message interface implementation.
func (msg *MsgGetBlockLocator) KaspaDecode(r io.Reader, pver uint32) error {
	msg.HighHash = &daghash.Hash{}
	err := ReadElement(r, msg.HighHash)
	if err != nil {
		return err
	}

	msg.LowHash = &daghash.Hash{}
	err = ReadElement(r, msg.LowHash)
	if err != nil {
		return err
	}
	return nil
}

// KaspaEncode encodes the receiver to w using the kaspa protocol encoding.
// This is part of the Message interface implementation.
func (msg *MsgGetBlockLocator) KaspaEncode(w io.Writer, pver uint32) error {
	err := WriteElement(w, msg.HighHash)
	if err != nil {
		return err
	}

	err = WriteElement(w, msg.LowHash)
	if err != nil {
		return err
	}
	return nil
}

// Command returns the protocol command string for the message. This is part
// of the Message interface implementation.
func (msg *MsgGetBlockLocator) Command() string {
	return CmdGetBlockLocator
}

// MaxPayloadLength returns the maximum length the payload can be for the
// receiver. This is part of the Message interface implementation.
func (msg *MsgGetBlockLocator) MaxPayloadLength(pver uint32) uint32 {
	return daghash.HashSize * 2
}

// NewMsgGetBlockLocator returns a new getlocator message that conforms to the
// Message interface using the passed parameters and defaults for the remaining
// fields.
func NewMsgGetBlockLocator(highHash, lowHash *daghash.Hash) *MsgGetBlockLocator {
	return &MsgGetBlockLocator{
		HighHash: highHash,
		LowHash:  lowHash,
	}
}
