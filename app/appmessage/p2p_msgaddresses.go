// Copyright (c) 2013-2015 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package appmessage

import (
	"fmt"

	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// MaxAddressesPerMsg is the maximum number of addresses that can be in a single
// kaspa Addresses message (MsgAddresses).
const MaxAddressesPerMsg = 1000

// MsgAddresses implements the Message interface and represents a kaspa
// Addresses message. It is used to provide a list of known active peers on the
// network. An active peer is considered one that has transmitted a message
// within the last 3 hours. Nodes which have not transmitted in that time
// frame should be forgotten. Each message is limited to a maximum number of
// addresses, which is currently 1000. As a result, multiple messages must
// be used to relay the full list.
//
// Use the AddAddress function to build up the list of known addresses when
// sending an Addresses message to another peer.
type MsgAddresses struct {
	baseMessage
	IncludeAllSubnetworks bool
	SubnetworkID          *externalapi.DomainSubnetworkID
	AddrList              []*NetAddress
}

// AddAddress adds a known active peer to the message.
func (msg *MsgAddresses) AddAddress(na *NetAddress) error {
	if len(msg.AddrList)+1 > MaxAddressesPerMsg {
		str := fmt.Sprintf("too many addresses in message [max %d]",
			MaxAddressesPerMsg)
		return messageError("MsgAddresses.AddAddress", str)
	}

	msg.AddrList = append(msg.AddrList, na)
	return nil
}

// AddAddresses adds multiple known active peers to the message.
func (msg *MsgAddresses) AddAddresses(netAddrs ...*NetAddress) error {
	for _, na := range netAddrs {
		err := msg.AddAddress(na)
		if err != nil {
			return err
		}
	}
	return nil
}

// ClearAddresses removes all addresses from the message.
func (msg *MsgAddresses) ClearAddresses() {
	msg.AddrList = []*NetAddress{}
}

// Command returns the protocol command string for the message. This is part
// of the Message interface implementation.
func (msg *MsgAddresses) Command() MessageCommand {
	return CmdAddresses
}

// NewMsgAddresses returns a new kaspa Addresses message that conforms to the
// Message interface. See MsgAddresses for details.
func NewMsgAddresses(includeAllSubnetworks bool, subnetworkID *externalapi.DomainSubnetworkID) *MsgAddresses {
	return &MsgAddresses{
		IncludeAllSubnetworks: includeAllSubnetworks,
		SubnetworkID:          subnetworkID,
		AddrList:              make([]*NetAddress, 0, MaxAddressesPerMsg),
	}
}
