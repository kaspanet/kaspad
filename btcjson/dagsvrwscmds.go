// Copyright (c) 2014-2017 The btcsuite developers
// Copyright (c) 2015-2017 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

// NOTE: This file is intended to house the RPC commands that are supported by
// a dag server, but are only available via websockets.

package btcjson

// AuthenticateCmd defines the authenticate JSON-RPC command.
type AuthenticateCmd struct {
	Username   string
	Passphrase string
}

// NewAuthenticateCmd returns a new instance which can be used to issue an
// authenticate JSON-RPC command.
func NewAuthenticateCmd(username, passphrase string) *AuthenticateCmd {
	return &AuthenticateCmd{
		Username:   username,
		Passphrase: passphrase,
	}
}

// NotifyBlocksCmd defines the notifyBlocks JSON-RPC command.
type NotifyBlocksCmd struct{}

// NewNotifyBlocksCmd returns a new instance which can be used to issue a
// notifyBlocks JSON-RPC command.
func NewNotifyBlocksCmd() *NotifyBlocksCmd {
	return &NotifyBlocksCmd{}
}

// StopNotifyBlocksCmd defines the stopNotifyBlocks JSON-RPC command.
type StopNotifyBlocksCmd struct{}

// NewStopNotifyBlocksCmd returns a new instance which can be used to issue a
// stopNotifyBlocks JSON-RPC command.
func NewStopNotifyBlocksCmd() *StopNotifyBlocksCmd {
	return &StopNotifyBlocksCmd{}
}

// NotifyNewTransactionsCmd defines the notifyNewTransactions JSON-RPC command.
type NotifyNewTransactionsCmd struct {
	Verbose *bool `jsonrpcdefault:"false"`
}

// NewNotifyNewTransactionsCmd returns a new instance which can be used to issue
// a notifyNewTransactions JSON-RPC command.
//
// The parameters which are pointers indicate they are optional.  Passing nil
// for optional parameters will use the default value.
func NewNotifyNewTransactionsCmd(verbose *bool) *NotifyNewTransactionsCmd {
	return &NotifyNewTransactionsCmd{
		Verbose: verbose,
	}
}

// SessionCmd defines the session JSON-RPC command.
type SessionCmd struct{}

// NewSessionCmd returns a new instance which can be used to issue a session
// JSON-RPC command.
func NewSessionCmd() *SessionCmd {
	return &SessionCmd{}
}

// StopNotifyNewTransactionsCmd defines the stopNotifyNewTransactions JSON-RPC command.
type StopNotifyNewTransactionsCmd struct{}

// NewStopNotifyNewTransactionsCmd returns a new instance which can be used to issue
// a stopNotifyNewTransactions JSON-RPC command.
//
// The parameters which are pointers indicate they are optional.  Passing nil
// for optional parameters will use the default value.
func NewStopNotifyNewTransactionsCmd() *StopNotifyNewTransactionsCmd {
	return &StopNotifyNewTransactionsCmd{}
}

// NotifyReceivedCmd defines the notifyReceived JSON-RPC command.
//
// NOTE: Deprecated. Use LoadTxFilterCmd instead.
type NotifyReceivedCmd struct {
	Addresses []string
}

// NewNotifyReceivedCmd returns a new instance which can be used to issue a
// notifyReceived JSON-RPC command.
//
// NOTE: Deprecated. Use NewLoadTxFilterCmd instead.
func NewNotifyReceivedCmd(addresses []string) *NotifyReceivedCmd {
	return &NotifyReceivedCmd{
		Addresses: addresses,
	}
}

// OutPoint describes a transaction outpoint that will be marshalled to and
// from JSON.
type OutPoint struct {
	Hash  string `json:"hash"`
	Index uint32 `json:"index"`
}

// LoadTxFilterCmd defines the loadTxFilter request parameters to load or
// reload a transaction filter.
//
// NOTE: This is a btcd extension ported from github.com/decred/dcrd/dcrjson
// and requires a websocket connection.
type LoadTxFilterCmd struct {
	Reload    bool
	Addresses []string
	OutPoints []OutPoint
}

// NewLoadTxFilterCmd returns a new instance which can be used to issue a
// loadTxFilter JSON-RPC command.
//
// NOTE: This is a btcd extension ported from github.com/decred/dcrd/dcrjson
// and requires a websocket connection.
func NewLoadTxFilterCmd(reload bool, addresses []string, outPoints []OutPoint) *LoadTxFilterCmd {
	return &LoadTxFilterCmd{
		Reload:    reload,
		Addresses: addresses,
		OutPoints: outPoints,
	}
}

// NotifySpentCmd defines the notifySpent JSON-RPC command.
//
// NOTE: Deprecated. Use LoadTxFilterCmd instead.
type NotifySpentCmd struct {
	OutPoints []OutPoint
}

// NewNotifySpentCmd returns a new instance which can be used to issue a
// notifySpent JSON-RPC command.
//
// NOTE: Deprecated. Use NewLoadTxFilterCmd instead.
func NewNotifySpentCmd(outPoints []OutPoint) *NotifySpentCmd {
	return &NotifySpentCmd{
		OutPoints: outPoints,
	}
}

// StopNotifyReceivedCmd defines the stopNotifyReceived JSON-RPC command.
//
// NOTE: Deprecated. Use LoadTxFilterCmd instead.
type StopNotifyReceivedCmd struct {
	Addresses []string
}

// NewStopNotifyReceivedCmd returns a new instance which can be used to issue a
// stopNotifyReceived JSON-RPC command.
//
// NOTE: Deprecated. Use NewLoadTxFilterCmd instead.
func NewStopNotifyReceivedCmd(addresses []string) *StopNotifyReceivedCmd {
	return &StopNotifyReceivedCmd{
		Addresses: addresses,
	}
}

// StopNotifySpentCmd defines the stopNotifySpent JSON-RPC command.
//
// NOTE: Deprecated. Use LoadTxFilterCmd instead.
type StopNotifySpentCmd struct {
	OutPoints []OutPoint
}

// NewStopNotifySpentCmd returns a new instance which can be used to issue a
// stopNotifySpent JSON-RPC command.
//
// NOTE: Deprecated. Use NewLoadTxFilterCmd instead.
func NewStopNotifySpentCmd(outPoints []OutPoint) *StopNotifySpentCmd {
	return &StopNotifySpentCmd{
		OutPoints: outPoints,
	}
}

// RescanBlocksCmd defines the rescan JSON-RPC command.
//
// NOTE: This is a btcd extension ported from github.com/decred/dcrd/dcrjson
// and requires a websocket connection.
type RescanBlocksCmd struct {
	// Block hashes as a string array.
	BlockHashes []string
}

// NewRescanBlocksCmd returns a new instance which can be used to issue a rescan
// JSON-RPC command.
//
// NOTE: This is a btcd extension ported from github.com/decred/dcrd/dcrjson
// and requires a websocket connection.
func NewRescanBlocksCmd(blockHashes []string) *RescanBlocksCmd {
	return &RescanBlocksCmd{BlockHashes: blockHashes}
}

func init() {
	// The commands in this file are only usable by websockets.
	flags := UFWebsocketOnly

	MustRegisterCmd("authenticate", (*AuthenticateCmd)(nil), flags)
	MustRegisterCmd("loadTxFilter", (*LoadTxFilterCmd)(nil), flags)
	MustRegisterCmd("notifyBlocks", (*NotifyBlocksCmd)(nil), flags)
	MustRegisterCmd("notifyNewTransactions", (*NotifyNewTransactionsCmd)(nil), flags)
	MustRegisterCmd("notifyReceived", (*NotifyReceivedCmd)(nil), flags)
	MustRegisterCmd("notifySpent", (*NotifySpentCmd)(nil), flags)
	MustRegisterCmd("session", (*SessionCmd)(nil), flags)
	MustRegisterCmd("stopNotifyBlocks", (*StopNotifyBlocksCmd)(nil), flags)
	MustRegisterCmd("stopNotifyNewTransactions", (*StopNotifyNewTransactionsCmd)(nil), flags)
	MustRegisterCmd("stopNotifySpent", (*StopNotifySpentCmd)(nil), flags)
	MustRegisterCmd("stopNotifyReceived", (*StopNotifyReceivedCmd)(nil), flags)
	MustRegisterCmd("rescanBlocks", (*RescanBlocksCmd)(nil), flags)
}
