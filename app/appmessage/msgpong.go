// Copyright (c) 2013-2015 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package appmessage

// MsgPong implements the Message interface and represents a kaspa pong
// message which is used primarily to confirm that a connection is still valid
// in response to a kaspa ping message (MsgPing).
//
// This message was not added until protocol versions AFTER BIP0031Version.
type MsgPong struct {
	baseMessage
	// Unique value associated with message that is used to identify
	// specific ping message.
	Nonce uint64
}

// Command returns the protocol command string for the message. This is part
// of the Message interface implementation.
func (msg *MsgPong) Command() MessageCommand {
	return CmdPong
}

// NewMsgPong returns a new kaspa pong message that conforms to the Message
// interface. See MsgPong for details.
func NewMsgPong(nonce uint64) *MsgPong {
	return &MsgPong{
		Nonce: nonce,
	}
}
