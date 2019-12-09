// Copyright (c) 2014 The btcsuite developers
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
	"github.com/kaspanet/kaspad/wire"
)

// TestRPCServerCommands tests all of the kaspa rpc server commands marshal and unmarshal
// into valid results include handling of optional fields being omitted in the
// marshalled command, while optional fields with defaults have the default
// assigned on unmarshalled commands.
func TestRPCServerCommands(t *testing.T) {
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
				return jsonrpc.NewCommand("addManualNode", "127.0.0.1")
			},
			staticCmd: func() interface{} {
				return jsonrpc.NewAddManualNodeCmd("127.0.0.1", nil)
			},
			marshalled:   `{"jsonrpc":"1.0","method":"addManualNode","params":["127.0.0.1"],"id":1}`,
			unmarshalled: &jsonrpc.AddManualNodeCmd{Addr: "127.0.0.1", OneTry: jsonrpc.Bool(false)},
		},
		{
			name: "createRawTransaction",
			newCmd: func() (interface{}, error) {
				return jsonrpc.NewCommand("createRawTransaction", `[{"txId":"123","vout":1}]`,
					`{"456":0.0123}`)
			},
			staticCmd: func() interface{} {
				txInputs := []jsonrpc.TransactionInput{
					{TxID: "123", Vout: 1},
				}
				amounts := map[string]float64{"456": .0123}
				return jsonrpc.NewCreateRawTransactionCmd(txInputs, amounts, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"createRawTransaction","params":[[{"txId":"123","vout":1}],{"456":0.0123}],"id":1}`,
			unmarshalled: &jsonrpc.CreateRawTransactionCmd{
				Inputs:  []jsonrpc.TransactionInput{{TxID: "123", Vout: 1}},
				Amounts: map[string]float64{"456": .0123},
			},
		},
		{
			name: "createRawTransaction optional",
			newCmd: func() (interface{}, error) {
				return jsonrpc.NewCommand("createRawTransaction", `[{"txId":"123","vout":1}]`,
					`{"456":0.0123}`, int64(12312333333))
			},
			staticCmd: func() interface{} {
				txInputs := []jsonrpc.TransactionInput{
					{TxID: "123", Vout: 1},
				}
				amounts := map[string]float64{"456": .0123}
				return jsonrpc.NewCreateRawTransactionCmd(txInputs, amounts, jsonrpc.Uint64(12312333333))
			},
			marshalled: `{"jsonrpc":"1.0","method":"createRawTransaction","params":[[{"txId":"123","vout":1}],{"456":0.0123},12312333333],"id":1}`,
			unmarshalled: &jsonrpc.CreateRawTransactionCmd{
				Inputs:   []jsonrpc.TransactionInput{{TxID: "123", Vout: 1}},
				Amounts:  map[string]float64{"456": .0123},
				LockTime: jsonrpc.Uint64(12312333333),
			},
		},

		{
			name: "decodeRawTransaction",
			newCmd: func() (interface{}, error) {
				return jsonrpc.NewCommand("decodeRawTransaction", "123")
			},
			staticCmd: func() interface{} {
				return jsonrpc.NewDecodeRawTransactionCmd("123")
			},
			marshalled:   `{"jsonrpc":"1.0","method":"decodeRawTransaction","params":["123"],"id":1}`,
			unmarshalled: &jsonrpc.DecodeRawTransactionCmd{HexTx: "123"},
		},
		{
			name: "decodeScript",
			newCmd: func() (interface{}, error) {
				return jsonrpc.NewCommand("decodeScript", "00")
			},
			staticCmd: func() interface{} {
				return jsonrpc.NewDecodeScriptCmd("00")
			},
			marshalled:   `{"jsonrpc":"1.0","method":"decodeScript","params":["00"],"id":1}`,
			unmarshalled: &jsonrpc.DecodeScriptCmd{HexScript: "00"},
		},
		{
			name: "getAllManualNodesInfo",
			newCmd: func() (interface{}, error) {
				return jsonrpc.NewCommand("getAllManualNodesInfo")
			},
			staticCmd: func() interface{} {
				return jsonrpc.NewGetAllManualNodesInfoCmd(nil)
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getAllManualNodesInfo","params":[],"id":1}`,
			unmarshalled: &jsonrpc.GetAllManualNodesInfoCmd{Details: jsonrpc.Bool(true)},
		},
		{
			name: "getSelectedTipHash",
			newCmd: func() (interface{}, error) {
				return jsonrpc.NewCommand("getSelectedTipHash")
			},
			staticCmd: func() interface{} {
				return jsonrpc.NewGetSelectedTipHashCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getSelectedTipHash","params":[],"id":1}`,
			unmarshalled: &jsonrpc.GetSelectedTipHashCmd{},
		},
		{
			name: "getBlock",
			newCmd: func() (interface{}, error) {
				return jsonrpc.NewCommand("getBlock", "123")
			},
			staticCmd: func() interface{} {
				return jsonrpc.NewGetBlockCmd("123", nil, nil, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getBlock","params":["123"],"id":1}`,
			unmarshalled: &jsonrpc.GetBlockCmd{
				Hash:      "123",
				Verbose:   jsonrpc.Bool(true),
				VerboseTx: jsonrpc.Bool(false),
			},
		},
		{
			name: "getBlock required optional1",
			newCmd: func() (interface{}, error) {
				// Intentionally use a source param that is
				// more pointers than the destination to
				// exercise that path.
				verbosePtr := jsonrpc.Bool(true)
				return jsonrpc.NewCommand("getBlock", "123", &verbosePtr)
			},
			staticCmd: func() interface{} {
				return jsonrpc.NewGetBlockCmd("123", jsonrpc.Bool(true), nil, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getBlock","params":["123",true],"id":1}`,
			unmarshalled: &jsonrpc.GetBlockCmd{
				Hash:      "123",
				Verbose:   jsonrpc.Bool(true),
				VerboseTx: jsonrpc.Bool(false),
			},
		},
		{
			name: "getBlock required optional2",
			newCmd: func() (interface{}, error) {
				return jsonrpc.NewCommand("getBlock", "123", true, true)
			},
			staticCmd: func() interface{} {
				return jsonrpc.NewGetBlockCmd("123", jsonrpc.Bool(true), jsonrpc.Bool(true), nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getBlock","params":["123",true,true],"id":1}`,
			unmarshalled: &jsonrpc.GetBlockCmd{
				Hash:      "123",
				Verbose:   jsonrpc.Bool(true),
				VerboseTx: jsonrpc.Bool(true),
			},
		},
		{
			name: "getBlock required optional3",
			newCmd: func() (interface{}, error) {
				return jsonrpc.NewCommand("getBlock", "123", true, true, "456")
			},
			staticCmd: func() interface{} {
				return jsonrpc.NewGetBlockCmd("123", jsonrpc.Bool(true), jsonrpc.Bool(true), jsonrpc.String("456"))
			},
			marshalled: `{"jsonrpc":"1.0","method":"getBlock","params":["123",true,true,"456"],"id":1}`,
			unmarshalled: &jsonrpc.GetBlockCmd{
				Hash:       "123",
				Verbose:    jsonrpc.Bool(true),
				VerboseTx:  jsonrpc.Bool(true),
				Subnetwork: jsonrpc.String("456"),
			},
		},
		{
			name: "getBlocks",
			newCmd: func() (interface{}, error) {
				return jsonrpc.NewCommand("getBlocks", true, true, "123")
			},
			staticCmd: func() interface{} {
				return jsonrpc.NewGetBlocksCmd(true, true, jsonrpc.String("123"))
			},
			marshalled: `{"jsonrpc":"1.0","method":"getBlocks","params":[true,true,"123"],"id":1}`,
			unmarshalled: &jsonrpc.GetBlocksCmd{
				IncludeRawBlockData:     true,
				IncludeVerboseBlockData: true,
				StartHash:               jsonrpc.String("123"),
			},
		},
		{
			name: "getBlockDagInfo",
			newCmd: func() (interface{}, error) {
				return jsonrpc.NewCommand("getBlockDagInfo")
			},
			staticCmd: func() interface{} {
				return jsonrpc.NewGetBlockDAGInfoCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getBlockDagInfo","params":[],"id":1}`,
			unmarshalled: &jsonrpc.GetBlockDAGInfoCmd{},
		},
		{
			name: "getBlockCount",
			newCmd: func() (interface{}, error) {
				return jsonrpc.NewCommand("getBlockCount")
			},
			staticCmd: func() interface{} {
				return jsonrpc.NewGetBlockCountCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getBlockCount","params":[],"id":1}`,
			unmarshalled: &jsonrpc.GetBlockCountCmd{},
		},
		{
			name: "getBlockHeader",
			newCmd: func() (interface{}, error) {
				return jsonrpc.NewCommand("getBlockHeader", "123")
			},
			staticCmd: func() interface{} {
				return jsonrpc.NewGetBlockHeaderCmd("123", nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getBlockHeader","params":["123"],"id":1}`,
			unmarshalled: &jsonrpc.GetBlockHeaderCmd{
				Hash:    "123",
				Verbose: jsonrpc.Bool(true),
			},
		},
		{
			name: "getBlockTemplate",
			newCmd: func() (interface{}, error) {
				return jsonrpc.NewCommand("getBlockTemplate")
			},
			staticCmd: func() interface{} {
				return jsonrpc.NewGetBlockTemplateCmd(nil)
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getBlockTemplate","params":[],"id":1}`,
			unmarshalled: &jsonrpc.GetBlockTemplateCmd{Request: nil},
		},
		{
			name: "getBlockTemplate optional - template request",
			newCmd: func() (interface{}, error) {
				return jsonrpc.NewCommand("getBlockTemplate", `{"mode":"template","capabilities":["longpoll","coinbasetxn"]}`)
			},
			staticCmd: func() interface{} {
				template := jsonrpc.TemplateRequest{
					Mode:         "template",
					Capabilities: []string{"longpoll", "coinbasetxn"},
				}
				return jsonrpc.NewGetBlockTemplateCmd(&template)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getBlockTemplate","params":[{"mode":"template","capabilities":["longpoll","coinbasetxn"]}],"id":1}`,
			unmarshalled: &jsonrpc.GetBlockTemplateCmd{
				Request: &jsonrpc.TemplateRequest{
					Mode:         "template",
					Capabilities: []string{"longpoll", "coinbasetxn"},
				},
			},
		},
		{
			name: "getBlockTemplate optional - template request with tweaks",
			newCmd: func() (interface{}, error) {
				return jsonrpc.NewCommand("getBlockTemplate", `{"mode":"template","capabilities":["longPoll","coinbaseTxn"],"sigOpLimit":500,"massLimit":100000000,"maxVersion":1}`)
			},
			staticCmd: func() interface{} {
				template := jsonrpc.TemplateRequest{
					Mode:         "template",
					Capabilities: []string{"longPoll", "coinbaseTxn"},
					SigOpLimit:   500,
					MassLimit:    100000000,
					MaxVersion:   1,
				}
				return jsonrpc.NewGetBlockTemplateCmd(&template)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getBlockTemplate","params":[{"mode":"template","capabilities":["longPoll","coinbaseTxn"],"sigOpLimit":500,"massLimit":100000000,"maxVersion":1}],"id":1}`,
			unmarshalled: &jsonrpc.GetBlockTemplateCmd{
				Request: &jsonrpc.TemplateRequest{
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
				return jsonrpc.NewCommand("getBlockTemplate", `{"mode":"template","capabilities":["longPoll","coinbaseTxn"],"sigOpLimit":true,"massLimit":100000000,"maxVersion":1}`)
			},
			staticCmd: func() interface{} {
				template := jsonrpc.TemplateRequest{
					Mode:         "template",
					Capabilities: []string{"longPoll", "coinbaseTxn"},
					SigOpLimit:   true,
					MassLimit:    100000000,
					MaxVersion:   1,
				}
				return jsonrpc.NewGetBlockTemplateCmd(&template)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getBlockTemplate","params":[{"mode":"template","capabilities":["longPoll","coinbaseTxn"],"sigOpLimit":true,"massLimit":100000000,"maxVersion":1}],"id":1}`,
			unmarshalled: &jsonrpc.GetBlockTemplateCmd{
				Request: &jsonrpc.TemplateRequest{
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
				return jsonrpc.NewCommand("getCFilter", "123",
					wire.GCSFilterExtended)
			},
			staticCmd: func() interface{} {
				return jsonrpc.NewGetCFilterCmd("123",
					wire.GCSFilterExtended)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getCFilter","params":["123",1],"id":1}`,
			unmarshalled: &jsonrpc.GetCFilterCmd{
				Hash:       "123",
				FilterType: wire.GCSFilterExtended,
			},
		},
		{
			name: "getCFilterHeader",
			newCmd: func() (interface{}, error) {
				return jsonrpc.NewCommand("getCFilterHeader", "123",
					wire.GCSFilterExtended)
			},
			staticCmd: func() interface{} {
				return jsonrpc.NewGetCFilterHeaderCmd("123",
					wire.GCSFilterExtended)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getCFilterHeader","params":["123",1],"id":1}`,
			unmarshalled: &jsonrpc.GetCFilterHeaderCmd{
				Hash:       "123",
				FilterType: wire.GCSFilterExtended,
			},
		},
		{
			name: "getChainFromBlock",
			newCmd: func() (interface{}, error) {
				return jsonrpc.NewCommand("getChainFromBlock", true, "123")
			},
			staticCmd: func() interface{} {
				return jsonrpc.NewGetChainFromBlockCmd(true, jsonrpc.String("123"))
			},
			marshalled: `{"jsonrpc":"1.0","method":"getChainFromBlock","params":[true,"123"],"id":1}`,
			unmarshalled: &jsonrpc.GetChainFromBlockCmd{
				IncludeBlocks: true,
				StartHash:     jsonrpc.String("123"),
			},
		},
		{
			name: "getDagTips",
			newCmd: func() (interface{}, error) {
				return jsonrpc.NewCommand("getDagTips")
			},
			staticCmd: func() interface{} {
				return jsonrpc.NewGetDAGTipsCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getDagTips","params":[],"id":1}`,
			unmarshalled: &jsonrpc.GetDAGTipsCmd{},
		},
		{
			name: "getConnectionCount",
			newCmd: func() (interface{}, error) {
				return jsonrpc.NewCommand("getConnectionCount")
			},
			staticCmd: func() interface{} {
				return jsonrpc.NewGetConnectionCountCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getConnectionCount","params":[],"id":1}`,
			unmarshalled: &jsonrpc.GetConnectionCountCmd{},
		},
		{
			name: "getDifficulty",
			newCmd: func() (interface{}, error) {
				return jsonrpc.NewCommand("getDifficulty")
			},
			staticCmd: func() interface{} {
				return jsonrpc.NewGetDifficultyCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getDifficulty","params":[],"id":1}`,
			unmarshalled: &jsonrpc.GetDifficultyCmd{},
		},
		{
			name: "getGenerate",
			newCmd: func() (interface{}, error) {
				return jsonrpc.NewCommand("getGenerate")
			},
			staticCmd: func() interface{} {
				return jsonrpc.NewGetGenerateCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getGenerate","params":[],"id":1}`,
			unmarshalled: &jsonrpc.GetGenerateCmd{},
		},
		{
			name: "getHashesPerSec",
			newCmd: func() (interface{}, error) {
				return jsonrpc.NewCommand("getHashesPerSec")
			},
			staticCmd: func() interface{} {
				return jsonrpc.NewGetHashesPerSecCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getHashesPerSec","params":[],"id":1}`,
			unmarshalled: &jsonrpc.GetHashesPerSecCmd{},
		},
		{
			name: "getInfo",
			newCmd: func() (interface{}, error) {
				return jsonrpc.NewCommand("getInfo")
			},
			staticCmd: func() interface{} {
				return jsonrpc.NewGetInfoCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getInfo","params":[],"id":1}`,
			unmarshalled: &jsonrpc.GetInfoCmd{},
		},
		{
			name: "getManualNodeInfo",
			newCmd: func() (interface{}, error) {
				return jsonrpc.NewCommand("getManualNodeInfo", "127.0.0.1")
			},
			staticCmd: func() interface{} {
				return jsonrpc.NewGetManualNodeInfoCmd("127.0.0.1", nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getManualNodeInfo","params":["127.0.0.1"],"id":1}`,
			unmarshalled: &jsonrpc.GetManualNodeInfoCmd{
				Node:    "127.0.0.1",
				Details: jsonrpc.Bool(true),
			},
		},
		{
			name: "getMempoolEntry",
			newCmd: func() (interface{}, error) {
				return jsonrpc.NewCommand("getMempoolEntry", "txhash")
			},
			staticCmd: func() interface{} {
				return jsonrpc.NewGetMempoolEntryCmd("txhash")
			},
			marshalled: `{"jsonrpc":"1.0","method":"getMempoolEntry","params":["txhash"],"id":1}`,
			unmarshalled: &jsonrpc.GetMempoolEntryCmd{
				TxID: "txhash",
			},
		},
		{
			name: "getMempoolInfo",
			newCmd: func() (interface{}, error) {
				return jsonrpc.NewCommand("getMempoolInfo")
			},
			staticCmd: func() interface{} {
				return jsonrpc.NewGetMempoolInfoCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getMempoolInfo","params":[],"id":1}`,
			unmarshalled: &jsonrpc.GetMempoolInfoCmd{},
		},
		{
			name: "getMiningInfo",
			newCmd: func() (interface{}, error) {
				return jsonrpc.NewCommand("getMiningInfo")
			},
			staticCmd: func() interface{} {
				return jsonrpc.NewGetMiningInfoCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getMiningInfo","params":[],"id":1}`,
			unmarshalled: &jsonrpc.GetMiningInfoCmd{},
		},
		{
			name: "getNetworkInfo",
			newCmd: func() (interface{}, error) {
				return jsonrpc.NewCommand("getNetworkInfo")
			},
			staticCmd: func() interface{} {
				return jsonrpc.NewGetNetworkInfoCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getNetworkInfo","params":[],"id":1}`,
			unmarshalled: &jsonrpc.GetNetworkInfoCmd{},
		},
		{
			name: "getNetTotals",
			newCmd: func() (interface{}, error) {
				return jsonrpc.NewCommand("getNetTotals")
			},
			staticCmd: func() interface{} {
				return jsonrpc.NewGetNetTotalsCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getNetTotals","params":[],"id":1}`,
			unmarshalled: &jsonrpc.GetNetTotalsCmd{},
		},
		{
			name: "getNetworkHashPs",
			newCmd: func() (interface{}, error) {
				return jsonrpc.NewCommand("getNetworkHashPs")
			},
			staticCmd: func() interface{} {
				return jsonrpc.NewGetNetworkHashPSCmd(nil, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getNetworkHashPs","params":[],"id":1}`,
			unmarshalled: &jsonrpc.GetNetworkHashPSCmd{
				Blocks: jsonrpc.Int(120),
				Height: jsonrpc.Int(-1),
			},
		},
		{
			name: "getNetworkHashPs optional1",
			newCmd: func() (interface{}, error) {
				return jsonrpc.NewCommand("getNetworkHashPs", 200)
			},
			staticCmd: func() interface{} {
				return jsonrpc.NewGetNetworkHashPSCmd(jsonrpc.Int(200), nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getNetworkHashPs","params":[200],"id":1}`,
			unmarshalled: &jsonrpc.GetNetworkHashPSCmd{
				Blocks: jsonrpc.Int(200),
				Height: jsonrpc.Int(-1),
			},
		},
		{
			name: "getNetworkHashPs optional2",
			newCmd: func() (interface{}, error) {
				return jsonrpc.NewCommand("getNetworkHashPs", 200, 123)
			},
			staticCmd: func() interface{} {
				return jsonrpc.NewGetNetworkHashPSCmd(jsonrpc.Int(200), jsonrpc.Int(123))
			},
			marshalled: `{"jsonrpc":"1.0","method":"getNetworkHashPs","params":[200,123],"id":1}`,
			unmarshalled: &jsonrpc.GetNetworkHashPSCmd{
				Blocks: jsonrpc.Int(200),
				Height: jsonrpc.Int(123),
			},
		},
		{
			name: "getPeerInfo",
			newCmd: func() (interface{}, error) {
				return jsonrpc.NewCommand("getPeerInfo")
			},
			staticCmd: func() interface{} {
				return jsonrpc.NewGetPeerInfoCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getPeerInfo","params":[],"id":1}`,
			unmarshalled: &jsonrpc.GetPeerInfoCmd{},
		},
		{
			name: "getRawMempool",
			newCmd: func() (interface{}, error) {
				return jsonrpc.NewCommand("getRawMempool")
			},
			staticCmd: func() interface{} {
				return jsonrpc.NewGetRawMempoolCmd(nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getRawMempool","params":[],"id":1}`,
			unmarshalled: &jsonrpc.GetRawMempoolCmd{
				Verbose: jsonrpc.Bool(false),
			},
		},
		{
			name: "getRawMempool optional",
			newCmd: func() (interface{}, error) {
				return jsonrpc.NewCommand("getRawMempool", false)
			},
			staticCmd: func() interface{} {
				return jsonrpc.NewGetRawMempoolCmd(jsonrpc.Bool(false))
			},
			marshalled: `{"jsonrpc":"1.0","method":"getRawMempool","params":[false],"id":1}`,
			unmarshalled: &jsonrpc.GetRawMempoolCmd{
				Verbose: jsonrpc.Bool(false),
			},
		},
		{
			name: "getRawTransaction",
			newCmd: func() (interface{}, error) {
				return jsonrpc.NewCommand("getRawTransaction", "123")
			},
			staticCmd: func() interface{} {
				return jsonrpc.NewGetRawTransactionCmd("123", nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getRawTransaction","params":["123"],"id":1}`,
			unmarshalled: &jsonrpc.GetRawTransactionCmd{
				TxID:    "123",
				Verbose: jsonrpc.Int(0),
			},
		},
		{
			name: "getRawTransaction optional",
			newCmd: func() (interface{}, error) {
				return jsonrpc.NewCommand("getRawTransaction", "123", 1)
			},
			staticCmd: func() interface{} {
				return jsonrpc.NewGetRawTransactionCmd("123", jsonrpc.Int(1))
			},
			marshalled: `{"jsonrpc":"1.0","method":"getRawTransaction","params":["123",1],"id":1}`,
			unmarshalled: &jsonrpc.GetRawTransactionCmd{
				TxID:    "123",
				Verbose: jsonrpc.Int(1),
			},
		},
		{
			name: "getSubnetwork",
			newCmd: func() (interface{}, error) {
				return jsonrpc.NewCommand("getSubnetwork", "123")
			},
			staticCmd: func() interface{} {
				return jsonrpc.NewGetSubnetworkCmd("123")
			},
			marshalled: `{"jsonrpc":"1.0","method":"getSubnetwork","params":["123"],"id":1}`,
			unmarshalled: &jsonrpc.GetSubnetworkCmd{
				SubnetworkID: "123",
			},
		},
		{
			name: "getTxOut",
			newCmd: func() (interface{}, error) {
				return jsonrpc.NewCommand("getTxOut", "123", 1)
			},
			staticCmd: func() interface{} {
				return jsonrpc.NewGetTxOutCmd("123", 1, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getTxOut","params":["123",1],"id":1}`,
			unmarshalled: &jsonrpc.GetTxOutCmd{
				TxID:           "123",
				Vout:           1,
				IncludeMempool: jsonrpc.Bool(true),
			},
		},
		{
			name: "getTxOut optional",
			newCmd: func() (interface{}, error) {
				return jsonrpc.NewCommand("getTxOut", "123", 1, true)
			},
			staticCmd: func() interface{} {
				return jsonrpc.NewGetTxOutCmd("123", 1, jsonrpc.Bool(true))
			},
			marshalled: `{"jsonrpc":"1.0","method":"getTxOut","params":["123",1,true],"id":1}`,
			unmarshalled: &jsonrpc.GetTxOutCmd{
				TxID:           "123",
				Vout:           1,
				IncludeMempool: jsonrpc.Bool(true),
			},
		},
		{
			name: "getTxOutSetInfo",
			newCmd: func() (interface{}, error) {
				return jsonrpc.NewCommand("getTxOutSetInfo")
			},
			staticCmd: func() interface{} {
				return jsonrpc.NewGetTxOutSetInfoCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getTxOutSetInfo","params":[],"id":1}`,
			unmarshalled: &jsonrpc.GetTxOutSetInfoCmd{},
		},
		{
			name: "help",
			newCmd: func() (interface{}, error) {
				return jsonrpc.NewCommand("help")
			},
			staticCmd: func() interface{} {
				return jsonrpc.NewHelpCmd(nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"help","params":[],"id":1}`,
			unmarshalled: &jsonrpc.HelpCmd{
				Command: nil,
			},
		},
		{
			name: "help optional",
			newCmd: func() (interface{}, error) {
				return jsonrpc.NewCommand("help", "getBlock")
			},
			staticCmd: func() interface{} {
				return jsonrpc.NewHelpCmd(jsonrpc.String("getBlock"))
			},
			marshalled: `{"jsonrpc":"1.0","method":"help","params":["getBlock"],"id":1}`,
			unmarshalled: &jsonrpc.HelpCmd{
				Command: jsonrpc.String("getBlock"),
			},
		},
		{
			name: "invalidateBlock",
			newCmd: func() (interface{}, error) {
				return jsonrpc.NewCommand("invalidateBlock", "123")
			},
			staticCmd: func() interface{} {
				return jsonrpc.NewInvalidateBlockCmd("123")
			},
			marshalled: `{"jsonrpc":"1.0","method":"invalidateBlock","params":["123"],"id":1}`,
			unmarshalled: &jsonrpc.InvalidateBlockCmd{
				BlockHash: "123",
			},
		},
		{
			name: "ping",
			newCmd: func() (interface{}, error) {
				return jsonrpc.NewCommand("ping")
			},
			staticCmd: func() interface{} {
				return jsonrpc.NewPingCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"ping","params":[],"id":1}`,
			unmarshalled: &jsonrpc.PingCmd{},
		},
		{
			name: "preciousBlock",
			newCmd: func() (interface{}, error) {
				return jsonrpc.NewCommand("preciousBlock", "0123")
			},
			staticCmd: func() interface{} {
				return jsonrpc.NewPreciousBlockCmd("0123")
			},
			marshalled: `{"jsonrpc":"1.0","method":"preciousBlock","params":["0123"],"id":1}`,
			unmarshalled: &jsonrpc.PreciousBlockCmd{
				BlockHash: "0123",
			},
		},
		{
			name: "reconsiderBlock",
			newCmd: func() (interface{}, error) {
				return jsonrpc.NewCommand("reconsiderBlock", "123")
			},
			staticCmd: func() interface{} {
				return jsonrpc.NewReconsiderBlockCmd("123")
			},
			marshalled: `{"jsonrpc":"1.0","method":"reconsiderBlock","params":["123"],"id":1}`,
			unmarshalled: &jsonrpc.ReconsiderBlockCmd{
				BlockHash: "123",
			},
		},
		{
			name: "removeManualNode",
			newCmd: func() (interface{}, error) {
				return jsonrpc.NewCommand("removeManualNode", "127.0.0.1")
			},
			staticCmd: func() interface{} {
				return jsonrpc.NewRemoveManualNodeCmd("127.0.0.1")
			},
			marshalled:   `{"jsonrpc":"1.0","method":"removeManualNode","params":["127.0.0.1"],"id":1}`,
			unmarshalled: &jsonrpc.RemoveManualNodeCmd{Addr: "127.0.0.1"},
		},
		{
			name: "searchRawTransactions",
			newCmd: func() (interface{}, error) {
				return jsonrpc.NewCommand("searchRawTransactions", "1Address")
			},
			staticCmd: func() interface{} {
				return jsonrpc.NewSearchRawTransactionsCmd("1Address", nil, nil, nil, nil, nil, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"searchRawTransactions","params":["1Address"],"id":1}`,
			unmarshalled: &jsonrpc.SearchRawTransactionsCmd{
				Address:     "1Address",
				Verbose:     jsonrpc.Bool(true),
				Skip:        jsonrpc.Int(0),
				Count:       jsonrpc.Int(100),
				VinExtra:    jsonrpc.Bool(false),
				Reverse:     jsonrpc.Bool(false),
				FilterAddrs: nil,
			},
		},
		{
			name: "searchRawTransactions",
			newCmd: func() (interface{}, error) {
				return jsonrpc.NewCommand("searchRawTransactions", "1Address", false)
			},
			staticCmd: func() interface{} {
				return jsonrpc.NewSearchRawTransactionsCmd("1Address",
					jsonrpc.Bool(false), nil, nil, nil, nil, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"searchRawTransactions","params":["1Address",false],"id":1}`,
			unmarshalled: &jsonrpc.SearchRawTransactionsCmd{
				Address:     "1Address",
				Verbose:     jsonrpc.Bool(false),
				Skip:        jsonrpc.Int(0),
				Count:       jsonrpc.Int(100),
				VinExtra:    jsonrpc.Bool(false),
				Reverse:     jsonrpc.Bool(false),
				FilterAddrs: nil,
			},
		},
		{
			name: "searchRawTransactions",
			newCmd: func() (interface{}, error) {
				return jsonrpc.NewCommand("searchRawTransactions", "1Address", false, 5)
			},
			staticCmd: func() interface{} {
				return jsonrpc.NewSearchRawTransactionsCmd("1Address",
					jsonrpc.Bool(false), jsonrpc.Int(5), nil, nil, nil, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"searchRawTransactions","params":["1Address",false,5],"id":1}`,
			unmarshalled: &jsonrpc.SearchRawTransactionsCmd{
				Address:     "1Address",
				Verbose:     jsonrpc.Bool(false),
				Skip:        jsonrpc.Int(5),
				Count:       jsonrpc.Int(100),
				VinExtra:    jsonrpc.Bool(false),
				Reverse:     jsonrpc.Bool(false),
				FilterAddrs: nil,
			},
		},
		{
			name: "searchRawTransactions",
			newCmd: func() (interface{}, error) {
				return jsonrpc.NewCommand("searchRawTransactions", "1Address", false, 5, 10)
			},
			staticCmd: func() interface{} {
				return jsonrpc.NewSearchRawTransactionsCmd("1Address",
					jsonrpc.Bool(false), jsonrpc.Int(5), jsonrpc.Int(10), nil, nil, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"searchRawTransactions","params":["1Address",false,5,10],"id":1}`,
			unmarshalled: &jsonrpc.SearchRawTransactionsCmd{
				Address:     "1Address",
				Verbose:     jsonrpc.Bool(false),
				Skip:        jsonrpc.Int(5),
				Count:       jsonrpc.Int(10),
				VinExtra:    jsonrpc.Bool(false),
				Reverse:     jsonrpc.Bool(false),
				FilterAddrs: nil,
			},
		},
		{
			name: "searchRawTransactions",
			newCmd: func() (interface{}, error) {
				return jsonrpc.NewCommand("searchRawTransactions", "1Address", false, 5, 10, true)
			},
			staticCmd: func() interface{} {
				return jsonrpc.NewSearchRawTransactionsCmd("1Address",
					jsonrpc.Bool(false), jsonrpc.Int(5), jsonrpc.Int(10), jsonrpc.Bool(true), nil, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"searchRawTransactions","params":["1Address",false,5,10,true],"id":1}`,
			unmarshalled: &jsonrpc.SearchRawTransactionsCmd{
				Address:     "1Address",
				Verbose:     jsonrpc.Bool(false),
				Skip:        jsonrpc.Int(5),
				Count:       jsonrpc.Int(10),
				VinExtra:    jsonrpc.Bool(true),
				Reverse:     jsonrpc.Bool(false),
				FilterAddrs: nil,
			},
		},
		{
			name: "searchRawTransactions",
			newCmd: func() (interface{}, error) {
				return jsonrpc.NewCommand("searchRawTransactions", "1Address", false, 5, 10, true, true)
			},
			staticCmd: func() interface{} {
				return jsonrpc.NewSearchRawTransactionsCmd("1Address",
					jsonrpc.Bool(false), jsonrpc.Int(5), jsonrpc.Int(10), jsonrpc.Bool(true), jsonrpc.Bool(true), nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"searchRawTransactions","params":["1Address",false,5,10,true,true],"id":1}`,
			unmarshalled: &jsonrpc.SearchRawTransactionsCmd{
				Address:     "1Address",
				Verbose:     jsonrpc.Bool(false),
				Skip:        jsonrpc.Int(5),
				Count:       jsonrpc.Int(10),
				VinExtra:    jsonrpc.Bool(true),
				Reverse:     jsonrpc.Bool(true),
				FilterAddrs: nil,
			},
		},
		{
			name: "searchRawTransactions",
			newCmd: func() (interface{}, error) {
				return jsonrpc.NewCommand("searchRawTransactions", "1Address", false, 5, 10, true, true, []string{"1Address"})
			},
			staticCmd: func() interface{} {
				return jsonrpc.NewSearchRawTransactionsCmd("1Address",
					jsonrpc.Bool(false), jsonrpc.Int(5), jsonrpc.Int(10), jsonrpc.Bool(true), jsonrpc.Bool(true), &[]string{"1Address"})
			},
			marshalled: `{"jsonrpc":"1.0","method":"searchRawTransactions","params":["1Address",false,5,10,true,true,["1Address"]],"id":1}`,
			unmarshalled: &jsonrpc.SearchRawTransactionsCmd{
				Address:     "1Address",
				Verbose:     jsonrpc.Bool(false),
				Skip:        jsonrpc.Int(5),
				Count:       jsonrpc.Int(10),
				VinExtra:    jsonrpc.Bool(true),
				Reverse:     jsonrpc.Bool(true),
				FilterAddrs: &[]string{"1Address"},
			},
		},
		{
			name: "sendRawTransaction",
			newCmd: func() (interface{}, error) {
				return jsonrpc.NewCommand("sendRawTransaction", "1122")
			},
			staticCmd: func() interface{} {
				return jsonrpc.NewSendRawTransactionCmd("1122", nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"sendRawTransaction","params":["1122"],"id":1}`,
			unmarshalled: &jsonrpc.SendRawTransactionCmd{
				HexTx:         "1122",
				AllowHighFees: jsonrpc.Bool(false),
			},
		},
		{
			name: "sendRawTransaction optional",
			newCmd: func() (interface{}, error) {
				return jsonrpc.NewCommand("sendRawTransaction", "1122", false)
			},
			staticCmd: func() interface{} {
				return jsonrpc.NewSendRawTransactionCmd("1122", jsonrpc.Bool(false))
			},
			marshalled: `{"jsonrpc":"1.0","method":"sendRawTransaction","params":["1122",false],"id":1}`,
			unmarshalled: &jsonrpc.SendRawTransactionCmd{
				HexTx:         "1122",
				AllowHighFees: jsonrpc.Bool(false),
			},
		},
		{
			name: "setGenerate",
			newCmd: func() (interface{}, error) {
				return jsonrpc.NewCommand("setGenerate", true)
			},
			staticCmd: func() interface{} {
				return jsonrpc.NewSetGenerateCmd(true, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"setGenerate","params":[true],"id":1}`,
			unmarshalled: &jsonrpc.SetGenerateCmd{
				Generate:     true,
				GenProcLimit: jsonrpc.Int(-1),
			},
		},
		{
			name: "setGenerate optional",
			newCmd: func() (interface{}, error) {
				return jsonrpc.NewCommand("setGenerate", true, 6)
			},
			staticCmd: func() interface{} {
				return jsonrpc.NewSetGenerateCmd(true, jsonrpc.Int(6))
			},
			marshalled: `{"jsonrpc":"1.0","method":"setGenerate","params":[true,6],"id":1}`,
			unmarshalled: &jsonrpc.SetGenerateCmd{
				Generate:     true,
				GenProcLimit: jsonrpc.Int(6),
			},
		},
		{
			name: "stop",
			newCmd: func() (interface{}, error) {
				return jsonrpc.NewCommand("stop")
			},
			staticCmd: func() interface{} {
				return jsonrpc.NewStopCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"stop","params":[],"id":1}`,
			unmarshalled: &jsonrpc.StopCmd{},
		},
		{
			name: "submitBlock",
			newCmd: func() (interface{}, error) {
				return jsonrpc.NewCommand("submitBlock", "112233")
			},
			staticCmd: func() interface{} {
				return jsonrpc.NewSubmitBlockCmd("112233", nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"submitBlock","params":["112233"],"id":1}`,
			unmarshalled: &jsonrpc.SubmitBlockCmd{
				HexBlock: "112233",
				Options:  nil,
			},
		},
		{
			name: "submitBlock optional",
			newCmd: func() (interface{}, error) {
				return jsonrpc.NewCommand("submitBlock", "112233", `{"workId":"12345"}`)
			},
			staticCmd: func() interface{} {
				options := jsonrpc.SubmitBlockOptions{
					WorkID: "12345",
				}
				return jsonrpc.NewSubmitBlockCmd("112233", &options)
			},
			marshalled: `{"jsonrpc":"1.0","method":"submitBlock","params":["112233",{"workId":"12345"}],"id":1}`,
			unmarshalled: &jsonrpc.SubmitBlockCmd{
				HexBlock: "112233",
				Options: &jsonrpc.SubmitBlockOptions{
					WorkID: "12345",
				},
			},
		},
		{
			name: "uptime",
			newCmd: func() (interface{}, error) {
				return jsonrpc.NewCommand("uptime")
			},
			staticCmd: func() interface{} {
				return jsonrpc.NewUptimeCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"uptime","params":[],"id":1}`,
			unmarshalled: &jsonrpc.UptimeCmd{},
		},
		{
			name: "validateAddress",
			newCmd: func() (interface{}, error) {
				return jsonrpc.NewCommand("validateAddress", "1Address")
			},
			staticCmd: func() interface{} {
				return jsonrpc.NewValidateAddressCmd("1Address")
			},
			marshalled: `{"jsonrpc":"1.0","method":"validateAddress","params":["1Address"],"id":1}`,
			unmarshalled: &jsonrpc.ValidateAddressCmd{
				Address: "1Address",
			},
		},
		{
			name: "debugLevel",
			newCmd: func() (interface{}, error) {
				return jsonrpc.NewCommand("debugLevel", "trace")
			},
			staticCmd: func() interface{} {
				return jsonrpc.NewDebugLevelCmd("trace")
			},
			marshalled: `{"jsonrpc":"1.0","method":"debugLevel","params":["trace"],"id":1}`,
			unmarshalled: &jsonrpc.DebugLevelCmd{
				LevelSpec: "trace",
			},
		},
		{
			name: "node",
			newCmd: func() (interface{}, error) {
				return jsonrpc.NewCommand("node", jsonrpc.NRemove, "1.1.1.1")
			},
			staticCmd: func() interface{} {
				return jsonrpc.NewNodeCmd("remove", "1.1.1.1", nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"node","params":["remove","1.1.1.1"],"id":1}`,
			unmarshalled: &jsonrpc.NodeCmd{
				SubCmd: jsonrpc.NRemove,
				Target: "1.1.1.1",
			},
		},
		{
			name: "node",
			newCmd: func() (interface{}, error) {
				return jsonrpc.NewCommand("node", jsonrpc.NDisconnect, "1.1.1.1")
			},
			staticCmd: func() interface{} {
				return jsonrpc.NewNodeCmd("disconnect", "1.1.1.1", nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"node","params":["disconnect","1.1.1.1"],"id":1}`,
			unmarshalled: &jsonrpc.NodeCmd{
				SubCmd: jsonrpc.NDisconnect,
				Target: "1.1.1.1",
			},
		},
		{
			name: "node",
			newCmd: func() (interface{}, error) {
				return jsonrpc.NewCommand("node", jsonrpc.NConnect, "1.1.1.1", "perm")
			},
			staticCmd: func() interface{} {
				return jsonrpc.NewNodeCmd("connect", "1.1.1.1", jsonrpc.String("perm"))
			},
			marshalled: `{"jsonrpc":"1.0","method":"node","params":["connect","1.1.1.1","perm"],"id":1}`,
			unmarshalled: &jsonrpc.NodeCmd{
				SubCmd:        jsonrpc.NConnect,
				Target:        "1.1.1.1",
				ConnectSubCmd: jsonrpc.String("perm"),
			},
		},
		{
			name: "node",
			newCmd: func() (interface{}, error) {
				return jsonrpc.NewCommand("node", jsonrpc.NConnect, "1.1.1.1", "temp")
			},
			staticCmd: func() interface{} {
				return jsonrpc.NewNodeCmd("connect", "1.1.1.1", jsonrpc.String("temp"))
			},
			marshalled: `{"jsonrpc":"1.0","method":"node","params":["connect","1.1.1.1","temp"],"id":1}`,
			unmarshalled: &jsonrpc.NodeCmd{
				SubCmd:        jsonrpc.NConnect,
				Target:        "1.1.1.1",
				ConnectSubCmd: jsonrpc.String("temp"),
			},
		},
		{
			name: "generate",
			newCmd: func() (interface{}, error) {
				return jsonrpc.NewCommand("generate", 1)
			},
			staticCmd: func() interface{} {
				return jsonrpc.NewGenerateCmd(1)
			},
			marshalled: `{"jsonrpc":"1.0","method":"generate","params":[1],"id":1}`,
			unmarshalled: &jsonrpc.GenerateCmd{
				NumBlocks: 1,
			},
		},
		{
			name: "getSelectedTip",
			newCmd: func() (interface{}, error) {
				return jsonrpc.NewCommand("getSelectedTip")
			},
			staticCmd: func() interface{} {
				return jsonrpc.NewGetSelectedTipCmd(nil, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getSelectedTip","params":[],"id":1}`,
			unmarshalled: &jsonrpc.GetSelectedTipCmd{
				Verbose:   jsonrpc.Bool(true),
				VerboseTx: jsonrpc.Bool(false),
			},
		},
		{
			name: "getCurrentNet",
			newCmd: func() (interface{}, error) {
				return jsonrpc.NewCommand("getCurrentNet")
			},
			staticCmd: func() interface{} {
				return jsonrpc.NewGetCurrentNetCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getCurrentNet","params":[],"id":1}`,
			unmarshalled: &jsonrpc.GetCurrentNetCmd{},
		},
		{
			name: "getHeaders",
			newCmd: func() (interface{}, error) {
				return jsonrpc.NewCommand("getHeaders", "", "")
			},
			staticCmd: func() interface{} {
				return jsonrpc.NewGetHeadersCmd(
					"",
					"",
				)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getHeaders","params":["",""],"id":1}`,
			unmarshalled: &jsonrpc.GetHeadersCmd{
				StartHash: "",
				StopHash:  "",
			},
		},
		{
			name: "getHeaders - with arguments",
			newCmd: func() (interface{}, error) {
				return jsonrpc.NewCommand("getHeaders", "000000000000000001f1739002418e2f9a84c47a4fd2a0eb7a787a6b7dc12f16", "000000000000000000ba33b33e1fad70b69e234fc24414dd47113bff38f523f7")
			},
			staticCmd: func() interface{} {
				return jsonrpc.NewGetHeadersCmd(
					"000000000000000001f1739002418e2f9a84c47a4fd2a0eb7a787a6b7dc12f16",
					"000000000000000000ba33b33e1fad70b69e234fc24414dd47113bff38f523f7",
				)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getHeaders","params":["000000000000000001f1739002418e2f9a84c47a4fd2a0eb7a787a6b7dc12f16","000000000000000000ba33b33e1fad70b69e234fc24414dd47113bff38f523f7"],"id":1}`,
			unmarshalled: &jsonrpc.GetHeadersCmd{
				StartHash: "000000000000000001f1739002418e2f9a84c47a4fd2a0eb7a787a6b7dc12f16",
				StopHash:  "000000000000000000ba33b33e1fad70b69e234fc24414dd47113bff38f523f7",
			},
		},
		{
			name: "getTopHeaders",
			newCmd: func() (interface{}, error) {
				return jsonrpc.NewCommand("getTopHeaders")
			},
			staticCmd: func() interface{} {
				return jsonrpc.NewGetTopHeadersCmd(
					nil,
				)
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getTopHeaders","params":[],"id":1}`,
			unmarshalled: &jsonrpc.GetTopHeadersCmd{},
		},
		{
			name: "getTopHeaders - with start hash",
			newCmd: func() (interface{}, error) {
				return jsonrpc.NewCommand("getTopHeaders", "000000000000000000ba33b33e1fad70b69e234fc24414dd47113bff38f523f7")
			},
			staticCmd: func() interface{} {
				return jsonrpc.NewGetTopHeadersCmd(
					jsonrpc.String("000000000000000000ba33b33e1fad70b69e234fc24414dd47113bff38f523f7"),
				)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getTopHeaders","params":["000000000000000000ba33b33e1fad70b69e234fc24414dd47113bff38f523f7"],"id":1}`,
			unmarshalled: &jsonrpc.GetTopHeadersCmd{
				StartHash: jsonrpc.String("000000000000000000ba33b33e1fad70b69e234fc24414dd47113bff38f523f7"),
			},
		},
		{
			name: "version",
			newCmd: func() (interface{}, error) {
				return jsonrpc.NewCommand("version")
			},
			staticCmd: func() interface{} {
				return jsonrpc.NewVersionCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"version","params":[],"id":1}`,
			unmarshalled: &jsonrpc.VersionCmd{},
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
			t.Errorf("\n%s\n%s", marshalled, test.marshalled)
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

// TestRPCServerCommandErrors ensures any errors that occur in the command during
// custom mashal and unmarshal are as expected.
func TestRPCServerCommandErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		result     interface{}
		marshalled string
		err        error
	}{
		{
			name:       "template request with invalid type",
			result:     &jsonrpc.TemplateRequest{},
			marshalled: `{"mode":1}`,
			err:        &json.UnmarshalTypeError{},
		},
		{
			name:       "invalid template request sigoplimit field",
			result:     &jsonrpc.TemplateRequest{},
			marshalled: `{"sigoplimit":"invalid"}`,
			err:        jsonrpc.Error{ErrorCode: jsonrpc.ErrInvalidType},
		},
		{
			name:       "invalid template request masslimit field",
			result:     &jsonrpc.TemplateRequest{},
			marshalled: `{"masslimit":"invalid"}`,
			err:        jsonrpc.Error{ErrorCode: jsonrpc.ErrInvalidType},
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

		if terr, ok := test.err.(jsonrpc.Error); ok {
			gotErrorCode := err.(jsonrpc.Error).ErrorCode
			if gotErrorCode != terr.ErrorCode {
				t.Errorf("Test #%d (%s) mismatched error code "+
					"- got %v (%v), want %v", i, test.name,
					gotErrorCode, terr, terr.ErrorCode)
				continue
			}
		}
	}
}
