// Copyright (c) 2014-2017 The btcsuite developers
// Copyright (c) 2015-2017 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package kaspajson_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/kaspanet/kaspad/kaspajson"
)

// TestDAGSvrWsCmds tests all of the kaspa rpc server websocket-specific commands
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
				return kaspajson.NewCmd("authenticate", "user", "pass")
			},
			staticCmd: func() interface{} {
				return kaspajson.NewAuthenticateCmd("user", "pass")
			},
			marshalled:   `{"jsonrpc":"1.0","method":"authenticate","params":["user","pass"],"id":1}`,
			unmarshalled: &kaspajson.AuthenticateCmd{Username: "user", Passphrase: "pass"},
		},
		{
			name: "notifyBlocks",
			newCmd: func() (interface{}, error) {
				return kaspajson.NewCmd("notifyBlocks")
			},
			staticCmd: func() interface{} {
				return kaspajson.NewNotifyBlocksCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"notifyBlocks","params":[],"id":1}`,
			unmarshalled: &kaspajson.NotifyBlocksCmd{},
		},
		{
			name: "stopNotifyBlocks",
			newCmd: func() (interface{}, error) {
				return kaspajson.NewCmd("stopNotifyBlocks")
			},
			staticCmd: func() interface{} {
				return kaspajson.NewStopNotifyBlocksCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"stopNotifyBlocks","params":[],"id":1}`,
			unmarshalled: &kaspajson.StopNotifyBlocksCmd{},
		},
		{
			name: "notifyChainChanges",
			newCmd: func() (interface{}, error) {
				return kaspajson.NewCmd("notifyChainChanges")
			},
			staticCmd: func() interface{} {
				return kaspajson.NewNotifyChainChangesCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"notifyChainChanges","params":[],"id":1}`,
			unmarshalled: &kaspajson.NotifyChainChangesCmd{},
		},
		{
			name: "stopNotifyChainChanges",
			newCmd: func() (interface{}, error) {
				return kaspajson.NewCmd("stopNotifyChainChanges")
			},
			staticCmd: func() interface{} {
				return kaspajson.NewStopNotifyChainChangesCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"stopNotifyChainChanges","params":[],"id":1}`,
			unmarshalled: &kaspajson.StopNotifyChainChangesCmd{},
		},
		{
			name: "notifyNewTransactions",
			newCmd: func() (interface{}, error) {
				return kaspajson.NewCmd("notifyNewTransactions")
			},
			staticCmd: func() interface{} {
				return kaspajson.NewNotifyNewTransactionsCmd(nil, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"notifyNewTransactions","params":[],"id":1}`,
			unmarshalled: &kaspajson.NotifyNewTransactionsCmd{
				Verbose: kaspajson.Bool(false),
			},
		},
		{
			name: "notifyNewTransactions optional",
			newCmd: func() (interface{}, error) {
				return kaspajson.NewCmd("notifyNewTransactions", true)
			},
			staticCmd: func() interface{} {
				return kaspajson.NewNotifyNewTransactionsCmd(kaspajson.Bool(true), nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"notifyNewTransactions","params":[true],"id":1}`,
			unmarshalled: &kaspajson.NotifyNewTransactionsCmd{
				Verbose: kaspajson.Bool(true),
			},
		},
		{
			name: "notifyNewTransactions optional 2",
			newCmd: func() (interface{}, error) {
				return kaspajson.NewCmd("notifyNewTransactions", true, "0000000000000000000000000000000000000123")
			},
			staticCmd: func() interface{} {
				return kaspajson.NewNotifyNewTransactionsCmd(kaspajson.Bool(true), kaspajson.String("0000000000000000000000000000000000000123"))
			},
			marshalled: `{"jsonrpc":"1.0","method":"notifyNewTransactions","params":[true,"0000000000000000000000000000000000000123"],"id":1}`,
			unmarshalled: &kaspajson.NotifyNewTransactionsCmd{
				Verbose:    kaspajson.Bool(true),
				Subnetwork: kaspajson.String("0000000000000000000000000000000000000123"),
			},
		},
		{
			name: "stopNotifyNewTransactions",
			newCmd: func() (interface{}, error) {
				return kaspajson.NewCmd("stopNotifyNewTransactions")
			},
			staticCmd: func() interface{} {
				return kaspajson.NewStopNotifyNewTransactionsCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"stopNotifyNewTransactions","params":[],"id":1}`,
			unmarshalled: &kaspajson.StopNotifyNewTransactionsCmd{},
		},
		{
			name: "loadTxFilter",
			newCmd: func() (interface{}, error) {
				return kaspajson.NewCmd("loadTxFilter", false, `["1Address"]`, `[{"txid":"0000000000000000000000000000000000000000000000000000000000000123","index":0}]`)
			},
			staticCmd: func() interface{} {
				addrs := []string{"1Address"}
				ops := []kaspajson.Outpoint{{
					TxID:  "0000000000000000000000000000000000000000000000000000000000000123",
					Index: 0,
				}}
				return kaspajson.NewLoadTxFilterCmd(false, addrs, ops)
			},
			marshalled: `{"jsonrpc":"1.0","method":"loadTxFilter","params":[false,["1Address"],[{"txid":"0000000000000000000000000000000000000000000000000000000000000123","index":0}]],"id":1}`,
			unmarshalled: &kaspajson.LoadTxFilterCmd{
				Reload:    false,
				Addresses: []string{"1Address"},
				Outpoints: []kaspajson.Outpoint{{TxID: "0000000000000000000000000000000000000000000000000000000000000123", Index: 0}},
			},
		},
		{
			name: "rescanBlocks",
			newCmd: func() (interface{}, error) {
				return kaspajson.NewCmd("rescanBlocks", `["0000000000000000000000000000000000000000000000000000000000000123"]`)
			},
			staticCmd: func() interface{} {
				blockhashes := []string{"0000000000000000000000000000000000000000000000000000000000000123"}
				return kaspajson.NewRescanBlocksCmd(blockhashes)
			},
			marshalled: `{"jsonrpc":"1.0","method":"rescanBlocks","params":[["0000000000000000000000000000000000000000000000000000000000000123"]],"id":1}`,
			unmarshalled: &kaspajson.RescanBlocksCmd{
				BlockHashes: []string{"0000000000000000000000000000000000000000000000000000000000000123"},
			},
		},
	}

	t.Logf("Running %d tests", len(tests))
	for i, test := range tests {
		// Marshal the command as created by the new static command
		// creation function.
		marshalled, err := kaspajson.MarshalCmd(testID, test.staticCmd())
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
		marshalled, err = kaspajson.MarshalCmd(testID, cmd)
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

		var request kaspajson.Request
		if err := json.Unmarshal(marshalled, &request); err != nil {
			t.Errorf("Test #%d (%s) unexpected error while "+
				"unmarshalling JSON-RPC request: %v", i,
				test.name, err)
			continue
		}

		cmd, err = kaspajson.UnmarshalCmd(&request)
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
