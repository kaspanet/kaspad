// Copyright (c) 2013-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package wire

import (
	"io"

	"github.com/daglabs/btcd/util/daghash"
)

// MsgGetBlocks implements the Message interface and represents a bitcoin
// getblocks message.  It is used to request a list of blocks starting after the
// last known hash in the slice of block locator hashes.  The list is returned
// via an inv message (MsgInv) and is limited by a specific hash to stop at or
// the maximum number of blocks per message, which is currently 500.
//
// Set the HashStop field to the hash at which to stop and use
// AddBlockLocatorHash to build up the list of block locator hashes.
//
// The algorithm for building the block locator hashes should be to add the
// hashes in reverse order until you reach the genesis block.  In order to keep
// the list of locator hashes to a reasonable number of entries, first add the
// most recent 10 block hashes, then double the step each loop iteration to
// exponentially decrease the number of hashes the further away from head and
// closer to the genesis block you get.
type MsgGetBlocks struct {
	HashStart       *daghash.Hash
	HashStop        *daghash.Hash
}

// BtcDecode decodes r using the bitcoin protocol encoding into the receiver.
// This is part of the Message interface implementation.
func (msg *MsgGetBlocks) BtcDecode(r io.Reader, pver uint32) error {
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
func (msg *MsgGetBlocks) BtcEncode(w io.Writer, pver uint32) error {
	err := WriteElement(w, msg.HashStart)
	if err != nil {
		return err
	}

	return WriteElement(w, msg.HashStop)
}

// Command returns the protocol command string for the message.  This is part
// of the Message interface implementation.
func (msg *MsgGetBlocks) Command() string {
	return CmdGetBlocks
}

// MaxPayloadLength returns the maximum length the payload can be for the
// receiver.  This is part of the Message interface implementation.
func (msg *MsgGetBlocks) MaxPayloadLength(pver uint32) uint32 {
	// start hash + hash stop.
	return 2 * daghash.HashSize
}

// NewMsgGetBlocks returns a new bitcoin getblocks message that conforms to the
// Message interface using the passed parameters and defaults for the remaining
// fields.
func NewMsgGetBlocks(hashStart, hashStop *daghash.Hash) *MsgGetBlocks {
	return &MsgGetBlocks{
		HashStart:       hashStart,
		HashStop:        hashStop,
	}
}
