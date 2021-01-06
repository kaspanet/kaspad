// Copyright (c) 2013-2015 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package appmessage

// MaxAddressesPerMsg is the maximum number of addresses that can be in a single
// kaspa Addresses message (MsgAddresses).
const MaxAddressesPerMsg = 1000

// MsgAddresses implements the Message interface and represents a kaspa
// Addresses message.
type MsgAddresses struct {
	baseMessage
	AddressList []*NetAddress
}

// Command returns the protocol command string for the message. This is part
// of the Message interface implementation.
func (msg *MsgAddresses) Command() MessageCommand {
	return CmdAddresses
}

// NewMsgAddresses returns a new kaspa Addresses message that conforms to the
// Message interface. See MsgAddresses for details.
func NewMsgAddresses(addressList []*NetAddress) *MsgAddresses {
	return &MsgAddresses{
		AddressList: addressList,
	}
}
