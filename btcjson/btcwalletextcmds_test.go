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

// TestBtcWalletExtCmds tests all of the btcwallet extended commands marshal and
// unmarshal into valid results include handling of optional fields being
// omitted in the marshalled command, while optional fields with defaults have
// the default assigned on unmarshalled commands.
func TestBtcWalletExtCmds(t *testing.T) {
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
			name: "createNewAccount",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("createNewAccount", "acct")
			},
			staticCmd: func() interface{} {
				return btcjson.NewCreateNewAccountCmd("acct")
			},
			marshalled: `{"jsonRpc":"1.0","method":"createNewAccount","params":["acct"],"id":1}`,
			unmarshalled: &btcjson.CreateNewAccountCmd{
				Account: "acct",
			},
		},
		{
			name: "dumpWallet",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("dumpWallet", "filename")
			},
			staticCmd: func() interface{} {
				return btcjson.NewDumpWalletCmd("filename")
			},
			marshalled: `{"jsonRpc":"1.0","method":"dumpWallet","params":["filename"],"id":1}`,
			unmarshalled: &btcjson.DumpWalletCmd{
				Filename: "filename",
			},
		},
		{
			name: "importAddress",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("importAddress", "1Address")
			},
			staticCmd: func() interface{} {
				return btcjson.NewImportAddressCmd("1Address", nil)
			},
			marshalled: `{"jsonRpc":"1.0","method":"importAddress","params":["1Address"],"id":1}`,
			unmarshalled: &btcjson.ImportAddressCmd{
				Address: "1Address",
				Rescan:  btcjson.Bool(true),
			},
		},
		{
			name: "importAddress optional",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("importAddress", "1Address", false)
			},
			staticCmd: func() interface{} {
				return btcjson.NewImportAddressCmd("1Address", btcjson.Bool(false))
			},
			marshalled: `{"jsonRpc":"1.0","method":"importAddress","params":["1Address",false],"id":1}`,
			unmarshalled: &btcjson.ImportAddressCmd{
				Address: "1Address",
				Rescan:  btcjson.Bool(false),
			},
		},
		{
			name: "importPubKey",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("importPubKey", "031234")
			},
			staticCmd: func() interface{} {
				return btcjson.NewImportPubKeyCmd("031234", nil)
			},
			marshalled: `{"jsonRpc":"1.0","method":"importPubKey","params":["031234"],"id":1}`,
			unmarshalled: &btcjson.ImportPubKeyCmd{
				PubKey: "031234",
				Rescan: btcjson.Bool(true),
			},
		},
		{
			name: "importPubKey optional",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("importPubKey", "031234", false)
			},
			staticCmd: func() interface{} {
				return btcjson.NewImportPubKeyCmd("031234", btcjson.Bool(false))
			},
			marshalled: `{"jsonRpc":"1.0","method":"importPubKey","params":["031234",false],"id":1}`,
			unmarshalled: &btcjson.ImportPubKeyCmd{
				PubKey: "031234",
				Rescan: btcjson.Bool(false),
			},
		},
		{
			name: "importWallet",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("importWallet", "filename")
			},
			staticCmd: func() interface{} {
				return btcjson.NewImportWalletCmd("filename")
			},
			marshalled: `{"jsonRpc":"1.0","method":"importWallet","params":["filename"],"id":1}`,
			unmarshalled: &btcjson.ImportWalletCmd{
				Filename: "filename",
			},
		},
		{
			name: "renameAccount",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("renameAccount", "oldacct", "newacct")
			},
			staticCmd: func() interface{} {
				return btcjson.NewRenameAccountCmd("oldacct", "newacct")
			},
			marshalled: `{"jsonRpc":"1.0","method":"renameAccount","params":["oldacct","newacct"],"id":1}`,
			unmarshalled: &btcjson.RenameAccountCmd{
				OldAccount: "oldacct",
				NewAccount: "newacct",
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
