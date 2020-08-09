// Copyright (c) 2013-2015 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package wire

import (
	"github.com/kaspanet/kaspad/util/subnetworkid"
)

// MsgRequestAddresses implements the Message interface and represents a kaspa
// RequestAddresses message. It is used to request a list of known active peers on the
// network from a peer to help identify potential nodes. The list is returned
// via one or more addr messages (MsgAddresses).
//
// This message has no payload.
type MsgRequestAddresses struct {
	IncludeAllSubnetworks bool
	SubnetworkID          *subnetworkid.SubnetworkID
}

// Command returns the protocol command string for the message. This is part
// of the Message interface implementation.
func (msg *MsgRequestAddresses) Command() MessageCommand {
	return CmdRequestAddresses
}

// NewMsgRequestAddresses returns a new kaspa RequestAddresses message that conforms to the
// Message interface. See MsgRequestAddresses for details.
func NewMsgRequestAddresses(includeAllSubnetworks bool, subnetworkID *subnetworkid.SubnetworkID) *MsgRequestAddresses {
	return &MsgRequestAddresses{
		IncludeAllSubnetworks: includeAllSubnetworks,
		SubnetworkID:          subnetworkID,
	}
}
