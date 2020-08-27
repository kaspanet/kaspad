// Copyright (c) 2014-2017 The btcsuite developers
// Copyright (c) 2015-2017 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

// NOTE: This file is intended to house the RPC commands that are supported by
// a kaspa rpc server, but are only available via websockets.

package model

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

// NotifyChainChangesCmd defines the notifyChainChanges JSON-RPC command.
type NotifyChainChangesCmd struct{}

// NewNotifyChainChangesCmd returns a new instance which can be used to issue a
// notifyChainChanges JSON-RPC command.
func NewNotifyChainChangesCmd() *NotifyChainChangesCmd {
	return &NotifyChainChangesCmd{}
}

// StopNotifyChainChangesCmd defines the stopNotifyChainChanges JSON-RPC command.
type StopNotifyChainChangesCmd struct{}

// NewStopNotifyChainChangesCmd returns a new instance which can be used to issue a
// stopNotifyChainChanges JSON-RPC command.
func NewStopNotifyChainChangesCmd() *StopNotifyChainChangesCmd {
	return &StopNotifyChainChangesCmd{}
}

// NotifyNewTransactionsCmd defines the notifyNewTransactions JSON-RPC command.
type NotifyNewTransactionsCmd struct {
	Verbose    *bool `jsonrpcdefault:"false"`
	Subnetwork *string
}

// NewNotifyNewTransactionsCmd returns a new instance which can be used to issue
// a notifyNewTransactions JSON-RPC command.
//
// The parameters which are pointers indicate they are optional. Passing nil
// for optional parameters will use the default value.
func NewNotifyNewTransactionsCmd(verbose *bool, subnetworkID *string) *NotifyNewTransactionsCmd {
	return &NotifyNewTransactionsCmd{
		Verbose:    verbose,
		Subnetwork: subnetworkID,
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
// The parameters which are pointers indicate they are optional. Passing nil
// for optional parameters will use the default value.
func NewStopNotifyNewTransactionsCmd() *StopNotifyNewTransactionsCmd {
	return &StopNotifyNewTransactionsCmd{}
}

// Outpoint describes a transaction outpoint that will be marshalled to and
// from JSON.
type Outpoint struct {
	TxID  string `json:"txid"`
	Index uint32 `json:"index"`
}

// LoadTxFilterCmd defines the loadTxFilter request parameters to load or
// reload a transaction filter.
type LoadTxFilterCmd struct {
	Reload    bool
	Addresses []string
	Outpoints []Outpoint
}

// NewLoadTxFilterCmd returns a new instance which can be used to issue a
// loadTxFilter JSON-RPC command.
func NewLoadTxFilterCmd(reload bool, addresses []string, outpoints []Outpoint) *LoadTxFilterCmd {
	return &LoadTxFilterCmd{
		Reload:    reload,
		Addresses: addresses,
		Outpoints: outpoints,
	}
}

// NotifyFinalityConflictCmd defines the notifyFinalityConflicts JSON-RPC command.
type NotifyFinalityConflictCmd struct{}

// NewNotifyFinalityConflictsCmd returns a new instance which can be used to issue
// a notifyFinalityConflicts JSON-RPC command.
func NewNotifyFinalityConflictsCmd() *NotifyFinalityConflictCmd {
	return &NotifyFinalityConflictCmd{}
}

// StopNotifyFinalityConflictCmd defines the stopNotifyFinalityConflicts JSON-RPC command.
type StopNotifyFinalityConflictCmd struct{}

// NewStopNotifyFinalityConflictsCmd returns a new instance which can be used to issue
// a stopNotifyFinalityConflicts JSON-RPC command.
func NewStopNotifyFinalityConflictsCmd() *NotifyFinalityConflictCmd {
	return &NotifyFinalityConflictCmd{}
}

func init() {
	// The commands in this file are only usable by websockets.
	flags := UFWebsocketOnly

	MustRegisterCommand("authenticate", (*AuthenticateCmd)(nil), flags)
	MustRegisterCommand("loadTxFilter", (*LoadTxFilterCmd)(nil), flags)
	MustRegisterCommand("notifyBlocks", (*NotifyBlocksCmd)(nil), flags)
	MustRegisterCommand("notifyChainChanges", (*NotifyChainChangesCmd)(nil), flags)
	MustRegisterCommand("notifyNewTransactions", (*NotifyNewTransactionsCmd)(nil), flags)
	MustRegisterCommand("session", (*SessionCmd)(nil), flags)
	MustRegisterCommand("stopNotifyBlocks", (*StopNotifyBlocksCmd)(nil), flags)
	MustRegisterCommand("stopNotifyChainChanges", (*StopNotifyChainChangesCmd)(nil), flags)
	MustRegisterCommand("stopNotifyNewTransactions", (*StopNotifyNewTransactionsCmd)(nil), flags)
	MustRegisterCommand("notifyFinalityConflicts", (*NotifyFinalityConflictCmd)(nil), flags)
	MustRegisterCommand("stopNotifyFinalityConflicts", (*StopNotifyFinalityConflictCmd)(nil), flags)
}
