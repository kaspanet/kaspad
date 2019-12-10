// Copyright (c) 2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package wire

import (
	"fmt"
	"io"

	"github.com/kaspanet/kaspad/util/daghash"
)

// FilterType is used to represent a filter type.
type FilterType uint8

const (
	// GCSFilterRegular is the regular filter type.
	GCSFilterRegular FilterType = iota

	// GCSFilterExtended is the extended filter type.
	GCSFilterExtended
)

const (
	// MaxCFilterDataSize is the maximum byte size of a committed filter.
	// The maximum size is currently defined as 256KiB.
	MaxCFilterDataSize = 256 * 1024
)

// MsgCFilter implements the Message interface and represents a kaspa cfilter
// message. It is used to deliver a committed filter in response to a
// getcfilters (MsgGetCFilters) message.
type MsgCFilter struct {
	FilterType FilterType
	BlockHash  daghash.Hash
	Data       []byte
}

// KaspaDecode decodes r using the kaspa protocol encoding into the receiver.
// This is part of the Message interface implementation.
func (msg *MsgCFilter) KaspaDecode(r io.Reader, pver uint32) error {
	// Read filter type
	err := ReadElement(r, &msg.FilterType)
	if err != nil {
		return err
	}

	// Read the hash of the filter's block
	err = ReadElement(r, &msg.BlockHash)
	if err != nil {
		return err
	}

	// Read filter data
	msg.Data, err = ReadVarBytes(r, pver, MaxCFilterDataSize,
		"cfilter data")
	return err
}

// KaspaEncode encodes the receiver to w using the kaspa protocol encoding.
// This is part of the Message interface implementation.
func (msg *MsgCFilter) KaspaEncode(w io.Writer, pver uint32) error {
	size := len(msg.Data)
	if size > MaxCFilterDataSize {
		str := fmt.Sprintf("cfilter size too large for message "+
			"[size %d, max %d]", size, MaxCFilterDataSize)
		return messageError("MsgCFilter.KaspaEncode", str)
	}

	err := WriteElement(w, msg.FilterType)
	if err != nil {
		return err
	}

	err = WriteElement(w, msg.BlockHash)
	if err != nil {
		return err
	}

	return WriteVarBytes(w, pver, msg.Data)
}

// Deserialize decodes a filter from r into the receiver using a format that is
// suitable for long-term storage such as a database. This function differs
// from KaspaDecode in that KaspaDecode decodes from the kaspa wire protocol as
// it was sent across the network. The wire encoding can technically differ
// depending on the protocol version and doesn't even really need to match the
// format of a stored filter at all. As of the time this comment was written,
// the encoded filter is the same in both instances, but there is a distinct
// difference and separating the two allows the API to be flexible enough to
// deal with changes.
func (msg *MsgCFilter) Deserialize(r io.Reader) error {
	// At the current time, there is no difference between the wire encoding
	// and the stable long-term storage format. As a result, make use of
	// KaspaDecode.
	return msg.KaspaDecode(r, 0)
}

// Command returns the protocol command string for the message. This is part
// of the Message interface implementation.
func (msg *MsgCFilter) Command() string {
	return CmdCFilter
}

// MaxPayloadLength returns the maximum length the payload can be for the
// receiver. This is part of the Message interface implementation.
func (msg *MsgCFilter) MaxPayloadLength(pver uint32) uint32 {
	return uint32(VarIntSerializeSize(MaxCFilterDataSize)) +
		MaxCFilterDataSize + daghash.HashSize + 1
}

// NewMsgCFilter returns a new kaspa cfilter message that conforms to the
// Message interface. See MsgCFilter for details.
func NewMsgCFilter(filterType FilterType, blockHash *daghash.Hash,
	data []byte) *MsgCFilter {
	return &MsgCFilter{
		FilterType: filterType,
		BlockHash:  *blockHash,
		Data:       data,
	}
}
