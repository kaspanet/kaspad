// Copyright (c) 2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package wire

import (
	"io"
)

// MsgSendHeaders implements the Message interface and represents a kaspa
// sendheaders message. It is used to request the peer send block headers
// rather than inventory vectors.
//
// This message has no payload.
type MsgSendHeaders struct{}

// KaspaDecode decodes r using the kaspa protocol encoding into the receiver.
// This is part of the Message interface implementation.
func (msg *MsgSendHeaders) KaspaDecode(r io.Reader, pver uint32) error {
	return nil
}

// KaspaEncode encodes the receiver to w using the kaspa protocol encoding.
// This is part of the Message interface implementation.
func (msg *MsgSendHeaders) KaspaEncode(w io.Writer, pver uint32) error {
	return nil
}

// Command returns the protocol command string for the message. This is part
// of the Message interface implementation.
func (msg *MsgSendHeaders) Command() string {
	return CmdSendHeaders
}

// MaxPayloadLength returns the maximum length the payload can be for the
// receiver. This is part of the Message interface implementation.
func (msg *MsgSendHeaders) MaxPayloadLength(pver uint32) uint32 {
	return 0
}

// NewMsgSendHeaders returns a new kaspa sendheaders message that conforms to
// the Message interface. See MsgSendHeaders for details.
func NewMsgSendHeaders() *MsgSendHeaders {
	return &MsgSendHeaders{}
}
