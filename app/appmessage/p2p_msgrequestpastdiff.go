// Copyright (c) 2013-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package appmessage

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// MsgRequestPastDiff implements the Message interface and represents a kaspa
// RequestHeaders message. It is used to request a past diff past(RequestedHash) \setminus past(HasHash)
type MsgRequestPastDiff struct {
	baseMessage
	HasHash       *externalapi.DomainHash
	RequestedHash *externalapi.DomainHash
}

// Command returns the protocol command string for the message. This is part
// of the Message interface implementation.
func (msg *MsgRequestPastDiff) Command() MessageCommand {
	return CmdRequestPastDiff
}

// NewMsgRequestPastDiff returns a new kaspa RequestPastDiff message that conforms to the
// Message interface using the passed parameters and defaults for the remaining
// fields.
func NewMsgRequestPastDiff(hasHash, requestedHash *externalapi.DomainHash) *MsgRequestPastDiff {
	return &MsgRequestPastDiff{
		HasHash:       hasHash,
		RequestedHash: requestedHash,
	}
}
