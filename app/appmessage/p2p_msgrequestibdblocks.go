// Copyright (c) 2013-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package appmessage

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// MsgRequestIBDBlocks implements the Message interface and represents a kaspa
// RequestIBDBlocks message. It is used to request a list of blocks starting after the
// low hash and until the high hash.
type MsgRequestIBDBlocks struct {
	baseMessage
	LowHash  *externalapi.DomainHash
	HighHash *externalapi.DomainHash
}

// Command returns the protocol command string for the message. This is part
// of the Message interface implementation.
func (msg *MsgRequestIBDBlocks) Command() MessageCommand {
	return CmdRequestIBDBlocks
}

// NewMsgRequstIBDBlocks returns a new kaspa RequestIBDBlocks message that conforms to the
// Message interface using the passed parameters and defaults for the remaining
// fields.
func NewMsgRequstIBDBlocks(lowHash, highHash *externalapi.DomainHash) *MsgRequestIBDBlocks {
	return &MsgRequestIBDBlocks{
		LowHash:  lowHash,
		HighHash: highHash,
	}
}
