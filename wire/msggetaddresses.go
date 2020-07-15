// Copyright (c) 2013-2015 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package wire

import (
	"io"

	"github.com/kaspanet/kaspad/util/subnetworkid"
)

// MsgGetAddresses implements the Message interface and represents a kaspa
// getaddr message. It is used to request a list of known active peers on the
// network from a peer to help identify potential nodes. The list is returned
// via one or more addr messages (MsgAddresses).
//
// This message has no payload.
type MsgGetAddresses struct {
	IncludeAllSubnetworks bool
	SubnetworkID          *subnetworkid.SubnetworkID
}

// KaspaDecode decodes r using the kaspa protocol encoding into the receiver.
// This is part of the Message interface implementation.
func (msg *MsgGetAddresses) KaspaDecode(r io.Reader, pver uint32) error {
	msg.SubnetworkID = nil

	err := ReadElement(r, &msg.IncludeAllSubnetworks)
	if err != nil {
		return err
	}
	if msg.IncludeAllSubnetworks {
		return nil
	}

	var isFullNode bool
	err = ReadElement(r, &isFullNode)
	if err != nil {
		return err
	}
	if isFullNode {
		return nil
	}

	var subnetworkID subnetworkid.SubnetworkID
	err = ReadElement(r, &subnetworkID)
	if err != nil {
		return err
	}
	msg.SubnetworkID = &subnetworkID

	return nil
}

// KaspaEncode encodes the receiver to w using the kaspa protocol encoding.
// This is part of the Message interface implementation.
func (msg *MsgGetAddresses) KaspaEncode(w io.Writer, pver uint32) error {
	err := WriteElement(w, msg.IncludeAllSubnetworks)
	if err != nil {
		return err
	}

	if msg.IncludeAllSubnetworks {
		return nil
	}

	isFullNode := msg.SubnetworkID == nil
	err = WriteElement(w, isFullNode)
	if err != nil {
		return err
	}

	if !isFullNode {
		err = WriteElement(w, msg.SubnetworkID)
		if err != nil {
			return err
		}
	}

	return nil
}

// Command returns the protocol command string for the message. This is part
// of the Message interface implementation.
func (msg *MsgGetAddresses) Command() MessageCommand {
	return CmdGetAddresses
}

// MaxPayloadLength returns the maximum length the payload can be for the
// receiver. This is part of the Message interface implementation.
func (msg *MsgGetAddresses) MaxPayloadLength(pver uint32) uint32 {
	// SubnetworkID length + IncludeAllSubnetworks (1) + isFullNode (1)
	return subnetworkid.IDLength + 2
}

// NewMsgGetAddresses returns a new kaspa getaddr message that conforms to the
// Message interface. See MsgGetAddresses for details.
func NewMsgGetAddresses(includeAllSubnetworks bool, subnetworkID *subnetworkid.SubnetworkID) *MsgGetAddresses {
	return &MsgGetAddresses{
		IncludeAllSubnetworks: includeAllSubnetworks,
		SubnetworkID:          subnetworkID,
	}
}
