// Copyright (c) 2014-2017 The btcsuite developers
// Copyright (c) 2015-2017 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package btcjson_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/daglabs/btcd/btcjson"
)

// TestDAGSvrWsCmds tests all of the dag server websocket-specific commands
// marshal and unmarshal into valid results include handling of optional fields
// being omitted in the marshalled command, while optional fields with defaults
// have the default assigned on unmarshalled commands.
func TestDAGSvrWsCmds(t *testing.T) {
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
				return btcjson.NewCmd("authenticate", "user", "pass")
			},
			staticCmd: func() interface{} {
				return btcjson.NewAuthenticateCmd("user", "pass")
			},
			marshalled:   `{"jsonrpc":"1.0","method":"authenticate","params":["user","pass"],"id":1}`,
			unmarshalled: &btcjson.AuthenticateCmd{Username: "user", Passphrase: "pass"},
		},
		{
			name: "notifyBlocks",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("notifyBlocks")
			},
			staticCmd: func() interface{} {
				return btcjson.NewNotifyBlocksCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"notifyBlocks","params":[],"id":1}`,
			unmarshalled: &btcjson.NotifyBlocksCmd{},
		},
		{
			name: "stopNotifyBlocks",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("stopNotifyBlocks")
			},
			staticCmd: func() interface{} {
				return btcjson.NewStopNotifyBlocksCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"stopNotifyBlocks","params":[],"id":1}`,
			unmarshalled: &btcjson.StopNotifyBlocksCmd{},
		},
		{
			name: "notifyNewTransactions",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("notifyNewTransactions")
			},
			staticCmd: func() interface{} {
				return btcjson.NewNotifyNewTransactionsCmd(nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"notifyNewTransactions","params":[],"id":1}`,
			unmarshalled: &btcjson.NotifyNewTransactionsCmd{
				Verbose: btcjson.Bool(false),
			},
		},
		{
			name: "notifyNewTransactions optional",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("notifyNewTransactions", true)
			},
			staticCmd: func() interface{} {
				return btcjson.NewNotifyNewTransactionsCmd(btcjson.Bool(true))
			},
			marshalled: `{"jsonrpc":"1.0","method":"notifyNewTransactions","params":[true],"id":1}`,
			unmarshalled: &btcjson.NotifyNewTransactionsCmd{
				Verbose: btcjson.Bool(true),
			},
		},
		{
			name: "stopNotifyNewTransactions",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("stopNotifyNewTransactions")
			},
			staticCmd: func() interface{} {
				return btcjson.NewStopNotifyNewTransactionsCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"stopNotifyNewTransactions","params":[],"id":1}`,
			unmarshalled: &btcjson.StopNotifyNewTransactionsCmd{},
		},
		{
			name: "notifyReceived",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("notifyReceived", []string{"1Address"})
			},
			staticCmd: func() interface{} {
				return btcjson.NewNotifyReceivedCmd([]string{"1Address"})
			},
			marshalled: `{"jsonrpc":"1.0","method":"notifyReceived","params":[["1Address"]],"id":1}`,
			unmarshalled: &btcjson.NotifyReceivedCmd{
				Addresses: []string{"1Address"},
			},
		},
		{
			name: "stopNotifyReceived",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("stopNotifyReceived", []string{"1Address"})
			},
			staticCmd: func() interface{} {
				return btcjson.NewStopNotifyReceivedCmd([]string{"1Address"})
			},
			marshalled: `{"jsonrpc":"1.0","method":"stopNotifyReceived","params":[["1Address"]],"id":1}`,
			unmarshalled: &btcjson.StopNotifyReceivedCmd{
				Addresses: []string{"1Address"},
			},
		},
		{
			name: "notifySpent",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("notifySpent", `[{"txid":"123","index":0}]`)
			},
			staticCmd: func() interface{} {
				ops := []btcjson.OutPoint{{TxID: "123", Index: 0}}
				return btcjson.NewNotifySpentCmd(ops)
			},
			marshalled: `{"jsonrpc":"1.0","method":"notifySpent","params":[[{"txid":"123","index":0}]],"id":1}`,
			unmarshalled: &btcjson.NotifySpentCmd{
				OutPoints: []btcjson.OutPoint{{TxID: "123", Index: 0}},
			},
		},
		{
			name: "stopNotifySpent",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("stopNotifySpent", `[{"txid":"123","index":0}]`)
			},
			staticCmd: func() interface{} {
				ops := []btcjson.OutPoint{{TxID: "123", Index: 0}}
				return btcjson.NewStopNotifySpentCmd(ops)
			},
			marshalled: `{"jsonrpc":"1.0","method":"stopNotifySpent","params":[[{"txid":"123","index":0}]],"id":1}`,
			unmarshalled: &btcjson.StopNotifySpentCmd{
				OutPoints: []btcjson.OutPoint{{TxID: "123", Index: 0}},
			},
		},
		{
			name: "loadTxFilter",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("loadTxFilter", false, `["1Address"]`, `[{"txid":"0000000000000000000000000000000000000000000000000000000000000123","index":0}]`)
			},
			staticCmd: func() interface{} {
				addrs := []string{"1Address"}
				ops := []btcjson.OutPoint{{
					TxID:  "0000000000000000000000000000000000000000000000000000000000000123",
					Index: 0,
				}}
				return btcjson.NewLoadTxFilterCmd(false, addrs, ops)
			},
			marshalled: `{"jsonrpc":"1.0","method":"loadTxFilter","params":[false,["1Address"],[{"txid":"0000000000000000000000000000000000000000000000000000000000000123","index":0}]],"id":1}`,
			unmarshalled: &btcjson.LoadTxFilterCmd{
				Reload:    false,
				Addresses: []string{"1Address"},
				OutPoints: []btcjson.OutPoint{{TxID: "0000000000000000000000000000000000000000000000000000000000000123", Index: 0}},
			},
		},
		{
			name: "rescanBlocks",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("rescanBlocks", `["0000000000000000000000000000000000000000000000000000000000000123"]`)
			},
			staticCmd: func() interface{} {
				blockhashes := []string{"0000000000000000000000000000000000000000000000000000000000000123"}
				return btcjson.NewRescanBlocksCmd(blockhashes)
			},
			marshalled: `{"jsonrpc":"1.0","method":"rescanBlocks","params":[["0000000000000000000000000000000000000000000000000000000000000123"]],"id":1}`,
			unmarshalled: &btcjson.RescanBlocksCmd{
				BlockHashes: []string{"0000000000000000000000000000000000000000000000000000000000000123"},
			},
		},
	}

	t.Logf("Running %d tests", len(tests))
	for i, test := range tests {
		// Marshal the command as created by the new static command
		// creation function.
		marshalled, err := btcjson.MarshalCmd(testID, test.staticCmd())
		if err != nil {
			t.Errorf("MarshalCmd #%d (%s) unexpected error: %v", i,
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
			t.Errorf("Test #%d (%s) unexpected NewCmd error: %v ",
				i, test.name, err)
		}

		// Marshal the command as created by the generic new command
		// creation function.
		marshalled, err = btcjson.MarshalCmd(testID, cmd)
		if err != nil {
			t.Errorf("MarshalCmd #%d (%s) unexpected error: %v", i,
				test.name, err)
			continue
		}

		if !bytes.Equal(marshalled, []byte(test.marshalled)) {
			t.Errorf("Test #%d (%s) unexpected marshalled data - "+
				"got %s, want %s", i, test.name, marshalled,
				test.marshalled)
			continue
		}

		var request btcjson.Request
		if err := json.Unmarshal(marshalled, &request); err != nil {
			t.Errorf("Test #%d (%s) unexpected error while "+
				"unmarshalling JSON-RPC request: %v", i,
				test.name, err)
			continue
		}

		cmd, err = btcjson.UnmarshalCmd(&request)
		if err != nil {
			t.Errorf("UnmarshalCmd #%d (%s) unexpected error: %v", i,
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
