// Copyright (c) 2014-2017 The btcsuite developers
// Copyright (c) 2015-2017 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package rpcmodel_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/kaspanet/kaspad/util/copytopointer"
	"reflect"
	"testing"

	"github.com/kaspanet/kaspad/rpcmodel"
)

// TestRPCServerWebsocketCommands tests all of the kaspa rpc server websocket-specific commands
// marshal and unmarshal into valid results include handling of optional fields
// being omitted in the marshalled command, while optional fields with defaults
// have the default assigned on unmarshalled commands.
func TestRPCServerWebsocketCommands(t *testing.T) {
	t.Parallel()

	testID := int(1)
	tests := []struct {
		name         string
		newCmd       func() (interface{}, error)
		staticCmd    func() interface{}
		marshalled   string
		unmarshalled interface{}
	}{
		{
			name: "authenticate",
			newCmd: func() (interface{}, error) {
				return rpcmodel.NewCommand("authenticate", "user", "pass")
			},
			staticCmd: func() interface{} {
				return rpcmodel.NewAuthenticateCmd("user", "pass")
			},
			marshalled:   `{"jsonrpc":"1.0","method":"authenticate","params":["user","pass"],"id":1}`,
			unmarshalled: &rpcmodel.AuthenticateCmd{Username: "user", Passphrase: "pass"},
		},
		{
			name: "notifyBlocks",
			newCmd: func() (interface{}, error) {
				return rpcmodel.NewCommand("notifyBlocks")
			},
			staticCmd: func() interface{} {
				return rpcmodel.NewNotifyBlocksCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"notifyBlocks","params":[],"id":1}`,
			unmarshalled: &rpcmodel.NotifyBlocksCmd{},
		},
		{
			name: "stopNotifyBlocks",
			newCmd: func() (interface{}, error) {
				return rpcmodel.NewCommand("stopNotifyBlocks")
			},
			staticCmd: func() interface{} {
				return rpcmodel.NewStopNotifyBlocksCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"stopNotifyBlocks","params":[],"id":1}`,
			unmarshalled: &rpcmodel.StopNotifyBlocksCmd{},
		},
		{
			name: "notifyChainChanges",
			newCmd: func() (interface{}, error) {
				return rpcmodel.NewCommand("notifyChainChanges")
			},
			staticCmd: func() interface{} {
				return rpcmodel.NewNotifyChainChangesCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"notifyChainChanges","params":[],"id":1}`,
			unmarshalled: &rpcmodel.NotifyChainChangesCmd{},
		},
		{
			name: "stopNotifyChainChanges",
			newCmd: func() (interface{}, error) {
				return rpcmodel.NewCommand("stopNotifyChainChanges")
			},
			staticCmd: func() interface{} {
				return rpcmodel.NewStopNotifyChainChangesCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"stopNotifyChainChanges","params":[],"id":1}`,
			unmarshalled: &rpcmodel.StopNotifyChainChangesCmd{},
		},
		{
			name: "notifyNewTransactions",
			newCmd: func() (interface{}, error) {
				return rpcmodel.NewCommand("notifyNewTransactions")
			},
			staticCmd: func() interface{} {
				return rpcmodel.NewNotifyNewTransactionsCmd(nil, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"notifyNewTransactions","params":[],"id":1}`,
			unmarshalled: &rpcmodel.NotifyNewTransactionsCmd{
				Verbose: copytopointer.Bool(false),
			},
		},
		{
			name: "notifyNewTransactions optional",
			newCmd: func() (interface{}, error) {
				return rpcmodel.NewCommand("notifyNewTransactions", true)
			},
			staticCmd: func() interface{} {
				return rpcmodel.NewNotifyNewTransactionsCmd(copytopointer.Bool(true), nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"notifyNewTransactions","params":[true],"id":1}`,
			unmarshalled: &rpcmodel.NotifyNewTransactionsCmd{
				Verbose: copytopointer.Bool(true),
			},
		},
		{
			name: "notifyNewTransactions optional 2",
			newCmd: func() (interface{}, error) {
				return rpcmodel.NewCommand("notifyNewTransactions", true, "0000000000000000000000000000000000000123")
			},
			staticCmd: func() interface{} {
				return rpcmodel.NewNotifyNewTransactionsCmd(copytopointer.Bool(true), copytopointer.String("0000000000000000000000000000000000000123"))
			},
			marshalled: `{"jsonrpc":"1.0","method":"notifyNewTransactions","params":[true,"0000000000000000000000000000000000000123"],"id":1}`,
			unmarshalled: &rpcmodel.NotifyNewTransactionsCmd{
				Verbose:    copytopointer.Bool(true),
				Subnetwork: copytopointer.String("0000000000000000000000000000000000000123"),
			},
		},
		{
			name: "stopNotifyNewTransactions",
			newCmd: func() (interface{}, error) {
				return rpcmodel.NewCommand("stopNotifyNewTransactions")
			},
			staticCmd: func() interface{} {
				return rpcmodel.NewStopNotifyNewTransactionsCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"stopNotifyNewTransactions","params":[],"id":1}`,
			unmarshalled: &rpcmodel.StopNotifyNewTransactionsCmd{},
		},
		{
			name: "loadTxFilter",
			newCmd: func() (interface{}, error) {
				return rpcmodel.NewCommand("loadTxFilter", false, `["1Address"]`, `[{"txid":"0000000000000000000000000000000000000000000000000000000000000123","index":0}]`)
			},
			staticCmd: func() interface{} {
				addrs := []string{"1Address"}
				ops := []rpcmodel.Outpoint{{
					TxID:  "0000000000000000000000000000000000000000000000000000000000000123",
					Index: 0,
				}}
				return rpcmodel.NewLoadTxFilterCmd(false, addrs, ops)
			},
			marshalled: `{"jsonrpc":"1.0","method":"loadTxFilter","params":[false,["1Address"],[{"txid":"0000000000000000000000000000000000000000000000000000000000000123","index":0}]],"id":1}`,
			unmarshalled: &rpcmodel.LoadTxFilterCmd{
				Reload:    false,
				Addresses: []string{"1Address"},
				Outpoints: []rpcmodel.Outpoint{{TxID: "0000000000000000000000000000000000000000000000000000000000000123", Index: 0}},
			},
		},
		{
			name: "rescanBlocks",
			newCmd: func() (interface{}, error) {
				return rpcmodel.NewCommand("rescanBlocks", `["0000000000000000000000000000000000000000000000000000000000000123"]`)
			},
			staticCmd: func() interface{} {
				blockhashes := []string{"0000000000000000000000000000000000000000000000000000000000000123"}
				return rpcmodel.NewRescanBlocksCmd(blockhashes)
			},
			marshalled: `{"jsonrpc":"1.0","method":"rescanBlocks","params":[["0000000000000000000000000000000000000000000000000000000000000123"]],"id":1}`,
			unmarshalled: &rpcmodel.RescanBlocksCmd{
				BlockHashes: []string{"0000000000000000000000000000000000000000000000000000000000000123"},
			},
		},
	}

	t.Logf("Running %d tests", len(tests))
	for i, test := range tests {
		// Marshal the command as created by the new static command
		// creation function.
		marshalled, err := rpcmodel.MarshalCommand(testID, test.staticCmd())
		if err != nil {
			t.Errorf("MarshalCommand #%d (%s) unexpected error: %v", i,
				test.name, err)
			continue
		}

		if !bytes.Equal(marshalled, []byte(test.marshalled)) {
			t.Errorf("Test #%d (%s) unexpected marshalled data - "+
				"got %s, want %s", i, test.name, marshalled,
				test.marshalled)
			continue
		}

		// Ensure the command is created without error via the generic
		// new command creation function.
		cmd, err := test.newCmd()
		if err != nil {
			t.Errorf("Test #%d (%s) unexpected NewCommand error: %v ",
				i, test.name, err)
		}

		// Marshal the command as created by the generic new command
		// creation function.
		marshalled, err = rpcmodel.MarshalCommand(testID, cmd)
		if err != nil {
			t.Errorf("MarshalCommand #%d (%s) unexpected error: %v", i,
				test.name, err)
			continue
		}

		if !bytes.Equal(marshalled, []byte(test.marshalled)) {
			t.Errorf("Test #%d (%s) unexpected marshalled data - "+
				"got %s, want %s", i, test.name, marshalled,
				test.marshalled)
			continue
		}

		var request rpcmodel.Request
		if err := json.Unmarshal(marshalled, &request); err != nil {
			t.Errorf("Test #%d (%s) unexpected error while "+
				"unmarshalling JSON-RPC request: %v", i,
				test.name, err)
			continue
		}

		cmd, err = rpcmodel.UnmarshalCommand(&request)
		if err != nil {
			t.Errorf("UnmarshalCommand #%d (%s) unexpected error: %v", i,
				test.name, err)
			continue
		}

		if !reflect.DeepEqual(cmd, test.unmarshalled) {
			t.Errorf("Test #%d (%s) unexpected unmarshalled command "+
				"- got %s, want %s", i, test.name,
				fmt.Sprintf("(%T) %+[1]v", cmd),
				fmt.Sprintf("(%T) %+[1]v\n", test.unmarshalled))
			continue
		}
	}
}
