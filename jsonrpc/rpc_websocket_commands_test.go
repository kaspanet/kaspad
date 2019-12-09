// Copyright (c) 2014-2017 The btcsuite developers
// Copyright (c) 2015-2017 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package jsonrpc_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/kaspanet/kaspad/jsonrpc"
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
				return jsonrpc.NewCommand("authenticate", "user", "pass")
			},
			staticCmd: func() interface{} {
				return jsonrpc.NewAuthenticateCmd("user", "pass")
			},
			marshalled:   `{"jsonrpc":"1.0","method":"authenticate","params":["user","pass"],"id":1}`,
			unmarshalled: &jsonrpc.AuthenticateCmd{Username: "user", Passphrase: "pass"},
		},
		{
			name: "notifyBlocks",
			newCmd: func() (interface{}, error) {
				return jsonrpc.NewCommand("notifyBlocks")
			},
			staticCmd: func() interface{} {
				return jsonrpc.NewNotifyBlocksCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"notifyBlocks","params":[],"id":1}`,
			unmarshalled: &jsonrpc.NotifyBlocksCmd{},
		},
		{
			name: "stopNotifyBlocks",
			newCmd: func() (interface{}, error) {
				return jsonrpc.NewCommand("stopNotifyBlocks")
			},
			staticCmd: func() interface{} {
				return jsonrpc.NewStopNotifyBlocksCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"stopNotifyBlocks","params":[],"id":1}`,
			unmarshalled: &jsonrpc.StopNotifyBlocksCmd{},
		},
		{
			name: "notifyChainChanges",
			newCmd: func() (interface{}, error) {
				return jsonrpc.NewCommand("notifyChainChanges")
			},
			staticCmd: func() interface{} {
				return jsonrpc.NewNotifyChainChangesCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"notifyChainChanges","params":[],"id":1}`,
			unmarshalled: &jsonrpc.NotifyChainChangesCmd{},
		},
		{
			name: "stopNotifyChainChanges",
			newCmd: func() (interface{}, error) {
				return jsonrpc.NewCommand("stopNotifyChainChanges")
			},
			staticCmd: func() interface{} {
				return jsonrpc.NewStopNotifyChainChangesCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"stopNotifyChainChanges","params":[],"id":1}`,
			unmarshalled: &jsonrpc.StopNotifyChainChangesCmd{},
		},
		{
			name: "notifyNewTransactions",
			newCmd: func() (interface{}, error) {
				return jsonrpc.NewCommand("notifyNewTransactions")
			},
			staticCmd: func() interface{} {
				return jsonrpc.NewNotifyNewTransactionsCmd(nil, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"notifyNewTransactions","params":[],"id":1}`,
			unmarshalled: &jsonrpc.NotifyNewTransactionsCmd{
				Verbose: jsonrpc.Bool(false),
			},
		},
		{
			name: "notifyNewTransactions optional",
			newCmd: func() (interface{}, error) {
				return jsonrpc.NewCommand("notifyNewTransactions", true)
			},
			staticCmd: func() interface{} {
				return jsonrpc.NewNotifyNewTransactionsCmd(jsonrpc.Bool(true), nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"notifyNewTransactions","params":[true],"id":1}`,
			unmarshalled: &jsonrpc.NotifyNewTransactionsCmd{
				Verbose: jsonrpc.Bool(true),
			},
		},
		{
			name: "notifyNewTransactions optional 2",
			newCmd: func() (interface{}, error) {
				return jsonrpc.NewCommand("notifyNewTransactions", true, "0000000000000000000000000000000000000123")
			},
			staticCmd: func() interface{} {
				return jsonrpc.NewNotifyNewTransactionsCmd(jsonrpc.Bool(true), jsonrpc.String("0000000000000000000000000000000000000123"))
			},
			marshalled: `{"jsonrpc":"1.0","method":"notifyNewTransactions","params":[true,"0000000000000000000000000000000000000123"],"id":1}`,
			unmarshalled: &jsonrpc.NotifyNewTransactionsCmd{
				Verbose:    jsonrpc.Bool(true),
				Subnetwork: jsonrpc.String("0000000000000000000000000000000000000123"),
			},
		},
		{
			name: "stopNotifyNewTransactions",
			newCmd: func() (interface{}, error) {
				return jsonrpc.NewCommand("stopNotifyNewTransactions")
			},
			staticCmd: func() interface{} {
				return jsonrpc.NewStopNotifyNewTransactionsCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"stopNotifyNewTransactions","params":[],"id":1}`,
			unmarshalled: &jsonrpc.StopNotifyNewTransactionsCmd{},
		},
		{
			name: "loadTxFilter",
			newCmd: func() (interface{}, error) {
				return jsonrpc.NewCommand("loadTxFilter", false, `["1Address"]`, `[{"txid":"0000000000000000000000000000000000000000000000000000000000000123","index":0}]`)
			},
			staticCmd: func() interface{} {
				addrs := []string{"1Address"}
				ops := []jsonrpc.Outpoint{{
					TxID:  "0000000000000000000000000000000000000000000000000000000000000123",
					Index: 0,
				}}
				return jsonrpc.NewLoadTxFilterCmd(false, addrs, ops)
			},
			marshalled: `{"jsonrpc":"1.0","method":"loadTxFilter","params":[false,["1Address"],[{"txid":"0000000000000000000000000000000000000000000000000000000000000123","index":0}]],"id":1}`,
			unmarshalled: &jsonrpc.LoadTxFilterCmd{
				Reload:    false,
				Addresses: []string{"1Address"},
				Outpoints: []jsonrpc.Outpoint{{TxID: "0000000000000000000000000000000000000000000000000000000000000123", Index: 0}},
			},
		},
		{
			name: "rescanBlocks",
			newCmd: func() (interface{}, error) {
				return jsonrpc.NewCommand("rescanBlocks", `["0000000000000000000000000000000000000000000000000000000000000123"]`)
			},
			staticCmd: func() interface{} {
				blockhashes := []string{"0000000000000000000000000000000000000000000000000000000000000123"}
				return jsonrpc.NewRescanBlocksCmd(blockhashes)
			},
			marshalled: `{"jsonrpc":"1.0","method":"rescanBlocks","params":[["0000000000000000000000000000000000000000000000000000000000000123"]],"id":1}`,
			unmarshalled: &jsonrpc.RescanBlocksCmd{
				BlockHashes: []string{"0000000000000000000000000000000000000000000000000000000000000123"},
			},
		},
	}

	t.Logf("Running %d tests", len(tests))
	for i, test := range tests {
		// Marshal the command as created by the new static command
		// creation function.
		marshalled, err := jsonrpc.MarshalCommand(testID, test.staticCmd())
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
		marshalled, err = jsonrpc.MarshalCommand(testID, cmd)
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

		var request jsonrpc.Request
		if err := json.Unmarshal(marshalled, &request); err != nil {
			t.Errorf("Test #%d (%s) unexpected error while "+
				"unmarshalling JSON-RPC request: %v", i,
				test.name, err)
			continue
		}

		cmd, err = jsonrpc.UnmarshalCommand(&request)
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
