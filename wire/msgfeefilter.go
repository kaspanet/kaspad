// Copyright (c) 2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package wire

import (
	"io"
)

// MsgFeeFilter implements the Message interface and represents a kaspa
// feefilter message. It is used to request the receiving peer does not
// announce any transactions below the specified minimum fee rate.
//
// This message was not added until protocol versions starting with
// FeeFilterVersion.
type MsgFeeFilter struct {
	MinFee int64
}

// KaspaDecode decodes r using the kaspa protocol encoding into the receiver.
// This is part of the Message interface implementation.
func (msg *MsgFeeFilter) KaspaDecode(r io.Reader, pver uint32) error {
	return ReadElement(r, &msg.MinFee)
}

// KaspaEncode encodes the receiver to w using the kaspa protocol encoding.
// This is part of the Message interface implementation.
func (msg *MsgFeeFilter) KaspaEncode(w io.Writer, pver uint32) error {
	return WriteElement(w, msg.MinFee)
}

// Command returns the protocol command string for the message. This is part
// of the Message interface implementation.
func (msg *MsgFeeFilter) Command() string {
	return CmdFeeFilter
}

// MaxPayloadLength returns the maximum length the payload can be for the
// receiver. This is part of the Message interface implementation.
func (msg *MsgFeeFilter) MaxPayloadLength(pver uint32) uint32 {
	return 8
}

// NewMsgFeeFilter returns a new kaspa feefilter message that conforms to
// the Message interface. See MsgFeeFilter for details.
func NewMsgFeeFilter(minfee int64) *MsgFeeFilter {
	return &MsgFeeFilter{
		MinFee: minfee,
	}
}
