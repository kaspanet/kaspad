package wire

import (
	"github.com/kaspanet/kaspad/util/daghash"
	"io"
)

// MsgSelectedTip implements the Message interface and represents a kaspa
// selectedtip message. It is used to answer getseltip messages and tell
// the asking peer what is the selected tip of this peer.
type MsgSelectedTip struct {
	// The selected tip hash of the generator of the message.
	SelectedTipHash *daghash.Hash
}

// KaspaDecode decodes r using the kaspa protocol encoding into the receiver.
// This is part of the Message interface implementation.
func (msg *MsgSelectedTip) KaspaDecode(r io.Reader, pver uint32) error {
	msg.SelectedTipHash = &daghash.Hash{}
	err := ReadElement(r, msg.SelectedTipHash)
	if err != nil {
		return err
	}

	return nil
}

// KaspaEncode encodes the receiver to w using the kaspa protocol encoding.
// This is part of the Message interface implementation.
func (msg *MsgSelectedTip) KaspaEncode(w io.Writer, pver uint32) error {
	return WriteElement(w, msg.SelectedTipHash)
}

// Command returns the protocol command string for the message. This is part
// of the Message interface implementation.
func (msg *MsgSelectedTip) Command() MessageCommand {
	return CmdSelectedTip
}

// MaxPayloadLength returns the maximum length the payload can be for the
// receiver. This is part of the Message interface implementation.
func (msg *MsgSelectedTip) MaxPayloadLength(_ uint32) uint32 {
	// selected tip hash 32 bytes
	return daghash.HashSize
}

// NewMsgSelectedTip returns a new kaspa selectedtip message that conforms to the
// Message interface.
func NewMsgSelectedTip(selectedTipHash *daghash.Hash) *MsgSelectedTip {
	return &MsgSelectedTip{
		SelectedTipHash: selectedTipHash,
	}
}
