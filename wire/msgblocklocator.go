// Copyright (c) 2013-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package wire

import (
	"fmt"
	"io"

	"github.com/daglabs/btcd/util/daghash"
)

// MaxBlockLocatorsPerMsg is the maximum number of block locator hashes allowed
// per message.
const MaxBlockLocatorsPerMsg = 500

// MsgBlockLocator implements the Message interface and represents a bitcoin
// blocklocator message.  It is used to request a list of block headers for
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
type MsgBlockLocator struct {
	BlockLocatorHashes []*daghash.Hash
	HashStop           *daghash.Hash
}

// AddBlockLocatorHash adds a new block locator hash to the message.
func (msg *MsgBlockLocator) AddBlockLocatorHash(hash *daghash.Hash) error {
	if len(msg.BlockLocatorHashes)+1 > MaxBlockLocatorsPerMsg {
		str := fmt.Sprintf("too many block locator hashes for message [max %d]",
			MaxBlockLocatorsPerMsg)
		return messageError("MsgBlockLocator.AddBlockLocatorHash", str)
	}

	msg.BlockLocatorHashes = append(msg.BlockLocatorHashes, hash)
	return nil
}

// BtcDecode decodes r using the bitcoin protocol encoding into the receiver.
// This is part of the Message interface implementation.
func (msg *MsgBlockLocator) BtcDecode(r io.Reader, pver uint32) error {

	// Read num block locator hashes and limit to max.
	count, err := ReadVarInt(r)
	if err != nil {
		return err
	}
	if count > MaxBlockLocatorsPerMsg {
		str := fmt.Sprintf("too many block locator hashes for message "+
			"[count %d, max %d]", count, MaxBlockLocatorsPerMsg)
		return messageError("MsgBlockLocator.BtcDecode", str)
	}

	// Create a contiguous slice of hashes to deserialize into in order to
	// reduce the number of allocations.
	locatorHashes := make([]daghash.Hash, count)
	msg.BlockLocatorHashes = make([]*daghash.Hash, 0, count)
	for i := uint64(0); i < count; i++ {
		hash := &locatorHashes[i]
		err := ReadElement(r, hash)
		if err != nil {
			return err
		}
		err = msg.AddBlockLocatorHash(hash)
		if err != nil {
			return err
		}
	}

	msg.HashStop = &daghash.Hash{}
	return ReadElement(r, msg.HashStop)
}

// BtcEncode encodes the receiver to w using the bitcoin protocol encoding.
// This is part of the Message interface implementation.
func (msg *MsgBlockLocator) BtcEncode(w io.Writer, pver uint32) error {
	// Limit to max block locator hashes per message.
	count := len(msg.BlockLocatorHashes)
	if count > MaxBlockLocatorsPerMsg {
		str := fmt.Sprintf("too many block locator hashes for message "+
			"[count %d, max %d]", count, MaxBlockLocatorsPerMsg)
		return messageError("MsgBlockLocator.BtcEncode", str)
	}

	err := WriteVarInt(w, uint64(count))
	if err != nil {
		return err
	}

	for _, hash := range msg.BlockLocatorHashes {
		err := WriteElement(w, hash)
		if err != nil {
			return err
		}
	}

	return WriteElement(w, msg.HashStop)
}

// Command returns the protocol command string for the message.  This is part
// of the Message interface implementation.
func (msg *MsgBlockLocator) Command() string {
	return CmdBlockLocator
}

// MaxPayloadLength returns the maximum length the payload can be for the
// receiver.  This is part of the Message interface implementation.
func (msg *MsgBlockLocator) MaxPayloadLength(pver uint32) uint32 {
	// Num block locator hashes (varInt) + max allowed block
	// locators + hash stop.
	return MaxVarIntPayload + (MaxBlockLocatorsPerMsg *
		daghash.HashSize) + daghash.HashSize
}

// NewMsgBlockLocator returns a new bitcoin getheaders message that conforms to
// the Message interface.  See MsgBlockLocator for details.
func NewMsgBlockLocator() *MsgBlockLocator {
	return &MsgBlockLocator{
		BlockLocatorHashes: make([]*daghash.Hash, 0,
			MaxBlockLocatorsPerMsg),
		HashStop: &daghash.ZeroHash,
	}
}
