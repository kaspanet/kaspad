package wire

import (
	"github.com/daglabs/btcd/util/daghash"
	"io"
)

type MsgGetBlockLocator struct {
	StartHash *daghash.Hash
	StopHash *daghash.Hash
}

// BtcDecode decodes r using the bitcoin protocol encoding into the receiver.
// This is part of the Message interface implementation.
func (msg *MsgGetBlockLocator) BtcDecode(r io.Reader, pver uint32) error {
	msg.StartHash = &daghash.Hash{}
	err := ReadElement(r, msg.StartHash)
	if err != nil {
		return err
	}

	msg.StopHash = &daghash.Hash{}
	err = ReadElement(r, msg.StopHash)
	if err != nil {
		return err
	}
	return nil
}

// BtcEncode encodes the receiver to w using the bitcoin protocol encoding.
// This is part of the Message interface implementation.
func (msg *MsgGetBlockLocator) BtcEncode(w io.Writer, pver uint32) error {
	err := WriteElement(w, msg.StartHash)
	if err != nil {
		return err
	}

	err = WriteElement(w, msg.StopHash)
	if err != nil {
		return err
	}
	return nil
}

// Command returns the protocol command string for the message.  This is part
// of the Message interface implementation.
func (msg *MsgGetBlockLocator) Command() string {
	return CmdGetBlockLocator
}

// MaxPayloadLength returns the maximum length the payload can be for the
// receiver.  This is part of the Message interface implementation.
func (msg *MsgGetBlockLocator) MaxPayloadLength(pver uint32) uint32 {
	return daghash.HashSize * 2
}

// NewMsgGetBlockLocator returns a new getblocklocator message that conforms to the
// Message interface using the passed parameters and defaults for the remaining
// fields.
func NewMsgGetBlockLocator(startHash, stopHash *daghash.Hash) *MsgGetBlockLocator {
	return &MsgGetBlockLocator{
		StartHash:startHash,
		StopHash:stopHash,
	}
}