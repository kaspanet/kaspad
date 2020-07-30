package wire

import (
	"io"

	"github.com/kaspanet/kaspad/util/daghash"
)

// MsgGetRelayBlocksHashes is the maximum number of hashes that can
// be in a single RequestRelayBlocks message.
const MsgGetRelayBlocksHashes = MaxInvPerMsg

// MsgRequestRelayBlocks implements the Message interface and represents a kaspa
// RequestRelayBlocks message. It is used to request blocks as part of the block
// relay protocol.
type MsgRequestRelayBlocks struct {
	Hashes []*daghash.Hash
}

// KaspaDecode decodes r using the kaspa protocol encoding into the receiver.
// This is part of the Message interface implementation.
func (msg *MsgRequestRelayBlocks) KaspaDecode(r io.Reader, pver uint32) error {
	numHashes, err := ReadVarInt(r)
	if err != nil {
		return err
	}

	msg.Hashes = make([]*daghash.Hash, numHashes)
	for i := uint64(0); i < numHashes; i++ {
		msg.Hashes[i] = &daghash.Hash{}
		err := ReadElement(r, msg.Hashes[i])
		if err != nil {
			return err
		}
	}

	return nil
}

// KaspaEncode encodes the receiver to w using the kaspa protocol encoding.
// This is part of the Message interface implementation.
func (msg *MsgRequestRelayBlocks) KaspaEncode(w io.Writer, pver uint32) error {
	err := WriteVarInt(w, uint64(len(msg.Hashes)))
	if err != nil {
		return err
	}
	for _, hash := range msg.Hashes {
		err := WriteElement(w, hash)
		if err != nil {
			return err
		}
	}

	return nil
}

// Command returns the protocol command string for the message. This is part
// of the Message interface implementation.
func (msg *MsgRequestRelayBlocks) Command() MessageCommand {
	return CmdRequestRelayBlocks
}

// MaxPayloadLength returns the maximum length the payload can be for the
// receiver. This is part of the Message interface implementation.
func (msg *MsgRequestRelayBlocks) MaxPayloadLength(pver uint32) uint32 {
	return daghash.HashSize*MsgGetRelayBlocksHashes + uint32(VarIntSerializeSize(MsgGetRelayBlocksHashes))
}

// NewMsgGetRelayBlocks returns a new kaspa RequestRelayBlocks message that conforms to
// the Message interface. See MsgRequestRelayBlocks for details.
func NewMsgGetRelayBlocks(hashes []*daghash.Hash) *MsgRequestRelayBlocks {
	return &MsgRequestRelayBlocks{
		Hashes: hashes,
	}
}
