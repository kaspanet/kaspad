// Copyright (c) 2013-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package appmessage

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// MsgRequestAnticone implements the Message interface and represents a kaspa
// RequestHeaders message. It is used to request the set past(ContextHash) \cap anticone(BlockHash)
type MsgRequestAnticone struct {
	baseMessage
	BlockHash   *externalapi.DomainHash
	ContextHash *externalapi.DomainHash
}

// Command returns the protocol command string for the message. This is part
// of the Message interface implementation.
func (msg *MsgRequestAnticone) Command() MessageCommand {
	return CmdRequestAnticone
}

// NewMsgRequestAnticone returns a new kaspa RequestPastDiff message that conforms to the
// Message interface using the passed parameters and defaults for the remaining
// fields.
func NewMsgRequestAnticone(blockHash, contextHash *externalapi.DomainHash) *MsgRequestAnticone {
	return &MsgRequestAnticone{
		BlockHash:   blockHash,
		ContextHash: contextHash,
	}
}
