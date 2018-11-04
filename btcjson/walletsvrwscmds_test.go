// Copyright (c) 2014 The btcsuite developers
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

// TestWalletSvrWsCmds tests all of the wallet server websocket-specific
// commands marshal and unmarshal into valid results include handling of
// optional fields being omitted in the marshalled command, while optional
// fields with defaults have the default assigned on unmarshalled commands.
func TestWalletSvrWsCmds(t *testing.T) {
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
			name: "createEncryptedWallet",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("createEncryptedWallet", "pass")
			},
			staticCmd: func() interface{} {
				return btcjson.NewCreateEncryptedWalletCmd("pass")
			},
			marshalled:   `{"jsonRpc":"1.0","method":"createEncryptedWallet","params":["pass"],"id":1}`,
			unmarshalled: &btcjson.CreateEncryptedWalletCmd{Passphrase: "pass"},
		},
		{
			name: "exportWatchingWallet",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("exportWatchingWallet")
			},
			staticCmd: func() interface{} {
				return btcjson.NewExportWatchingWalletCmd(nil, nil)
			},
			marshalled: `{"jsonRpc":"1.0","method":"exportWatchingWallet","params":[],"id":1}`,
			unmarshalled: &btcjson.ExportWatchingWalletCmd{
				Account:  nil,
				Download: btcjson.Bool(false),
			},
		},
		{
			name: "exportWatchingWallet optional1",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("exportWatchingWallet", "acct")
			},
			staticCmd: func() interface{} {
				return btcjson.NewExportWatchingWalletCmd(btcjson.String("acct"), nil)
			},
			marshalled: `{"jsonRpc":"1.0","method":"exportWatchingWallet","params":["acct"],"id":1}`,
			unmarshalled: &btcjson.ExportWatchingWalletCmd{
				Account:  btcjson.String("acct"),
				Download: btcjson.Bool(false),
			},
		},
		{
			name: "exportWatchingWallet optional2",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("exportWatchingWallet", "acct", true)
			},
			staticCmd: func() interface{} {
				return btcjson.NewExportWatchingWalletCmd(btcjson.String("acct"),
					btcjson.Bool(true))
			},
			marshalled: `{"jsonRpc":"1.0","method":"exportWatchingWallet","params":["acct",true],"id":1}`,
			unmarshalled: &btcjson.ExportWatchingWalletCmd{
				Account:  btcjson.String("acct"),
				Download: btcjson.Bool(true),
			},
		},
		{
			name: "getUnconfirmedBalance",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("getUnconfirmedBalance")
			},
			staticCmd: func() interface{} {
				return btcjson.NewGetUnconfirmedBalanceCmd(nil)
			},
			marshalled: `{"jsonRpc":"1.0","method":"getUnconfirmedBalance","params":[],"id":1}`,
			unmarshalled: &btcjson.GetUnconfirmedBalanceCmd{
				Account: nil,
			},
		},
		{
			name: "getUnconfirmedBalance optional1",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("getUnconfirmedBalance", "acct")
			},
			staticCmd: func() interface{} {
				return btcjson.NewGetUnconfirmedBalanceCmd(btcjson.String("acct"))
			},
			marshalled: `{"jsonRpc":"1.0","method":"getUnconfirmedBalance","params":["acct"],"id":1}`,
			unmarshalled: &btcjson.GetUnconfirmedBalanceCmd{
				Account: btcjson.String("acct"),
			},
		},
		{
			name: "listAddressTransactions",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("listAddressTransactions", `["1Address"]`)
			},
			staticCmd: func() interface{} {
				return btcjson.NewListAddressTransactionsCmd([]string{"1Address"}, nil)
			},
			marshalled: `{"jsonRpc":"1.0","method":"listAddressTransactions","params":[["1Address"]],"id":1}`,
			unmarshalled: &btcjson.ListAddressTransactionsCmd{
				Addresses: []string{"1Address"},
				Account:   nil,
			},
		},
		{
			name: "listAddressTransactions optional1",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("listAddressTransactions", `["1Address"]`, "acct")
			},
			staticCmd: func() interface{} {
				return btcjson.NewListAddressTransactionsCmd([]string{"1Address"},
					btcjson.String("acct"))
			},
			marshalled: `{"jsonRpc":"1.0","method":"listAddressTransactions","params":[["1Address"],"acct"],"id":1}`,
			unmarshalled: &btcjson.ListAddressTransactionsCmd{
				Addresses: []string{"1Address"},
				Account:   btcjson.String("acct"),
			},
		},
		{
			name: "listAllTransactions",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("listAllTransactions")
			},
			staticCmd: func() interface{} {
				return btcjson.NewListAllTransactionsCmd(nil)
			},
			marshalled: `{"jsonRpc":"1.0","method":"listAllTransactions","params":[],"id":1}`,
			unmarshalled: &btcjson.ListAllTransactionsCmd{
				Account: nil,
			},
		},
		{
			name: "listAllTransactions optional",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("listAllTransactions", "acct")
			},
			staticCmd: func() interface{} {
				return btcjson.NewListAllTransactionsCmd(btcjson.String("acct"))
			},
			marshalled: `{"jsonRpc":"1.0","method":"listAllTransactions","params":["acct"],"id":1}`,
			unmarshalled: &btcjson.ListAllTransactionsCmd{
				Account: btcjson.String("acct"),
			},
		},
		{
			name: "recoverAddresses",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("recoverAddresses", "acct", 10)
			},
			staticCmd: func() interface{} {
				return btcjson.NewRecoverAddressesCmd("acct", 10)
			},
			marshalled: `{"jsonRpc":"1.0","method":"recoverAddresses","params":["acct",10],"id":1}`,
			unmarshalled: &btcjson.RecoverAddressesCmd{
				Account: "acct",
				N:       10,
			},
		},
		{
			name: "walletIsLocked",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("walletIsLocked")
			},
			staticCmd: func() interface{} {
				return btcjson.NewWalletIsLockedCmd()
			},
			marshalled:   `{"jsonRpc":"1.0","method":"walletIsLocked","params":[],"id":1}`,
			unmarshalled: &btcjson.WalletIsLockedCmd{},
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
