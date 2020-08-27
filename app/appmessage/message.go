// Copyright (c) 2013-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package appmessage

import (
	"fmt"
	"time"
)

// MaxMessagePayload is the maximum bytes a message can be regardless of other
// individual limits imposed by messages themselves.
const MaxMessagePayload = (1024 * 1024 * 32) // 32MB

// MessageCommand is a number in the header of a message that represents its type.
type MessageCommand uint32

func (cmd MessageCommand) String() string {
	cmdString, ok := MessageCommandToString[cmd]
	if !ok {
		cmdString = "unknown command"
	}
	return fmt.Sprintf("%s [code %d]", cmdString, uint8(cmd))
}

// Commands used in kaspa message headers which describe the type of message.
const (
	// protocol
	CmdVersion MessageCommand = iota
	CmdVerAck
	CmdRequestAddresses
	CmdAddresses
	CmdRequestIBDBlocks
	CmdBlock
	CmdTx
	CmdPing
	CmdPong
	CmdRequestBlockLocator
	CmdBlockLocator
	CmdSelectedTip
	CmdRequestSelectedTip
	CmdInvRelayBlock
	CmdRequestRelayBlocks
	CmdInvTransaction
	CmdRequestTransactions
	CmdIBDBlock
	CmdRequestNextIBDBlocks
	CmdDoneIBDBlocks
	CmdTransactionNotFound
	CmdReject

	// rpc
	CmdGetCurrentNetworkRequestMessage
	CmdGetCurrentNetworkResponseMessage
	CmdSubmitBlockRequestMessage
	CmdSubmitBlockResponseMessage
	CmdGetBlockTemplateRequestMessage
	CmdGetBlockTemplateResponseMessage
	CmdGetBlockTemplateTransactionMessage
	CmdRPCErrorMessage
)

// MessageCommandToString maps all MessageCommands to their string representation
var MessageCommandToString = map[MessageCommand]string{
	// protocol
	CmdVersion:              "Version",
	CmdVerAck:               "VerAck",
	CmdRequestAddresses:     "RequestAddresses",
	CmdAddresses:            "Addresses",
	CmdRequestIBDBlocks:     "RequestBlocks",
	CmdBlock:                "Block",
	CmdTx:                   "Tx",
	CmdPing:                 "Ping",
	CmdPong:                 "Pong",
	CmdRequestBlockLocator:  "RequestBlockLocator",
	CmdBlockLocator:         "BlockLocator",
	CmdSelectedTip:          "SelectedTip",
	CmdRequestSelectedTip:   "RequestSelectedTip",
	CmdInvRelayBlock:        "InvRelayBlock",
	CmdRequestRelayBlocks:   "RequestRelayBlocks",
	CmdInvTransaction:       "InvTransaction",
	CmdRequestTransactions:  "RequestTransactions",
	CmdIBDBlock:             "IBDBlock",
	CmdRequestNextIBDBlocks: "RequestNextIBDBlocks",
	CmdDoneIBDBlocks:        "DoneIBDBlocks",
	CmdTransactionNotFound:  "TransactionNotFound",
	CmdReject:               "Reject",

	// rpc
	CmdGetCurrentNetworkRequestMessage:    "GetCurrentNetworkRequest",
	CmdGetCurrentNetworkResponseMessage:   "GetCurrentNetworkResponse",
	CmdSubmitBlockRequestMessage:          "SubmitBlockRequest",
	CmdSubmitBlockResponseMessage:         "SubmitBlockResponse",
	CmdGetBlockTemplateRequestMessage:     "GetBlockTemplateRequest",
	CmdGetBlockTemplateResponseMessage:    "GetBlockTemplateResponse",
	CmdGetBlockTemplateTransactionMessage: "CmdGetBlockTemplateTransaction",
	CmdRPCErrorMessage:                    "CmdRPCError",
}

// Message is an interface that describes a kaspa message. A type that
// implements Message has complete control over the representation of its data
// and may therefore contain additional or fewer fields than those which
// are used directly in the protocol encoded message.
type Message interface {
	Command() MessageCommand
	MessageNumber() uint64
	SetMessageNumber(index uint64)
	ReceivedAt() time.Time
	SetReceivedAt(receivedAt time.Time)
}
