// Copyright (c) 2013-2015 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package wire

import (
	"bytes"
	"fmt"
	"io"

	"github.com/daglabs/btcd/util/subnetworkid"
)

// MsgGetAddr implements the Message interface and represents a bitcoin
// getaddr message.  It is used to request a list of known active peers on the
// network from a peer to help identify potential nodes.  The list is returned
// via one or more addr messages (MsgAddr).
//
// This message has no payload.
type MsgGetAddr struct {
	SubnetworkID *subnetworkid.SubnetworkID
}

// BtcDecode decodes r using the bitcoin protocol encoding into the receiver.
// This is part of the Message interface implementation.
func (msg *MsgGetAddr) BtcDecode(r io.Reader, pver uint32) error {
	var allSubnetworks bool
	err := readElement(r, &allSubnetworks)
	if err != nil {
		return err
	}
	if allSubnetworks {
		msg.SubnetworkID = nil
	} else {
		buf, ok := r.(*bytes.Buffer)
		if !ok {
			return fmt.Errorf("MsgAddr.BtcDecode reader is not a *bytes.Buffer")
		}
		var subnetworkID subnetworkid.SubnetworkID
		err = readElement(buf, &subnetworkID)
		if err != nil {
			return err
		}
		msg.SubnetworkID = &subnetworkID
	}

	return nil
}

// BtcEncode encodes the receiver to w using the bitcoin protocol encoding.
// This is part of the Message interface implementation.
func (msg *MsgGetAddr) BtcEncode(w io.Writer, pver uint32) error {
	allSubnetworks := msg.SubnetworkID == nil
	err := writeElement(w, allSubnetworks)
	if err != nil {
		return err
	}
	if !allSubnetworks {
		err = writeElement(w, msg.SubnetworkID)
		if err != nil {
			return err
		}
	}
	return nil
}

// Command returns the protocol command string for the message.  This is part
// of the Message interface implementation.
func (msg *MsgGetAddr) Command() string {
	return CmdGetAddr
}

// MaxPayloadLength returns the maximum length the payload can be for the
// receiver.  This is part of the Message interface implementation.
func (msg *MsgGetAddr) MaxPayloadLength(pver uint32) uint32 {
	return subnetworkid.IDLength + 1
}

// NewMsgGetAddr returns a new bitcoin getaddr message that conforms to the
// Message interface.  See MsgGetAddr for details.
func NewMsgGetAddr(subnetworkID *subnetworkid.SubnetworkID) *MsgGetAddr {
	return &MsgGetAddr{
		SubnetworkID: subnetworkID,
	}
}
