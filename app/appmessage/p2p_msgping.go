// Copyright (c) 2013-2015 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package appmessage

// MsgPing implements the Message interface and represents a kaspa ping
// message.
//
// For versions BIP0031Version and earlier, it is used primarily to confirm
// that a connection is still valid. A transmission error is typically
// interpreted as a closed connection and that the peer should be removed.
// For versions AFTER BIP0031Version it contains an identifier which can be
// returned in the pong message to determine network timing.
//
// The payload for this message just consists of a nonce used for identifying
// it later.
type MsgPing struct {
	baseMessage
	// Unique value associated with message that is used to identify
	// specific ping message.
	Nonce uint64
}

// Command returns the protocol command string for the message. This is part
// of the Message interface implementation.
func (msg *MsgPing) Command() MessageCommand {
	return CmdPing
}

// NewMsgPing returns a new kaspa ping message that conforms to the Message
// interface. See MsgPing for details.
func NewMsgPing(nonce uint64) *MsgPing {
	return &MsgPing{
		Nonce: nonce,
	}
}
