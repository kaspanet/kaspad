// Copyright (c) 2013-2015 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package appmessage

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// MsgTransactionNotFound defines a kaspa TransactionNotFound message which is sent in response to
// a RequestTransactions message if any of the requested data in not available on the peer.
type MsgTransactionNotFound struct {
	baseMessage
	ID *externalapi.DomainTransactionID
}

// Command returns the protocol command string for the message. This is part
// of the Message interface implementation.
func (msg *MsgTransactionNotFound) Command() MessageCommand {
	return CmdTransactionNotFound
}

// NewMsgTransactionNotFound returns a new kaspa transactionsnotfound message that conforms to the
// Message interface. See MsgTransactionNotFound for details.
func NewMsgTransactionNotFound(id *externalapi.DomainTransactionID) *MsgTransactionNotFound {
	return &MsgTransactionNotFound{
		ID: id,
	}
}
