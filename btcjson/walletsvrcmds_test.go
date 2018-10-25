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

// TestWalletSvrCmds tests all of the wallet server commands marshal and
// unmarshal into valid results include handling of optional fields being
// omitted in the marshalled command, while optional fields with defaults have
// the default assigned on unmarshalled commands.
func TestWalletSvrCmds(t *testing.T) {
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
			name: "addMultisigAddress",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("addMultisigAddress", 2, []string{"031234", "035678"})
			},
			staticCmd: func() interface{} {
				keys := []string{"031234", "035678"}
				return btcjson.NewAddMultisigAddressCmd(2, keys, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"addMultisigAddress","params":[2,["031234","035678"]],"id":1}`,
			unmarshalled: &btcjson.AddMultisigAddressCmd{
				NRequired: 2,
				Keys:      []string{"031234", "035678"},
				Account:   nil,
			},
		},
		{
			name: "addMultisigAddress optional",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("addMultisigAddress", 2, []string{"031234", "035678"}, "test")
			},
			staticCmd: func() interface{} {
				keys := []string{"031234", "035678"}
				return btcjson.NewAddMultisigAddressCmd(2, keys, btcjson.String("test"))
			},
			marshalled: `{"jsonrpc":"1.0","method":"addMultisigAddress","params":[2,["031234","035678"],"test"],"id":1}`,
			unmarshalled: &btcjson.AddMultisigAddressCmd{
				NRequired: 2,
				Keys:      []string{"031234", "035678"},
				Account:   btcjson.String("test"),
			},
		},
		{
			name: "createMultisig",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("createMultisig", 2, []string{"031234", "035678"})
			},
			staticCmd: func() interface{} {
				keys := []string{"031234", "035678"}
				return btcjson.NewCreateMultisigCmd(2, keys)
			},
			marshalled: `{"jsonrpc":"1.0","method":"createMultisig","params":[2,["031234","035678"]],"id":1}`,
			unmarshalled: &btcjson.CreateMultisigCmd{
				NRequired: 2,
				Keys:      []string{"031234", "035678"},
			},
		},
		{
			name: "dumpPrivKey",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("dumpPrivKey", "1Address")
			},
			staticCmd: func() interface{} {
				return btcjson.NewDumpPrivKeyCmd("1Address")
			},
			marshalled: `{"jsonrpc":"1.0","method":"dumpPrivKey","params":["1Address"],"id":1}`,
			unmarshalled: &btcjson.DumpPrivKeyCmd{
				Address: "1Address",
			},
		},
		{
			name: "encryptWallet",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("encryptWallet", "pass")
			},
			staticCmd: func() interface{} {
				return btcjson.NewEncryptWalletCmd("pass")
			},
			marshalled: `{"jsonrpc":"1.0","method":"encryptWallet","params":["pass"],"id":1}`,
			unmarshalled: &btcjson.EncryptWalletCmd{
				Passphrase: "pass",
			},
		},
		{
			name: "estimateFee",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("estimateFee", 6)
			},
			staticCmd: func() interface{} {
				return btcjson.NewEstimateFeeCmd(6)
			},
			marshalled: `{"jsonrpc":"1.0","method":"estimateFee","params":[6],"id":1}`,
			unmarshalled: &btcjson.EstimateFeeCmd{
				NumBlocks: 6,
			},
		},
		{
			name: "estimatePriority",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("estimatePriority", 6)
			},
			staticCmd: func() interface{} {
				return btcjson.NewEstimatePriorityCmd(6)
			},
			marshalled: `{"jsonrpc":"1.0","method":"estimatePriority","params":[6],"id":1}`,
			unmarshalled: &btcjson.EstimatePriorityCmd{
				NumBlocks: 6,
			},
		},
		{
			name: "getAccount",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("getAccount", "1Address")
			},
			staticCmd: func() interface{} {
				return btcjson.NewGetAccountCmd("1Address")
			},
			marshalled: `{"jsonrpc":"1.0","method":"getAccount","params":["1Address"],"id":1}`,
			unmarshalled: &btcjson.GetAccountCmd{
				Address: "1Address",
			},
		},
		{
			name: "getAccountAddress",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("getAccountAddress", "acct")
			},
			staticCmd: func() interface{} {
				return btcjson.NewGetAccountAddressCmd("acct")
			},
			marshalled: `{"jsonrpc":"1.0","method":"getAccountAddress","params":["acct"],"id":1}`,
			unmarshalled: &btcjson.GetAccountAddressCmd{
				Account: "acct",
			},
		},
		{
			name: "getAddressesByAccount",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("getAddressesByAccount", "acct")
			},
			staticCmd: func() interface{} {
				return btcjson.NewGetAddressesByAccountCmd("acct")
			},
			marshalled: `{"jsonrpc":"1.0","method":"getAddressesByAccount","params":["acct"],"id":1}`,
			unmarshalled: &btcjson.GetAddressesByAccountCmd{
				Account: "acct",
			},
		},
		{
			name: "getBalance",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("getBalance")
			},
			staticCmd: func() interface{} {
				return btcjson.NewGetBalanceCmd(nil, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getBalance","params":[],"id":1}`,
			unmarshalled: &btcjson.GetBalanceCmd{
				Account: nil,
				MinConf: btcjson.Int(1),
			},
		},
		{
			name: "getBalance optional1",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("getBalance", "acct")
			},
			staticCmd: func() interface{} {
				return btcjson.NewGetBalanceCmd(btcjson.String("acct"), nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getBalance","params":["acct"],"id":1}`,
			unmarshalled: &btcjson.GetBalanceCmd{
				Account: btcjson.String("acct"),
				MinConf: btcjson.Int(1),
			},
		},
		{
			name: "getBalance optional2",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("getBalance", "acct", 6)
			},
			staticCmd: func() interface{} {
				return btcjson.NewGetBalanceCmd(btcjson.String("acct"), btcjson.Int(6))
			},
			marshalled: `{"jsonrpc":"1.0","method":"getBalance","params":["acct",6],"id":1}`,
			unmarshalled: &btcjson.GetBalanceCmd{
				Account: btcjson.String("acct"),
				MinConf: btcjson.Int(6),
			},
		},
		{
			name: "getNewAddress",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("getNewAddress")
			},
			staticCmd: func() interface{} {
				return btcjson.NewGetNewAddressCmd(nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getNewAddress","params":[],"id":1}`,
			unmarshalled: &btcjson.GetNewAddressCmd{
				Account: nil,
			},
		},
		{
			name: "getNewAddress optional",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("getNewAddress", "acct")
			},
			staticCmd: func() interface{} {
				return btcjson.NewGetNewAddressCmd(btcjson.String("acct"))
			},
			marshalled: `{"jsonrpc":"1.0","method":"getNewAddress","params":["acct"],"id":1}`,
			unmarshalled: &btcjson.GetNewAddressCmd{
				Account: btcjson.String("acct"),
			},
		},
		{
			name: "getRawChangeAddress",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("getRawChangeAddress")
			},
			staticCmd: func() interface{} {
				return btcjson.NewGetRawChangeAddressCmd(nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getRawChangeAddress","params":[],"id":1}`,
			unmarshalled: &btcjson.GetRawChangeAddressCmd{
				Account: nil,
			},
		},
		{
			name: "getRawChangeAddress optional",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("getRawChangeAddress", "acct")
			},
			staticCmd: func() interface{} {
				return btcjson.NewGetRawChangeAddressCmd(btcjson.String("acct"))
			},
			marshalled: `{"jsonrpc":"1.0","method":"getRawChangeAddress","params":["acct"],"id":1}`,
			unmarshalled: &btcjson.GetRawChangeAddressCmd{
				Account: btcjson.String("acct"),
			},
		},
		{
			name: "getReceivedByAccount",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("getReceivedByAccount", "acct")
			},
			staticCmd: func() interface{} {
				return btcjson.NewGetReceivedByAccountCmd("acct", nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getReceivedByAccount","params":["acct"],"id":1}`,
			unmarshalled: &btcjson.GetReceivedByAccountCmd{
				Account: "acct",
				MinConf: btcjson.Int(1),
			},
		},
		{
			name: "getReceivedByAccount optional",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("getReceivedByAccount", "acct", 6)
			},
			staticCmd: func() interface{} {
				return btcjson.NewGetReceivedByAccountCmd("acct", btcjson.Int(6))
			},
			marshalled: `{"jsonrpc":"1.0","method":"getReceivedByAccount","params":["acct",6],"id":1}`,
			unmarshalled: &btcjson.GetReceivedByAccountCmd{
				Account: "acct",
				MinConf: btcjson.Int(6),
			},
		},
		{
			name: "getReceivedByAddress",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("getReceivedByAddress", "1Address")
			},
			staticCmd: func() interface{} {
				return btcjson.NewGetReceivedByAddressCmd("1Address", nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getReceivedByAddress","params":["1Address"],"id":1}`,
			unmarshalled: &btcjson.GetReceivedByAddressCmd{
				Address: "1Address",
				MinConf: btcjson.Int(1),
			},
		},
		{
			name: "getReceivedByAddress optional",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("getReceivedByAddress", "1Address", 6)
			},
			staticCmd: func() interface{} {
				return btcjson.NewGetReceivedByAddressCmd("1Address", btcjson.Int(6))
			},
			marshalled: `{"jsonrpc":"1.0","method":"getReceivedByAddress","params":["1Address",6],"id":1}`,
			unmarshalled: &btcjson.GetReceivedByAddressCmd{
				Address: "1Address",
				MinConf: btcjson.Int(6),
			},
		},
		{
			name: "getTransaction",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("getTransaction", "123")
			},
			staticCmd: func() interface{} {
				return btcjson.NewGetTransactionCmd("123", nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getTransaction","params":["123"],"id":1}`,
			unmarshalled: &btcjson.GetTransactionCmd{
				Txid:             "123",
				IncludeWatchOnly: btcjson.Bool(false),
			},
		},
		{
			name: "getTransaction optional",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("getTransaction", "123", true)
			},
			staticCmd: func() interface{} {
				return btcjson.NewGetTransactionCmd("123", btcjson.Bool(true))
			},
			marshalled: `{"jsonrpc":"1.0","method":"getTransaction","params":["123",true],"id":1}`,
			unmarshalled: &btcjson.GetTransactionCmd{
				Txid:             "123",
				IncludeWatchOnly: btcjson.Bool(true),
			},
		},
		{
			name: "getWalletInfo",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("getWalletInfo")
			},
			staticCmd: func() interface{} {
				return btcjson.NewGetWalletInfoCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getWalletInfo","params":[],"id":1}`,
			unmarshalled: &btcjson.GetWalletInfoCmd{},
		},
		{
			name: "importPrivKey",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("importPrivKey", "abc")
			},
			staticCmd: func() interface{} {
				return btcjson.NewImportPrivKeyCmd("abc", nil, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"importPrivKey","params":["abc"],"id":1}`,
			unmarshalled: &btcjson.ImportPrivKeyCmd{
				PrivKey: "abc",
				Label:   nil,
				Rescan:  btcjson.Bool(true),
			},
		},
		{
			name: "importPrivKey optional1",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("importPrivKey", "abc", "label")
			},
			staticCmd: func() interface{} {
				return btcjson.NewImportPrivKeyCmd("abc", btcjson.String("label"), nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"importPrivKey","params":["abc","label"],"id":1}`,
			unmarshalled: &btcjson.ImportPrivKeyCmd{
				PrivKey: "abc",
				Label:   btcjson.String("label"),
				Rescan:  btcjson.Bool(true),
			},
		},
		{
			name: "importPrivKey optional2",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("importPrivKey", "abc", "label", false)
			},
			staticCmd: func() interface{} {
				return btcjson.NewImportPrivKeyCmd("abc", btcjson.String("label"), btcjson.Bool(false))
			},
			marshalled: `{"jsonrpc":"1.0","method":"importPrivKey","params":["abc","label",false],"id":1}`,
			unmarshalled: &btcjson.ImportPrivKeyCmd{
				PrivKey: "abc",
				Label:   btcjson.String("label"),
				Rescan:  btcjson.Bool(false),
			},
		},
		{
			name: "keyPoolRefill",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("keyPoolRefill")
			},
			staticCmd: func() interface{} {
				return btcjson.NewKeyPoolRefillCmd(nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"keyPoolRefill","params":[],"id":1}`,
			unmarshalled: &btcjson.KeyPoolRefillCmd{
				NewSize: btcjson.Uint(100),
			},
		},
		{
			name: "keyPoolRefill optional",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("keyPoolRefill", 200)
			},
			staticCmd: func() interface{} {
				return btcjson.NewKeyPoolRefillCmd(btcjson.Uint(200))
			},
			marshalled: `{"jsonrpc":"1.0","method":"keyPoolRefill","params":[200],"id":1}`,
			unmarshalled: &btcjson.KeyPoolRefillCmd{
				NewSize: btcjson.Uint(200),
			},
		},
		{
			name: "listAccounts",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("listAccounts")
			},
			staticCmd: func() interface{} {
				return btcjson.NewListAccountsCmd(nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"listAccounts","params":[],"id":1}`,
			unmarshalled: &btcjson.ListAccountsCmd{
				MinConf: btcjson.Int(1),
			},
		},
		{
			name: "listAccounts optional",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("listAccounts", 6)
			},
			staticCmd: func() interface{} {
				return btcjson.NewListAccountsCmd(btcjson.Int(6))
			},
			marshalled: `{"jsonrpc":"1.0","method":"listAccounts","params":[6],"id":1}`,
			unmarshalled: &btcjson.ListAccountsCmd{
				MinConf: btcjson.Int(6),
			},
		},
		{
			name: "listAddressGroupings",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("listAddressGroupings")
			},
			staticCmd: func() interface{} {
				return btcjson.NewListAddressGroupingsCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"listAddressGroupings","params":[],"id":1}`,
			unmarshalled: &btcjson.ListAddressGroupingsCmd{},
		},
		{
			name: "listLockUnspent",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("listLockUnspent")
			},
			staticCmd: func() interface{} {
				return btcjson.NewListLockUnspentCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"listLockUnspent","params":[],"id":1}`,
			unmarshalled: &btcjson.ListLockUnspentCmd{},
		},
		{
			name: "listReceivedByAccount",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("listReceivedByAccount")
			},
			staticCmd: func() interface{} {
				return btcjson.NewListReceivedByAccountCmd(nil, nil, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"listReceivedByAccount","params":[],"id":1}`,
			unmarshalled: &btcjson.ListReceivedByAccountCmd{
				MinConf:          btcjson.Int(1),
				IncludeEmpty:     btcjson.Bool(false),
				IncludeWatchOnly: btcjson.Bool(false),
			},
		},
		{
			name: "listReceivedByAccount optional1",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("listReceivedByAccount", 6)
			},
			staticCmd: func() interface{} {
				return btcjson.NewListReceivedByAccountCmd(btcjson.Int(6), nil, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"listReceivedByAccount","params":[6],"id":1}`,
			unmarshalled: &btcjson.ListReceivedByAccountCmd{
				MinConf:          btcjson.Int(6),
				IncludeEmpty:     btcjson.Bool(false),
				IncludeWatchOnly: btcjson.Bool(false),
			},
		},
		{
			name: "listReceivedByAccount optional2",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("listReceivedByAccount", 6, true)
			},
			staticCmd: func() interface{} {
				return btcjson.NewListReceivedByAccountCmd(btcjson.Int(6), btcjson.Bool(true), nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"listReceivedByAccount","params":[6,true],"id":1}`,
			unmarshalled: &btcjson.ListReceivedByAccountCmd{
				MinConf:          btcjson.Int(6),
				IncludeEmpty:     btcjson.Bool(true),
				IncludeWatchOnly: btcjson.Bool(false),
			},
		},
		{
			name: "listReceivedByAccount optional3",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("listReceivedByAccount", 6, true, false)
			},
			staticCmd: func() interface{} {
				return btcjson.NewListReceivedByAccountCmd(btcjson.Int(6), btcjson.Bool(true), btcjson.Bool(false))
			},
			marshalled: `{"jsonrpc":"1.0","method":"listReceivedByAccount","params":[6,true,false],"id":1}`,
			unmarshalled: &btcjson.ListReceivedByAccountCmd{
				MinConf:          btcjson.Int(6),
				IncludeEmpty:     btcjson.Bool(true),
				IncludeWatchOnly: btcjson.Bool(false),
			},
		},
		{
			name: "listReceivedByAddress",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("listReceivedByAddress")
			},
			staticCmd: func() interface{} {
				return btcjson.NewListReceivedByAddressCmd(nil, nil, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"listReceivedByAddress","params":[],"id":1}`,
			unmarshalled: &btcjson.ListReceivedByAddressCmd{
				MinConf:          btcjson.Int(1),
				IncludeEmpty:     btcjson.Bool(false),
				IncludeWatchOnly: btcjson.Bool(false),
			},
		},
		{
			name: "listReceivedByAddress optional1",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("listReceivedByAddress", 6)
			},
			staticCmd: func() interface{} {
				return btcjson.NewListReceivedByAddressCmd(btcjson.Int(6), nil, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"listReceivedByAddress","params":[6],"id":1}`,
			unmarshalled: &btcjson.ListReceivedByAddressCmd{
				MinConf:          btcjson.Int(6),
				IncludeEmpty:     btcjson.Bool(false),
				IncludeWatchOnly: btcjson.Bool(false),
			},
		},
		{
			name: "listReceivedByAddress optional2",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("listReceivedByAddress", 6, true)
			},
			staticCmd: func() interface{} {
				return btcjson.NewListReceivedByAddressCmd(btcjson.Int(6), btcjson.Bool(true), nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"listReceivedByAddress","params":[6,true],"id":1}`,
			unmarshalled: &btcjson.ListReceivedByAddressCmd{
				MinConf:          btcjson.Int(6),
				IncludeEmpty:     btcjson.Bool(true),
				IncludeWatchOnly: btcjson.Bool(false),
			},
		},
		{
			name: "listReceivedByAddress optional3",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("listReceivedByAddress", 6, true, false)
			},
			staticCmd: func() interface{} {
				return btcjson.NewListReceivedByAddressCmd(btcjson.Int(6), btcjson.Bool(true), btcjson.Bool(false))
			},
			marshalled: `{"jsonrpc":"1.0","method":"listReceivedByAddress","params":[6,true,false],"id":1}`,
			unmarshalled: &btcjson.ListReceivedByAddressCmd{
				MinConf:          btcjson.Int(6),
				IncludeEmpty:     btcjson.Bool(true),
				IncludeWatchOnly: btcjson.Bool(false),
			},
		},
		{
			name: "listSinceBlock",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("listSinceBlock")
			},
			staticCmd: func() interface{} {
				return btcjson.NewListSinceBlockCmd(nil, nil, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"listSinceBlock","params":[],"id":1}`,
			unmarshalled: &btcjson.ListSinceBlockCmd{
				BlockHash:           nil,
				TargetConfirmations: btcjson.Int(1),
				IncludeWatchOnly:    btcjson.Bool(false),
			},
		},
		{
			name: "listSinceBlock optional1",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("listSinceBlock", "123")
			},
			staticCmd: func() interface{} {
				return btcjson.NewListSinceBlockCmd(btcjson.String("123"), nil, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"listSinceBlock","params":["123"],"id":1}`,
			unmarshalled: &btcjson.ListSinceBlockCmd{
				BlockHash:           btcjson.String("123"),
				TargetConfirmations: btcjson.Int(1),
				IncludeWatchOnly:    btcjson.Bool(false),
			},
		},
		{
			name: "listSinceBlock optional2",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("listSinceBlock", "123", 6)
			},
			staticCmd: func() interface{} {
				return btcjson.NewListSinceBlockCmd(btcjson.String("123"), btcjson.Int(6), nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"listSinceBlock","params":["123",6],"id":1}`,
			unmarshalled: &btcjson.ListSinceBlockCmd{
				BlockHash:           btcjson.String("123"),
				TargetConfirmations: btcjson.Int(6),
				IncludeWatchOnly:    btcjson.Bool(false),
			},
		},
		{
			name: "listSinceBlock optional3",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("listSinceBlock", "123", 6, true)
			},
			staticCmd: func() interface{} {
				return btcjson.NewListSinceBlockCmd(btcjson.String("123"), btcjson.Int(6), btcjson.Bool(true))
			},
			marshalled: `{"jsonrpc":"1.0","method":"listSinceBlock","params":["123",6,true],"id":1}`,
			unmarshalled: &btcjson.ListSinceBlockCmd{
				BlockHash:           btcjson.String("123"),
				TargetConfirmations: btcjson.Int(6),
				IncludeWatchOnly:    btcjson.Bool(true),
			},
		},
		{
			name: "listTransactions",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("listTransactions")
			},
			staticCmd: func() interface{} {
				return btcjson.NewListTransactionsCmd(nil, nil, nil, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"listTransactions","params":[],"id":1}`,
			unmarshalled: &btcjson.ListTransactionsCmd{
				Account:          nil,
				Count:            btcjson.Int(10),
				From:             btcjson.Int(0),
				IncludeWatchOnly: btcjson.Bool(false),
			},
		},
		{
			name: "listTransactions optional1",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("listTransactions", "acct")
			},
			staticCmd: func() interface{} {
				return btcjson.NewListTransactionsCmd(btcjson.String("acct"), nil, nil, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"listTransactions","params":["acct"],"id":1}`,
			unmarshalled: &btcjson.ListTransactionsCmd{
				Account:          btcjson.String("acct"),
				Count:            btcjson.Int(10),
				From:             btcjson.Int(0),
				IncludeWatchOnly: btcjson.Bool(false),
			},
		},
		{
			name: "listTransactions optional2",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("listTransactions", "acct", 20)
			},
			staticCmd: func() interface{} {
				return btcjson.NewListTransactionsCmd(btcjson.String("acct"), btcjson.Int(20), nil, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"listTransactions","params":["acct",20],"id":1}`,
			unmarshalled: &btcjson.ListTransactionsCmd{
				Account:          btcjson.String("acct"),
				Count:            btcjson.Int(20),
				From:             btcjson.Int(0),
				IncludeWatchOnly: btcjson.Bool(false),
			},
		},
		{
			name: "listTransactions optional3",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("listTransactions", "acct", 20, 1)
			},
			staticCmd: func() interface{} {
				return btcjson.NewListTransactionsCmd(btcjson.String("acct"), btcjson.Int(20),
					btcjson.Int(1), nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"listTransactions","params":["acct",20,1],"id":1}`,
			unmarshalled: &btcjson.ListTransactionsCmd{
				Account:          btcjson.String("acct"),
				Count:            btcjson.Int(20),
				From:             btcjson.Int(1),
				IncludeWatchOnly: btcjson.Bool(false),
			},
		},
		{
			name: "listTransactions optional4",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("listTransactions", "acct", 20, 1, true)
			},
			staticCmd: func() interface{} {
				return btcjson.NewListTransactionsCmd(btcjson.String("acct"), btcjson.Int(20),
					btcjson.Int(1), btcjson.Bool(true))
			},
			marshalled: `{"jsonrpc":"1.0","method":"listTransactions","params":["acct",20,1,true],"id":1}`,
			unmarshalled: &btcjson.ListTransactionsCmd{
				Account:          btcjson.String("acct"),
				Count:            btcjson.Int(20),
				From:             btcjson.Int(1),
				IncludeWatchOnly: btcjson.Bool(true),
			},
		},
		{
			name: "listUnspent",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("listUnspent")
			},
			staticCmd: func() interface{} {
				return btcjson.NewListUnspentCmd(nil, nil, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"listUnspent","params":[],"id":1}`,
			unmarshalled: &btcjson.ListUnspentCmd{
				MinConf:   btcjson.Int(1),
				MaxConf:   btcjson.Int(9999999),
				Addresses: nil,
			},
		},
		{
			name: "listUnspent optional1",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("listUnspent", 6)
			},
			staticCmd: func() interface{} {
				return btcjson.NewListUnspentCmd(btcjson.Int(6), nil, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"listUnspent","params":[6],"id":1}`,
			unmarshalled: &btcjson.ListUnspentCmd{
				MinConf:   btcjson.Int(6),
				MaxConf:   btcjson.Int(9999999),
				Addresses: nil,
			},
		},
		{
			name: "listUnspent optional2",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("listUnspent", 6, 100)
			},
			staticCmd: func() interface{} {
				return btcjson.NewListUnspentCmd(btcjson.Int(6), btcjson.Int(100), nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"listUnspent","params":[6,100],"id":1}`,
			unmarshalled: &btcjson.ListUnspentCmd{
				MinConf:   btcjson.Int(6),
				MaxConf:   btcjson.Int(100),
				Addresses: nil,
			},
		},
		{
			name: "listUnspent optional3",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("listUnspent", 6, 100, []string{"1Address", "1Address2"})
			},
			staticCmd: func() interface{} {
				return btcjson.NewListUnspentCmd(btcjson.Int(6), btcjson.Int(100),
					&[]string{"1Address", "1Address2"})
			},
			marshalled: `{"jsonrpc":"1.0","method":"listUnspent","params":[6,100,["1Address","1Address2"]],"id":1}`,
			unmarshalled: &btcjson.ListUnspentCmd{
				MinConf:   btcjson.Int(6),
				MaxConf:   btcjson.Int(100),
				Addresses: &[]string{"1Address", "1Address2"},
			},
		},
		{
			name: "lockUnspent",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("lockUnspent", true, `[{"txid":"123","vout":1}]`)
			},
			staticCmd: func() interface{} {
				txInputs := []btcjson.TransactionInput{
					{Txid: "123", Vout: 1},
				}
				return btcjson.NewLockUnspentCmd(true, txInputs)
			},
			marshalled: `{"jsonrpc":"1.0","method":"lockUnspent","params":[true,[{"txid":"123","vout":1}]],"id":1}`,
			unmarshalled: &btcjson.LockUnspentCmd{
				Unlock: true,
				Transactions: []btcjson.TransactionInput{
					{Txid: "123", Vout: 1},
				},
			},
		},
		{
			name: "move",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("move", "from", "to", 0.5)
			},
			staticCmd: func() interface{} {
				return btcjson.NewMoveCmd("from", "to", 0.5, nil, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"move","params":["from","to",0.5],"id":1}`,
			unmarshalled: &btcjson.MoveCmd{
				FromAccount: "from",
				ToAccount:   "to",
				Amount:      0.5,
				MinConf:     btcjson.Int(1),
				Comment:     nil,
			},
		},
		{
			name: "move optional1",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("move", "from", "to", 0.5, 6)
			},
			staticCmd: func() interface{} {
				return btcjson.NewMoveCmd("from", "to", 0.5, btcjson.Int(6), nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"move","params":["from","to",0.5,6],"id":1}`,
			unmarshalled: &btcjson.MoveCmd{
				FromAccount: "from",
				ToAccount:   "to",
				Amount:      0.5,
				MinConf:     btcjson.Int(6),
				Comment:     nil,
			},
		},
		{
			name: "move optional2",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("move", "from", "to", 0.5, 6, "comment")
			},
			staticCmd: func() interface{} {
				return btcjson.NewMoveCmd("from", "to", 0.5, btcjson.Int(6), btcjson.String("comment"))
			},
			marshalled: `{"jsonrpc":"1.0","method":"move","params":["from","to",0.5,6,"comment"],"id":1}`,
			unmarshalled: &btcjson.MoveCmd{
				FromAccount: "from",
				ToAccount:   "to",
				Amount:      0.5,
				MinConf:     btcjson.Int(6),
				Comment:     btcjson.String("comment"),
			},
		},
		{
			name: "sendFrom",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("sendFrom", "from", "1Address", 0.5)
			},
			staticCmd: func() interface{} {
				return btcjson.NewSendFromCmd("from", "1Address", 0.5, nil, nil, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"sendFrom","params":["from","1Address",0.5],"id":1}`,
			unmarshalled: &btcjson.SendFromCmd{
				FromAccount: "from",
				ToAddress:   "1Address",
				Amount:      0.5,
				MinConf:     btcjson.Int(1),
				Comment:     nil,
				CommentTo:   nil,
			},
		},
		{
			name: "sendFrom optional1",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("sendFrom", "from", "1Address", 0.5, 6)
			},
			staticCmd: func() interface{} {
				return btcjson.NewSendFromCmd("from", "1Address", 0.5, btcjson.Int(6), nil, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"sendFrom","params":["from","1Address",0.5,6],"id":1}`,
			unmarshalled: &btcjson.SendFromCmd{
				FromAccount: "from",
				ToAddress:   "1Address",
				Amount:      0.5,
				MinConf:     btcjson.Int(6),
				Comment:     nil,
				CommentTo:   nil,
			},
		},
		{
			name: "sendFrom optional2",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("sendFrom", "from", "1Address", 0.5, 6, "comment")
			},
			staticCmd: func() interface{} {
				return btcjson.NewSendFromCmd("from", "1Address", 0.5, btcjson.Int(6),
					btcjson.String("comment"), nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"sendFrom","params":["from","1Address",0.5,6,"comment"],"id":1}`,
			unmarshalled: &btcjson.SendFromCmd{
				FromAccount: "from",
				ToAddress:   "1Address",
				Amount:      0.5,
				MinConf:     btcjson.Int(6),
				Comment:     btcjson.String("comment"),
				CommentTo:   nil,
			},
		},
		{
			name: "sendFrom optional3",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("sendFrom", "from", "1Address", 0.5, 6, "comment", "commentto")
			},
			staticCmd: func() interface{} {
				return btcjson.NewSendFromCmd("from", "1Address", 0.5, btcjson.Int(6),
					btcjson.String("comment"), btcjson.String("commentto"))
			},
			marshalled: `{"jsonrpc":"1.0","method":"sendFrom","params":["from","1Address",0.5,6,"comment","commentto"],"id":1}`,
			unmarshalled: &btcjson.SendFromCmd{
				FromAccount: "from",
				ToAddress:   "1Address",
				Amount:      0.5,
				MinConf:     btcjson.Int(6),
				Comment:     btcjson.String("comment"),
				CommentTo:   btcjson.String("commentto"),
			},
		},
		{
			name: "sendMany",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("sendMany", "from", `{"1Address":0.5}`)
			},
			staticCmd: func() interface{} {
				amounts := map[string]float64{"1Address": 0.5}
				return btcjson.NewSendManyCmd("from", amounts, nil, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"sendMany","params":["from",{"1Address":0.5}],"id":1}`,
			unmarshalled: &btcjson.SendManyCmd{
				FromAccount: "from",
				Amounts:     map[string]float64{"1Address": 0.5},
				MinConf:     btcjson.Int(1),
				Comment:     nil,
			},
		},
		{
			name: "sendMany optional1",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("sendMany", "from", `{"1Address":0.5}`, 6)
			},
			staticCmd: func() interface{} {
				amounts := map[string]float64{"1Address": 0.5}
				return btcjson.NewSendManyCmd("from", amounts, btcjson.Int(6), nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"sendMany","params":["from",{"1Address":0.5},6],"id":1}`,
			unmarshalled: &btcjson.SendManyCmd{
				FromAccount: "from",
				Amounts:     map[string]float64{"1Address": 0.5},
				MinConf:     btcjson.Int(6),
				Comment:     nil,
			},
		},
		{
			name: "sendMany optional2",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("sendMany", "from", `{"1Address":0.5}`, 6, "comment")
			},
			staticCmd: func() interface{} {
				amounts := map[string]float64{"1Address": 0.5}
				return btcjson.NewSendManyCmd("from", amounts, btcjson.Int(6), btcjson.String("comment"))
			},
			marshalled: `{"jsonrpc":"1.0","method":"sendMany","params":["from",{"1Address":0.5},6,"comment"],"id":1}`,
			unmarshalled: &btcjson.SendManyCmd{
				FromAccount: "from",
				Amounts:     map[string]float64{"1Address": 0.5},
				MinConf:     btcjson.Int(6),
				Comment:     btcjson.String("comment"),
			},
		},
		{
			name: "sendToAddress",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("sendToAddress", "1Address", 0.5)
			},
			staticCmd: func() interface{} {
				return btcjson.NewSendToAddressCmd("1Address", 0.5, nil, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"sendToAddress","params":["1Address",0.5],"id":1}`,
			unmarshalled: &btcjson.SendToAddressCmd{
				Address:   "1Address",
				Amount:    0.5,
				Comment:   nil,
				CommentTo: nil,
			},
		},
		{
			name: "sendToAddress optional1",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("sendToAddress", "1Address", 0.5, "comment", "commentto")
			},
			staticCmd: func() interface{} {
				return btcjson.NewSendToAddressCmd("1Address", 0.5, btcjson.String("comment"),
					btcjson.String("commentto"))
			},
			marshalled: `{"jsonrpc":"1.0","method":"sendToAddress","params":["1Address",0.5,"comment","commentto"],"id":1}`,
			unmarshalled: &btcjson.SendToAddressCmd{
				Address:   "1Address",
				Amount:    0.5,
				Comment:   btcjson.String("comment"),
				CommentTo: btcjson.String("commentto"),
			},
		},
		{
			name: "setAccount",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("setAccount", "1Address", "acct")
			},
			staticCmd: func() interface{} {
				return btcjson.NewSetAccountCmd("1Address", "acct")
			},
			marshalled: `{"jsonrpc":"1.0","method":"setAccount","params":["1Address","acct"],"id":1}`,
			unmarshalled: &btcjson.SetAccountCmd{
				Address: "1Address",
				Account: "acct",
			},
		},
		{
			name: "setTxFee",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("setTxFee", 0.0001)
			},
			staticCmd: func() interface{} {
				return btcjson.NewSetTxFeeCmd(0.0001)
			},
			marshalled: `{"jsonrpc":"1.0","method":"setTxFee","params":[0.0001],"id":1}`,
			unmarshalled: &btcjson.SetTxFeeCmd{
				Amount: 0.0001,
			},
		},
		{
			name: "signMessage",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("signMessage", "1Address", "message")
			},
			staticCmd: func() interface{} {
				return btcjson.NewSignMessageCmd("1Address", "message")
			},
			marshalled: `{"jsonrpc":"1.0","method":"signMessage","params":["1Address","message"],"id":1}`,
			unmarshalled: &btcjson.SignMessageCmd{
				Address: "1Address",
				Message: "message",
			},
		},
		{
			name: "signRawTransaction",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("signRawTransaction", "001122")
			},
			staticCmd: func() interface{} {
				return btcjson.NewSignRawTransactionCmd("001122", nil, nil, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"signRawTransaction","params":["001122"],"id":1}`,
			unmarshalled: &btcjson.SignRawTransactionCmd{
				RawTx:    "001122",
				Inputs:   nil,
				PrivKeys: nil,
				Flags:    btcjson.String("ALL"),
			},
		},
		{
			name: "signRawTransaction optional1",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("signRawTransaction", "001122", `[{"txid":"123","vout":1,"scriptPubKey":"00","redeemScript":"01"}]`)
			},
			staticCmd: func() interface{} {
				txInputs := []btcjson.RawTxInput{
					{
						Txid:         "123",
						Vout:         1,
						ScriptPubKey: "00",
						RedeemScript: "01",
					},
				}

				return btcjson.NewSignRawTransactionCmd("001122", &txInputs, nil, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"signRawTransaction","params":["001122",[{"txid":"123","vout":1,"scriptPubKey":"00","redeemScript":"01"}]],"id":1}`,
			unmarshalled: &btcjson.SignRawTransactionCmd{
				RawTx: "001122",
				Inputs: &[]btcjson.RawTxInput{
					{
						Txid:         "123",
						Vout:         1,
						ScriptPubKey: "00",
						RedeemScript: "01",
					},
				},
				PrivKeys: nil,
				Flags:    btcjson.String("ALL"),
			},
		},
		{
			name: "signRawTransaction optional2",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("signRawTransaction", "001122", `[]`, `["abc"]`)
			},
			staticCmd: func() interface{} {
				txInputs := []btcjson.RawTxInput{}
				privKeys := []string{"abc"}
				return btcjson.NewSignRawTransactionCmd("001122", &txInputs, &privKeys, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"signRawTransaction","params":["001122",[],["abc"]],"id":1}`,
			unmarshalled: &btcjson.SignRawTransactionCmd{
				RawTx:    "001122",
				Inputs:   &[]btcjson.RawTxInput{},
				PrivKeys: &[]string{"abc"},
				Flags:    btcjson.String("ALL"),
			},
		},
		{
			name: "signRawTransaction optional3",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("signRawTransaction", "001122", `[]`, `[]`, "ALL")
			},
			staticCmd: func() interface{} {
				txInputs := []btcjson.RawTxInput{}
				privKeys := []string{}
				return btcjson.NewSignRawTransactionCmd("001122", &txInputs, &privKeys,
					btcjson.String("ALL"))
			},
			marshalled: `{"jsonrpc":"1.0","method":"signRawTransaction","params":["001122",[],[],"ALL"],"id":1}`,
			unmarshalled: &btcjson.SignRawTransactionCmd{
				RawTx:    "001122",
				Inputs:   &[]btcjson.RawTxInput{},
				PrivKeys: &[]string{},
				Flags:    btcjson.String("ALL"),
			},
		},
		{
			name: "walletLock",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("walletLock")
			},
			staticCmd: func() interface{} {
				return btcjson.NewWalletLockCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"walletLock","params":[],"id":1}`,
			unmarshalled: &btcjson.WalletLockCmd{},
		},
		{
			name: "walletPassphrase",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("walletPassphrase", "pass", 60)
			},
			staticCmd: func() interface{} {
				return btcjson.NewWalletPassphraseCmd("pass", 60)
			},
			marshalled: `{"jsonrpc":"1.0","method":"walletPassphrase","params":["pass",60],"id":1}`,
			unmarshalled: &btcjson.WalletPassphraseCmd{
				Passphrase: "pass",
				Timeout:    60,
			},
		},
		{
			name: "walletPassphraseChange",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("walletPassphraseChange", "old", "new")
			},
			staticCmd: func() interface{} {
				return btcjson.NewWalletPassphraseChangeCmd("old", "new")
			},
			marshalled: `{"jsonrpc":"1.0","method":"walletPassphraseChange","params":["old","new"],"id":1}`,
			unmarshalled: &btcjson.WalletPassphraseChangeCmd{
				OldPassphrase: "old",
				NewPassphrase: "new",
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
