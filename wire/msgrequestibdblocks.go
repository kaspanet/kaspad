// Copyright (c) 2013-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package wire

import (
	"io"

	"github.com/kaspanet/kaspad/util/daghash"
)

// MsgRequestIBDBlocks implements the Message interface and represents a kaspa
// RequestIBDBlocks message. It is used to request a list of blocks starting after the
// low hash and until the high hash.
type MsgRequestIBDBlocks struct {
	LowHash  *daghash.Hash
	HighHash *daghash.Hash
}

// KaspaDecode decodes r using the kaspa protocol encoding into the receiver.
// This is part of the Message interface implementation.
func (msg *MsgRequestIBDBlocks) KaspaDecode(r io.Reader, pver uint32) error {
	msg.LowHash = &daghash.Hash{}
	err := ReadElement(r, msg.LowHash)
	if err != nil {
		return err
	}

	msg.HighHash = &daghash.Hash{}
	return ReadElement(r, msg.HighHash)
}

// KaspaEncode encodes the receiver to w using the kaspa protocol encoding.
// This is part of the Message interface implementation.
func (msg *MsgRequestIBDBlocks) KaspaEncode(w io.Writer, pver uint32) error {
	err := WriteElement(w, msg.LowHash)
	if err != nil {
		return err
	}

	return WriteElement(w, msg.HighHash)
}

// Command returns the protocol command string for the message. This is part
// of the Message interface implementation.
func (msg *MsgRequestIBDBlocks) Command() MessageCommand {
	return CmdRequestIBDBlocks
}

// MaxPayloadLength returns the maximum length the payload can be for the
// receiver. This is part of the Message interface implementation.
func (msg *MsgRequestIBDBlocks) MaxPayloadLength(pver uint32) uint32 {
	// low hash + high hash.
	return 2 * daghash.HashSize
}

// NewMsgGetBlocks returns a new kaspa RequestIBDBlocks message that conforms to the
// Message interface using the passed parameters and defaults for the remaining
// fields.
func NewMsgGetBlocks(lowHash, highHash *daghash.Hash) *MsgRequestIBDBlocks {
	return &MsgRequestIBDBlocks{
		LowHash:  lowHash,
		HighHash: highHash,
	}
}
