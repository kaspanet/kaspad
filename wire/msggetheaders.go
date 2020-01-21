// Copyright (c) 2013-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package wire

import (
	"io"

	"github.com/kaspanet/kaspad/util/daghash"
)

// MsgGetHeaders implements the Message interface and represents a kaspa
// getheaders message. It is used to request a list of block headers for
// blocks starting after the last known hash in the slice of block locator
// hashes. The list is returned via a headers message (MsgHeaders) and is
// limited by a specific hash to stop at or the maximum number of block headers
// per message, which is currently 2000.
//
// Set the HighHash field to the hash at which to stop and use
// AddBlockLocatorHash to build up the list of block locator hashes.
//
// The algorithm for building the block locator hashes should be to add the
// hashes in reverse order until you reach the genesis block. In order to keep
// the list of locator hashes to a resonable number of entries, first add the
// most recent 10 block hashes, then double the step each loop iteration to
// exponentially decrease the number of hashes the further away from head and
// closer to the genesis block you get.
type MsgGetHeaders struct {
	StartHash *daghash.Hash
	StopHash  *daghash.Hash
}

// KaspaDecode decodes r using the kaspa protocol encoding into the receiver.
// This is part of the Message interface implementation.
func (msg *MsgGetHeaders) KaspaDecode(r io.Reader, pver uint32) error {
	msg.StartHash = &daghash.Hash{}
	err := ReadElement(r, msg.StartHash)
	if err != nil {
		return err
	}

	msg.StopHash = &daghash.Hash{}
	return ReadElement(r, msg.StopHash)
}

// KaspaEncode encodes the receiver to w using the kaspa protocol encoding.
// This is part of the Message interface implementation.
func (msg *MsgGetHeaders) KaspaEncode(w io.Writer, pver uint32) error {
	err := WriteElement(w, msg.StartHash)
	if err != nil {
		return err
	}

	return WriteElement(w, msg.StopHash)
}

// Command returns the protocol command string for the message. This is part
// of the Message interface implementation.
func (msg *MsgGetHeaders) Command() string {
	return CmdGetHeaders
}

// MaxPayloadLength returns the maximum length the payload can be for the
// receiver. This is part of the Message interface implementation.
func (msg *MsgGetHeaders) MaxPayloadLength(pver uint32) uint32 {
	// start hash + stop hash.
	return 2 * daghash.HashSize
}

// NewMsgGetHeaders returns a new kaspa getheaders message that conforms to
// the Message interface. See MsgGetHeaders for details.
func NewMsgGetHeaders(startHash, stopHash *daghash.Hash) *MsgGetHeaders {
	return &MsgGetHeaders{
		StartHash: startHash,
		StopHash:  stopHash,
	}
}
