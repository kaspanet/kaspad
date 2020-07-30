// Copyright (c) 2013-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package wire

import (
	"fmt"
)

// MaxMessagePayload is the maximum bytes a message can be regardless of other
// individual limits imposed by messages themselves.
const MaxMessagePayload = (1024 * 1024 * 32) // 32MB

// MessageCommand is a number in the header of a message that represents its type.
type MessageCommand uint32

func (cmd MessageCommand) String() string {
	cmdString, ok := messageCommandToString[cmd]
	if !ok {
		cmdString = "unknown command"
	}
	return fmt.Sprintf("%s [code %d]", cmdString, uint8(cmd))
}

// Commands used in kaspa message headers which describe the type of message.
const (
	CmdVersion             MessageCommand = 0
	CmdVerAck              MessageCommand = 1
	CmdRequestAddresses    MessageCommand = 2
	CmdAddresses           MessageCommand = 3
	CmdRequestIBDBlocks    MessageCommand = 4
	CmdBlock               MessageCommand = 5
	CmdTx                  MessageCommand = 6
	CmdPing                MessageCommand = 7
	CmdPong                MessageCommand = 8
	CmdRequestBlockLocator MessageCommand = 9
	CmdBlockLocator        MessageCommand = 10
	CmdSelectedTip         MessageCommand = 11
	CmdRequestSelectedTip  MessageCommand = 12
	CmdInvRelayBlock       MessageCommand = 13
	CmdRequestRelayBlocks  MessageCommand = 14
	CmdInvTransaction      MessageCommand = 15
	CmdRequestTransactions MessageCommand = 16
	CmdIBDBlock            MessageCommand = 17
)

var messageCommandToString = map[MessageCommand]string{
	CmdVersion:             "Version",
	CmdVerAck:              "VerAck",
	CmdRequestAddresses:    "RequestAddresses",
	CmdAddresses:           "Addresses",
	CmdRequestIBDBlocks:    "RequestBlocks",
	CmdBlock:               "Block",
	CmdTx:                  "Tx",
	CmdPing:                "Ping",
	CmdPong:                "Pong",
	CmdRequestBlockLocator: "RequestBlockLocator",
	CmdBlockLocator:        "BlockLocator",
	CmdSelectedTip:         "SelectedTip",
	CmdRequestSelectedTip:  "RequestSelectedTip",
	CmdInvRelayBlock:       "InvRelayBlock",
	CmdRequestRelayBlocks:  "RequestRelayBlocks",
	CmdInvTransaction:      "InvTransaction",
	CmdRequestTransactions: "RequestTransactions",
	CmdIBDBlock:            "IBDBlock",
}

// Message is an interface that describes a kaspa message. A type that
// implements Message has complete control over the representation of its data
// and may therefore contain additional or fewer fields than those which
// are used directly in the protocol encoded message.
type Message interface {
	Command() MessageCommand
}
