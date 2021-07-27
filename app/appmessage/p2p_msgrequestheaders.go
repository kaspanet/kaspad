// Copyright (c) 2013-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package appmessage

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// MsgRequestHeaders implements the Message interface and represents a kaspa
// RequestHeaders message. It is used to request a list of blocks starting after the
// low hash and until the high hash.
type MsgRequestHeaders struct {
	baseMessage
	LowHash  *externalapi.DomainHash
	HighHash *externalapi.DomainHash
}

// Command returns the protocol command string for the message. This is part
// of the Message interface implementation.
func (msg *MsgRequestHeaders) Command() MessageCommand {
	return CmdRequestHeaders
}

// NewMsgRequstHeaders returns a new kaspa RequestHeaders message that conforms to the
// Message interface using the passed parameters and defaults for the remaining
// fields.
func NewMsgRequstHeaders(lowHash, highHash *externalapi.DomainHash) *MsgRequestHeaders {
	return &MsgRequestHeaders{
		LowHash:  lowHash,
		HighHash: highHash,
	}
}
