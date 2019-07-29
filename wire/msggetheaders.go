// Copyright (c) 2013-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package wire

import (
	"io"

	"github.com/daglabs/btcd/util/daghash"
)

// MsgGetHeaders implements the Message interface and represents a bitcoin
// getheaders message.  It is used to request a list of block headers for
// blocks starting after the last known hash in the slice of block locator
// hashes.  The list is returned via a headers message (MsgHeaders) and is
// limited by a specific hash to stop at or the maximum number of block headers
// per message, which is currently 2000.
//
// Set the HashStop field to the hash at which to stop and use
// AddBlockLocatorHash to build up the list of block locator hashes.
//
// The algorithm for building the block locator hashes should be to add the
// hashes in reverse order until you reach the genesis block.  In order to keep
// the list of locator hashes to a resonable number of entries, first add the
// most recent 10 block hashes, then double the step each loop iteration to
// exponentially decrease the number of hashes the further away from head and
// closer to the genesis block you get.
type MsgGetHeaders struct {
	HashStart *daghash.Hash
	HashStop  *daghash.Hash
}

// BtcDecode decodes r using the bitcoin protocol encoding into the receiver.
// This is part of the Message interface implementation.
func (msg *MsgGetHeaders) BtcDecode(r io.Reader, pver uint32) error {
	msg.HashStart = &daghash.Hash{}
	err := ReadElement(r, msg.HashStart)
	if err != nil {
		return err
	}

	msg.HashStop = &daghash.Hash{}
	return ReadElement(r, msg.HashStop)
}

// BtcEncode encodes the receiver to w using the bitcoin protocol encoding.
// This is part of the Message interface implementation.
func (msg *MsgGetHeaders) BtcEncode(w io.Writer, pver uint32) error {
	err := WriteElement(w, msg.HashStart)
	if err != nil {
		return err
	}

	return WriteElement(w, msg.HashStop)
}

// Command returns the protocol command string for the message.  This is part
// of the Message interface implementation.
func (msg *MsgGetHeaders) Command() string {
	return CmdGetHeaders
}

// MaxPayloadLength returns the maximum length the payload can be for the
// receiver.  This is part of the Message interface implementation.
func (msg *MsgGetHeaders) MaxPayloadLength(pver uint32) uint32 {
	// start hash + hash stop.
	return 2 * daghash.HashSize
}

// NewMsgGetHeaders returns a new bitcoin getheaders message that conforms to
// the Message interface.  See MsgGetHeaders for details.
func NewMsgGetHeaders(hashStart, hashStop *daghash.Hash) *MsgGetHeaders {
	return &MsgGetHeaders{
		HashStart: hashStart,
		HashStop:  hashStop,
	}
}
