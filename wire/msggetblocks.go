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

// MsgGetBlockInvs implements the Message interface and represents a bitcoin
// getblockinvss message.  It is used to request a list of blocks starting after
// the last known hash in the slice of block locator hashes.  The list is
// returned via an inv message (MsgInv) and is limited by a specific hash to
// stop at or the maximum number of blocks per message, which is currently 500.
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
type MsgGetBlockInvs struct {
	ProtocolVersion    uint32
	BlockLocatorHashes []*daghash.Hash
	HashStop           *daghash.Hash
}

// AddBlockLocatorHash adds a new block locator hash to the message.
func (msg *MsgGetBlockInvs) AddBlockLocatorHash(hash *daghash.Hash) error {
	if len(msg.BlockLocatorHashes)+1 > MaxBlockLocatorsPerMsg {
		str := fmt.Sprintf("too many block locator hashes for message [max %d]",
			MaxBlockLocatorsPerMsg)
		return messageError("MsgGetBlockInvs.AddBlockLocatorHash", str)
	}

	msg.BlockLocatorHashes = append(msg.BlockLocatorHashes, hash)
	return nil
}

// BtcDecode decodes r using the bitcoin protocol encoding into the receiver.
// This is part of the Message interface implementation.
func (msg *MsgGetBlockInvs) BtcDecode(r io.Reader, pver uint32) error {
	err := ReadElement(r, &msg.ProtocolVersion)
	if err != nil {
		return err
	}

	// Read num block locator hashes and limit to max.
	count, err := ReadVarInt(r)
	if err != nil {
		return err
	}
	if count > MaxBlockLocatorsPerMsg {
		str := fmt.Sprintf("too many block locator hashes for message "+
			"[count %d, max %d]", count, MaxBlockLocatorsPerMsg)
		return messageError("MsgGetBlockInvs.BtcDecode", str)
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
		msg.AddBlockLocatorHash(hash)
	}

	msg.HashStop = &daghash.Hash{}
	return ReadElement(r, msg.HashStop)
}

// BtcEncode encodes the receiver to w using the bitcoin protocol encoding.
// This is part of the Message interface implementation.
func (msg *MsgGetBlockInvs) BtcEncode(w io.Writer, pver uint32) error {
	count := len(msg.BlockLocatorHashes)
	if count > MaxBlockLocatorsPerMsg {
		str := fmt.Sprintf("too many block locator hashes for message "+
			"[count %d, max %d]", count, MaxBlockLocatorsPerMsg)
		return messageError("MsgGetBlockInvs.BtcEncode", str)
	}

	err := WriteElement(w, msg.ProtocolVersion)
	if err != nil {
		return err
	}

	err = WriteVarInt(w, uint64(count))
	if err != nil {
		return err
	}

	for _, hash := range msg.BlockLocatorHashes {
		err = WriteElement(w, hash)
		if err != nil {
			return err
		}
	}

	return WriteElement(w, msg.HashStop)
}

// Command returns the protocol command string for the message.  This is part
// of the Message interface implementation.
func (msg *MsgGetBlockInvs) Command() string {
	return CmdGetBlockInvs
}

// MaxPayloadLength returns the maximum length the payload can be for the
// receiver.  This is part of the Message interface implementation.
func (msg *MsgGetBlockInvs) MaxPayloadLength(pver uint32) uint32 {
	// Protocol version 4 bytes + num hashes (varInt) + max block locator
	// hashes + hash stop.
	return 4 + MaxVarIntPayload + (MaxBlockLocatorsPerMsg * daghash.HashSize) + daghash.HashSize
}

// NewMsgGetBlockInvs returns a new bitcoin getblockinvs message that conforms
// to the Message interface using the passed parameters and defaults for the
// remaining fields.
func NewMsgGetBlockInvs(hashStop *daghash.Hash) *MsgGetBlockInvs {
	return &MsgGetBlockInvs{
		ProtocolVersion:    ProtocolVersion,
		BlockLocatorHashes: make([]*daghash.Hash, 0, MaxBlockLocatorsPerMsg),
		HashStop:           hashStop,
	}
}
