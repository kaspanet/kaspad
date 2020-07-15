// Copyright (c) 2013-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package wire

import "io"

// MsgIBDBlock implements the Message interface and represents a kaspa
// ibdblock message. It is used to deliver block and transaction information in
// response to a getblocks message (MsgGetBlocks).
type MsgIBDBlock struct {
	MsgBlock
}

// KaspaDecode decodes r using the kaspa protocol encoding into the receiver.
// This is part of the Message interface implementation.
func (msg *MsgIBDBlock) KaspaDecode(r io.Reader, pver uint32) error {
	return msg.MsgBlock.KaspaDecode(r, pver)
}

// KaspaEncode encodes the receiver to w using the kaspa protocol encoding.
// This is part of the Message interface implementation.
func (msg *MsgIBDBlock) KaspaEncode(w io.Writer, pver uint32) error {
	return msg.MsgBlock.KaspaEncode(w, pver)
}

// Command returns the protocol command string for the message. This is part
// of the Message interface implementation.
func (msg *MsgIBDBlock) Command() string {
	return CmdIBDBlock
}

// MaxPayloadLength returns the maximum length the payload can be for the
// receiver. This is part of the Message interface implementation.
func (msg *MsgIBDBlock) MaxPayloadLength(pver uint32) uint32 {
	return MaxMessagePayload
}

// NewMsgIBDBlock returns a new kaspa ibdblock message that conforms to the
// Message interface. See MsgIBDBlock for details.
func NewMsgIBDBlock(msgBlock *MsgBlock) *MsgIBDBlock {
	return &MsgIBDBlock{*msgBlock}
}
