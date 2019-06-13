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
	"github.com/daglabs/btcd/wire"
)

// TestDAGSvrCmds tests all of the dag server commands marshal and unmarshal
// into valid results include handling of optional fields being omitted in the
// marshalled command, while optional fields with defaults have the default
// assigned on unmarshalled commands.
func TestDAGSvrCmds(t *testing.T) {
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
			name: "addManualNode",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("addManualNode", "127.0.0.1")
			},
			staticCmd: func() interface{} {
				return btcjson.NewAddManualNodeCmd("127.0.0.1", nil)
			},
			marshalled:   `{"jsonrpc":"1.0","method":"addManualNode","params":["127.0.0.1"],"id":1}`,
			unmarshalled: &btcjson.AddManualNodeCmd{Addr: "127.0.0.1", OneTry: btcjson.Bool(false)},
		},
		{
			name: "createRawTransaction",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("createRawTransaction", `[{"txId":"123","vout":1}]`,
					`{"456":0.0123}`)
			},
			staticCmd: func() interface{} {
				txInputs := []btcjson.TransactionInput{
					{TxID: "123", Vout: 1},
				}
				amounts := map[string]float64{"456": .0123}
				return btcjson.NewCreateRawTransactionCmd(txInputs, amounts, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"createRawTransaction","params":[[{"txId":"123","vout":1}],{"456":0.0123}],"id":1}`,
			unmarshalled: &btcjson.CreateRawTransactionCmd{
				Inputs:  []btcjson.TransactionInput{{TxID: "123", Vout: 1}},
				Amounts: map[string]float64{"456": .0123},
			},
		},
		{
			name: "createRawTransaction optional",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("createRawTransaction", `[{"txId":"123","vout":1}]`,
					`{"456":0.0123}`, int64(12312333333))
			},
			staticCmd: func() interface{} {
				txInputs := []btcjson.TransactionInput{
					{TxID: "123", Vout: 1},
				}
				amounts := map[string]float64{"456": .0123}
				return btcjson.NewCreateRawTransactionCmd(txInputs, amounts, btcjson.Uint64(12312333333))
			},
			marshalled: `{"jsonrpc":"1.0","method":"createRawTransaction","params":[[{"txId":"123","vout":1}],{"456":0.0123},12312333333],"id":1}`,
			unmarshalled: &btcjson.CreateRawTransactionCmd{
				Inputs:   []btcjson.TransactionInput{{TxID: "123", Vout: 1}},
				Amounts:  map[string]float64{"456": .0123},
				LockTime: btcjson.Uint64(12312333333),
			},
		},

		{
			name: "decodeRawTransaction",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("decodeRawTransaction", "123")
			},
			staticCmd: func() interface{} {
				return btcjson.NewDecodeRawTransactionCmd("123")
			},
			marshalled:   `{"jsonrpc":"1.0","method":"decodeRawTransaction","params":["123"],"id":1}`,
			unmarshalled: &btcjson.DecodeRawTransactionCmd{HexTx: "123"},
		},
		{
			name: "decodeScript",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("decodeScript", "00")
			},
			staticCmd: func() interface{} {
				return btcjson.NewDecodeScriptCmd("00")
			},
			marshalled:   `{"jsonrpc":"1.0","method":"decodeScript","params":["00"],"id":1}`,
			unmarshalled: &btcjson.DecodeScriptCmd{HexScript: "00"},
		},
		{
			name: "getAllManualNodesInfo",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("getAllManualNodesInfo")
			},
			staticCmd: func() interface{} {
				return btcjson.NewGetAllManualNodesInfoCmd(nil)
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getAllManualNodesInfo","params":[],"id":1}`,
			unmarshalled: &btcjson.GetAllManualNodesInfoCmd{Details: btcjson.Bool(true)},
		},
		{
			name: "getBestBlockHash",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("getBestBlockHash")
			},
			staticCmd: func() interface{} {
				return btcjson.NewGetBestBlockHashCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getBestBlockHash","params":[],"id":1}`,
			unmarshalled: &btcjson.GetBestBlockHashCmd{},
		},
		{
			name: "getBlock",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("getBlock", "123")
			},
			staticCmd: func() interface{} {
				return btcjson.NewGetBlockCmd("123", nil, nil, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getBlock","params":["123"],"id":1}`,
			unmarshalled: &btcjson.GetBlockCmd{
				Hash:      "123",
				Verbose:   btcjson.Bool(true),
				VerboseTx: btcjson.Bool(false),
			},
		},
		{
			name: "getBlock required optional1",
			newCmd: func() (interface{}, error) {
				// Intentionally use a source param that is
				// more pointers than the destination to
				// exercise that path.
				verbosePtr := btcjson.Bool(true)
				return btcjson.NewCmd("getBlock", "123", &verbosePtr)
			},
			staticCmd: func() interface{} {
				return btcjson.NewGetBlockCmd("123", btcjson.Bool(true), nil, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getBlock","params":["123",true],"id":1}`,
			unmarshalled: &btcjson.GetBlockCmd{
				Hash:      "123",
				Verbose:   btcjson.Bool(true),
				VerboseTx: btcjson.Bool(false),
			},
		},
		{
			name: "getBlock required optional2",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("getBlock", "123", true, true)
			},
			staticCmd: func() interface{} {
				return btcjson.NewGetBlockCmd("123", btcjson.Bool(true), btcjson.Bool(true), nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getBlock","params":["123",true,true],"id":1}`,
			unmarshalled: &btcjson.GetBlockCmd{
				Hash:      "123",
				Verbose:   btcjson.Bool(true),
				VerboseTx: btcjson.Bool(true),
			},
		},
		{
			name: "getBlock required optional3",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("getBlock", "123", true, true, "456")
			},
			staticCmd: func() interface{} {
				return btcjson.NewGetBlockCmd("123", btcjson.Bool(true), btcjson.Bool(true), btcjson.String("456"))
			},
			marshalled: `{"jsonrpc":"1.0","method":"getBlock","params":["123",true,true,"456"],"id":1}`,
			unmarshalled: &btcjson.GetBlockCmd{
				Hash:       "123",
				Verbose:    btcjson.Bool(true),
				VerboseTx:  btcjson.Bool(true),
				Subnetwork: btcjson.String("456"),
			},
		},
		{
			name: "getBlockDagInfo",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("getBlockDagInfo")
			},
			staticCmd: func() interface{} {
				return btcjson.NewGetBlockDAGInfoCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getBlockDagInfo","params":[],"id":1}`,
			unmarshalled: &btcjson.GetBlockDAGInfoCmd{},
		},
		{
			name: "getBlockCount",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("getBlockCount")
			},
			staticCmd: func() interface{} {
				return btcjson.NewGetBlockCountCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getBlockCount","params":[],"id":1}`,
			unmarshalled: &btcjson.GetBlockCountCmd{},
		},
		{
			name: "getBlockHash",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("getBlockHash", 123)
			},
			staticCmd: func() interface{} {
				return btcjson.NewGetBlockHashCmd(123)
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getBlockHash","params":[123],"id":1}`,
			unmarshalled: &btcjson.GetBlockHashCmd{Index: 123},
		},
		{
			name: "getBlockHeader",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("getBlockHeader", "123")
			},
			staticCmd: func() interface{} {
				return btcjson.NewGetBlockHeaderCmd("123", nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getBlockHeader","params":["123"],"id":1}`,
			unmarshalled: &btcjson.GetBlockHeaderCmd{
				Hash:    "123",
				Verbose: btcjson.Bool(true),
			},
		},
		{
			name: "getBlockTemplate",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("getBlockTemplate")
			},
			staticCmd: func() interface{} {
				return btcjson.NewGetBlockTemplateCmd(nil)
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getBlockTemplate","params":[],"id":1}`,
			unmarshalled: &btcjson.GetBlockTemplateCmd{Request: nil},
		},
		{
			name: "getBlockTemplate optional - template request",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("getBlockTemplate", `{"mode":"template","capabilities":["longpoll","coinbasetxn"]}`)
			},
			staticCmd: func() interface{} {
				template := btcjson.TemplateRequest{
					Mode:         "template",
					Capabilities: []string{"longpoll", "coinbasetxn"},
				}
				return btcjson.NewGetBlockTemplateCmd(&template)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getBlockTemplate","params":[{"mode":"template","capabilities":["longpoll","coinbasetxn"]}],"id":1}`,
			unmarshalled: &btcjson.GetBlockTemplateCmd{
				Request: &btcjson.TemplateRequest{
					Mode:         "template",
					Capabilities: []string{"longpoll", "coinbasetxn"},
				},
			},
		},
		{
			name: "getBlockTemplate optional - template request with tweaks",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("getBlockTemplate", `{"mode":"template","capabilities":["longPoll","coinbaseTxn"],"sigOpLimit":500,"sizeLimit":100000000,"maxVersion":1}`)
			},
			staticCmd: func() interface{} {
				template := btcjson.TemplateRequest{
					Mode:         "template",
					Capabilities: []string{"longPoll", "coinbaseTxn"},
					SigOpLimit:   500,
					SizeLimit:    100000000,
					MaxVersion:   1,
				}
				return btcjson.NewGetBlockTemplateCmd(&template)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getBlockTemplate","params":[{"mode":"template","capabilities":["longPoll","coinbaseTxn"],"sigOpLimit":500,"sizeLimit":100000000,"maxVersion":1}],"id":1}`,
			unmarshalled: &btcjson.GetBlockTemplateCmd{
				Request: &btcjson.TemplateRequest{
					Mode:         "template",
					Capabilities: []string{"longPoll", "coinbaseTxn"},
					SigOpLimit:   int64(500),
					SizeLimit:    int64(100000000),
					MaxVersion:   1,
				},
			},
		},
		{
			name: "getBlockTemplate optional - template request with tweaks 2",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("getBlockTemplate", `{"mode":"template","capabilities":["longPoll","coinbaseTxn"],"sigOpLimit":true,"sizeLimit":100000000,"maxVersion":1}`)
			},
			staticCmd: func() interface{} {
				template := btcjson.TemplateRequest{
					Mode:         "template",
					Capabilities: []string{"longPoll", "coinbaseTxn"},
					SigOpLimit:   true,
					SizeLimit:    100000000,
					MaxVersion:   1,
				}
				return btcjson.NewGetBlockTemplateCmd(&template)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getBlockTemplate","params":[{"mode":"template","capabilities":["longPoll","coinbaseTxn"],"sigOpLimit":true,"sizeLimit":100000000,"maxVersion":1}],"id":1}`,
			unmarshalled: &btcjson.GetBlockTemplateCmd{
				Request: &btcjson.TemplateRequest{
					Mode:         "template",
					Capabilities: []string{"longPoll", "coinbaseTxn"},
					SigOpLimit:   true,
					SizeLimit:    int64(100000000),
					MaxVersion:   1,
				},
			},
		},
		{
			name: "getCFilter",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("getCFilter", "123",
					wire.GCSFilterExtended)
			},
			staticCmd: func() interface{} {
				return btcjson.NewGetCFilterCmd("123",
					wire.GCSFilterExtended)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getCFilter","params":["123",1],"id":1}`,
			unmarshalled: &btcjson.GetCFilterCmd{
				Hash:       "123",
				FilterType: wire.GCSFilterExtended,
			},
		},
		{
			name: "getCFilterHeader",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("getCFilterHeader", "123",
					wire.GCSFilterExtended)
			},
			staticCmd: func() interface{} {
				return btcjson.NewGetCFilterHeaderCmd("123",
					wire.GCSFilterExtended)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getCFilterHeader","params":["123",1],"id":1}`,
			unmarshalled: &btcjson.GetCFilterHeaderCmd{
				Hash:       "123",
				FilterType: wire.GCSFilterExtended,
			},
		},
		{
			name: "getDagTips",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("getDagTips")
			},
			staticCmd: func() interface{} {
				return btcjson.NewGetDAGTipsCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getDagTips","params":[],"id":1}`,
			unmarshalled: &btcjson.GetDAGTipsCmd{},
		},
		{
			name: "getConnectionCount",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("getConnectionCount")
			},
			staticCmd: func() interface{} {
				return btcjson.NewGetConnectionCountCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getConnectionCount","params":[],"id":1}`,
			unmarshalled: &btcjson.GetConnectionCountCmd{},
		},
		{
			name: "getDifficulty",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("getDifficulty")
			},
			staticCmd: func() interface{} {
				return btcjson.NewGetDifficultyCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getDifficulty","params":[],"id":1}`,
			unmarshalled: &btcjson.GetDifficultyCmd{},
		},
		{
			name: "getGenerate",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("getGenerate")
			},
			staticCmd: func() interface{} {
				return btcjson.NewGetGenerateCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getGenerate","params":[],"id":1}`,
			unmarshalled: &btcjson.GetGenerateCmd{},
		},
		{
			name: "getHashesPerSec",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("getHashesPerSec")
			},
			staticCmd: func() interface{} {
				return btcjson.NewGetHashesPerSecCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getHashesPerSec","params":[],"id":1}`,
			unmarshalled: &btcjson.GetHashesPerSecCmd{},
		},
		{
			name: "getInfo",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("getInfo")
			},
			staticCmd: func() interface{} {
				return btcjson.NewGetInfoCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getInfo","params":[],"id":1}`,
			unmarshalled: &btcjson.GetInfoCmd{},
		},
		{
			name: "getManualNodeInfo",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("getManualNodeInfo", "127.0.0.1")
			},
			staticCmd: func() interface{} {
				return btcjson.NewGetManualNodeInfoCmd("127.0.0.1", nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getManualNodeInfo","params":["127.0.0.1"],"id":1}`,
			unmarshalled: &btcjson.GetManualNodeInfoCmd{
				Node:    "127.0.0.1",
				Details: btcjson.Bool(true),
			},
		},
		{
			name: "getMempoolEntry",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("getMempoolEntry", "txhash")
			},
			staticCmd: func() interface{} {
				return btcjson.NewGetMempoolEntryCmd("txhash")
			},
			marshalled: `{"jsonrpc":"1.0","method":"getMempoolEntry","params":["txhash"],"id":1}`,
			unmarshalled: &btcjson.GetMempoolEntryCmd{
				TxID: "txhash",
			},
		},
		{
			name: "getMempoolInfo",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("getMempoolInfo")
			},
			staticCmd: func() interface{} {
				return btcjson.NewGetMempoolInfoCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getMempoolInfo","params":[],"id":1}`,
			unmarshalled: &btcjson.GetMempoolInfoCmd{},
		},
		{
			name: "getMiningInfo",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("getMiningInfo")
			},
			staticCmd: func() interface{} {
				return btcjson.NewGetMiningInfoCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getMiningInfo","params":[],"id":1}`,
			unmarshalled: &btcjson.GetMiningInfoCmd{},
		},
		{
			name: "getNetworkInfo",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("getNetworkInfo")
			},
			staticCmd: func() interface{} {
				return btcjson.NewGetNetworkInfoCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getNetworkInfo","params":[],"id":1}`,
			unmarshalled: &btcjson.GetNetworkInfoCmd{},
		},
		{
			name: "getNetTotals",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("getNetTotals")
			},
			staticCmd: func() interface{} {
				return btcjson.NewGetNetTotalsCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getNetTotals","params":[],"id":1}`,
			unmarshalled: &btcjson.GetNetTotalsCmd{},
		},
		{
			name: "getNetworkHashPs",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("getNetworkHashPs")
			},
			staticCmd: func() interface{} {
				return btcjson.NewGetNetworkHashPSCmd(nil, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getNetworkHashPs","params":[],"id":1}`,
			unmarshalled: &btcjson.GetNetworkHashPSCmd{
				Blocks: btcjson.Int(120),
				Height: btcjson.Int(-1),
			},
		},
		{
			name: "getNetworkHashPs optional1",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("getNetworkHashPs", 200)
			},
			staticCmd: func() interface{} {
				return btcjson.NewGetNetworkHashPSCmd(btcjson.Int(200), nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getNetworkHashPs","params":[200],"id":1}`,
			unmarshalled: &btcjson.GetNetworkHashPSCmd{
				Blocks: btcjson.Int(200),
				Height: btcjson.Int(-1),
			},
		},
		{
			name: "getNetworkHashPs optional2",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("getNetworkHashPs", 200, 123)
			},
			staticCmd: func() interface{} {
				return btcjson.NewGetNetworkHashPSCmd(btcjson.Int(200), btcjson.Int(123))
			},
			marshalled: `{"jsonrpc":"1.0","method":"getNetworkHashPs","params":[200,123],"id":1}`,
			unmarshalled: &btcjson.GetNetworkHashPSCmd{
				Blocks: btcjson.Int(200),
				Height: btcjson.Int(123),
			},
		},
		{
			name: "getPeerInfo",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("getPeerInfo")
			},
			staticCmd: func() interface{} {
				return btcjson.NewGetPeerInfoCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getPeerInfo","params":[],"id":1}`,
			unmarshalled: &btcjson.GetPeerInfoCmd{},
		},
		{
			name: "getRawMempool",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("getRawMempool")
			},
			staticCmd: func() interface{} {
				return btcjson.NewGetRawMempoolCmd(nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getRawMempool","params":[],"id":1}`,
			unmarshalled: &btcjson.GetRawMempoolCmd{
				Verbose: btcjson.Bool(false),
			},
		},
		{
			name: "getRawMempool optional",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("getRawMempool", false)
			},
			staticCmd: func() interface{} {
				return btcjson.NewGetRawMempoolCmd(btcjson.Bool(false))
			},
			marshalled: `{"jsonrpc":"1.0","method":"getRawMempool","params":[false],"id":1}`,
			unmarshalled: &btcjson.GetRawMempoolCmd{
				Verbose: btcjson.Bool(false),
			},
		},
		{
			name: "getRawTransaction",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("getRawTransaction", "123")
			},
			staticCmd: func() interface{} {
				return btcjson.NewGetRawTransactionCmd("123", nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getRawTransaction","params":["123"],"id":1}`,
			unmarshalled: &btcjson.GetRawTransactionCmd{
				TxID:    "123",
				Verbose: btcjson.Int(0),
			},
		},
		{
			name: "getRawTransaction optional",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("getRawTransaction", "123", 1)
			},
			staticCmd: func() interface{} {
				return btcjson.NewGetRawTransactionCmd("123", btcjson.Int(1))
			},
			marshalled: `{"jsonrpc":"1.0","method":"getRawTransaction","params":["123",1],"id":1}`,
			unmarshalled: &btcjson.GetRawTransactionCmd{
				TxID:    "123",
				Verbose: btcjson.Int(1),
			},
		},
		{
			name: "getSubnetwork",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("getSubnetwork", "123")
			},
			staticCmd: func() interface{} {
				return btcjson.NewGetSubnetworkCmd("123")
			},
			marshalled: `{"jsonrpc":"1.0","method":"getSubnetwork","params":["123"],"id":1}`,
			unmarshalled: &btcjson.GetSubnetworkCmd{
				SubnetworkID: "123",
			},
		},
		{
			name: "getTxOut",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("getTxOut", "123", 1)
			},
			staticCmd: func() interface{} {
				return btcjson.NewGetTxOutCmd("123", 1, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getTxOut","params":["123",1],"id":1}`,
			unmarshalled: &btcjson.GetTxOutCmd{
				TxID:           "123",
				Vout:           1,
				IncludeMempool: btcjson.Bool(true),
			},
		},
		{
			name: "getTxOut optional",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("getTxOut", "123", 1, true)
			},
			staticCmd: func() interface{} {
				return btcjson.NewGetTxOutCmd("123", 1, btcjson.Bool(true))
			},
			marshalled: `{"jsonrpc":"1.0","method":"getTxOut","params":["123",1,true],"id":1}`,
			unmarshalled: &btcjson.GetTxOutCmd{
				TxID:           "123",
				Vout:           1,
				IncludeMempool: btcjson.Bool(true),
			},
		},
		{
			name: "getTxOutProof",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("getTxOutProof", []string{"123", "456"})
			},
			staticCmd: func() interface{} {
				return btcjson.NewGetTxOutProofCmd([]string{"123", "456"}, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getTxOutProof","params":[["123","456"]],"id":1}`,
			unmarshalled: &btcjson.GetTxOutProofCmd{
				TxIDs: []string{"123", "456"},
			},
		},
		{
			name: "getTxOutProof optional",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("getTxOutProof", []string{"123", "456"},
					btcjson.String("000000000000034a7dedef4a161fa058a2d67a173a90155f3a2fe6fc132e0ebf"))
			},
			staticCmd: func() interface{} {
				return btcjson.NewGetTxOutProofCmd([]string{"123", "456"},
					btcjson.String("000000000000034a7dedef4a161fa058a2d67a173a90155f3a2fe6fc132e0ebf"))
			},
			marshalled: `{"jsonrpc":"1.0","method":"getTxOutProof","params":[["123","456"],` +
				`"000000000000034a7dedef4a161fa058a2d67a173a90155f3a2fe6fc132e0ebf"],"id":1}`,
			unmarshalled: &btcjson.GetTxOutProofCmd{
				TxIDs:     []string{"123", "456"},
				BlockHash: btcjson.String("000000000000034a7dedef4a161fa058a2d67a173a90155f3a2fe6fc132e0ebf"),
			},
		},
		{
			name: "getTxOutSetInfo",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("getTxOutSetInfo")
			},
			staticCmd: func() interface{} {
				return btcjson.NewGetTxOutSetInfoCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getTxOutSetInfo","params":[],"id":1}`,
			unmarshalled: &btcjson.GetTxOutSetInfoCmd{},
		},
		{
			name: "help",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("help")
			},
			staticCmd: func() interface{} {
				return btcjson.NewHelpCmd(nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"help","params":[],"id":1}`,
			unmarshalled: &btcjson.HelpCmd{
				Command: nil,
			},
		},
		{
			name: "help optional",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("help", "getBlock")
			},
			staticCmd: func() interface{} {
				return btcjson.NewHelpCmd(btcjson.String("getBlock"))
			},
			marshalled: `{"jsonrpc":"1.0","method":"help","params":["getBlock"],"id":1}`,
			unmarshalled: &btcjson.HelpCmd{
				Command: btcjson.String("getBlock"),
			},
		},
		{
			name: "invalidateBlock",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("invalidateBlock", "123")
			},
			staticCmd: func() interface{} {
				return btcjson.NewInvalidateBlockCmd("123")
			},
			marshalled: `{"jsonrpc":"1.0","method":"invalidateBlock","params":["123"],"id":1}`,
			unmarshalled: &btcjson.InvalidateBlockCmd{
				BlockHash: "123",
			},
		},
		{
			name: "ping",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("ping")
			},
			staticCmd: func() interface{} {
				return btcjson.NewPingCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"ping","params":[],"id":1}`,
			unmarshalled: &btcjson.PingCmd{},
		},
		{
			name: "preciousBlock",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("preciousBlock", "0123")
			},
			staticCmd: func() interface{} {
				return btcjson.NewPreciousBlockCmd("0123")
			},
			marshalled: `{"jsonrpc":"1.0","method":"preciousBlock","params":["0123"],"id":1}`,
			unmarshalled: &btcjson.PreciousBlockCmd{
				BlockHash: "0123",
			},
		},
		{
			name: "reconsiderBlock",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("reconsiderBlock", "123")
			},
			staticCmd: func() interface{} {
				return btcjson.NewReconsiderBlockCmd("123")
			},
			marshalled: `{"jsonrpc":"1.0","method":"reconsiderBlock","params":["123"],"id":1}`,
			unmarshalled: &btcjson.ReconsiderBlockCmd{
				BlockHash: "123",
			},
		},
		{
			name: "removeManualNode",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("removeManualNode", "127.0.0.1")
			},
			staticCmd: func() interface{} {
				return btcjson.NewRemoveManualNodeCmd("127.0.0.1")
			},
			marshalled:   `{"jsonrpc":"1.0","method":"removeManualNode","params":["127.0.0.1"],"id":1}`,
			unmarshalled: &btcjson.RemoveManualNodeCmd{Addr: "127.0.0.1"},
		},
		{
			name: "searchRawTransactions",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("searchRawTransactions", "1Address")
			},
			staticCmd: func() interface{} {
				return btcjson.NewSearchRawTransactionsCmd("1Address", nil, nil, nil, nil, nil, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"searchRawTransactions","params":["1Address"],"id":1}`,
			unmarshalled: &btcjson.SearchRawTransactionsCmd{
				Address:     "1Address",
				Verbose:     btcjson.Bool(true),
				Skip:        btcjson.Int(0),
				Count:       btcjson.Int(100),
				VinExtra:    btcjson.Bool(false),
				Reverse:     btcjson.Bool(false),
				FilterAddrs: nil,
			},
		},
		{
			name: "searchRawTransactions",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("searchRawTransactions", "1Address", false)
			},
			staticCmd: func() interface{} {
				return btcjson.NewSearchRawTransactionsCmd("1Address",
					btcjson.Bool(false), nil, nil, nil, nil, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"searchRawTransactions","params":["1Address",false],"id":1}`,
			unmarshalled: &btcjson.SearchRawTransactionsCmd{
				Address:     "1Address",
				Verbose:     btcjson.Bool(false),
				Skip:        btcjson.Int(0),
				Count:       btcjson.Int(100),
				VinExtra:    btcjson.Bool(false),
				Reverse:     btcjson.Bool(false),
				FilterAddrs: nil,
			},
		},
		{
			name: "searchRawTransactions",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("searchRawTransactions", "1Address", false, 5)
			},
			staticCmd: func() interface{} {
				return btcjson.NewSearchRawTransactionsCmd("1Address",
					btcjson.Bool(false), btcjson.Int(5), nil, nil, nil, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"searchRawTransactions","params":["1Address",false,5],"id":1}`,
			unmarshalled: &btcjson.SearchRawTransactionsCmd{
				Address:     "1Address",
				Verbose:     btcjson.Bool(false),
				Skip:        btcjson.Int(5),
				Count:       btcjson.Int(100),
				VinExtra:    btcjson.Bool(false),
				Reverse:     btcjson.Bool(false),
				FilterAddrs: nil,
			},
		},
		{
			name: "searchRawTransactions",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("searchRawTransactions", "1Address", false, 5, 10)
			},
			staticCmd: func() interface{} {
				return btcjson.NewSearchRawTransactionsCmd("1Address",
					btcjson.Bool(false), btcjson.Int(5), btcjson.Int(10), nil, nil, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"searchRawTransactions","params":["1Address",false,5,10],"id":1}`,
			unmarshalled: &btcjson.SearchRawTransactionsCmd{
				Address:     "1Address",
				Verbose:     btcjson.Bool(false),
				Skip:        btcjson.Int(5),
				Count:       btcjson.Int(10),
				VinExtra:    btcjson.Bool(false),
				Reverse:     btcjson.Bool(false),
				FilterAddrs: nil,
			},
		},
		{
			name: "searchRawTransactions",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("searchRawTransactions", "1Address", false, 5, 10, true)
			},
			staticCmd: func() interface{} {
				return btcjson.NewSearchRawTransactionsCmd("1Address",
					btcjson.Bool(false), btcjson.Int(5), btcjson.Int(10), btcjson.Bool(true), nil, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"searchRawTransactions","params":["1Address",false,5,10,true],"id":1}`,
			unmarshalled: &btcjson.SearchRawTransactionsCmd{
				Address:     "1Address",
				Verbose:     btcjson.Bool(false),
				Skip:        btcjson.Int(5),
				Count:       btcjson.Int(10),
				VinExtra:    btcjson.Bool(true),
				Reverse:     btcjson.Bool(false),
				FilterAddrs: nil,
			},
		},
		{
			name: "searchRawTransactions",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("searchRawTransactions", "1Address", false, 5, 10, true, true)
			},
			staticCmd: func() interface{} {
				return btcjson.NewSearchRawTransactionsCmd("1Address",
					btcjson.Bool(false), btcjson.Int(5), btcjson.Int(10), btcjson.Bool(true), btcjson.Bool(true), nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"searchRawTransactions","params":["1Address",false,5,10,true,true],"id":1}`,
			unmarshalled: &btcjson.SearchRawTransactionsCmd{
				Address:     "1Address",
				Verbose:     btcjson.Bool(false),
				Skip:        btcjson.Int(5),
				Count:       btcjson.Int(10),
				VinExtra:    btcjson.Bool(true),
				Reverse:     btcjson.Bool(true),
				FilterAddrs: nil,
			},
		},
		{
			name: "searchRawTransactions",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("searchRawTransactions", "1Address", false, 5, 10, true, true, []string{"1Address"})
			},
			staticCmd: func() interface{} {
				return btcjson.NewSearchRawTransactionsCmd("1Address",
					btcjson.Bool(false), btcjson.Int(5), btcjson.Int(10), btcjson.Bool(true), btcjson.Bool(true), &[]string{"1Address"})
			},
			marshalled: `{"jsonrpc":"1.0","method":"searchRawTransactions","params":["1Address",false,5,10,true,true,["1Address"]],"id":1}`,
			unmarshalled: &btcjson.SearchRawTransactionsCmd{
				Address:     "1Address",
				Verbose:     btcjson.Bool(false),
				Skip:        btcjson.Int(5),
				Count:       btcjson.Int(10),
				VinExtra:    btcjson.Bool(true),
				Reverse:     btcjson.Bool(true),
				FilterAddrs: &[]string{"1Address"},
			},
		},
		{
			name: "sendRawTransaction",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("sendRawTransaction", "1122")
			},
			staticCmd: func() interface{} {
				return btcjson.NewSendRawTransactionCmd("1122", nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"sendRawTransaction","params":["1122"],"id":1}`,
			unmarshalled: &btcjson.SendRawTransactionCmd{
				HexTx:         "1122",
				AllowHighFees: btcjson.Bool(false),
			},
		},
		{
			name: "sendRawTransaction optional",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("sendRawTransaction", "1122", false)
			},
			staticCmd: func() interface{} {
				return btcjson.NewSendRawTransactionCmd("1122", btcjson.Bool(false))
			},
			marshalled: `{"jsonrpc":"1.0","method":"sendRawTransaction","params":["1122",false],"id":1}`,
			unmarshalled: &btcjson.SendRawTransactionCmd{
				HexTx:         "1122",
				AllowHighFees: btcjson.Bool(false),
			},
		},
		{
			name: "setGenerate",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("setGenerate", true)
			},
			staticCmd: func() interface{} {
				return btcjson.NewSetGenerateCmd(true, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"setGenerate","params":[true],"id":1}`,
			unmarshalled: &btcjson.SetGenerateCmd{
				Generate:     true,
				GenProcLimit: btcjson.Int(-1),
			},
		},
		{
			name: "setGenerate optional",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("setGenerate", true, 6)
			},
			staticCmd: func() interface{} {
				return btcjson.NewSetGenerateCmd(true, btcjson.Int(6))
			},
			marshalled: `{"jsonrpc":"1.0","method":"setGenerate","params":[true,6],"id":1}`,
			unmarshalled: &btcjson.SetGenerateCmd{
				Generate:     true,
				GenProcLimit: btcjson.Int(6),
			},
		},
		{
			name: "stop",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("stop")
			},
			staticCmd: func() interface{} {
				return btcjson.NewStopCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"stop","params":[],"id":1}`,
			unmarshalled: &btcjson.StopCmd{},
		},
		{
			name: "submitBlock",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("submitBlock", "112233")
			},
			staticCmd: func() interface{} {
				return btcjson.NewSubmitBlockCmd("112233", nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"submitBlock","params":["112233"],"id":1}`,
			unmarshalled: &btcjson.SubmitBlockCmd{
				HexBlock: "112233",
				Options:  nil,
			},
		},
		{
			name: "submitBlock optional",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("submitBlock", "112233", `{"workId":"12345"}`)
			},
			staticCmd: func() interface{} {
				options := btcjson.SubmitBlockOptions{
					WorkID: "12345",
				}
				return btcjson.NewSubmitBlockCmd("112233", &options)
			},
			marshalled: `{"jsonrpc":"1.0","method":"submitBlock","params":["112233",{"workId":"12345"}],"id":1}`,
			unmarshalled: &btcjson.SubmitBlockCmd{
				HexBlock: "112233",
				Options: &btcjson.SubmitBlockOptions{
					WorkID: "12345",
				},
			},
		},
		{
			name: "uptime",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("uptime")
			},
			staticCmd: func() interface{} {
				return btcjson.NewUptimeCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"uptime","params":[],"id":1}`,
			unmarshalled: &btcjson.UptimeCmd{},
		},
		{
			name: "validateAddress",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("validateAddress", "1Address")
			},
			staticCmd: func() interface{} {
				return btcjson.NewValidateAddressCmd("1Address")
			},
			marshalled: `{"jsonrpc":"1.0","method":"validateAddress","params":["1Address"],"id":1}`,
			unmarshalled: &btcjson.ValidateAddressCmd{
				Address: "1Address",
			},
		},
		{
			name: "verifyMessage",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("verifyMessage", "1Address", "301234", "test")
			},
			staticCmd: func() interface{} {
				return btcjson.NewVerifyMessageCmd("1Address", "301234", "test")
			},
			marshalled: `{"jsonrpc":"1.0","method":"verifyMessage","params":["1Address","301234","test"],"id":1}`,
			unmarshalled: &btcjson.VerifyMessageCmd{
				Address:   "1Address",
				Signature: "301234",
				Message:   "test",
			},
		},
		{
			name: "verifyTxOutProof",
			newCmd: func() (interface{}, error) {
				return btcjson.NewCmd("verifyTxOutProof", "test")
			},
			staticCmd: func() interface{} {
				return btcjson.NewVerifyTxOutProofCmd("test")
			},
			marshalled: `{"jsonrpc":"1.0","method":"verifyTxOutProof","params":["test"],"id":1}`,
			unmarshalled: &btcjson.VerifyTxOutProofCmd{
				Proof: "test",
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
			result:     &btcjson.TemplateRequest{},
			marshalled: `{"mode":1}`,
			err:        &json.UnmarshalTypeError{},
		},
		{
			name:       "invalid template request sigoplimit field",
			result:     &btcjson.TemplateRequest{},
			marshalled: `{"sigoplimit":"invalid"}`,
			err:        btcjson.Error{ErrorCode: btcjson.ErrInvalidType},
		},
		{
			name:       "invalid template request sizelimit field",
			result:     &btcjson.TemplateRequest{},
			marshalled: `{"sizelimit":"invalid"}`,
			err:        btcjson.Error{ErrorCode: btcjson.ErrInvalidType},
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

		if terr, ok := test.err.(btcjson.Error); ok {
			gotErrorCode := err.(btcjson.Error).ErrorCode
			if gotErrorCode != terr.ErrorCode {
				t.Errorf("Test #%d (%s) mismatched error code "+
					"- got %v (%v), want %v", i, test.name,
					gotErrorCode, terr, terr.ErrorCode)
				continue
			}
		}
	}
}
