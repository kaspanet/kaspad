// Copyright (c) 2013-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package wire

import (
	"github.com/kaspanet/kaspad/util/daghash"
	"io"
)

// MsgSelectedTip implements the Message interface and represents a kaspa
// selectedtip message. It is used to answer getseltip messages and tell
// the asking peer what is your selected tip.
type MsgSelectedTip struct {
	// The selected tip hash of the generator of the message.
	SelectedTipHash *daghash.Hash
}

// KaspaDecode decodes r using the kaspa protocol encoding into the receiver.
// The version message is special in that the protocol version hasn't been
// negotiated yet. As a result, the pver field is ignored and any fields which
// are added in new versions are optional. This also mean that r must be a
// *bytes.Buffer so the number of remaining bytes can be ascertained.
//
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
func (msg *MsgSelectedTip) Command() string {
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
