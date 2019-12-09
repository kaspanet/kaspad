// Copyright (c) 2014 The btcsuite developers
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
	"github.com/kaspanet/kaspad/wire"
)

// TestDAGSvrCmds tests all of the kaspa rpc server commands marshal and unmarshal
// into valid results include handling of optional fields being omitted in the
// marshalled command, while optional fields with defaults have the default
// assigned on unmarshalled commands.
func TestDAGSvrCmds(t *testing.T) {
	t.Parallel()

	testID := 1
	tests := []struct {
		name         string
		newCmd       func() (interface{}, error)
		staticCmd    func() interface{}
		marshalled   string
		unmarshalled interface{}
	}{
		{
			name: "addManualNode",
			newCmd: func() (interface{}, error) {
				return kaspajson.NewCmd("addManualNode", "127.0.0.1")
			},
			staticCmd: func() interface{} {
				return kaspajson.NewAddManualNodeCmd("127.0.0.1", nil)
			},
			marshalled:   `{"jsonrpc":"1.0","method":"addManualNode","params":["127.0.0.1"],"id":1}`,
			unmarshalled: &kaspajson.AddManualNodeCmd{Addr: "127.0.0.1", OneTry: kaspajson.Bool(false)},
		},
		{
			name: "createRawTransaction",
			newCmd: func() (interface{}, error) {
				return kaspajson.NewCmd("createRawTransaction", `[{"txId":"123","vout":1}]`,
					`{"456":0.0123}`)
			},
			staticCmd: func() interface{} {
				txInputs := []kaspajson.TransactionInput{
					{TxID: "123", Vout: 1},
				}
				amounts := map[string]float64{"456": .0123}
				return kaspajson.NewCreateRawTransactionCmd(txInputs, amounts, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"createRawTransaction","params":[[{"txId":"123","vout":1}],{"456":0.0123}],"id":1}`,
			unmarshalled: &kaspajson.CreateRawTransactionCmd{
				Inputs:  []kaspajson.TransactionInput{{TxID: "123", Vout: 1}},
				Amounts: map[string]float64{"456": .0123},
			},
		},
		{
			name: "createRawTransaction optional",
			newCmd: func() (interface{}, error) {
				return kaspajson.NewCmd("createRawTransaction", `[{"txId":"123","vout":1}]`,
					`{"456":0.0123}`, int64(12312333333))
			},
			staticCmd: func() interface{} {
				txInputs := []kaspajson.TransactionInput{
					{TxID: "123", Vout: 1},
				}
				amounts := map[string]float64{"456": .0123}
				return kaspajson.NewCreateRawTransactionCmd(txInputs, amounts, kaspajson.Uint64(12312333333))
			},
			marshalled: `{"jsonrpc":"1.0","method":"createRawTransaction","params":[[{"txId":"123","vout":1}],{"456":0.0123},12312333333],"id":1}`,
			unmarshalled: &kaspajson.CreateRawTransactionCmd{
				Inputs:   []kaspajson.TransactionInput{{TxID: "123", Vout: 1}},
				Amounts:  map[string]float64{"456": .0123},
				LockTime: kaspajson.Uint64(12312333333),
			},
		},

		{
			name: "decodeRawTransaction",
			newCmd: func() (interface{}, error) {
				return kaspajson.NewCmd("decodeRawTransaction", "123")
			},
			staticCmd: func() interface{} {
				return kaspajson.NewDecodeRawTransactionCmd("123")
			},
			marshalled:   `{"jsonrpc":"1.0","method":"decodeRawTransaction","params":["123"],"id":1}`,
			unmarshalled: &kaspajson.DecodeRawTransactionCmd{HexTx: "123"},
		},
		{
			name: "decodeScript",
			newCmd: func() (interface{}, error) {
				return kaspajson.NewCmd("decodeScript", "00")
			},
			staticCmd: func() interface{} {
				return kaspajson.NewDecodeScriptCmd("00")
			},
			marshalled:   `{"jsonrpc":"1.0","method":"decodeScript","params":["00"],"id":1}`,
			unmarshalled: &kaspajson.DecodeScriptCmd{HexScript: "00"},
		},
		{
			name: "getAllManualNodesInfo",
			newCmd: func() (interface{}, error) {
				return kaspajson.NewCmd("getAllManualNodesInfo")
			},
			staticCmd: func() interface{} {
				return kaspajson.NewGetAllManualNodesInfoCmd(nil)
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getAllManualNodesInfo","params":[],"id":1}`,
			unmarshalled: &kaspajson.GetAllManualNodesInfoCmd{Details: kaspajson.Bool(true)},
		},
		{
			name: "getSelectedTipHash",
			newCmd: func() (interface{}, error) {
				return kaspajson.NewCmd("getSelectedTipHash")
			},
			staticCmd: func() interface{} {
				return kaspajson.NewGetSelectedTipHashCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getSelectedTipHash","params":[],"id":1}`,
			unmarshalled: &kaspajson.GetSelectedTipHashCmd{},
		},
		{
			name: "getBlock",
			newCmd: func() (interface{}, error) {
				return kaspajson.NewCmd("getBlock", "123")
			},
			staticCmd: func() interface{} {
				return kaspajson.NewGetBlockCmd("123", nil, nil, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getBlock","params":["123"],"id":1}`,
			unmarshalled: &kaspajson.GetBlockCmd{
				Hash:      "123",
				Verbose:   kaspajson.Bool(true),
				VerboseTx: kaspajson.Bool(false),
			},
		},
		{
			name: "getBlock required optional1",
			newCmd: func() (interface{}, error) {
				// Intentionally use a source param that is
				// more pointers than the destination to
				// exercise that path.
				verbosePtr := kaspajson.Bool(true)
				return kaspajson.NewCmd("getBlock", "123", &verbosePtr)
			},
			staticCmd: func() interface{} {
				return kaspajson.NewGetBlockCmd("123", kaspajson.Bool(true), nil, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getBlock","params":["123",true],"id":1}`,
			unmarshalled: &kaspajson.GetBlockCmd{
				Hash:      "123",
				Verbose:   kaspajson.Bool(true),
				VerboseTx: kaspajson.Bool(false),
			},
		},
		{
			name: "getBlock required optional2",
			newCmd: func() (interface{}, error) {
				return kaspajson.NewCmd("getBlock", "123", true, true)
			},
			staticCmd: func() interface{} {
				return kaspajson.NewGetBlockCmd("123", kaspajson.Bool(true), kaspajson.Bool(true), nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getBlock","params":["123",true,true],"id":1}`,
			unmarshalled: &kaspajson.GetBlockCmd{
				Hash:      "123",
				Verbose:   kaspajson.Bool(true),
				VerboseTx: kaspajson.Bool(true),
			},
		},
		{
			name: "getBlock required optional3",
			newCmd: func() (interface{}, error) {
				return kaspajson.NewCmd("getBlock", "123", true, true, "456")
			},
			staticCmd: func() interface{} {
				return kaspajson.NewGetBlockCmd("123", kaspajson.Bool(true), kaspajson.Bool(true), kaspajson.String("456"))
			},
			marshalled: `{"jsonrpc":"1.0","method":"getBlock","params":["123",true,true,"456"],"id":1}`,
			unmarshalled: &kaspajson.GetBlockCmd{
				Hash:       "123",
				Verbose:    kaspajson.Bool(true),
				VerboseTx:  kaspajson.Bool(true),
				Subnetwork: kaspajson.String("456"),
			},
		},
		{
			name: "getBlocks",
			newCmd: func() (interface{}, error) {
				return kaspajson.NewCmd("getBlocks", true, true, "123")
			},
			staticCmd: func() interface{} {
				return kaspajson.NewGetBlocksCmd(true, true, kaspajson.String("123"))
			},
			marshalled: `{"jsonrpc":"1.0","method":"getBlocks","params":[true,true,"123"],"id":1}`,
			unmarshalled: &kaspajson.GetBlocksCmd{
				IncludeRawBlockData:     true,
				IncludeVerboseBlockData: true,
				StartHash:               kaspajson.String("123"),
			},
		},
		{
			name: "getBlockDagInfo",
			newCmd: func() (interface{}, error) {
				return kaspajson.NewCmd("getBlockDagInfo")
			},
			staticCmd: func() interface{} {
				return kaspajson.NewGetBlockDAGInfoCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getBlockDagInfo","params":[],"id":1}`,
			unmarshalled: &kaspajson.GetBlockDAGInfoCmd{},
		},
		{
			name: "getBlockCount",
			newCmd: func() (interface{}, error) {
				return kaspajson.NewCmd("getBlockCount")
			},
			staticCmd: func() interface{} {
				return kaspajson.NewGetBlockCountCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getBlockCount","params":[],"id":1}`,
			unmarshalled: &kaspajson.GetBlockCountCmd{},
		},
		{
			name: "getBlockHeader",
			newCmd: func() (interface{}, error) {
				return kaspajson.NewCmd("getBlockHeader", "123")
			},
			staticCmd: func() interface{} {
				return kaspajson.NewGetBlockHeaderCmd("123", nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getBlockHeader","params":["123"],"id":1}`,
			unmarshalled: &kaspajson.GetBlockHeaderCmd{
				Hash:    "123",
				Verbose: kaspajson.Bool(true),
			},
		},
		{
			name: "getBlockTemplate",
			newCmd: func() (interface{}, error) {
				return kaspajson.NewCmd("getBlockTemplate")
			},
			staticCmd: func() interface{} {
				return kaspajson.NewGetBlockTemplateCmd(nil)
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getBlockTemplate","params":[],"id":1}`,
			unmarshalled: &kaspajson.GetBlockTemplateCmd{Request: nil},
		},
		{
			name: "getBlockTemplate optional - template request",
			newCmd: func() (interface{}, error) {
				return kaspajson.NewCmd("getBlockTemplate", `{"mode":"template","capabilities":["longpoll","coinbasetxn"]}`)
			},
			staticCmd: func() interface{} {
				template := kaspajson.TemplateRequest{
					Mode:         "template",
					Capabilities: []string{"longpoll", "coinbasetxn"},
				}
				return kaspajson.NewGetBlockTemplateCmd(&template)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getBlockTemplate","params":[{"mode":"template","capabilities":["longpoll","coinbasetxn"]}],"id":1}`,
			unmarshalled: &kaspajson.GetBlockTemplateCmd{
				Request: &kaspajson.TemplateRequest{
					Mode:         "template",
					Capabilities: []string{"longpoll", "coinbasetxn"},
				},
			},
		},
		{
			name: "getBlockTemplate optional - template request with tweaks",
			newCmd: func() (interface{}, error) {
				return kaspajson.NewCmd("getBlockTemplate", `{"mode":"template","capabilities":["longPoll","coinbaseTxn"],"sigOpLimit":500,"massLimit":100000000,"maxVersion":1}`)
			},
			staticCmd: func() interface{} {
				template := kaspajson.TemplateRequest{
					Mode:         "template",
					Capabilities: []string{"longPoll", "coinbaseTxn"},
					SigOpLimit:   500,
					MassLimit:    100000000,
					MaxVersion:   1,
				}
				return kaspajson.NewGetBlockTemplateCmd(&template)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getBlockTemplate","params":[{"mode":"template","capabilities":["longPoll","coinbaseTxn"],"sigOpLimit":500,"massLimit":100000000,"maxVersion":1}],"id":1}`,
			unmarshalled: &kaspajson.GetBlockTemplateCmd{
				Request: &kaspajson.TemplateRequest{
					Mode:         "template",
					Capabilities: []string{"longPoll", "coinbaseTxn"},
					SigOpLimit:   int64(500),
					MassLimit:    int64(100000000),
					MaxVersion:   1,
				},
			},
		},
		{
			name: "getBlockTemplate optional - template request with tweaks 2",
			newCmd: func() (interface{}, error) {
				return kaspajson.NewCmd("getBlockTemplate", `{"mode":"template","capabilities":["longPoll","coinbaseTxn"],"sigOpLimit":true,"massLimit":100000000,"maxVersion":1}`)
			},
			staticCmd: func() interface{} {
				template := kaspajson.TemplateRequest{
					Mode:         "template",
					Capabilities: []string{"longPoll", "coinbaseTxn"},
					SigOpLimit:   true,
					MassLimit:    100000000,
					MaxVersion:   1,
				}
				return kaspajson.NewGetBlockTemplateCmd(&template)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getBlockTemplate","params":[{"mode":"template","capabilities":["longPoll","coinbaseTxn"],"sigOpLimit":true,"massLimit":100000000,"maxVersion":1}],"id":1}`,
			unmarshalled: &kaspajson.GetBlockTemplateCmd{
				Request: &kaspajson.TemplateRequest{
					Mode:         "template",
					Capabilities: []string{"longPoll", "coinbaseTxn"},
					SigOpLimit:   true,
					MassLimit:    int64(100000000),
					MaxVersion:   1,
				},
			},
		},
		{
			name: "getCFilter",
			newCmd: func() (interface{}, error) {
				return kaspajson.NewCmd("getCFilter", "123",
					wire.GCSFilterExtended)
			},
			staticCmd: func() interface{} {
				return kaspajson.NewGetCFilterCmd("123",
					wire.GCSFilterExtended)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getCFilter","params":["123",1],"id":1}`,
			unmarshalled: &kaspajson.GetCFilterCmd{
				Hash:       "123",
				FilterType: wire.GCSFilterExtended,
			},
		},
		{
			name: "getCFilterHeader",
			newCmd: func() (interface{}, error) {
				return kaspajson.NewCmd("getCFilterHeader", "123",
					wire.GCSFilterExtended)
			},
			staticCmd: func() interface{} {
				return kaspajson.NewGetCFilterHeaderCmd("123",
					wire.GCSFilterExtended)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getCFilterHeader","params":["123",1],"id":1}`,
			unmarshalled: &kaspajson.GetCFilterHeaderCmd{
				Hash:       "123",
				FilterType: wire.GCSFilterExtended,
			},
		},
		{
			name: "getChainFromBlock",
			newCmd: func() (interface{}, error) {
				return kaspajson.NewCmd("getChainFromBlock", true, "123")
			},
			staticCmd: func() interface{} {
				return kaspajson.NewGetChainFromBlockCmd(true, kaspajson.String("123"))
			},
			marshalled: `{"jsonrpc":"1.0","method":"getChainFromBlock","params":[true,"123"],"id":1}`,
			unmarshalled: &kaspajson.GetChainFromBlockCmd{
				IncludeBlocks: true,
				StartHash:     kaspajson.String("123"),
			},
		},
		{
			name: "getDagTips",
			newCmd: func() (interface{}, error) {
				return kaspajson.NewCmd("getDagTips")
			},
			staticCmd: func() interface{} {
				return kaspajson.NewGetDAGTipsCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getDagTips","params":[],"id":1}`,
			unmarshalled: &kaspajson.GetDAGTipsCmd{},
		},
		{
			name: "getConnectionCount",
			newCmd: func() (interface{}, error) {
				return kaspajson.NewCmd("getConnectionCount")
			},
			staticCmd: func() interface{} {
				return kaspajson.NewGetConnectionCountCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getConnectionCount","params":[],"id":1}`,
			unmarshalled: &kaspajson.GetConnectionCountCmd{},
		},
		{
			name: "getDifficulty",
			newCmd: func() (interface{}, error) {
				return kaspajson.NewCmd("getDifficulty")
			},
			staticCmd: func() interface{} {
				return kaspajson.NewGetDifficultyCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getDifficulty","params":[],"id":1}`,
			unmarshalled: &kaspajson.GetDifficultyCmd{},
		},
		{
			name: "getGenerate",
			newCmd: func() (interface{}, error) {
				return kaspajson.NewCmd("getGenerate")
			},
			staticCmd: func() interface{} {
				return kaspajson.NewGetGenerateCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getGenerate","params":[],"id":1}`,
			unmarshalled: &kaspajson.GetGenerateCmd{},
		},
		{
			name: "getHashesPerSec",
			newCmd: func() (interface{}, error) {
				return kaspajson.NewCmd("getHashesPerSec")
			},
			staticCmd: func() interface{} {
				return kaspajson.NewGetHashesPerSecCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getHashesPerSec","params":[],"id":1}`,
			unmarshalled: &kaspajson.GetHashesPerSecCmd{},
		},
		{
			name: "getInfo",
			newCmd: func() (interface{}, error) {
				return kaspajson.NewCmd("getInfo")
			},
			staticCmd: func() interface{} {
				return kaspajson.NewGetInfoCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getInfo","params":[],"id":1}`,
			unmarshalled: &kaspajson.GetInfoCmd{},
		},
		{
			name: "getManualNodeInfo",
			newCmd: func() (interface{}, error) {
				return kaspajson.NewCmd("getManualNodeInfo", "127.0.0.1")
			},
			staticCmd: func() interface{} {
				return kaspajson.NewGetManualNodeInfoCmd("127.0.0.1", nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getManualNodeInfo","params":["127.0.0.1"],"id":1}`,
			unmarshalled: &kaspajson.GetManualNodeInfoCmd{
				Node:    "127.0.0.1",
				Details: kaspajson.Bool(true),
			},
		},
		{
			name: "getMempoolEntry",
			newCmd: func() (interface{}, error) {
				return kaspajson.NewCmd("getMempoolEntry", "txhash")
			},
			staticCmd: func() interface{} {
				return kaspajson.NewGetMempoolEntryCmd("txhash")
			},
			marshalled: `{"jsonrpc":"1.0","method":"getMempoolEntry","params":["txhash"],"id":1}`,
			unmarshalled: &kaspajson.GetMempoolEntryCmd{
				TxID: "txhash",
			},
		},
		{
			name: "getMempoolInfo",
			newCmd: func() (interface{}, error) {
				return kaspajson.NewCmd("getMempoolInfo")
			},
			staticCmd: func() interface{} {
				return kaspajson.NewGetMempoolInfoCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getMempoolInfo","params":[],"id":1}`,
			unmarshalled: &kaspajson.GetMempoolInfoCmd{},
		},
		{
			name: "getMiningInfo",
			newCmd: func() (interface{}, error) {
				return kaspajson.NewCmd("getMiningInfo")
			},
			staticCmd: func() interface{} {
				return kaspajson.NewGetMiningInfoCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getMiningInfo","params":[],"id":1}`,
			unmarshalled: &kaspajson.GetMiningInfoCmd{},
		},
		{
			name: "getNetworkInfo",
			newCmd: func() (interface{}, error) {
				return kaspajson.NewCmd("getNetworkInfo")
			},
			staticCmd: func() interface{} {
				return kaspajson.NewGetNetworkInfoCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getNetworkInfo","params":[],"id":1}`,
			unmarshalled: &kaspajson.GetNetworkInfoCmd{},
		},
		{
			name: "getNetTotals",
			newCmd: func() (interface{}, error) {
				return kaspajson.NewCmd("getNetTotals")
			},
			staticCmd: func() interface{} {
				return kaspajson.NewGetNetTotalsCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getNetTotals","params":[],"id":1}`,
			unmarshalled: &kaspajson.GetNetTotalsCmd{},
		},
		{
			name: "getNetworkHashPs",
			newCmd: func() (interface{}, error) {
				return kaspajson.NewCmd("getNetworkHashPs")
			},
			staticCmd: func() interface{} {
				return kaspajson.NewGetNetworkHashPSCmd(nil, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getNetworkHashPs","params":[],"id":1}`,
			unmarshalled: &kaspajson.GetNetworkHashPSCmd{
				Blocks: kaspajson.Int(120),
				Height: kaspajson.Int(-1),
			},
		},
		{
			name: "getNetworkHashPs optional1",
			newCmd: func() (interface{}, error) {
				return kaspajson.NewCmd("getNetworkHashPs", 200)
			},
			staticCmd: func() interface{} {
				return kaspajson.NewGetNetworkHashPSCmd(kaspajson.Int(200), nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getNetworkHashPs","params":[200],"id":1}`,
			unmarshalled: &kaspajson.GetNetworkHashPSCmd{
				Blocks: kaspajson.Int(200),
				Height: kaspajson.Int(-1),
			},
		},
		{
			name: "getNetworkHashPs optional2",
			newCmd: func() (interface{}, error) {
				return kaspajson.NewCmd("getNetworkHashPs", 200, 123)
			},
			staticCmd: func() interface{} {
				return kaspajson.NewGetNetworkHashPSCmd(kaspajson.Int(200), kaspajson.Int(123))
			},
			marshalled: `{"jsonrpc":"1.0","method":"getNetworkHashPs","params":[200,123],"id":1}`,
			unmarshalled: &kaspajson.GetNetworkHashPSCmd{
				Blocks: kaspajson.Int(200),
				Height: kaspajson.Int(123),
			},
		},
		{
			name: "getPeerInfo",
			newCmd: func() (interface{}, error) {
				return kaspajson.NewCmd("getPeerInfo")
			},
			staticCmd: func() interface{} {
				return kaspajson.NewGetPeerInfoCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getPeerInfo","params":[],"id":1}`,
			unmarshalled: &kaspajson.GetPeerInfoCmd{},
		},
		{
			name: "getRawMempool",
			newCmd: func() (interface{}, error) {
				return kaspajson.NewCmd("getRawMempool")
			},
			staticCmd: func() interface{} {
				return kaspajson.NewGetRawMempoolCmd(nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getRawMempool","params":[],"id":1}`,
			unmarshalled: &kaspajson.GetRawMempoolCmd{
				Verbose: kaspajson.Bool(false),
			},
		},
		{
			name: "getRawMempool optional",
			newCmd: func() (interface{}, error) {
				return kaspajson.NewCmd("getRawMempool", false)
			},
			staticCmd: func() interface{} {
				return kaspajson.NewGetRawMempoolCmd(kaspajson.Bool(false))
			},
			marshalled: `{"jsonrpc":"1.0","method":"getRawMempool","params":[false],"id":1}`,
			unmarshalled: &kaspajson.GetRawMempoolCmd{
				Verbose: kaspajson.Bool(false),
			},
		},
		{
			name: "getRawTransaction",
			newCmd: func() (interface{}, error) {
				return kaspajson.NewCmd("getRawTransaction", "123")
			},
			staticCmd: func() interface{} {
				return kaspajson.NewGetRawTransactionCmd("123", nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getRawTransaction","params":["123"],"id":1}`,
			unmarshalled: &kaspajson.GetRawTransactionCmd{
				TxID:    "123",
				Verbose: kaspajson.Int(0),
			},
		},
		{
			name: "getRawTransaction optional",
			newCmd: func() (interface{}, error) {
				return kaspajson.NewCmd("getRawTransaction", "123", 1)
			},
			staticCmd: func() interface{} {
				return kaspajson.NewGetRawTransactionCmd("123", kaspajson.Int(1))
			},
			marshalled: `{"jsonrpc":"1.0","method":"getRawTransaction","params":["123",1],"id":1}`,
			unmarshalled: &kaspajson.GetRawTransactionCmd{
				TxID:    "123",
				Verbose: kaspajson.Int(1),
			},
		},
		{
			name: "getSubnetwork",
			newCmd: func() (interface{}, error) {
				return kaspajson.NewCmd("getSubnetwork", "123")
			},
			staticCmd: func() interface{} {
				return kaspajson.NewGetSubnetworkCmd("123")
			},
			marshalled: `{"jsonrpc":"1.0","method":"getSubnetwork","params":["123"],"id":1}`,
			unmarshalled: &kaspajson.GetSubnetworkCmd{
				SubnetworkID: "123",
			},
		},
		{
			name: "getTxOut",
			newCmd: func() (interface{}, error) {
				return kaspajson.NewCmd("getTxOut", "123", 1)
			},
			staticCmd: func() interface{} {
				return kaspajson.NewGetTxOutCmd("123", 1, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getTxOut","params":["123",1],"id":1}`,
			unmarshalled: &kaspajson.GetTxOutCmd{
				TxID:           "123",
				Vout:           1,
				IncludeMempool: kaspajson.Bool(true),
			},
		},
		{
			name: "getTxOut optional",
			newCmd: func() (interface{}, error) {
				return kaspajson.NewCmd("getTxOut", "123", 1, true)
			},
			staticCmd: func() interface{} {
				return kaspajson.NewGetTxOutCmd("123", 1, kaspajson.Bool(true))
			},
			marshalled: `{"jsonrpc":"1.0","method":"getTxOut","params":["123",1,true],"id":1}`,
			unmarshalled: &kaspajson.GetTxOutCmd{
				TxID:           "123",
				Vout:           1,
				IncludeMempool: kaspajson.Bool(true),
			},
		},
		{
			name: "getTxOutSetInfo",
			newCmd: func() (interface{}, error) {
				return kaspajson.NewCmd("getTxOutSetInfo")
			},
			staticCmd: func() interface{} {
				return kaspajson.NewGetTxOutSetInfoCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getTxOutSetInfo","params":[],"id":1}`,
			unmarshalled: &kaspajson.GetTxOutSetInfoCmd{},
		},
		{
			name: "help",
			newCmd: func() (interface{}, error) {
				return kaspajson.NewCmd("help")
			},
			staticCmd: func() interface{} {
				return kaspajson.NewHelpCmd(nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"help","params":[],"id":1}`,
			unmarshalled: &kaspajson.HelpCmd{
				Command: nil,
			},
		},
		{
			name: "help optional",
			newCmd: func() (interface{}, error) {
				return kaspajson.NewCmd("help", "getBlock")
			},
			staticCmd: func() interface{} {
				return kaspajson.NewHelpCmd(kaspajson.String("getBlock"))
			},
			marshalled: `{"jsonrpc":"1.0","method":"help","params":["getBlock"],"id":1}`,
			unmarshalled: &kaspajson.HelpCmd{
				Command: kaspajson.String("getBlock"),
			},
		},
		{
			name: "invalidateBlock",
			newCmd: func() (interface{}, error) {
				return kaspajson.NewCmd("invalidateBlock", "123")
			},
			staticCmd: func() interface{} {
				return kaspajson.NewInvalidateBlockCmd("123")
			},
			marshalled: `{"jsonrpc":"1.0","method":"invalidateBlock","params":["123"],"id":1}`,
			unmarshalled: &kaspajson.InvalidateBlockCmd{
				BlockHash: "123",
			},
		},
		{
			name: "ping",
			newCmd: func() (interface{}, error) {
				return kaspajson.NewCmd("ping")
			},
			staticCmd: func() interface{} {
				return kaspajson.NewPingCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"ping","params":[],"id":1}`,
			unmarshalled: &kaspajson.PingCmd{},
		},
		{
			name: "preciousBlock",
			newCmd: func() (interface{}, error) {
				return kaspajson.NewCmd("preciousBlock", "0123")
			},
			staticCmd: func() interface{} {
				return kaspajson.NewPreciousBlockCmd("0123")
			},
			marshalled: `{"jsonrpc":"1.0","method":"preciousBlock","params":["0123"],"id":1}`,
			unmarshalled: &kaspajson.PreciousBlockCmd{
				BlockHash: "0123",
			},
		},
		{
			name: "reconsiderBlock",
			newCmd: func() (interface{}, error) {
				return kaspajson.NewCmd("reconsiderBlock", "123")
			},
			staticCmd: func() interface{} {
				return kaspajson.NewReconsiderBlockCmd("123")
			},
			marshalled: `{"jsonrpc":"1.0","method":"reconsiderBlock","params":["123"],"id":1}`,
			unmarshalled: &kaspajson.ReconsiderBlockCmd{
				BlockHash: "123",
			},
		},
		{
			name: "removeManualNode",
			newCmd: func() (interface{}, error) {
				return kaspajson.NewCmd("removeManualNode", "127.0.0.1")
			},
			staticCmd: func() interface{} {
				return kaspajson.NewRemoveManualNodeCmd("127.0.0.1")
			},
			marshalled:   `{"jsonrpc":"1.0","method":"removeManualNode","params":["127.0.0.1"],"id":1}`,
			unmarshalled: &kaspajson.RemoveManualNodeCmd{Addr: "127.0.0.1"},
		},
		{
			name: "searchRawTransactions",
			newCmd: func() (interface{}, error) {
				return kaspajson.NewCmd("searchRawTransactions", "1Address")
			},
			staticCmd: func() interface{} {
				return kaspajson.NewSearchRawTransactionsCmd("1Address", nil, nil, nil, nil, nil, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"searchRawTransactions","params":["1Address"],"id":1}`,
			unmarshalled: &kaspajson.SearchRawTransactionsCmd{
				Address:     "1Address",
				Verbose:     kaspajson.Bool(true),
				Skip:        kaspajson.Int(0),
				Count:       kaspajson.Int(100),
				VinExtra:    kaspajson.Bool(false),
				Reverse:     kaspajson.Bool(false),
				FilterAddrs: nil,
			},
		},
		{
			name: "searchRawTransactions",
			newCmd: func() (interface{}, error) {
				return kaspajson.NewCmd("searchRawTransactions", "1Address", false)
			},
			staticCmd: func() interface{} {
				return kaspajson.NewSearchRawTransactionsCmd("1Address",
					kaspajson.Bool(false), nil, nil, nil, nil, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"searchRawTransactions","params":["1Address",false],"id":1}`,
			unmarshalled: &kaspajson.SearchRawTransactionsCmd{
				Address:     "1Address",
				Verbose:     kaspajson.Bool(false),
				Skip:        kaspajson.Int(0),
				Count:       kaspajson.Int(100),
				VinExtra:    kaspajson.Bool(false),
				Reverse:     kaspajson.Bool(false),
				FilterAddrs: nil,
			},
		},
		{
			name: "searchRawTransactions",
			newCmd: func() (interface{}, error) {
				return kaspajson.NewCmd("searchRawTransactions", "1Address", false, 5)
			},
			staticCmd: func() interface{} {
				return kaspajson.NewSearchRawTransactionsCmd("1Address",
					kaspajson.Bool(false), kaspajson.Int(5), nil, nil, nil, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"searchRawTransactions","params":["1Address",false,5],"id":1}`,
			unmarshalled: &kaspajson.SearchRawTransactionsCmd{
				Address:     "1Address",
				Verbose:     kaspajson.Bool(false),
				Skip:        kaspajson.Int(5),
				Count:       kaspajson.Int(100),
				VinExtra:    kaspajson.Bool(false),
				Reverse:     kaspajson.Bool(false),
				FilterAddrs: nil,
			},
		},
		{
			name: "searchRawTransactions",
			newCmd: func() (interface{}, error) {
				return kaspajson.NewCmd("searchRawTransactions", "1Address", false, 5, 10)
			},
			staticCmd: func() interface{} {
				return kaspajson.NewSearchRawTransactionsCmd("1Address",
					kaspajson.Bool(false), kaspajson.Int(5), kaspajson.Int(10), nil, nil, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"searchRawTransactions","params":["1Address",false,5,10],"id":1}`,
			unmarshalled: &kaspajson.SearchRawTransactionsCmd{
				Address:     "1Address",
				Verbose:     kaspajson.Bool(false),
				Skip:        kaspajson.Int(5),
				Count:       kaspajson.Int(10),
				VinExtra:    kaspajson.Bool(false),
				Reverse:     kaspajson.Bool(false),
				FilterAddrs: nil,
			},
		},
		{
			name: "searchRawTransactions",
			newCmd: func() (interface{}, error) {
				return kaspajson.NewCmd("searchRawTransactions", "1Address", false, 5, 10, true)
			},
			staticCmd: func() interface{} {
				return kaspajson.NewSearchRawTransactionsCmd("1Address",
					kaspajson.Bool(false), kaspajson.Int(5), kaspajson.Int(10), kaspajson.Bool(true), nil, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"searchRawTransactions","params":["1Address",false,5,10,true],"id":1}`,
			unmarshalled: &kaspajson.SearchRawTransactionsCmd{
				Address:     "1Address",
				Verbose:     kaspajson.Bool(false),
				Skip:        kaspajson.Int(5),
				Count:       kaspajson.Int(10),
				VinExtra:    kaspajson.Bool(true),
				Reverse:     kaspajson.Bool(false),
				FilterAddrs: nil,
			},
		},
		{
			name: "searchRawTransactions",
			newCmd: func() (interface{}, error) {
				return kaspajson.NewCmd("searchRawTransactions", "1Address", false, 5, 10, true, true)
			},
			staticCmd: func() interface{} {
				return kaspajson.NewSearchRawTransactionsCmd("1Address",
					kaspajson.Bool(false), kaspajson.Int(5), kaspajson.Int(10), kaspajson.Bool(true), kaspajson.Bool(true), nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"searchRawTransactions","params":["1Address",false,5,10,true,true],"id":1}`,
			unmarshalled: &kaspajson.SearchRawTransactionsCmd{
				Address:     "1Address",
				Verbose:     kaspajson.Bool(false),
				Skip:        kaspajson.Int(5),
				Count:       kaspajson.Int(10),
				VinExtra:    kaspajson.Bool(true),
				Reverse:     kaspajson.Bool(true),
				FilterAddrs: nil,
			},
		},
		{
			name: "searchRawTransactions",
			newCmd: func() (interface{}, error) {
				return kaspajson.NewCmd("searchRawTransactions", "1Address", false, 5, 10, true, true, []string{"1Address"})
			},
			staticCmd: func() interface{} {
				return kaspajson.NewSearchRawTransactionsCmd("1Address",
					kaspajson.Bool(false), kaspajson.Int(5), kaspajson.Int(10), kaspajson.Bool(true), kaspajson.Bool(true), &[]string{"1Address"})
			},
			marshalled: `{"jsonrpc":"1.0","method":"searchRawTransactions","params":["1Address",false,5,10,true,true,["1Address"]],"id":1}`,
			unmarshalled: &kaspajson.SearchRawTransactionsCmd{
				Address:     "1Address",
				Verbose:     kaspajson.Bool(false),
				Skip:        kaspajson.Int(5),
				Count:       kaspajson.Int(10),
				VinExtra:    kaspajson.Bool(true),
				Reverse:     kaspajson.Bool(true),
				FilterAddrs: &[]string{"1Address"},
			},
		},
		{
			name: "sendRawTransaction",
			newCmd: func() (interface{}, error) {
				return kaspajson.NewCmd("sendRawTransaction", "1122")
			},
			staticCmd: func() interface{} {
				return kaspajson.NewSendRawTransactionCmd("1122", nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"sendRawTransaction","params":["1122"],"id":1}`,
			unmarshalled: &kaspajson.SendRawTransactionCmd{
				HexTx:         "1122",
				AllowHighFees: kaspajson.Bool(false),
			},
		},
		{
			name: "sendRawTransaction optional",
			newCmd: func() (interface{}, error) {
				return kaspajson.NewCmd("sendRawTransaction", "1122", false)
			},
			staticCmd: func() interface{} {
				return kaspajson.NewSendRawTransactionCmd("1122", kaspajson.Bool(false))
			},
			marshalled: `{"jsonrpc":"1.0","method":"sendRawTransaction","params":["1122",false],"id":1}`,
			unmarshalled: &kaspajson.SendRawTransactionCmd{
				HexTx:         "1122",
				AllowHighFees: kaspajson.Bool(false),
			},
		},
		{
			name: "setGenerate",
			newCmd: func() (interface{}, error) {
				return kaspajson.NewCmd("setGenerate", true)
			},
			staticCmd: func() interface{} {
				return kaspajson.NewSetGenerateCmd(true, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"setGenerate","params":[true],"id":1}`,
			unmarshalled: &kaspajson.SetGenerateCmd{
				Generate:     true,
				GenProcLimit: kaspajson.Int(-1),
			},
		},
		{
			name: "setGenerate optional",
			newCmd: func() (interface{}, error) {
				return kaspajson.NewCmd("setGenerate", true, 6)
			},
			staticCmd: func() interface{} {
				return kaspajson.NewSetGenerateCmd(true, kaspajson.Int(6))
			},
			marshalled: `{"jsonrpc":"1.0","method":"setGenerate","params":[true,6],"id":1}`,
			unmarshalled: &kaspajson.SetGenerateCmd{
				Generate:     true,
				GenProcLimit: kaspajson.Int(6),
			},
		},
		{
			name: "stop",
			newCmd: func() (interface{}, error) {
				return kaspajson.NewCmd("stop")
			},
			staticCmd: func() interface{} {
				return kaspajson.NewStopCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"stop","params":[],"id":1}`,
			unmarshalled: &kaspajson.StopCmd{},
		},
		{
			name: "submitBlock",
			newCmd: func() (interface{}, error) {
				return kaspajson.NewCmd("submitBlock", "112233")
			},
			staticCmd: func() interface{} {
				return kaspajson.NewSubmitBlockCmd("112233", nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"submitBlock","params":["112233"],"id":1}`,
			unmarshalled: &kaspajson.SubmitBlockCmd{
				HexBlock: "112233",
				Options:  nil,
			},
		},
		{
			name: "submitBlock optional",
			newCmd: func() (interface{}, error) {
				return kaspajson.NewCmd("submitBlock", "112233", `{"workId":"12345"}`)
			},
			staticCmd: func() interface{} {
				options := kaspajson.SubmitBlockOptions{
					WorkID: "12345",
				}
				return kaspajson.NewSubmitBlockCmd("112233", &options)
			},
			marshalled: `{"jsonrpc":"1.0","method":"submitBlock","params":["112233",{"workId":"12345"}],"id":1}`,
			unmarshalled: &kaspajson.SubmitBlockCmd{
				HexBlock: "112233",
				Options: &kaspajson.SubmitBlockOptions{
					WorkID: "12345",
				},
			},
		},
		{
			name: "uptime",
			newCmd: func() (interface{}, error) {
				return kaspajson.NewCmd("uptime")
			},
			staticCmd: func() interface{} {
				return kaspajson.NewUptimeCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"uptime","params":[],"id":1}`,
			unmarshalled: &kaspajson.UptimeCmd{},
		},
		{
			name: "validateAddress",
			newCmd: func() (interface{}, error) {
				return kaspajson.NewCmd("validateAddress", "1Address")
			},
			staticCmd: func() interface{} {
				return kaspajson.NewValidateAddressCmd("1Address")
			},
			marshalled: `{"jsonrpc":"1.0","method":"validateAddress","params":["1Address"],"id":1}`,
			unmarshalled: &kaspajson.ValidateAddressCmd{
				Address: "1Address",
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
			t.Errorf("\n%s\n%s", marshalled, test.marshalled)
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

// TestDAGSvrCmdErrors ensures any errors that occur in the command during
// custom mashal and unmarshal are as expected.
func TestDAGSvrCmdErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		result     interface{}
		marshalled string
		err        error
	}{
		{
			name:       "template request with invalid type",
			result:     &kaspajson.TemplateRequest{},
			marshalled: `{"mode":1}`,
			err:        &json.UnmarshalTypeError{},
		},
		{
			name:       "invalid template request sigoplimit field",
			result:     &kaspajson.TemplateRequest{},
			marshalled: `{"sigoplimit":"invalid"}`,
			err:        kaspajson.Error{ErrorCode: kaspajson.ErrInvalidType},
		},
		{
			name:       "invalid template request masslimit field",
			result:     &kaspajson.TemplateRequest{},
			marshalled: `{"masslimit":"invalid"}`,
			err:        kaspajson.Error{ErrorCode: kaspajson.ErrInvalidType},
		},
	}

	t.Logf("Running %d tests", len(tests))
	for i, test := range tests {
		err := json.Unmarshal([]byte(test.marshalled), &test.result)
		if reflect.TypeOf(err) != reflect.TypeOf(test.err) {
			t.Errorf("Test #%d (%s) wrong error - got %T (%[2]v), "+
				"want %T", i, test.name, err, test.err)
			continue
		}

		if terr, ok := test.err.(kaspajson.Error); ok {
			gotErrorCode := err.(kaspajson.Error).ErrorCode
			if gotErrorCode != terr.ErrorCode {
				t.Errorf("Test #%d (%s) mismatched error code "+
					"- got %v (%v), want %v", i, test.name,
					gotErrorCode, terr, terr.ErrorCode)
				continue
			}
		}
	}
}

// TestBtcdExtCmds tests all of the btcd extended commands marshal and unmarshal
// into valid results include handling of optional fields being omitted in the
// marshalled command, while optional fields with defaults have the default
// assigned on unmarshalled commands.
func TestBtcdExtCmds(t *testing.T) {
	t.Parallel()

	testID := 1
	tests := []struct {
		name         string
		newCmd       func() (interface{}, error)
		staticCmd    func() interface{}
		marshalled   string
		unmarshalled interface{}
	}{
		{
			name: "debugLevel",
			newCmd: func() (interface{}, error) {
				return kaspajson.NewCmd("debugLevel", "trace")
			},
			staticCmd: func() interface{} {
				return kaspajson.NewDebugLevelCmd("trace")
			},
			marshalled: `{"jsonrpc":"1.0","method":"debugLevel","params":["trace"],"id":1}`,
			unmarshalled: &kaspajson.DebugLevelCmd{
				LevelSpec: "trace",
			},
		},
		{
			name: "node",
			newCmd: func() (interface{}, error) {
				return kaspajson.NewCmd("node", kaspajson.NRemove, "1.1.1.1")
			},
			staticCmd: func() interface{} {
				return kaspajson.NewNodeCmd("remove", "1.1.1.1", nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"node","params":["remove","1.1.1.1"],"id":1}`,
			unmarshalled: &kaspajson.NodeCmd{
				SubCmd: kaspajson.NRemove,
				Target: "1.1.1.1",
			},
		},
		{
			name: "node",
			newCmd: func() (interface{}, error) {
				return kaspajson.NewCmd("node", kaspajson.NDisconnect, "1.1.1.1")
			},
			staticCmd: func() interface{} {
				return kaspajson.NewNodeCmd("disconnect", "1.1.1.1", nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"node","params":["disconnect","1.1.1.1"],"id":1}`,
			unmarshalled: &kaspajson.NodeCmd{
				SubCmd: kaspajson.NDisconnect,
				Target: "1.1.1.1",
			},
		},
		{
			name: "node",
			newCmd: func() (interface{}, error) {
				return kaspajson.NewCmd("node", kaspajson.NConnect, "1.1.1.1", "perm")
			},
			staticCmd: func() interface{} {
				return kaspajson.NewNodeCmd("connect", "1.1.1.1", kaspajson.String("perm"))
			},
			marshalled: `{"jsonrpc":"1.0","method":"node","params":["connect","1.1.1.1","perm"],"id":1}`,
			unmarshalled: &kaspajson.NodeCmd{
				SubCmd:        kaspajson.NConnect,
				Target:        "1.1.1.1",
				ConnectSubCmd: kaspajson.String("perm"),
			},
		},
		{
			name: "node",
			newCmd: func() (interface{}, error) {
				return kaspajson.NewCmd("node", kaspajson.NConnect, "1.1.1.1", "temp")
			},
			staticCmd: func() interface{} {
				return kaspajson.NewNodeCmd("connect", "1.1.1.1", kaspajson.String("temp"))
			},
			marshalled: `{"jsonrpc":"1.0","method":"node","params":["connect","1.1.1.1","temp"],"id":1}`,
			unmarshalled: &kaspajson.NodeCmd{
				SubCmd:        kaspajson.NConnect,
				Target:        "1.1.1.1",
				ConnectSubCmd: kaspajson.String("temp"),
			},
		},
		{
			name: "generate",
			newCmd: func() (interface{}, error) {
				return kaspajson.NewCmd("generate", 1)
			},
			staticCmd: func() interface{} {
				return kaspajson.NewGenerateCmd(1)
			},
			marshalled: `{"jsonrpc":"1.0","method":"generate","params":[1],"id":1}`,
			unmarshalled: &kaspajson.GenerateCmd{
				NumBlocks: 1,
			},
		},
		{
			name: "getSelectedTip",
			newCmd: func() (interface{}, error) {
				return kaspajson.NewCmd("getSelectedTip")
			},
			staticCmd: func() interface{} {
				return kaspajson.NewGetSelectedTipCmd(nil, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getSelectedTip","params":[],"id":1}`,
			unmarshalled: &kaspajson.GetSelectedTipCmd{
				Verbose:   kaspajson.Bool(true),
				VerboseTx: kaspajson.Bool(false),
			},
		},
		{
			name: "getCurrentNet",
			newCmd: func() (interface{}, error) {
				return kaspajson.NewCmd("getCurrentNet")
			},
			staticCmd: func() interface{} {
				return kaspajson.NewGetCurrentNetCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getCurrentNet","params":[],"id":1}`,
			unmarshalled: &kaspajson.GetCurrentNetCmd{},
		},
		{
			name: "getHeaders",
			newCmd: func() (interface{}, error) {
				return kaspajson.NewCmd("getHeaders", "", "")
			},
			staticCmd: func() interface{} {
				return kaspajson.NewGetHeadersCmd(
					"",
					"",
				)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getHeaders","params":["",""],"id":1}`,
			unmarshalled: &kaspajson.GetHeadersCmd{
				StartHash: "",
				StopHash:  "",
			},
		},
		{
			name: "getHeaders - with arguments",
			newCmd: func() (interface{}, error) {
				return kaspajson.NewCmd("getHeaders", "000000000000000001f1739002418e2f9a84c47a4fd2a0eb7a787a6b7dc12f16", "000000000000000000ba33b33e1fad70b69e234fc24414dd47113bff38f523f7")
			},
			staticCmd: func() interface{} {
				return kaspajson.NewGetHeadersCmd(
					"000000000000000001f1739002418e2f9a84c47a4fd2a0eb7a787a6b7dc12f16",
					"000000000000000000ba33b33e1fad70b69e234fc24414dd47113bff38f523f7",
				)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getHeaders","params":["000000000000000001f1739002418e2f9a84c47a4fd2a0eb7a787a6b7dc12f16","000000000000000000ba33b33e1fad70b69e234fc24414dd47113bff38f523f7"],"id":1}`,
			unmarshalled: &kaspajson.GetHeadersCmd{
				StartHash: "000000000000000001f1739002418e2f9a84c47a4fd2a0eb7a787a6b7dc12f16",
				StopHash:  "000000000000000000ba33b33e1fad70b69e234fc24414dd47113bff38f523f7",
			},
		},
		{
			name: "getTopHeaders",
			newCmd: func() (interface{}, error) {
				return kaspajson.NewCmd("getTopHeaders")
			},
			staticCmd: func() interface{} {
				return kaspajson.NewGetTopHeadersCmd(
					nil,
				)
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getTopHeaders","params":[],"id":1}`,
			unmarshalled: &kaspajson.GetTopHeadersCmd{},
		},
		{
			name: "getTopHeaders - with start hash",
			newCmd: func() (interface{}, error) {
				return kaspajson.NewCmd("getTopHeaders", "000000000000000000ba33b33e1fad70b69e234fc24414dd47113bff38f523f7")
			},
			staticCmd: func() interface{} {
				return kaspajson.NewGetTopHeadersCmd(
					kaspajson.String("000000000000000000ba33b33e1fad70b69e234fc24414dd47113bff38f523f7"),
				)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getTopHeaders","params":["000000000000000000ba33b33e1fad70b69e234fc24414dd47113bff38f523f7"],"id":1}`,
			unmarshalled: &kaspajson.GetTopHeadersCmd{
				StartHash: kaspajson.String("000000000000000000ba33b33e1fad70b69e234fc24414dd47113bff38f523f7"),
			},
		},
		{
			name: "version",
			newCmd: func() (interface{}, error) {
				return kaspajson.NewCmd("version")
			},
			staticCmd: func() interface{} {
				return kaspajson.NewVersionCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"version","params":[],"id":1}`,
			unmarshalled: &kaspajson.VersionCmd{},
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
