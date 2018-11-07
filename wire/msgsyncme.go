package wire

import (
	"bytes"
	"fmt"
	"io"
	"math"

	"github.com/daglabs/btcd/dagconfig/daghash"
)

// MsgSyncMe implements the Message interface and represents a bitcoin version
// message.  It is used in order to initialize a sync session from a peer with
// unknown tip hashes.
type MsgSyncMe struct {
	ProtocolVersion     uint32
	PartialSelectedPath []*daghash.Hash
}

// BtcDecode decodes r using the bitcoin protocol encoding into the receiver.
// This is part of the Message interface implementation.
func (msg *MsgSyncMe) BtcDecode(r io.Reader, pver uint32) error {
	buf, ok := r.(*bytes.Buffer)
	if !ok {
		return fmt.Errorf("MsgSyncMe.BtcDecode reader is not a " +
			"*bytes.Buffer")
	}

	err := readElement(buf, &msg.ProtocolVersion)
	if err != nil {
		return err
	}

	var numHashes uint16
	err = readElements(r, &numHashes)
	if err != nil {
		return err
	}
	msg.PartialSelectedPath = make([]*daghash.Hash, numHashes)
	for i := uint16(0); i < numHashes; i++ {
		err := readElement(r, &msg.PartialSelectedPath[i])
		if err != nil {
			return err
		}
	}

	return nil
}

// BtcEncode encodes the receiver to w using the bitcoin protocol encoding.
// This is part of the Message interface implementation.
func (msg *MsgSyncMe) BtcEncode(w io.Writer, pver uint32) error {
	err := writeElement(w, msg.ProtocolVersion)
	if err != nil {
		return err
	}

	err = writeElements(w, uint16(len(msg.PartialSelectedPath)), &msg.PartialSelectedPath)
	if err != nil {
		return err
	}
	return nil
}

// Command returns the protocol command string for the message.  This is part
// of the Message interface implementation.
func (msg *MsgSyncMe) Command() string {
	return CmdSyncMe
}

// MaxPayloadLength returns the maximum length the payload can be for the
// receiver.  This is part of the Message interface implementation.
func (msg *MsgSyncMe) MaxPayloadLength(pver uint32) uint32 {
	// Protocol version 4 bytes + number of hashes 2 bytes + max uint16 * daghash.HashSize
	return 6 + math.MaxUint16*daghash.HashSize
}

// NewMsgSyncMe returns a new bitcoin version message that conforms to the
// Message interface using the passed parameters and defaults for the remaining
// fields.
func NewMsgSyncMe(PartialSelectedPath []*daghash.Hash) *MsgSyncMe {
	return &MsgSyncMe{
		ProtocolVersion:     ProtocolVersion,
		PartialSelectedPath: PartialSelectedPath,
	}
}
