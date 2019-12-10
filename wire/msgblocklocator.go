package wire

import (
	"fmt"
	"io"

	"github.com/kaspanet/kaspad/util/daghash"
)

// MaxBlockLocatorsPerMsg is the maximum number of block locator hashes allowed
// per message.
const MaxBlockLocatorsPerMsg = 500

// MsgBlockLocator implements the Message interface and represents a kaspa
// locator message. It is used to find the highest known chain block with
// a peer that is syncing with you.
type MsgBlockLocator struct {
	BlockLocatorHashes []*daghash.Hash
}

// AddBlockLocatorHash adds a new block locator hash to the message.
func (msg *MsgBlockLocator) AddBlockLocatorHash(hash *daghash.Hash) error {
	if len(msg.BlockLocatorHashes) >= MaxBlockLocatorsPerMsg {
		str := fmt.Sprintf("too many block locator hashes for message [max %d]",
			MaxBlockLocatorsPerMsg)
		return messageError("MsgBlockLocator.AddBlockLocatorHash", str)
	}

	msg.BlockLocatorHashes = append(msg.BlockLocatorHashes, hash)
	return nil
}

// KaspaDecode decodes r using the kaspa protocol encoding into the receiver.
// This is part of the Message interface implementation.
func (msg *MsgBlockLocator) KaspaDecode(r io.Reader, pver uint32) error {
	// Read num block locator hashes and limit to max.
	count, err := ReadVarInt(r)
	if err != nil {
		return err
	}
	if count > MaxBlockLocatorsPerMsg {
		str := fmt.Sprintf("too many block locator hashes for message "+
			"[count %d, max %d]", count, MaxBlockLocatorsPerMsg)
		return messageError("MsgBlockLocator.KaspaDecode", str)
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
	return nil
}

// KaspaEncode encodes the receiver to w using the kaspa protocol encoding.
// This is part of the Message interface implementation.
func (msg *MsgBlockLocator) KaspaEncode(w io.Writer, pver uint32) error {
	// Limit to max block locator hashes per message.
	count := len(msg.BlockLocatorHashes)
	if count > MaxBlockLocatorsPerMsg {
		str := fmt.Sprintf("too many block locator hashes for message "+
			"[count %d, max %d]", count, MaxBlockLocatorsPerMsg)
		return messageError("MsgBlockLocator.KaspaEncode", str)
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
	return nil
}

// Command returns the protocol command string for the message. This is part
// of the Message interface implementation.
func (msg *MsgBlockLocator) Command() string {
	return CmdBlockLocator
}

// MaxPayloadLength returns the maximum length the payload can be for the
// receiver. This is part of the Message interface implementation.
func (msg *MsgBlockLocator) MaxPayloadLength(pver uint32) uint32 {
	// Num block locator hashes (varInt) + max allowed block
	// locators.
	return MaxVarIntPayload + (MaxBlockLocatorsPerMsg *
		daghash.HashSize)
}

// NewMsgBlockLocator returns a new kaspa locator message that conforms to
// the Message interface. See MsgBlockLocator for details.
func NewMsgBlockLocator() *MsgBlockLocator {
	return &MsgBlockLocator{
		BlockLocatorHashes: make([]*daghash.Hash, 0,
			MaxBlockLocatorsPerMsg),
	}
}
