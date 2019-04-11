// Copyright (c) 2013-2015 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package wire

import (
	"fmt"
	"io"

	"github.com/daglabs/btcd/util/subnetworkid"
)

// MaxAddrPerMsg is the maximum number of addresses that can be in a single
// bitcoin addr message (MsgAddr).
const MaxAddrPerMsg = 1000

// MsgAddr implements the Message interface and represents a bitcoin
// addr message.  It is used to provide a list of known active peers on the
// network.  An active peer is considered one that has transmitted a message
// within the last 3 hours.  Nodes which have not transmitted in that time
// frame should be forgotten.  Each message is limited to a maximum number of
// addresses, which is currently 1000.  As a result, multiple messages must
// be used to relay the full list.
//
// Use the AddAddress function to build up the list of known addresses when
// sending an addr message to another peer.
type MsgAddr struct {
	IncludeAllSubnetworks bool
	SubnetworkID          *subnetworkid.SubnetworkID
	AddrList              []*NetAddress
}

// AddAddress adds a known active peer to the message.
func (msg *MsgAddr) AddAddress(na *NetAddress) error {
	if len(msg.AddrList)+1 > MaxAddrPerMsg {
		str := fmt.Sprintf("too many addresses in message [max %d]",
			MaxAddrPerMsg)
		return messageError("MsgAddr.AddAddress", str)
	}

	msg.AddrList = append(msg.AddrList, na)
	return nil
}

// AddAddresses adds multiple known active peers to the message.
func (msg *MsgAddr) AddAddresses(netAddrs ...*NetAddress) error {
	for _, na := range netAddrs {
		err := msg.AddAddress(na)
		if err != nil {
			return err
		}
	}
	return nil
}

// ClearAddresses removes all addresses from the message.
func (msg *MsgAddr) ClearAddresses() {
	msg.AddrList = []*NetAddress{}
}

// BtcDecode decodes r using the bitcoin protocol encoding into the receiver.
// This is part of the Message interface implementation.
func (msg *MsgAddr) BtcDecode(r io.Reader, pver uint32) error {
	msg.SubnetworkID = nil

	err := ReadElement(r, &msg.IncludeAllSubnetworks)
	if err != nil {
		return err
	}

	if !msg.IncludeAllSubnetworks {
		var isFullNode bool
		err := ReadElement(r, &isFullNode)
		if err != nil {
			return err
		}
		if !isFullNode {
			var subnetworkID subnetworkid.SubnetworkID
			err = ReadElement(r, &subnetworkID)
			if err != nil {
				return err
			}
			msg.SubnetworkID = &subnetworkID
		}
	}

	// Read addresses array
	count, err := ReadVarInt(r)
	if err != nil {
		return err
	}

	// Limit to max addresses per message.
	if count > MaxAddrPerMsg {
		str := fmt.Sprintf("too many addresses for message "+
			"[count %d, max %d]", count, MaxAddrPerMsg)
		return messageError("MsgAddr.BtcDecode", str)
	}

	addrList := make([]NetAddress, count)
	msg.AddrList = make([]*NetAddress, 0, count)
	for i := uint64(0); i < count; i++ {
		na := &addrList[i]
		err = readNetAddress(r, pver, na, true)
		if err != nil {
			return err
		}
		msg.AddAddress(na)
	}
	return nil
}

// BtcEncode encodes the receiver to w using the bitcoin protocol encoding.
// This is part of the Message interface implementation.
func (msg *MsgAddr) BtcEncode(w io.Writer, pver uint32) error {
	count := len(msg.AddrList)
	if count > MaxAddrPerMsg {
		str := fmt.Sprintf("too many addresses for message "+
			"[count %d, max %d]", count, MaxAddrPerMsg)
		return messageError("MsgAddr.BtcEncode", str)
	}

	err := WriteElement(w, msg.IncludeAllSubnetworks)
	if err != nil {
		return err
	}

	if !msg.IncludeAllSubnetworks {
		// Write subnetwork ID
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
	}

	err = WriteVarInt(w, uint64(count))
	if err != nil {
		return err
	}

	for _, na := range msg.AddrList {
		err = writeNetAddress(w, pver, na, true)
		if err != nil {
			return err
		}
	}

	return nil
}

// Command returns the protocol command string for the message.  This is part
// of the Message interface implementation.
func (msg *MsgAddr) Command() string {
	return CmdAddr
}

// MaxPayloadLength returns the maximum length the payload can be for the
// receiver.  This is part of the Message interface implementation.
func (msg *MsgAddr) MaxPayloadLength(pver uint32) uint32 {
	// IncludeAllSubnetworks flag 1 byte + isFullNode 1 byte + SubnetworkID length + Num addresses (varInt) + max allowed addresses.
	return 1 + 1 + subnetworkid.IDLength + MaxVarIntPayload + (MaxAddrPerMsg * maxNetAddressPayload(pver))
}

// NewMsgAddr returns a new bitcoin addr message that conforms to the
// Message interface.  See MsgAddr for details.
func NewMsgAddr(includeAllSubnetworks bool, subnetworkID *subnetworkid.SubnetworkID) *MsgAddr {
	return &MsgAddr{
		IncludeAllSubnetworks: includeAllSubnetworks,
		SubnetworkID:          subnetworkID,
		AddrList:              make([]*NetAddress, 0, MaxAddrPerMsg),
	}
}
