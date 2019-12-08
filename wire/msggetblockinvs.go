// Copyright (c) 2013-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package wire

import (
	"io"

	"github.com/daglabs/kaspad/util/daghash"
)

// MsgGetBlockInvs implements the Message interface and represents a bitcoin
// getblockinvs message.  It is used to request a list of blocks starting after the
// start hash and until the stop hash.
type MsgGetBlockInvs struct {
	StartHash *daghash.Hash
	StopHash  *daghash.Hash
}

// BtcDecode decodes r using the bitcoin protocol encoding into the receiver.
// This is part of the Message interface implementation.
func (msg *MsgGetBlockInvs) BtcDecode(r io.Reader, pver uint32) error {
	msg.StartHash = &daghash.Hash{}
	err := ReadElement(r, msg.StartHash)
	if err != nil {
		return err
	}

	msg.StopHash = &daghash.Hash{}
	return ReadElement(r, msg.StopHash)
}

// BtcEncode encodes the receiver to w using the bitcoin protocol encoding.
// This is part of the Message interface implementation.
func (msg *MsgGetBlockInvs) BtcEncode(w io.Writer, pver uint32) error {
	err := WriteElement(w, msg.StartHash)
	if err != nil {
		return err
	}

	return WriteElement(w, msg.StopHash)
}

// Command returns the protocol command string for the message.  This is part
// of the Message interface implementation.
func (msg *MsgGetBlockInvs) Command() string {
	return CmdGetBlockInvs
}

// MaxPayloadLength returns the maximum length the payload can be for the
// receiver.  This is part of the Message interface implementation.
func (msg *MsgGetBlockInvs) MaxPayloadLength(pver uint32) uint32 {
	// start hash + stop hash.
	return 2 * daghash.HashSize
}

// NewMsgGetBlockInvs returns a new bitcoin getblockinvs message that conforms to the
// Message interface using the passed parameters and defaults for the remaining
// fields.
func NewMsgGetBlockInvs(startHash, stopHash *daghash.Hash) *MsgGetBlockInvs {
	return &MsgGetBlockInvs{
		StartHash: startHash,
		StopHash:  stopHash,
	}
}
